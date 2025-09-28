package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/allentom/haruka"
	"github.com/jinzhu/copier"
	serializer2 "github.com/projectxpolaris/youcomic/api/httpapi/serializer"
	"github.com/projectxpolaris/youcomic/auth"
	appconfig "github.com/projectxpolaris/youcomic/config"
	"github.com/projectxpolaris/youcomic/database"
	ApiError "github.com/projectxpolaris/youcomic/error"
	ApplicationError "github.com/projectxpolaris/youcomic/error"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/permission"
	"github.com/projectxpolaris/youcomic/plugin"
	"github.com/projectxpolaris/youcomic/services"
	"github.com/projectxpolaris/youcomic/utils"
	"github.com/projectxpolaris/youcomic/validate"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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
	var claims auth.JwtClaims
	if _, ok := context.Param["claim"]; ok {
		claims = context.Param["claim"].(*model.User)
	} else {
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
	Id                int
	Name              string            `form:"name" json:"name" xml:"name"`
	Cover             string            `form:"cover" json:"cover" xml:"cover"` // 添加封面字段
	TitleTranslations map[string]string `json:"titleTranslations"`
	UpdateTags        []struct {
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

	var claims auth.JwtClaims
	if _, ok := context.Param["claim"]; ok {
		claims = context.Param["claim"].(*model.User)
	} else {
		ApiError.RaiseApiError(context, ApplicationError.UserAuthFailError, nil)
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
	// 至少要提供名称、封面或标签中的一个
	if requestBody.Name == "" && requestBody.Cover == "" && requestBody.UpdateTags == nil {
		ApiError.RaiseApiError(context, errors.New("至少要提供名称、封面或标签中的一个字段进行更新"), nil)
		return
	}

	if requestBody.Name != "" {
		if isValidate := validate.RunValidatorsAndRaiseApiError(context,
			&validate.StringLengthValidator{Value: requestBody.Name, LessThan: 256, GreaterThan: 0, FieldName: "BookName"},
		); !isValidate {
			return
		}
	}

	book := &model.Book{}
	err = AssignUpdateModel(&requestBody, book)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	book.ID = uint(id)

	// 确定要更新的字段
	fieldsToUpdate := []string{}

	// 只有在提供了名称时才更新名称
	if requestBody.Name != "" {
		fieldsToUpdate = append(fieldsToUpdate, "Name")
	}

	// 提供了标题翻译时更新
	if requestBody.TitleTranslations != nil {
		book.TitleTranslations = requestBody.TitleTranslations
		fieldsToUpdate = append(fieldsToUpdate, "TitleTranslations")
	}

	// 处理封面更新
	if requestBody.Cover != "" {
		// 验证文件名格式
		if strings.Contains(requestBody.Cover, "?") || strings.Contains(requestBody.Cover, "#") {
			logrus.Errorf("封面文件名包含非法字符: '%s'", requestBody.Cover)
			ApiError.RaiseApiError(context, errors.New("封面文件名不能包含URL参数或片段标识符: "+requestBody.Cover), nil)
			return
		}

		// 验证书籍是否存在
		err = services.GetBook(book)
		if err != nil {
			ApiError.RaiseApiError(context, err, nil)
			return
		}

		// 获取书籍路径并验证封面文件是否存在
		err, storePath := services.GetBookPath(book.Path, book.LibraryId)
		if err != nil {
			ApiError.RaiseApiError(context, err, nil)
			return
		}

		coverFilePath := filepath.Join(storePath, requestBody.Cover)
		if _, err := os.Stat(coverFilePath); os.IsNotExist(err) {
			logrus.Errorf("封面文件不存在: 请求的文件名='%s', 完整路径='%s'", requestBody.Cover, coverFilePath)
			ApiError.RaiseApiError(context, errors.New("封面文件不存在，请确认文件名正确: "+requestBody.Cover), nil)
			return
		}

		// 设置新的封面文件名
		book.Cover = requestBody.Cover
		fieldsToUpdate = append(fieldsToUpdate, "Cover")

		// 重新生成缩略图
		coverThumbnailStorePath := filepath.Join(appconfig.Instance.Store.Root, "generate", fmt.Sprintf("%d", book.ID))
		_, err = services.GenerateCoverThumbnail(coverFilePath, coverThumbnailStorePath)
		if err != nil {
			logrus.Error("生成缩略图失败:", err)
			// 不阻断流程，只记录错误
		}
	}

	err = services.UpdateBook(book, fieldsToUpdate...)
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
	err := context.BindingInput(&queryBuilder)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
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
		{
			Lookup: "noTags",
			Method: "SetNoTagsFilter",
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

	var claims auth.JwtClaims
	if _, ok := context.Param["claim"]; ok {
		claims = context.Param["claim"].(*model.User)
	} else {
		ApiError.RaiseApiError(context, ApplicationError.UserAuthFailError, nil)
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
	var claims auth.JwtClaims
	if _, ok := context.Param["claim"]; ok {
		claims = context.Param["claim"].(*model.User)
	} else {
		ApiError.RaiseApiError(context, ApplicationError.UserAuthFailError, nil)
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

		// 确定要更新的字段
		fieldsToUpdate := []string{}

		// 只有在提供了名称时才更新名称
		if updateBook.Name != "" {
			fieldsToUpdate = append(fieldsToUpdate, "Name")
		}

		// 提供了标题翻译时更新
		if updateBook.TitleTranslations != nil {
			book.TitleTranslations = updateBook.TitleTranslations
			fieldsToUpdate = append(fieldsToUpdate, "TitleTranslations")
		}

		// 处理封面更新
		if updateBook.Cover != "" {
			// 验证文件名格式
			if strings.Contains(updateBook.Cover, "?") || strings.Contains(updateBook.Cover, "#") {
				logrus.Errorf("封面文件名包含非法字符: '%s'", updateBook.Cover)
				ApiError.RaiseApiError(context, errors.New("封面文件名不能包含URL参数或片段标识符: "+updateBook.Cover), nil)
				return
			}

			// 验证书籍是否存在
			err = services.GetBook(book)
			if err != nil {
				ApiError.RaiseApiError(context, err, nil)
				return
			}

			// 获取书籍路径并验证封面文件是否存在
			err, storePath := services.GetBookPath(book.Path, book.LibraryId)
			if err != nil {
				ApiError.RaiseApiError(context, err, nil)
				return
			}

			coverFilePath := filepath.Join(storePath, updateBook.Cover)
			if _, err := os.Stat(coverFilePath); os.IsNotExist(err) {
				logrus.Errorf("封面文件不存在: 请求的文件名='%s', 完整路径='%s'", updateBook.Cover, coverFilePath)
				ApiError.RaiseApiError(context, errors.New("封面文件不存在，请确认文件名正确: "+updateBook.Cover), nil)
				return
			}

			// 设置新的封面文件名
			book.Cover = updateBook.Cover
			fieldsToUpdate = append(fieldsToUpdate, "Cover")

			// 重新生成缩略图
			coverThumbnailStorePath := filepath.Join(appconfig.Instance.Store.Root, "generate", fmt.Sprintf("%d", book.ID))
			_, err = services.GenerateCoverThumbnail(coverFilePath, coverThumbnailStorePath)
			if err != nil {
				logrus.Error("生成缩略图失败:", err)
				// 不阻断流程，只记录错误
			}
		} else {
			err = services.GetBook(book)
			if err != nil {
				ApiError.RaiseApiError(context, err, nil)
				return
			}
		}

		err = services.UpdateBook(book, fieldsToUpdate...)
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
	// 不需要再次批量更新，因为已经在循环中单独更新了每本书

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

// generateUniqueFileName generates a unique filename by adding suffix if file already exists
func generateUniqueFileName(storePath, fileName string) string {
	fullPath := filepath.Join(storePath, fileName)

	// If file doesn't exist, return original filename
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fileName
	}

	// Extract name and extension
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)

	// Try different suffixes
	for i := 1; i < 1000; i++ {
		newFileName := fmt.Sprintf("%s_%d%s", nameWithoutExt, i, ext)
		newFullPath := filepath.Join(storePath, newFileName)

		if _, err := os.Stat(newFullPath); os.IsNotExist(err) {
			return newFileName
		}
	}

	// If still can't find unique name, use timestamp
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s_%d%s", nameWithoutExt, timestamp, ext)
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

// CropBookCoverHandler handles cover image cropping and saves the result
var CropBookCoverHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	// Get form file
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
	defer file.Close()

	// Get filename parameter
	fileName := context.Request.FormValue("fileName")
	if fileName == "" {
		ApiError.RaiseApiError(context, errors.New("fileName parameter is required"), nil)
		return
	}

	// Get book information
	book := model.Book{Model: gorm.Model{ID: uint(id)}}
	err = services.GetBook(&book)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	// Get book storage path
	err, storePath := services.GetBookPath(book.Path, book.LibraryId)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	// Generate unique filename to avoid conflicts
	finalFileName := generateUniqueFileName(storePath, fileName)
	coverImageFilePath := filepath.Join(storePath, finalFileName)

	// Save the cropped image file
	f, err := os.OpenFile(coverImageFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	// Generate thumbnail for the new cover
	coverThumbnailStorePath := filepath.Join(appconfig.Instance.Store.Root, "generate", fmt.Sprintf("%d", book.ID))
	_, err = services.GenerateCoverThumbnail(coverImageFilePath, coverThumbnailStorePath)
	if err != nil {
		// Log the error but don't fail the entire operation
		fmt.Printf("Warning: Failed to generate thumbnail for cropped cover: %v\n", err)
	}

	// Update book cover in database
	book.Cover = finalFileName
	err = services.UpdateModel(&book, "Cover")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	// Create a new page for the cropped cover
	var maxPageOrder int
	database.Instance.Model(&model.Page{}).Where("book_id = ?", book.ID).Select("COALESCE(MAX(page_order), 0)").Scan(&maxPageOrder)

	newPage := &model.Page{
		Path:      finalFileName,
		PageOrder: maxPageOrder + 1,
		BookId:    int(book.ID),
	}

	err = services.CreatePage(newPage)
	if err != nil {
		// Log the error but don't fail the entire operation since cover was already saved
		fmt.Printf("Warning: Failed to create page for cropped cover: %v\n", err)
	}

	// Return success response
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
			claims := userClaimsInterface.(*model.User)
			err = services.AddBookHistory(claims.GetUserId(), uint(id))
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

type AnalyzeBookFolderRequestBody struct {
	UpdateFields []string `json:"updateFields"` // 指定要更新的字段，如 ["name", "tags"]
}

type AnalyzeBookFolderResponseBody struct {
	ID         uint                `json:"id"`
	FolderName string              `json:"folderName"`
	Analysis   *BookAnalysisResult `json:"analysis"`
	MatchTags  []MatchTagResult    `json:"matchTags"`
	Success    bool                `json:"success"`
	Message    string              `json:"message"`
}

type MatchTagResult struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Source string `json:"source"`
}

type BookAnalysisResult struct {
	Title      string   `json:"title"`
	Author     string   `json:"author"`
	Series     string   `json:"series"`
	Tags       []string `json:"tags"`
	Genre      string   `json:"genre"`
	Volume     string   `json:"volume"`
	Chapter    string   `json:"chapter"`
	Language   string   `json:"language"`
	Publisher  string   `json:"publisher"`
	Year       string   `json:"year"`
	Confidence float64  `json:"confidence"`
	Reasoning  string   `json:"reasoning"`
}

// analyze book folder name using LLM
//
// path: /book/:id/analyze-folder
//
// method: post
var AnalyzeBookFolderHandler haruka.RequestHandler = func(harukaCtx *haruka.Context) {
	id, err := GetLookUpId(harukaCtx, "id")
	if err != nil {
		ApiError.RaiseApiError(harukaCtx, ApiError.RequestPathError, nil)
		return
	}

	var claims auth.JwtClaims
	if _, ok := harukaCtx.Param["claim"]; ok {
		claims = harukaCtx.Param["claim"].(*model.User)
	} else {
		ApiError.RaiseApiError(harukaCtx, ApplicationError.UserAuthFailError, nil)
		return
	}

	// 检查权限 - 需要更新书籍的权限
	if hasPermission := permission.CheckPermissionAndServerError(harukaCtx,
		&permission.StandardPermissionChecker{PermissionName: permission.UpdateBookPermissionName, UserId: claims.GetUserId()},
	); !hasPermission {
		return
	}

	var requestBody AnalyzeBookFolderRequestBody
	err = DecodeJsonBody(harukaCtx, &requestBody)
	if err != nil {
		return
	}

	// 获取书籍信息
	book := &model.Book{Model: gorm.Model{ID: uint(id)}}
	err = services.GetBook(book)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(harukaCtx, err, nil)
		return
	}

	// 获取LLM客户端
	llmClient, err := plugin.LLM.GetClient()
	if err != nil {
		logrus.Error("LLM plugin not available:", err)
		harukaCtx.JSONWithStatus(AnalyzeBookFolderResponseBody{
			ID:         book.ID,
			FolderName: filepath.Base(book.Path),
			MatchTags:  []MatchTagResult{},
			Success:    false,
			Message:    "LLM功能不可用，请检查配置",
		}, http.StatusServiceUnavailable)
		return
	}

	// 构建LLM分析提示
	folderName := filepath.Base(book.Path)
	if folderName == "" || folderName == "." {
		folderName = book.OriginalName
	}
	if folderName == "" {
		folderName = book.Name
	}

	prompt := fmt.Sprintf(`请分析以下漫画/书籍文件夹名，提取相关信息。请以JSON格式返回结果，包含以下字段：

文件夹名: %s

请返回JSON格式的分析结果，包含这些字段：
{
  "title": "标题",
  "author": "作者",
  "series": "系列名",
  "tags": ["标签1", "标签2"],
  "genre": "类型/题材",
  "volume": "卷数",
  "chapter": "章节",
  "language": "语言",
  "publisher": "出版社",
  "year": "年份",
  "confidence": 0.8,
  "reasoning": "分析推理过程"
}

注意：
1. 如果某个字段无法确定，请设为空字符串或空数组
2. confidence表示分析的置信度(0-1之间)
3. reasoning字段说明你的分析推理过程
4. 标签应该包括题材、风格、内容特征等
5. 请尽量从文件夹名中提取有用信息，即使信息不完整也要尽力分析`, folderName)

	// 调用LLM
	ctx := harukaCtx.Request.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	llmResponse, err := llmClient.GenerateText(ctx, prompt)
	if err != nil {
		logrus.Error("LLM analysis failed:", err)
		harukaCtx.JSONWithStatus(AnalyzeBookFolderResponseBody{
			ID:         book.ID,
			FolderName: folderName,
			MatchTags:  []MatchTagResult{},
			Success:    false,
			Message:    fmt.Sprintf("LLM分析失败: %v", err),
		}, http.StatusInternalServerError)
		return
	}

	// 解析LLM响应
	var analysisResult BookAnalysisResult
	err = json.Unmarshal([]byte(llmResponse), &analysisResult)
	if err != nil {
		// 如果JSON解析失败，尝试从响应中提取JSON
		jsonStart := -1
		jsonEnd := -1
		for i, r := range llmResponse {
			if r == '{' && jsonStart == -1 {
				jsonStart = i
			}
			if r == '}' {
				jsonEnd = i + 1
			}
		}

		if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
			jsonStr := llmResponse[jsonStart:jsonEnd]
			err = json.Unmarshal([]byte(jsonStr), &analysisResult)
		}

		if err != nil {
			logrus.Error("Failed to parse LLM response:", err)
			logrus.Error("LLM response:", llmResponse)
			harukaCtx.JSONWithStatus(AnalyzeBookFolderResponseBody{
				ID:         book.ID,
				FolderName: folderName,
				MatchTags:  []MatchTagResult{},
				Success:    false,
				Message:    "LLM响应解析失败，请检查响应格式",
			}, http.StatusInternalServerError)
			return
		}
	}

	// 更新书籍信息（如果请求中指定了要更新的字段）
	if len(requestBody.UpdateFields) > 0 {
		updateColumns := make([]string, 0)

		for _, field := range requestBody.UpdateFields {
			switch field {
			case "name":
				if analysisResult.Title != "" {
					book.Name = analysisResult.Title
					updateColumns = append(updateColumns, "Name")
				}
			}
		}

		// 更新书籍基本信息
		if len(updateColumns) > 0 {
			err = services.UpdateBook(book, updateColumns...)
			if err != nil {
				logrus.Error("Failed to update book:", err)
				harukaCtx.JSONWithStatus(AnalyzeBookFolderResponseBody{
					ID:         book.ID,
					FolderName: folderName,
					Analysis:   &analysisResult,
					MatchTags:  []MatchTagResult{},
					Success:    false,
					Message:    fmt.Sprintf("更新书籍信息失败: %v", err),
				}, http.StatusInternalServerError)
				return
			}
		}

		// 处理标签更新
		if contains(requestBody.UpdateFields, "tags") && len(analysisResult.Tags) > 0 {
			tags := make([]*model.Tag, 0)
			for _, tagName := range analysisResult.Tags {
				if tagName != "" {
					tags = append(tags, &model.Tag{Name: tagName, Type: "genre"})
				}
			}
			if len(tags) > 0 {
				err = services.AddOrCreateTagToBook(book, tags, services.Overwrite)
				if err != nil {
					logrus.Error("Failed to update book tags:", err)
				}
			}
		}
	}

	// 生成MatchTags
	matchTags := generateMatchTags(&analysisResult)

	// 返回成功响应
	harukaCtx.JSONWithStatus(AnalyzeBookFolderResponseBody{
		ID:         book.ID,
		FolderName: folderName,
		Analysis:   &analysisResult,
		MatchTags:  matchTags,
		Success:    true,
		Message:    "分析完成",
	}, http.StatusOK)
}

// 辅助函数：检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// 生成MatchTag结果的辅助函数
func generateMatchTags(analysis *BookAnalysisResult) []MatchTagResult {
	var matchTags []MatchTagResult

	// 添加标题
	if analysis.Title != "" {
		matchTags = append(matchTags, MatchTagResult{
			ID:     generateTagId(),
			Name:   analysis.Title,
			Type:   "name",
			Source: "ai",
		})
	}

	// 添加作者
	if analysis.Author != "" {
		matchTags = append(matchTags, MatchTagResult{
			ID:     generateTagId(),
			Name:   analysis.Author,
			Type:   "artist",
			Source: "ai",
		})
	}

	// 添加系列
	if analysis.Series != "" {
		matchTags = append(matchTags, MatchTagResult{
			ID:     generateTagId(),
			Name:   analysis.Series,
			Type:   "series",
			Source: "ai",
		})
	}

	// 添加标签（作为theme处理）
	for _, tag := range analysis.Tags {
		if tag != "" {
			matchTags = append(matchTags, MatchTagResult{
				ID:     generateTagId(),
				Name:   tag,
				Type:   "theme",
				Source: "ai",
			})
		}
	}

	// 添加其他信息作为原始标签
	if analysis.Genre != "" {
		matchTags = append(matchTags, MatchTagResult{
			ID:     generateTagId(),
			Name:   analysis.Genre,
			Type:   "theme",
			Source: "ai",
		})
	}

	if analysis.Volume != "" {
		matchTags = append(matchTags, MatchTagResult{
			ID:     generateTagId(),
			Name:   "Vol." + analysis.Volume,
			Type:   "theme",
			Source: "ai",
		})
	}

	if analysis.Chapter != "" {
		matchTags = append(matchTags, MatchTagResult{
			ID:     generateTagId(),
			Name:   "Ch." + analysis.Chapter,
			Type:   "theme",
			Source: "ai",
		})
	}

	if analysis.Publisher != "" {
		matchTags = append(matchTags, MatchTagResult{
			ID:     generateTagId(),
			Name:   analysis.Publisher,
			Type:   "theme",
			Source: "ai",
		})
	}

	if analysis.Year != "" {
		matchTags = append(matchTags, MatchTagResult{
			ID:     generateTagId(),
			Name:   analysis.Year,
			Type:   "theme",
			Source: "ai",
		})
	}

	if analysis.Language != "" {
		matchTags = append(matchTags, MatchTagResult{
			ID:     generateTagId(),
			Name:   analysis.Language,
			Type:   "theme",
			Source: "ai",
		})
	}

	return matchTags
}

// 生成唯一ID的辅助函数
func generateTagId() string {
	return fmt.Sprintf("llm_%d", time.Now().UnixNano())
}

// Request and handler for translating book titles using LLM
type TranslateBookTitleRequest struct {
	BookIDs         []uint   `json:"bookIds"`
	TargetLanguages []string `json:"targetLanguages"`
	DryRun          bool     `json:"dryRun"`
}

var TranslateBookTitleHandler haruka.RequestHandler = func(context *haruka.Context) {
	var req TranslateBookTitleRequest
	if err := DecodeJsonBody(context, &req); err != nil {
		return
	}

	if len(req.BookIDs) == 0 || len(req.TargetLanguages) == 0 {
		ApiError.RaiseApiError(context, errors.New("bookIds 与 targetLanguages 不能为空"), nil)
		return
	}

	var claims auth.JwtClaims
	if _, ok := context.Param["claim"]; ok {
		claims = context.Param["claim"].(*model.User)
	} else {
		ApiError.RaiseApiError(context, ApplicationError.UserAuthFailError, nil)
		return
	}

	if hasPermission := permission.CheckPermissionAndServerError(context,
		&permission.StandardPermissionChecker{PermissionName: permission.UpdateBookPermissionName, UserId: claims.GetUserId()},
	); !hasPermission {
		return
	}

	// LLM client
	llmClient, err := plugin.LLM.GetClient()
	if err != nil {
		ApiError.RaiseApiError(context, errors.New("LLM 功能不可用"), nil)
		return
	}

	type ItemResult struct {
		ID      uint              `json:"id"`
		Success bool              `json:"success"`
		Error   string            `json:"error,omitempty"`
		Data    map[string]string `json:"data,omitempty"`
	}

	results := make([]ItemResult, 0)

	for _, id := range req.BookIDs {
		book := &model.Book{Model: gorm.Model{ID: id}}
		if err := services.GetBook(book); err != nil {
			results = append(results, ItemResult{ID: id, Success: false, Error: err.Error()})
			continue
		}

		if book.TitleTranslations == nil {
			book.TitleTranslations = make(map[string]string)
		}

		// Translate for each target language
		itemData := make(map[string]string)
		for _, lang := range req.TargetLanguages {
			if strings.TrimSpace(lang) == "" {
				continue
			}
			prompt := fmt.Sprintf("Translate the following title into %s. Output only the translated title without quotes or extra text.\n\nTitle: %s", lang, book.Name)
			resp, err := llmClient.GenerateText(context.Request.Context(), prompt)
			if err != nil {
				continue
			}
			translated := strings.TrimSpace(resp)
			if translated != "" {
				itemData[lang] = translated
				if !req.DryRun {
					book.TitleTranslations[lang] = translated
				}
			}
		}

		if !req.DryRun {
			if err := services.UpdateBook(book, "TitleTranslations"); err != nil {
				results = append(results, ItemResult{ID: id, Success: false, Error: err.Error()})
				continue
			}
		}

		results = append(results, ItemResult{ID: id, Success: true, Data: itemData})
	}

	context.JSON(map[string]interface{}{
		"results": results,
	})
}
