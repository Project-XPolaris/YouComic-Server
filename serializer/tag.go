package serializer

import (
	"github.com/jinzhu/copier"
	"time"
)

type BaseTagTemplate struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
}

func (t *BaseTagTemplate) Serializer(model interface{}, context map[string]interface{}) error {
	var err error
	err = copier.Copy(t, model)
	if err != nil {
		return err
	}
	return nil
}
