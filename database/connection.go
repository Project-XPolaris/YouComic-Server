package database

import (
	"fmt"
	"github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
)

var DB *gorm.DB

func MysqlConnector() (err error) {
	mysqlConfig := config.Config.Mysql
	connectString := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		mysqlConfig.Username,
		mysqlConfig.Password,
		mysqlConfig.Host,
		mysqlConfig.Port,
		mysqlConfig.Database,
	)
	DB, err = gorm.Open("mysql", connectString)
	if err != nil {
		return err
	}
	return nil
}
func SqliteConnector() (err error) {
	DB, err = gorm.Open("sqlite3", config.Config.Sqlite.Path)
	return
}
func ConnectDatabase() {
	var err error
	databaseType := config.Config.Database.Type
	if databaseType == "mysql" {
		err = MysqlConnector()
	} else if databaseType == "sqlite" {
		err = SqliteConnector()
	}

	if err != nil {
		log.Fatal(err)
	}

	DB.AutoMigrate(
		&model.Book{},
		&model.Tag{},
		&model.Page{},
		&model.User{},
		&model.Collection{},
		&model.UserGroup{},
		&model.Permission{},
	)
}
