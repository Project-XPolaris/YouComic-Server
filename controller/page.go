package controller

import (
	"fmt"
	"github.com/allentom/youcomic-api/config"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/serializer"
	"github.com/allentom/youcomic-api/services"
	"github.com/allentom/youcomic-api/utils"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type UploadPageForm struct {
	BookId int `form:"book_id"`
	Order  int `form:"order"`
}

var PageUploadHandler gin.HandlerFunc = func(context *gin.Context) {
	var err error
	//parse request form
	form := UploadPageForm{}
	err = context.ShouldBind(&form)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	//get file from request form
	file, _ := context.FormFile("file")

	//store file
	storePath, err := SavePageFile(context, file, form.BookId, form.Order)
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

	template := serializer.BasePageTemplate{}
	err = template.Serializer(&modelToCreate, nil)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	context.JSON(http.StatusOK, template)
}

func SavePageFile(ctx *gin.Context, file *multipart.FileHeader, bookId int, order int) (string, error) {
	var err error

	storePath := filepath.Join(config.Config.Store.Books,strconv.Itoa(bookId))
	if _, err = os.Stat(storePath); os.IsNotExist(err) {
		err = os.MkdirAll(storePath, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	saveFileName := fmt.Sprintf("%s/%d%s", storePath, order, filepath.Ext(file.Filename))
	err = ctx.SaveUploadedFile(file, saveFileName)
	return saveFileName, err
}

type UpdatePageRequestBody struct {
	Order int `json:"order"`
}

var UpdatePageHandler gin.HandlerFunc = func(context *gin.Context) {
	var err error
	requestBody := UpdatePageRequestBody{}
	DecodeJsonBody(context, &requestBody)

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
	template := serializer.BasePageTemplate{}
	err = template.Serializer(&modelToUpdate, nil)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	context.JSON(http.StatusOK, template)
}

var DeletePageHandler gin.HandlerFunc = func(context *gin.Context) {
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

var PageListHandler gin.HandlerFunc = func(context *gin.Context) {
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

	result := serializer.SerializeMultipleTemplate(pages, &serializer.BasePageTemplate{},nil)
	responseBody := serializer.DefaultListContainer{}
	responseBody.SerializeList(result, map[string]interface{}{
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
		"count":    count,
		"url":      context.Request.URL,
	})
	context.JSON(http.StatusOK, responseBody)
}

var BatchPageHandler gin.HandlerFunc = func(context *gin.Context) {
	view := ModelsBatchView{
		Context: context,
		CreateModel: func() interface{} {
			return model.Page{}
		},
		AllowUpdateField: []string{
			"order",
		},
		AllowOperations: []BatchOperation{
			Update,Delete,
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
