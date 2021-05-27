package serializer

import (
	"github.com/allentom/youcomic-api/utils"
	"net/url"
)

type ListContainerSerializer interface {
	SerializeList(result interface{}, context map[string]interface{})
}

type DefaultListContainer struct {
	Count    int64         `json:"count"`
	Next     string      `json:"next"`
	Previous string      `json:"previous"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Results  interface{} `json:"result"`
}

func (c *DefaultListContainer) SerializeList(result interface{}, context map[string]interface{}) {
	page := context["page"].(int)
	pageSize := context["pageSize"].(int)
	requestUrl := context["url"].(*url.URL)
	count := context["count"].(int64)
	c.Count = count
	c.Next = utils.GetNextPageURL(requestUrl, count, page, pageSize)
	c.Previous = utils.GetNextPreviousURL(requestUrl, count, page)
	c.Results = result
	c.PageSize = pageSize
	c.Page = page
	return
}
