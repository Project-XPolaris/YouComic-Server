package database

import (
	"fmt"
	"github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
)

var DB *gorm.DB

func ConnectDatabase() {
	var err error
	mysqlConfig := config.Config.Mysql
	log.Println(mysqlConfig)
	connectString := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		mysqlConfig.Username,
		mysqlConfig.Password,
		mysqlConfig.Host,
		mysqlConfig.Port,
		mysqlConfig.Database,
	)
	DB, err = gorm.Open("mysql", connectString)
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
