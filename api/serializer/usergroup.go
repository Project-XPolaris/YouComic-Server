package serializer

import "github.com/jinzhu/copier"

type BaseUserGroupTemplate struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func (t *BaseUserGroupTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	var err error
	err = copier.Copy(t, dataModel)
	if err != nil {
		return err
	}
	return nil
}

