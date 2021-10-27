package serializer

import "github.com/jinzhu/copier"

type BaseUserTemplate struct {
	ID       uint   `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}



func (b *BaseUserTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	var err error
	err = copier.Copy(b, dataModel)
	if err != nil {
		return err
	}
	return nil
}
type ManagerUserTemplate struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}
func (b *ManagerUserTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	var err error
	err = copier.Copy(b, dataModel)
	if err != nil {
		return err
	}
	return nil
}