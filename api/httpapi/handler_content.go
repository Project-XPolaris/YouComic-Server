package httpapi

import (
	"fmt"
	"github.com/allentom/haruka"
	appconfig "github.com/allentom/youcomic-api/config"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/services"
	"github.com/allentom/youcomic-api/utils"
	"net/http"
	"path"
	"strings"
)

var BookContentHandler haruka.RequestHandler = func(context *haruka.Context) {
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	fileName := context.GetPathParameterAsString("fileName")
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
		thumbnail := path.Join(appconfig.Instance.Store.Root, "generate", fmt.Sprintf("%d", book.ID), fmt.Sprintf("cover_thumbnail%s", thumbnailExt))
		if utils.CheckFileExist(thumbnail) {
			http.ServeFile(context.Writer, context.Request, thumbnail)
			return
		}
		// cover not generate,return original cover
		http.ServeFile(context.Writer, context.Request, path.Join(library.Path, book.Path, book.Cover))
		return
	}
	if fileName == path.Base(book.Cover) {
		http.ServeFile(context.Writer, context.Request, path.Join(library.Path, book.Path, book.Cover))
		return
	}

	//handle with page
	http.ServeFile(context.Writer, context.Request, path.Join(library.Path, book.Path, fileName))
}
