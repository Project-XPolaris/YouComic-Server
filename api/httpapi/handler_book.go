package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/allentom/haruka"
	"github.com/jinzhu/copier"
	serializer2 "github.com/projectxpolaris/youcomic/api/httpapi/serializer"
	"github.com/projectxpolaris/youcomic/auth"
	appconfig "github.com/projectxpolaris/youcomic/config"
	ApiError "github.com/projectxpolaris/youcomic/error"
	ApplicationError "github.com/projectxpolaris/youcomic/error"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/permission"
	"github.com/projectxpolaris/youcomic/services"
	"github.com/projectxpolaris/youcomic/utils"
	"github.com/projectxpolaris/youcomic/validate"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
)

type CreateBookRequestBody struct {
	Name    string `form:"name" json:"name" xml:"name"  binding:"required"`
	Library int    `form:"library" json:"library" xml:"library"`
}

// create book handler
//
// path: /books
//
// method: post
var CreateBookHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody CreateBookRequestBody
	err := DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}
	claims, err := auth.ParseAuthHeader(context)
	if err != nil {
		ApiError.RaiseApiError(context, ApplicationError.UserAuthFailError, nil)
		return
	}

	if hasPermission := permission.CheckPermissionAndServerError(context,
		&permission.StandardPermissionChecker{PermissionName: permission.CreateBookPermissionName, UserId: claims.GetUserId()},
	); !hasPermission {
		return
	}

	if isValidate := validate.RunValidatorsAndRaiseApiError(context,
		&validate.UniqBookNameValidator{Value: requestBody.Name},
	); !isValidate {
		return
	}

	err, book := services.CreateBook(requestBody.Name, uint(requestBody.Library))
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	//serializer response
	template := serializer2.BaseBookTemplate{}
	RenderTemplate(context, &template, *book)
	context.JSONWithStatus(template, http.StatusCreated)
}

type UpdateBookRequestBody struct {
	Id         int
	Name       string `form:"name" json:"name" xml:"name"  binding:"required"`
	UpdateTags []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"updateTags"`
	OverwriteTag bool `json:"overwriteTag"`
}

// update book handler
//
// path: /book/:id
//
// method: patch
var UpdateBookHandler haruka.RequestHandler = func(context *haruka.Context) {

	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}

	claims, err := auth.ParseAuthHeader(context)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.UserAuthFailError, nil)
		return
	}

	//check permission
	if hasPermission := permission.CheckPermissionAndServerError(context,
		&permission.StandardPermissionChecker{PermissionName: permission.UpdateBookPermissionName, UserId: claims.GetUserId()},
	); !hasPermission {
		return
	}

	requestBody := UpdateBookRequestBody{}
	err = DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}

	//validate
	if isValidate := validate.RunValidatorsAndRaiseApiError(context,
		&validate.StringLengthValidator{Value: requestBody.Name, LessThan: 256, GreaterThan: 0, FieldName: "BookName"},
	); !isValidate {
		return
	}

	book := &model.Book{}
	err = AssignUpdateModel(&requestBody, book)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	book.ID = uint(id)

	err = services.UpdateBook(book, "Name")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	err = services.GetBook(book)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	// update tags
	if requestBody.UpdateTags != nil {
		tags := make([]*model.Tag, 0)
		for _, rawTag := range requestBody.UpdateTags {
			tags = append(tags, &model.Tag{Name: rawTag.Name, Type: rawTag.Type})
		}
		err = services.AddOrCreateTagToBook(book, tags, services.Overwrite)
		if err != nil {
			ApiError.RaiseApiError(context, err, nil)
			return
		}
	}
	template := &serializer2.BaseBookTemplate{}
	RenderTemplate(context, template, *book)
	context.JSONWithStatus(template, http.StatusOK)
}

// get book list handler
//
// path: /books
//
// method: get
var BookListHandler haruka.RequestHandler = func(context *haruka.Context) {
	//get page
	pagination := DefaultPagination{}
	pagination.Read(context)
	//get filter
	var books []model.Book
	queryBuilder := services.BooksQueryBuilder{}
	queryBuilder.SetPageFilter(pagination.Page, pagination.PageSize)

	filterMapping := []FilterMapping{
		{
			Lookup: "id",
			Method: "InId",
			Many:   true,
		},
		{
			Lookup: "name",
			Method: "SetNameFilter",
			Many:   true,
		},
		{
			Lookup: "order",
			Method: "SetOrderFilter",
			Many:   false,
		},
		{
			Lookup: "collection",
			Method: "SetCollectionQueryFilter",
			Many:   true,
		},
		{
			Lookup: "tag",
			Method: "SetTagQueryFilter",
			Many:   true,
		},
		{
			Lookup: "startTime",
			Method: "SetStartTimeQueryFilter",
			Many:   false,
		},
		{
			Lookup: "endTime",
			Method: "SetEndTimeQueryFilter",
			Many:   false,
		},
		{
			Lookup: "nameSearch",
			Method: "SetNameSearchQueryFilter",
			Many:   false,
		},
		{
			Lookup: "library",
			Method: "SetLibraryQueryFilter",
			Many:   true,
		},
		{
			Lookup: "pathSearch",
			Method: "SetPathSearchQueryFilter",
			Many:   false,
		},
		{
			Lookup: "random",
			Method: "SetRandomQueryFilter",
			Many:   false,
		},
	}
	for _, filter := range filterMapping {
		utils.FilterByParam(context, filter.Lookup, &queryBuilder, filter.Method, filter.Many)
	}

	count, err := queryBuilder.ReadModels(&books)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	with := context.GetQueryStrings("with")
	result := serializer2.SerializeMultipleTemplate(books, &serializer2.BaseBookTemplate{}, map[string]interface{}{"with": with})
	responseBody := serializer2.DefaultListContainer{}
	responseBody.SerializeList(result, map[string]interface{}{
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
		"count":    count,
		"url":      context.Request.URL,
	})
	context.JSONWithStatus(responseBody, http.StatusOK)
}

// delete book handler
//
// path: /book/:id
//
// method: delete
var DeleteBookHandler haruka.RequestHandler = func(context *haruka.Context) {
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}

	claims, err := auth.ParseAuthHeader(context)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.UserAuthFailError, nil)
		return
	}

	//check permission
	if hasPermission := permission.CheckPermissionAndServerError(context,
		&permission.StandardPermissionChecker{PermissionName: permission.DeleteBookPermissionName, UserId: claims.GetUserId()},
	); !hasPermission {
		return
	}

	//permanently delete permission check
	permanently := context.GetQueryString("permanently") == "true"
	if permanently {
		if hasPermission := permission.CheckPermissionAndServerError(context,
			&permission.StandardPermissionChecker{PermissionName: permission.PermanentlyDeleteBookPermissionName, UserId: claims.GetUserId()},
		); !hasPermission {
			return
		}
	}

	book := &model.Book{}
	book.ID = uint(id)
	err = services.DeleteById(&book)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	if permanently {
		err = services.DeleteBookFile(uint(id))
		if err != nil {
			ApiError.RaiseApiError(context, err, nil)
			return
		}
	}
	ServerSuccessResponse(context)
}

type BatchRequestBody struct {
	Create []*CreateBookRequestBody `json:"create"`
	Update []*UpdateBookRequestBody `json:"update"`
	Delete []int                    `json:"delete"`
}

// books action handler
//
// path: /books/batch
//
// method: post
var BookBatchHandler haruka.RequestHandler = func(context *haruka.Context) {
	requestBody := BatchRequestBody{}
	err := DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}

	//create action
	claims, err := auth.ParseAuthHeader(context)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.UserAuthFailError, nil)
		return
	}

	if hasPermission := permission.CheckPermissionAndServerError(context,
		&permission.StandardPermissionChecker{PermissionName: permission.CreateBookPermissionName, UserId: claims.GetUserId()},
	); !hasPermission {
		return
	}
	booksToCreate := make([]model.Book, 0)
	for _, requestBook := range requestBody.Create {
		book := model.Book{}
		err = copier.Copy(&book, &requestBook)
		if err != nil {
			ApiError.RaiseApiError(context, err, nil)
			return
		}
		booksToCreate = append(booksToCreate, book)
	}
	err = services.CreateBooks(booksToCreate)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	//update
	if hasPermission := permission.CheckPermissionAndServerError(context,
		&permission.StandardPermissionChecker{PermissionName: permission.UpdateBookPermissionName, UserId: claims.GetUserId()},
	); !hasPermission {
		return
	}
	booksToUpdate := make([]model.Book, 0)
	for _, updateBook := range requestBody.Update {
		book := &model.Book{}
		err = AssignUpdateModel(&updateBook, book)
		if err != nil {
			ApiError.RaiseApiError(context, err, nil)
			return
		}
		book.ID = uint(updateBook.Id)

		err = services.UpdateBook(book, "Name")
		if err != nil {
			ApiError.RaiseApiError(context, err, nil)
			return
		}

		err = services.GetBook(book)
		if err != nil {
			ApiError.RaiseApiError(context, err, nil)
			return
		}
		booksToUpdate = append(booksToUpdate, *book)

		// update tags
		if updateBook.UpdateTags != nil {
			tags := make([]*model.Tag, 0)
			for _, rawTag := range updateBook.UpdateTags {
				tags = append(tags, &model.Tag{Name: rawTag.Name, Type: rawTag.Type})
			}
			err = services.AddOrCreateTagToBook(book, tags, services.Overwrite)
			if err != nil {
				ApiError.RaiseApiError(context, err, nil)
				return
			}
		}
	}
	err = services.UpdateBooks(booksToUpdate, "Name")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	//delete
	if hasPermission := permission.CheckPermissionAndServerError(context,
		&permission.StandardPermissionChecker{PermissionName: permission.DeleteBookPermissionName, UserId: claims.GetUserId()},
	); !hasPermission {
		return
	}
	err = services.DeleteBooks(requestBody.Delete...)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

type AddTagToBookRequestBody struct {
	Tags []int `json:"tags"`
}

var BookTagBatch haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	requestBody := AddTagToBookRequestBody{}
	err = DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}
	err = services.AddTagToBook(id, requestBody.Tags...)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

func SaveCover(book model.Book, file multipart.File, header *multipart.FileHeader) (error, string) {
	err, storePath := services.GetBookPath(book.Path, book.LibraryId)
	if err != nil {
		return err, ""
	}
	fileExt := filepath.Ext(header.Filename)
	coverImageFilePath := filepath.Join(storePath, fmt.Sprintf("cover%s", fileExt))
	f, err := os.OpenFile(coverImageFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		return err, ""
	}
	return nil, coverImageFilePath
}

var AddBookCover haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	file, header, err := context.Request.FormFile("file")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	if header == nil {
		ApiError.RaiseApiError(context, errors.New("form not found"), nil)
		return
	}
	if file == nil {
		ApiError.RaiseApiError(context, errors.New("no such file in form"), nil)
		return
	}
	//update database
	book := model.Book{Model: gorm.Model{ID: uint(id)}}
	err = services.GetBook(&book)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	defer file.Close()
	//save cover and generate thumbnail
	err, coverImageFilePath := SaveCover(book, file, header)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	coverThumbnailStorePath := filepath.Join(appconfig.Instance.Store.Root, "generate", fmt.Sprintf("%d", book.ID))
	_, err = services.GenerateCoverThumbnail(coverImageFilePath, coverThumbnailStorePath)

	// update cover
	book.Cover = filepath.Base(coverImageFilePath)
	err = services.UpdateModel(&book, "Cover")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	// render response
	template := &serializer2.BaseBookTemplate{}
	RenderTemplate(context, template, book)
	context.JSONWithStatus(template, http.StatusOK)
}

var AddBookPages haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	form := context.Request.MultipartForm
	if form == nil {
		ApiError.RaiseApiError(context, errors.New("request not a form"), nil)
		return
	}

	re, err := regexp.Compile(`^page_(\d+)$`)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	book, err := services.GetBookById(uint(id))
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	err, storePath := services.GetBookPath(book.Path, book.LibraryId)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	createPages := make([]model.Page, 0)
	for fileField, file := range form.File {
		if re.MatchString(fileField) {
			matchGroups := re.FindAllStringSubmatch(fileField, 1)
			if len(matchGroups) > 0 && len(matchGroups[0]) > 1 {
				orderStr := matchGroups[0][1]
				order, err := strconv.Atoi(orderStr)
				if err != nil {
					ApiError.RaiseApiError(context, err, nil)
					return
				}
				//store
				storeFileHeader := file[0]
				fileExt := path.Ext(storeFileHeader.Filename)
				storeFileName := fmt.Sprintf("page_%d%s", order, fileExt)
				file, err := storeFileHeader.Open()
				if err != nil {
					ApiError.RaiseApiError(context, err, nil)
					return
				}
				f, err := os.OpenFile(fmt.Sprintf("%s/%s", storePath, storeFileName), os.O_WRONLY|os.O_CREATE, 0666)
				if err != nil {
					ApiError.RaiseApiError(context, err, nil)
					return
				}
				_, err = io.Copy(f, file)
				if err != nil {
					ApiError.RaiseApiError(context, err, nil)
					return
				}
				f.Close()
				file.Close()
				page := &model.Page{Path: storeFileName, PageOrder: order, BookId: id}
				err = services.CreateModel(page)
				if err != nil {
					ApiError.RaiseApiError(context, err, nil)
					return
				}
				createPages = append(createPages, *page)
			}
		}
	}

	result := serializer2.SerializeMultipleTemplate(createPages, &serializer2.BasePageTemplate{}, nil)
	responseBody := serializer2.DefaultListContainer{}
	responseBody.SerializeList(result, map[string]interface{}{
		"page":     1,
		"pageSize": len(createPages),
		"count":    int64(len(createPages)),
		"url":      context.Request.URL,
	})
	context.JSONWithStatus(responseBody, http.StatusOK)
}

var GetBookTags haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	tags, err := services.GetBookTag(uint(id))
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	result := serializer2.SerializeMultipleTemplate(tags, &serializer2.BaseTagTemplate{}, nil)
	responseBody := serializer2.DefaultListContainer{}
	responseBody.SerializeList(result, map[string]interface{}{
		"page":     1,
		"pageSize": len(tags),
		"count":    int64(len(tags)),
		"url":      context.Request.URL,
	})
	context.JSONWithStatus(responseBody, http.StatusOK)
}

var DeleteBookTag haruka.RequestHandler = func(context *haruka.Context) {
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	tagId, err := GetLookUpId(context, "tag")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	err = services.RemoveTagFromBook(uint(id), uint(tagId))
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

type UploadBookRequestBody struct {
	Name    string `form:"name"`
	Library string `form:"library"`
	Tags    string `form:"tags"`
	Pages   string `form:"pages"`
	Cover   string `form:"cover"`
}

var CreateBook haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody UploadBookRequestBody
	err := DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}
	libraryId, err := strconv.Atoi(requestBody.Library)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	err, book := services.CreateBook(requestBody.Name, uint(libraryId))
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	tagToAdd := make([]*model.Tag, 0)
	err = json.Unmarshal([]byte(requestBody.Tags), &tagToAdd)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	err = services.AddOrCreateTagToBook(book, tagToAdd, services.Overwrite)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	//handle with pages
	form := context.Request.MultipartForm
	files := form.File["image"]
	pageFilenames := make([]string, 0)
	err = json.Unmarshal([]byte(requestBody.Pages), &pageFilenames)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	for _, pageFilename := range pageFilenames {
		for pageIdx, file := range files {
			if pageFilename == file.Filename {
				storePath, err := SavePageFile(file, int(book.ID), pageIdx)
				if err != nil {
					logrus.Error(err)
					ApiError.RaiseApiError(context, err, nil)
					return
				}
				err = services.CreatePage(&model.Page{PageOrder: pageIdx, Path: filepath.Base(storePath), BookId: int(book.ID)})
				if err != nil {
					logrus.Error(err)
					ApiError.RaiseApiError(context, err, nil)
					return
				}
			}
		}
	}

	for _, fileHeader := range files {
		if fileHeader.Filename == requestBody.Cover {
			//save cover
			file, err := fileHeader.Open()
			err, coverPath := SaveCover(*book, file, fileHeader)
			if err != nil {
				logrus.Error(err)
				ApiError.RaiseApiError(context, err, nil)
				return
			}
			book.Cover = filepath.Base(coverPath)
			err = services.UpdateBook(book, "Cover")
			if err != nil {
				logrus.Error(err)
				ApiError.RaiseApiError(context, err, nil)
				return
			}
		}
	}

	template := &serializer2.BaseBookTemplate{}
	RenderTemplate(context, template, *book)
	context.JSONWithStatus(template, http.StatusOK)
}

var GetBook haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}

	book := &model.Book{Model: gorm.Model{ID: uint(id)}}
	err = services.GetBook(book)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	// add query history
	if context.GetQueryString("history") == "True" {
		userClaimsInterface, exist := context.Param["claim"]
		if exist {
			claims := userClaimsInterface.(*auth.UserClaims)
			err = services.AddBookHistory(claims.UserId, uint(id))
			if err != nil {
				logrus.Error(err)
				ApiError.RaiseApiError(context, err, nil)
				return
			}
		}
	}

	template := &serializer2.BaseBookTemplate{}
	RenderTemplate(context, template, *book)
	context.JSONWithStatus(template, http.StatusOK)
}

type ImportLibraryRequestBody struct {
	LibraryPath string `form:"library_path" json:"library_path" xml:"library_path"  binding:"required"`
}

var ImportLibraryHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody ImportLibraryRequestBody
	err := DecodeJsonBody(context, &requestBody)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	err = services.ImportLibrary(requestBody.LibraryPath)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

type RenameBookDirectoryRequestBody struct {
	Pattern string                `json:"pattern"`
	Slots   []services.RenameSlot `json:"slots"`
}

var RenameBookDirectoryHandler haruka.RequestHandler = func(context *haruka.Context) {
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	var requestBody RenameBookDirectoryRequestBody
	err = DecodeJsonBody(context, &requestBody)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	book, err := services.RenameBookDirectoryById(id, requestBody.Pattern, requestBody.Slots)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	template := &serializer2.BaseBookTemplate{}
	RenderTemplate(context, template, *book)
	context.JSONWithStatus(template, http.StatusOK)
}

var GenerateCoverThumbnail haruka.RequestHandler = func(context *haruka.Context) {
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	err = services.GenerateBookCoverById(uint(id))
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}
