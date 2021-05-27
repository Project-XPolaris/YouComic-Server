package database

import (
	"fmt"
	"github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/model"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	DB, err = gorm.Open(mysql.Open(connectString), &gorm.Config{})
	if err != nil {
		return err
	}
	return nil
}
func SqliteConnector() (err error) {
	DB, err = gorm.Open(sqlite.Open(config.Config.Sqlite.Path) ,&gorm.Config{})
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
		&model.History{},
		&model.Library{},
	)
}
