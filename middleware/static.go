package middleware

import (
	"fmt"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/gin-gonic/gin"
	"regexp"
	"strconv"
	"strings"
)

func StaticRouter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.URL.Path,"/assets/books") {
			r, _ := regexp.Compile("^/assets/books/(\\d+)/(.*?)$")
			result := r.FindStringSubmatch(c.Request.URL.Path)
			fmt.Println(result)
			bookId,err := strconv.Atoi(result[1])
			fileName := result[2]
			if err != nil {
				c.Next()
			}
			book := &model.Book{}
			database.DB.First(book,bookId)
			url := ""
			if book.Cover == fileName{
				// get cover
				url = fmt.Sprintf("/assets/books/%d/%s",book.ID,book.Cover)
			}else{
				// get page
			}

			fmt.Println(book.Cover)
			c.Request.URL.Path = url
			c.Next()
			return
		}
		c.Next()
	}
}