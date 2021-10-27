package serializer

import (
	"github.com/allentom/youcomic-api/model"
	"github.com/jinzhu/copier"
)

type BaseCollectionTemplate struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Owner int    `json:"owner"`
}

func (t *BaseCollectionTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	var err error
	err = copier.Copy(t, dataModel)
	if err != nil {
		return err
	}
	return nil
}

// contain bool type property exist in collection model
//
// exist  =  specify book in collection
type CollectionWithBookContainTemplate struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Owner   int    `json:"owner"`
	Contain bool   `json:"contain"`
}

func (t *CollectionWithBookContainTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	collection := dataModel.(model.Collection)
	isExist := false
	bookExistCollectionInterface, isBookCollectionExist := context["bookCollections"]
	if isBookCollectionExist {
		bookCollections, typeCheck := bookExistCollectionInterface.([]model.Collection)
		if typeCheck {
			for _, containCollection := range bookCollections {
				if containCollection.ID == collection.ID {
					isExist = true
				}
			}
		}
	}
	t.Contain = isExist
	var err error
	err = copier.Copy(t, dataModel)
	if err != nil {
		return err
	}
	return nil
}
