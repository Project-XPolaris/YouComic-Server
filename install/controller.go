package install

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

var IndexController gin.HandlerFunc = func(context *gin.Context) {
	context.HTML(http.StatusOK, "index.html", gin.H{})
}

var SettingDatabaseController gin.HandlerFunc = func(context *gin.Context) {
	context.HTML(http.StatusOK, "database.html", gin.H{})
}
var SettingMysqlController gin.HandlerFunc = func(context *gin.Context) {
	context.HTML(http.StatusOK, "mysql.html", gin.H{})
}
var SettingSqliteController gin.HandlerFunc = func(context *gin.Context) {
	context.HTML(http.StatusOK, "sqlite.html", gin.H{})
}

type SqliteDatabaseSettingForm struct {
	Path string `form:"path"`
}
var SettingSqliteSubmitController gin.HandlerFunc = func(context *gin.Context) {
	var form SqliteDatabaseSettingForm
	err := context.ShouldBind(&form)
	if err != nil {
		panic(err)
	}
	PreinstallConfig.Sqlite.Path = form.Path
	PreinstallConfig.Database.Type = "sqlite"
	context.Redirect(http.StatusFound,"/store")
}

type MysqlDatabaseSettingForm struct {
	Host string `form:"host"`
	Port string `form:"port"`
	Database string `form:"database"`
	Username string `form:"username"`
	Password string `form:"password"`
}
var SettingMysqlSubmitController gin.HandlerFunc = func(context *gin.Context) {
	var form MysqlDatabaseSettingForm
	err := context.ShouldBind(&form)
	if err != nil {
		panic(err)
	}
	PreinstallConfig.Mysql.Host = form.Host
	PreinstallConfig.Mysql.Port = form.Port
	PreinstallConfig.Mysql.Database = form.Database
	PreinstallConfig.Mysql.Password = form.Password
	PreinstallConfig.Mysql.Username = form.Username
	context.Redirect(http.StatusFound,"/store")
}


var SettingStoreController gin.HandlerFunc = func(context *gin.Context) {
	context.HTML(http.StatusOK, "store.html", gin.H{})
}
type StoreSettingForm struct {
	Root string `form:"root"`
	Books string `form:"books"`
}
var SettingStoreSubmitController gin.HandlerFunc = func(context *gin.Context) {
	var form StoreSettingForm
	err := context.ShouldBind(&form)
	if err != nil {
		panic(err)
	}
	PreinstallConfig.Store.Root = form.Root
	PreinstallConfig.Store.Books = form.Books
	context.Redirect(http.StatusFound,"/security")
}

var SettingSecurityController gin.HandlerFunc = func(context *gin.Context) {
	context.HTML(http.StatusOK, "security.html", gin.H{})
}

type SecuritySettingForm struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

var SettingSecuritySubmitController gin.HandlerFunc = func(context *gin.Context) {
	var form SecuritySettingForm
	err := context.ShouldBind(&form)
	if err != nil {
		panic(err)
	}
	InitUsername = form.Username
	InitPassword = form.Password
	context.Redirect(http.StatusFound,"/application")

}

var SettingApplicationController gin.HandlerFunc = func(context *gin.Context) {
	context.HTML(http.StatusOK, "application.html", gin.H{})
}
type ApplicationSettingForm struct {
	Port string `form:"port"`
}
var SettingApplicationSubmitController gin.HandlerFunc = func(context *gin.Context) {
	var form ApplicationSettingForm
	err := context.ShouldBind(&form)
	if err != nil {
		panic(err)
	}
	PreinstallConfig.Application.Port = form.Port

	generateSettingFile()
	context.Redirect(http.StatusFound,"/complete")

}
var SettingCompleteController gin.HandlerFunc = func(context *gin.Context) {
	context.HTML(http.StatusOK, "complete.html", gin.H{})
}