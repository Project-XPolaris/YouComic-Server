package router

import (
	"github.com/allentom/youcomic-api/controller"
	"github.com/gin-gonic/gin"
)

func SetRouter(engine *gin.Engine) {
	engine.POST("/books", controller.CreateBookHandler)
	engine.POST("/books/upload", controller.CreateBook)
	engine.PATCH("/book/:id", controller.UpdateBookHandler)
	engine.GET("/book/:id", controller.GetBook)
	engine.PUT("/book/:id/tags", controller.BookTagBatch)
	engine.PUT("/book/:id/cover", controller.AddBookCover)
	engine.PUT("/book/:id/pages", controller.AddBookPages)
	engine.DELETE("/book/:id/tag/:tag", controller.DeleteBookTag)
	engine.GET("/books", controller.BookListHandler)
	engine.DELETE("/book/:id", controller.DeleteBookHandler)
	engine.POST("/books/batch", controller.BookBatchHandler)
	engine.POST("/pages", controller.PageUploadHandler)
	engine.PATCH("/page/:id", controller.UpdatePageHandler)
	engine.DELETE("/page/:id", controller.DeletePageHandler)
	engine.GET("/pages", controller.PageListHandler)
	engine.POST("/pages/batch", controller.BatchPageHandler)
	engine.POST("/tags", controller.CreateTagHandler)
	engine.POST("/tags/batch", controller.BatchTagHandler)
	engine.GET("/tags", controller.TagListHandler)
	engine.GET("/book/:id/tags", controller.GetBookTags)
	engine.GET("/tag/:id/books", controller.TagBooksHandler)
	engine.PUT("/tag/:id/books", controller.AddBooksToTagHandler)
	engine.PUT("/tag/:id/subscription", controller.AddSubscriptionUser)
	engine.DELETE("/tag/:id/subscription", controller.RemoveSubscriptionUser)
	engine.DELETE("/tag/:id/books", controller.RemoveBooksFromTagHandler)
	engine.GET("/tag/:id", controller.GetTag)
	engine.POST("/user/register", controller.RegisterUserHandler)
	engine.POST("/user/auth", controller.LoginUserHandler)
	engine.GET("/user/:id", controller.GetUserHandler)
	engine.GET("/user/:id/groups", controller.GetUserUserGroupsHandler)
	engine.GET("/users", controller.GetUserUserListHandler)
	engine.POST("/collections", controller.CreateCollectionHandler)
	engine.GET("/collections", controller.CollectionsListHandler)
	engine.PUT("/collection/:id/books", controller.AddToCollectionHandler)
	engine.DELETE("/collection/:id/books", controller.DeleteFromCollectionHandler)
	engine.PUT("/collection/:id/users", controller.AddUsersToCollectionHandler)
	engine.DELETE("/collection/:id/users", controller.DeleteUsersFromCollectionHandler)
	engine.DELETE("/collection/:id", controller.DeleteCollectionHandler)
	engine.PATCH("/collection/:id", controller.UpdateCollectionHandler)
	engine.GET("/permissions", controller.GetPermissionListHandler)
	engine.GET("/usergroups", controller.GetUserGroupListHandler)
	engine.POST("/usergroups", controller.CreateUserGroupHandler)
	engine.PUT("/usergroup/:id/users", controller.AddUserToUserGroupHandler)
	engine.PUT("/usergroup/:id/permissions", controller.AddPermissionToUserGroupHandler)
	engine.DELETE("/usergroup/:id/users", controller.RemoveUserFromUserGroupHandler)
	engine.DELETE("/usergroup/:id/permissions", controller.RemovePermissionFromUserGroupHandler)
	engine.PUT("/user/password", controller.ChangeUserPasswordHandler)
	engine.PUT("/user/nickname", controller.ChangeUserNicknameHandler)
	engine.GET("/histories", controller.HistoryListHandler)
	engine.DELETE("/history/:id", controller.DeleteHistoryHandler)
	engine.GET("/account/histories", controller.UserHistoryHandler)
	engine.DELETE("/account/histories", controller.DeleteUserHistoryHandler)
	engine.GET("/content/book/:id/:fileName", controller.BookContentHandler)
	engine.POST("/libraries", controller.CreateLibraryHandler)
	engine.POST("/library/import", controller.ImportLibraryHandler)
	engine.POST("/library/batch", controller.LibraryBatchHandler)
	engine.DELETE("/library/:id", controller.DeleteLibraryHandler)
	engine.GET("/library/:id", controller.LibraryObjectHandler)
	engine.GET("/libraries", controller.LibraryListHandler)
	engine.GET("/dashboard/book/daily", controller.BookCountDailySummaryHandler)
	engine.GET("/dashboard/tag/books", controller.TagBooksCountHandler)
	engine.GET("/dashboard/tag/types", controller.TagTypeCountHandler)
}
