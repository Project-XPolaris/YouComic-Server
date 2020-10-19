package utils

import (
	"github.com/karrick/godirwalk"
	"io/ioutil"
	"path/filepath"
)
var DefaultScanPageExt  = []string{
	".jpg",".png",".jpeg",
}

type ScannerResult struct {
	DirPath string
	CoverName string
	Pages []string
}
type Scanner struct {
	TargetPath string
	PageExt []string
	MinPageCount int
	Result []ScannerResult
}

func (s *Scanner)Scan() error{
	s.Result = []ScannerResult{}
	err := godirwalk.Walk(s.TargetPath, &godirwalk.Options{
		AllowNonDirectory: false,

		PostChildrenCallback: func(osPathname string, directoryEntry *godirwalk.Dirent) error {
			fileNames,_ := ioutil.ReadDir(osPathname)
			targetCount := 0
			coverName := ""
			pages := make([]string,0)
			for _, fileInfo := range fileNames {
				if !fileInfo.IsDir() {
					ext := filepath.Ext(fileInfo.Name())
					for _, targetExt := range s.PageExt {
						if targetExt == ext {
							if coverName == "" {
								coverName = fileInfo.Name()
							}
							targetCount += 1
							pages = append(pages, fileInfo.Name())
							break
						}
					}
				}
			}
			if targetCount >= s.MinPageCount {
				s.Result = append(s.Result, ScannerResult{
					DirPath:   osPathname,
					CoverName: coverName,
					Pages: pages,
				})
			}
			return nil
		},
		Callback: func(osPathname string, directoryEntry *godirwalk.Dirent) error {

			return nil
		},
	})
	return err
}