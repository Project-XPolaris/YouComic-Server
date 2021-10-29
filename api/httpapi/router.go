package httpapi

import (
	"github.com/allentom/haruka"
)

func SetRouter(engine *haruka.Engine) {
	engine.Router.POST("/books", CreateBookHandler)
	engine.Router.POST("/books/upload", CreateBook)
	engine.Router.PATCH("/book/{id:[0-9]+}", UpdateBookHandler)
	engine.Router.GET("/book/{id:[0-9]+}", GetBook)
	engine.Router.PUT("/book/{id:[0-9]+}/tags", BookTagBatch)
	engine.Router.PUT("/book/{id:[0-9]+}/cover", AddBookCover)
	engine.Router.POST("/book/{id:[0-9]+}/cover/generate", GenerateCoverThumbnail)
	engine.Router.PUT("/book/{id:[0-9]+}/pages", AddBookPages)
	engine.Router.PUT("/book/{id:[0-9]+}/dir/rename", RenameBookDirectoryHandler)
	engine.Router.DELETE("/book/{id:[0-9]+}/tag/{id:[0-9]+}", DeleteBookTag)
	engine.Router.GET("/books", BookListHandler)
	engine.Router.DELETE("/book/{id:[0-9]+}", DeleteBookHandler)
	engine.Router.POST("/books/batch", BookBatchHandler)
	engine.Router.POST("/pages", PageUploadHandler)
	engine.Router.PATCH("/page/{id:[0-9]+}", UpdatePageHandler)
	engine.Router.DELETE("/page/{id:[0-9]+}", DeletePageHandler)
	engine.Router.GET("/pages", PageListHandler)
	engine.Router.POST("/pages/batch", BatchPageHandler)
	engine.Router.POST("/tags", CreateTagHandler)
	engine.Router.POST("/tags/batch", BatchTagHandler)
	engine.Router.POST("/tags/addTag", AddTagBooksToTag)
	engine.Router.GET("/tags", TagListHandler)
	engine.Router.GET("/book/{id:[0-9]+}/tags", GetBookTags)
	engine.Router.GET("/tag/{id:[0-9]+}/books", TagBooksHandler)
	engine.Router.PUT("/tag/{id:[0-9]+}/books", AddBooksToTagHandler)
	engine.Router.PUT("/tag/{id:[0-9]+}/subscription", AddSubscriptionUser)
	engine.Router.DELETE("/tag/{id:[0-9]+}/subscription", RemoveSubscriptionUser)
	engine.Router.DELETE("/tag/{id:[0-9]+}/books", RemoveBooksFromTagHandler)
	engine.Router.POST("/tags/match", AnalyzeTagFromTextHandler)
	engine.Router.POST("/tags/clean", ClearEmptyTagHandler)
	engine.Router.GET("/tag/{id:[0-9]+}", GetTag)
	engine.Router.POST("/user/register", RegisterUserHandler)
	engine.Router.POST("/user/auth", LoginUserHandler)
	engine.Router.GET("/user/{id:[0-9]+}", GetUserHandler)
	engine.Router.GET("/user/{id:[0-9]+}/groups", GetUserUserGroupsHandler)
	engine.Router.GET("/users", GetUserUserListHandler)
	engine.Router.POST("/collections", CreateCollectionHandler)
	engine.Router.GET("/collections", CollectionsListHandler)
	engine.Router.PUT("/collection/{id:[0-9]+}/books", AddToCollectionHandler)
	engine.Router.DELETE("/collection/{id:[0-9]+}/books", DeleteFromCollectionHandler)
	engine.Router.PUT("/collection/{id:[0-9]+}/users", AddUsersToCollectionHandler)
	engine.Router.DELETE("/collection/{id:[0-9]+}/users", DeleteUsersFromCollectionHandler)
	engine.Router.DELETE("/collection/{id:[0-9]+}", DeleteCollectionHandler)
	engine.Router.PATCH("/collection/{id:[0-9]+}", UpdateCollectionHandler)
	engine.Router.GET("/permissions", GetPermissionListHandler)
	engine.Router.GET("/usergroups", GetUserGroupListHandler)
	engine.Router.POST("/usergroups", CreateUserGroupHandler)
	engine.Router.PUT("/usergroup/{id:[0-9]+}/users", AddUserToUserGroupHandler)
	engine.Router.PUT("/usergroup/{id:[0-9]+}/permissions", AddPermissionToUserGroupHandler)
	engine.Router.DELETE("/usergroup/{id:[0-9]+}/users", RemoveUserFromUserGroupHandler)
	engine.Router.DELETE("/usergroup/{id:[0-9]+}/permissions", RemovePermissionFromUserGroupHandler)
	engine.Router.PUT("/user/password", ChangeUserPasswordHandler)
	engine.Router.PUT("/user/nickname", ChangeUserNicknameHandler)
	engine.Router.GET("/histories", HistoryListHandler)
	engine.Router.DELETE("/history/{id:[0-9]+}", DeleteHistoryHandler)
	engine.Router.GET("/account/histories", UserHistoryHandler)
	engine.Router.DELETE("/account/histories", DeleteUserHistoryHandler)
	engine.Router.GET("/content/book/{id:[0-9]+}/{fileName}", BookContentHandler)
	engine.Router.POST("/libraries", CreateLibraryHandler)
	engine.Router.POST("/library/import", ImportLibraryHandler)
	engine.Router.POST("/library/batch", LibraryBatchHandler)
	engine.Router.DELETE("/library/{id:[0-9]+}", DeleteLibraryHandler)
	engine.Router.PUT("/library/{id:[0-9]+}/scan", ScanLibraryHandler)
	engine.Router.PUT("/library/{id:[0-9]+}/thumbnails", NewLibraryGenerateThumbnailsHandler)
	engine.Router.PUT("/library/{id:[0-9]+}/match", NewLibraryMatchTagHandler)
	engine.Router.POST("/library/{id:[0-9]+}/task/writemeta", WriteBookMetaTaskHandler)
	engine.Router.PUT("/library/{id:[0-9]+}/books/rename", NewRenameLibraryBookDirectoryHandler)
	engine.Router.GET("/library/{id:[0-9]+}", LibraryObjectHandler)
	engine.Router.GET("/libraries", LibraryListHandler)
	engine.Router.GET("/dashboard/book/daily", BookCountDailySummaryHandler)
	engine.Router.GET("/dashboard/tag/books", TagBooksCountHandler)
	engine.Router.GET("/dashboard/tag/types", TagTypeCountHandler)
	engine.Router.POST("/scan/tasks", NewScannerHandler)
	engine.Router.POST("/scan/stop", StopLibraryScanHandler)
	engine.Router.POST("/task/bookMove", NewMoveBookTaskHandler)
	engine.Router.GET("/explore/read", ReadDirectoryHandler)
	engine.Router.GET("/thumbnail/status", GetThumbnailGeneratorStatus)
	engine.Router.GET("/ws", WShandler)
}
