package serializer

import "github.com/jinzhu/copier"

type BaseCollectionTemplate struct {
	ID    int   `json:"id"`
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
