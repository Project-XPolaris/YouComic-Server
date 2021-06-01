package services

import "os"

func ReadDirectory(target string) ([]os.DirEntry,error){
	return os.ReadDir(target)
}