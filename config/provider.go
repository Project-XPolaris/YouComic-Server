package config

import (
	"errors"
	"github.com/allentom/harukap/config"
	"os"
	"path/filepath"
)

var DefaultConfigProvider *config.Provider

func InitConfigProvider() error {
	var err error
	customConfigPath := os.Getenv("YOUCOMIC_CONFIG_PATH")
	DefaultConfigProvider, err = config.NewProvider(func(provider *config.Provider) {
		ReadConfig(provider)
	}, customConfigPath)
	storeRootPath := Instance.Store.Root
	if _, err := os.Stat(filepath.Dir(storeRootPath)); os.IsNotExist(err) {
		return errors.New("store root path not exists,path = " + filepath.Dir(storeRootPath))
	}
	return err
}

var Instance Config

type EntityConfig struct {
	Enable  bool
	Name    string
	Version int64
}
type YouLibraryConfig struct {
	Enable bool
	Url    string
}
type ScannerConfig struct {
	MinPageCount int64
	MinPageSize  int64
	Extensions   []string
	MinWidth     int64
	MinHeight    int64
}
type Config struct {
	AuthEnable bool
	Database   string
	Thumbnail  struct {
		Type       string `json:"type"`
		Target     string `json:"target"`
		ServiceUrl string `json:"serviceUrl"`
	} `json:"thumbnail"`
	Store struct {
		Root  string `json:"root"`
		Books string `json:"books"`
	} `json:"store"`
	Security struct {
		Salt      string `json:"salt"`
		AppSecret string `json:"app_secret"`
	} `json:"security"`
	ScannerConfig ScannerConfig `json:"scanner"`
}

func ReadConfig(provider *config.Provider) {
	configer := provider.Manager
	configer.SetDefault("addr", ":7600")
	configer.SetDefault("application", "YouComic Core Service")
	configer.SetDefault("instance", "main")
	configer.SetDefault("scanner.minPageCount", 3)
	configer.SetDefault("scanner.minPageSize", 1024*10) //10kb
	configer.SetDefault("scanner.extensions", []string{".jpg", ".jpeg", ".png", ".bmp"})
	configer.SetDefault("scanner.minWidth", 1)
	configer.SetDefault("scanner.minHeight", 1)
	Instance = Config{
		AuthEnable: configer.GetBool("youplus.auth"),
		Database:   configer.GetString("datasource"),
		Thumbnail: struct {
			Type       string `json:"type"`
			Target     string `json:"target"`
			ServiceUrl string `json:"serviceUrl"`
		}{
			Type:       configer.GetString("thumbnail.type"),
			Target:     configer.GetString("thumbnail.target"),
			ServiceUrl: configer.GetString("thumbnail.service_url"),
		},
		Store: struct {
			Root  string `json:"root"`
			Books string `json:"books"`
		}{
			Root:  configer.GetString("store.root"),
			Books: configer.GetString("store.books"),
		},
		Security: struct {
			Salt      string `json:"salt"`
			AppSecret string `json:"app_secret"`
		}{
			Salt:      configer.GetString("security.salt"),
			AppSecret: configer.GetString("security.app_secret"),
		},
		ScannerConfig: ScannerConfig{
			MinPageCount: configer.GetInt64("scanner.minPageCount"),
			MinPageSize:  configer.GetInt64("scanner.minPageSize"),
			Extensions:   configer.GetStringSlice("scanner.extensions"),
			MinWidth:     configer.GetInt64("scanner.minWidth"),
			MinHeight:    configer.GetInt64("scanner.minHeight"),
		},
	}
	os.Mkdir(Instance.Store.Root, 0777)
}
