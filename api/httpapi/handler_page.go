package httpapi

import (
	"fmt"
	"github.com/allentom/haruka"
	serializer2 "github.com/allentom/youcomic-api/api/serializer"
	"github.com/allentom/youcomic-api/config"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/services"
	"github.com/allentom/youcomic-api/utils"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type UploadPageForm struct {
	BookId int             `hsource:"form" hname:"book_id"`
	Order  int             `hsource:"form" hname:"order"`
	File   haruka.FormFile `hsource:"form" hname:"file"`
}

var PageUploadHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	//parse request form
	form := UploadPageForm{}
	err = context.BindingInput(&form)
	if err != nil {
		return
	}

	//store file
	storePath, err := SavePageFile(form.File.Header, form.BookId, form.Order)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	//create page record in database
	modelToCreate := model.Page{}
	err = AssignRequestBodyToModel(&form, &modelToCreate)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	modelToCreate.Path = storePath
	err = services.CreatePage(&modelToCreate)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	template := serializer2.BasePageTemplate{}
	err = template.Serializer(&modelToCreate, nil)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	context.JSONWithStatus(template, http.StatusOK)
}

func SavePageFile(header *multipart.FileHeader, bookId int, order int) (string, error) {
	var err error

	storePath := filepath.Join(config.Instance.Store.Books, strconv.Itoa(bookId))
	if _, err = os.Stat(storePath); os.IsNotExist(err) {
		err = os.MkdirAll(storePath, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	file, err := header.Open()
	if err != nil {
		return "", err
	}
	saveFileName := fmt.Sprintf("%s/%d%s", storePath, order, filepath.Ext(header.Filename))
	f, err := os.OpenFile(saveFileName, os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()
	_, err = io.Copy(f, file)
	return saveFileName, err
}

type UpdatePageRequestBody struct {
	Order int `json:"order"`
}

var UpdatePageHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	requestBody := UpdatePageRequestBody{}
	err = DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	modelToUpdate := model.Page{}
	err = AssignRequestBodyToModel(&requestBody, &modelToUpdate)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	err = services.UpdateModel(&modelToUpdate, "Order")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	err = services.GetModelById(&modelToUpdate, id)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	//render
	template := serializer2.BasePageTemplate{}
	err = template.Serializer(&modelToUpdate, nil)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	context.JSONWithStatus(template, http.StatusOK)
}

var DeletePageHandler haruka.RequestHandler = func(context *haruka.Context) {
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	pageToDelete := model.Page{}
	pageToDelete.ID = uint(id)
	err = services.DeleteById(&pageToDelete)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

var PageListHandler haruka.RequestHandler = func(context *haruka.Context) {
	//get page
	pagination := DefaultPagination{}
	pagination.Read(context)

	//get filter
	//allowFilterParam := []string{"id",""}
	var pages []model.Page
	queryBuilder := services.PageQueryBuilder{}
	queryBuilder.SetPageFilter(pagination.Page, pagination.PageSize)

	filterMapping := []FilterMapping{
		{
			Lookup: "id",
			Method: "InId",
			Many:   true,
		},
		{
			Lookup: "book",
			Method: "SetBookIdFilter",
			Many:   true,
		},
		{
			Lookup: "order",
			Method: "SetOrderFilter",
			Many:   false,
		},
	}
	for _, filter := range filterMapping {
		utils.FilterByParam(context, filter.Lookup, &queryBuilder, filter.Method, filter.Many)
	}
	count, err := queryBuilder.ReadModels(&pages)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	var template serializer2.TemplateSerializer
	template = &serializer2.BasePageTemplate{}
	templateQueryParam := context.GetQueryString("template")
	if len(templateQueryParam) != 0 {
		if templateQueryParam == "withSize" {
			template = &serializer2.PageTemplateWithSize{}
		}
	}
	result := serializer2.SerializeMultipleTemplate(pages, template, nil)
	responseBody := serializer2.DefaultListContainer{}
	responseBody.SerializeList(result, map[string]interface{}{
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
		"count":    count,
		"url":      context.Request.URL,
	})
	context.JSONWithStatus(responseBody, http.StatusOK)
}

var BatchPageHandler haruka.RequestHandler = func(context *haruka.Context) {
	view := ModelsBatchView{
		Context: context,
		CreateModel: func() interface{} {
			return model.Page{}
		},
		AllowUpdateField: []string{
			"order",
		},
		AllowOperations: []BatchOperation{
			Update, Delete,
		},
		OperationFunc: map[BatchOperation]func(v *ModelsBatchView) error{
			Delete: func(v *ModelsBatchView) error {
				var err error
				err = services.DeletePages(v.RequestBody.Delete...)
				return err
			},
		},
	}

	view.Run()
}
