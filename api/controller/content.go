package controller

import (
	"fmt"
	appconfig "github.com/allentom/youcomic-api/config"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/services"
	"github.com/allentom/youcomic-api/utils"
	"github.com/gin-gonic/gin"
	"path"
	"strings"
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
	// handle with cover thumbnail
	if strings.Contains(fileName, "cover_thumbnail") {
		thumbnailExt := path.Ext(book.Cover)
		thumbnail := path.Join(appconfig.Config.Store.Root, "generate", fmt.Sprintf("%d", book.ID), fmt.Sprintf("cover_thumbnail%s", thumbnailExt))
		if utils.CheckFileExist(thumbnail) {
			context.File(thumbnail)
			return
		}
		// cover not generate,return original cover
		context.File(path.Join(library.Path, book.Path, book.Cover))
		return
	}
	if fileName == path.Base(book.Cover) {
		context.File(path.Join(library.Path, book.Path, book.Cover))
		return
	}

	//handle with page
	context.File(path.Join(library.Path, book.Path, fileName))
}
