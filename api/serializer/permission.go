package serializer

import "github.com/jinzhu/copier"

type BasePermissionTemplate struct {
	ID       uint   `json:"id"`
	Name string `json:"name"`
}

func (b *BasePermissionTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	var err error
	err = copier.Copy(b, dataModel)
	if err != nil {
		return err
	}
	return nil
}