package httpapi

import (
	"bytes"
	context2 "context"
	"fmt"
	"github.com/allentom/haruka"
	appconfig "github.com/projectxpolaris/youcomic/config"
	ApiError "github.com/projectxpolaris/youcomic/error"
	"github.com/projectxpolaris/youcomic/plugin"
	"github.com/projectxpolaris/youcomic/services"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"
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
		storage := plugin.GetDefaultStorage()
		isExist, err := storage.IsExist(context2.Background(), plugin.GetDefaultBucket(), thumbnail)
		if err != nil {
			ApiError.RaiseApiError(context, err, nil)
			return
		}
		if isExist {
			out, err := storage.Get(context2.Background(), plugin.GetDefaultBucket(), thumbnail)
			if err != nil {
				ApiError.RaiseApiError(context, err, nil)
				return
			}
			data, err := ioutil.ReadAll(out)
			if err != nil {
				ApiError.RaiseApiError(context, err, nil)
				return
			}
			http.ServeContent(context.Writer, context.Request, filepath.Base(thumbnail), time.Now(), bytes.NewReader(data))
			return
		}
		// cover not generate,return original cover
		http.ServeFile(context.Writer, context.Request, path.Join(library.Path, book.Path, book.Cover))
		return
	}

	//if fileName == path.Base(book.Cover) {
	//	coverPath := path.Join(library.Path, book.Path, book.Cover)
	//	http.ServeFile(context.Writer, context.Request, coverPath)
	//	return
	//}

	// for compress image
	compress, _ := context.GetQueryInt("compress")
	if compress > 0 {
		out, err := services.ResizeImageWithSizeCap(path.Join(library.Path, book.Path, fileName), int64(compress))
		if err != nil {
			ApiError.RaiseApiError(context, err, nil)
			return
		}
		context.Writer.Write(out)
		return
	}
	//handle with page
	http.ServeFile(context.Writer, context.Request, path.Join(library.Path, book.Path, fileName))
}
