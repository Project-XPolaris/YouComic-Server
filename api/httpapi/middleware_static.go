package httpapi

import (
	"fmt"
	"github.com/allentom/haruka"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"regexp"
	"strconv"
	"strings"
)

type StaticMiddleware struct {
}

func (m StaticMiddleware) OnRequest(c *haruka.Context) {
	if strings.Contains(c.Request.URL.Path, "/assets/books") {
		r, _ := regexp.Compile("^/assets/books/(\\d+)/(.*?)$")
		result := r.FindStringSubmatch(c.Request.URL.Path)
		fmt.Println(result)
		bookId, err := strconv.Atoi(result[1])
		fileName := result[2]
		if err != nil {
			return
		}
		book := &model.Book{}
		database.Instance.First(book, bookId)
		url := ""
		if book.Cover == fileName {
			// get cover
			url = fmt.Sprintf("/assets/books/%d/%s", book.ID, book.Cover)
		} else {
			// get page
		}
		fmt.Println(book.Cover)
		c.Request.URL.Path = url
		return
	}
}
