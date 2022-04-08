package utils

import (
	"github.com/karrick/godirwalk"
	"github.com/projectxpolaris/youcomic/config"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type ScannerResult struct {
	DirPath   string
	CoverName string
	Pages     []string
}
type Scanner struct {
	TargetPath string
	Result     []ScannerResult
}

func (s *Scanner) Scan(onScan func(result ScannerResult)) error {
	err := godirwalk.Walk(s.TargetPath, &godirwalk.Options{
		AllowNonDirectory: false,
		PostChildrenCallback: func(osPathname string, directoryEntry *godirwalk.Dirent) error {
			fileNames, _ := ioutil.ReadDir(osPathname)
			targetCount := 0
			coverName := ""
			pages := make([]string, 0)
			for _, fileInfo := range fileNames {
				if !fileInfo.IsDir() {
					ext := strings.ToLower(filepath.Ext(fileInfo.Name()))
					for _, targetExt := range config.Instance.ScannerConfig.Extensions {
						if targetExt == ext &&
							len(fileInfo.Name()) > 0 &&
							fileInfo.Size() > config.Instance.ScannerConfig.MinPageSize &&
							!strings.HasPrefix(fileInfo.Name(), ".") {
							if coverName == "" {
								coverName = fileInfo.Name()
							}
							targetCount += 1
							pages = append(pages, fileInfo.Name())
							fileInfo.Name()
							break
						}
					}
				}
			}
			if targetCount >= int(config.Instance.ScannerConfig.MinPageCount) {
				onScan(ScannerResult{
					DirPath:   osPathname,
					CoverName: coverName,
					Pages:     pages,
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
