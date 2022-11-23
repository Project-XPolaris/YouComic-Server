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
	} else {
		t.Type = "File"
	}
	t.Ext = filepath.Ext(data.Name())
	t.Name = data.Name()
	t.Path = filepath.Join(rootDir, data.Name())
	return nil
}

type BaseFileItemTemplate struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Path string `json:"path"`
}

func (t *BaseFileItemTemplate) Assign(info os.DirEntry, rootPath string) {
	if info.IsDir() {
		t.Type = "Directory"
	} else {
		t.Type = "File"
	}
	t.Name = info.Name()
	t.Path = filepath.Join(rootPath, info.Name())
}
