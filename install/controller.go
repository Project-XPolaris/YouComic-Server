package install

import (
	"github.com/allentom/haruka"
	"net/http"
)

var IndexController haruka.RequestHandler = func(context *haruka.Context) {
	context.HTML("assets/install/templates/index.html", map[string]interface{}{})
}

var SettingDatabaseController haruka.RequestHandler = func(context *haruka.Context) {
	context.HTML("assets/install/templates/database.html", map[string]interface{}{})
}
var SettingMysqlController haruka.RequestHandler = func(context *haruka.Context) {
	context.HTML("assets/install/templates/mysql.html", map[string]interface{}{})
}
var SettingSqliteController haruka.RequestHandler = func(context *haruka.Context) {
	context.HTML("assets/install/templates/sqlite.html", map[string]interface{}{})
}

type SqliteDatabaseSettingForm struct {
	Path string `hsource:"form" hname:"path"`
}

var SettingSqliteSubmitController haruka.RequestHandler = func(context *haruka.Context) {
	var form SqliteDatabaseSettingForm
	err := context.BindingInput(&form)
	if err != nil {
		panic(err)
	}
	PreinstallConfig.Sqlite.Path = form.Path
	PreinstallConfig.Database.Type = "sqlite"
	http.Redirect(context.Writer, context.Request, "/store", http.StatusFound)
}

type MysqlDatabaseSettingForm struct {
	Host     string `hsource:"form" hname:"host"`
	Port     string `hsource:"form" hname:"port"`
	Database string `hsource:"form" hname:"database"`
	Username string `hsource:"form" hname:"username"`
	Password string `hsource:"form" hname:"password"`
}

var SettingMysqlSubmitController haruka.RequestHandler = func(context *haruka.Context) {
	var form MysqlDatabaseSettingForm
	err := context.BindingInput(&form)
	if err != nil {
		panic(err)
	}
	PreinstallConfig.Mysql.Host = form.Host
	PreinstallConfig.Mysql.Port = form.Port
	PreinstallConfig.Mysql.Database = form.Database
	PreinstallConfig.Mysql.Password = form.Password
	PreinstallConfig.Mysql.Username = form.Username
	http.Redirect(context.Writer, context.Request, "/store", http.StatusFound)
}

var SettingStoreController haruka.RequestHandler = func(context *haruka.Context) {
	context.HTML("assets/install/templates/store.html", map[string]interface{}{})
}

type StoreSettingForm struct {
	Root  string `hsource:"form" hname:"root"`
	Books string `hsource:"form" hname:"books"`
}

var SettingStoreSubmitController haruka.RequestHandler = func(context *haruka.Context) {
	var form StoreSettingForm
	err := context.BindingInput(&form)
	if err != nil {
		panic(err)
	}
	PreinstallConfig.Store.Root = form.Root
	PreinstallConfig.Store.Books = form.Books
	http.Redirect(context.Writer, context.Request, "/security", http.StatusFound)
}

var SettingSecurityController haruka.RequestHandler = func(context *haruka.Context) {
	context.HTML("assets/install/templates/security.html", map[string]interface{}{})
}

type SecuritySettingForm struct {
	Username string `hsource:"form" hname:"username"`
	Password string `hsource:"form" hname:"password"`
}

var SettingSecuritySubmitController haruka.RequestHandler = func(context *haruka.Context) {
	var form SecuritySettingForm
	err := context.BindingInput(&form)
	if err != nil {
		panic(err)
	}
	InitUsername = form.Username
	InitPassword = form.Password
	http.Redirect(context.Writer, context.Request, "/application", http.StatusFound)
}

var SettingApplicationController haruka.RequestHandler = func(context *haruka.Context) {
	context.HTML("assets/install/templates/application.html", map[string]interface{}{})
}

type ApplicationSettingForm struct {
	Port string `hsource:"form" hname:"port"`
}

var SettingApplicationSubmitController haruka.RequestHandler = func(context *haruka.Context) {
	var form ApplicationSettingForm
	err := context.BindingInput(&form)
	if err != nil {
		panic(err)
	}
	PreinstallConfig.Application.Port = form.Port

	generateSettingFile()
	http.Redirect(context.Writer, context.Request, "/complete", http.StatusFound)
}
var SettingCompleteController haruka.RequestHandler = func(context *haruka.Context) {
	context.HTML("assets/install/templates/complete.html", map[string]interface{}{})
}
