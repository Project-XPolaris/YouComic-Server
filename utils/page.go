package utils

import (
	"math"
	"net/url"
	"strconv"
)

func GetNextPageURL(url *url.URL, count int, page int, pageSize int) string {
	totalPage := math.Ceil(float64(count) / float64(pageSize))
	query := url.Query()
	if totalPage > float64(page) {
		query.Set("page", strconv.Itoa(page+1))
		url.RawQuery = query.Encode()
		return url.String()
	} else {
		return ""
	}
}

func GetNextPreviousURL(url *url.URL, count int, page int) string {
	query := url.Query()
	if page > 2 {
		query.Set("page", strconv.Itoa(page-1))
		url.RawQuery = query.Encode()
		return url.String()
	} else {
		return ""
	}
}
