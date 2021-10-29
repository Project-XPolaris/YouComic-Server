package serializer

import (
	"os"
	"path/filepath"
)

type FileItemSerializer struct {
	Name string `json:"name"`
	Ext  string `json:"ext"`
	Path string `json:"path"`
	Type string `json:"type"`
}

func (t *FileItemSerializer) Serializer(dataModel interface{}, context map[string]interface{}) error {
	data := dataModel.(os.DirEntry)
	rootDir := context["root"].(string)
	if data.IsDir() {
		t.Type = "Directory"
	}else{
		t.Type = "File"
	}
	t.Ext = filepath.Ext(data.Name())
	t.Name = data.Name()
	t.Path = filepath.Join(rootDir,data.Name())
	return nil
}


