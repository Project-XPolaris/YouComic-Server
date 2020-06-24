package controller

import (
	"fmt"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/services"
	"github.com/gin-gonic/gin"
	"path"
)

var BookContentHandler gin.HandlerFunc = func(context *gin.Context) {
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	fileName := context.Param("fileName")
	book, err := services.GetBookById(uint(id))
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	//query library
	library, err := services.GetLibraryById(book.LibraryId)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	if fileName == book.Cover {
		context.File(path.Join(library.Path, fmt.Sprintf("%d", book.ID), book.Cover))
		return
	}

	//handle with page
	context.File(path.Join(library.Path, fmt.Sprintf("%d", book.ID), fileName))
}
