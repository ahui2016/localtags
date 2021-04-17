package main

import (
	"log"

	"github.com/labstack/echo/v4"
)

func main() {
	defer db.Close()

	e := echo.New()
	e.HTTPErrorHandler = errorHandler

	e.Static("/public", "public")
	e.Static("/temp", tempFolder)
	e.Static("/thumbs", thumbsFolder)
	e.Static("/mainbucket", mainBucket)
	e.File("/", "public/home.html")

	light := e.Group("/light")
	light.File("/waiting", "public/waiting.html")
	light.File("/files", "public/files.html")
	light.File("/search", "public/search.html")
	light.File("/info", "public/info.html")
	light.File("/backup", "public/backup.html")
	light.File("/tag-groups", "public/tag-groups.html")
	light.File("/tags", "public/tags.html")
	light.File("/tag", "public/tag.html")
	light.File("/home", "public/home.html")
	light.File("/add", "public/add.html")

	// api 只使用 GET, POST, 不采用 RESTful.
	api := e.Group("/api")
	api.Use(sleep)
	api.GET("/get-db-info", databaseInfo)
	api.GET("/force-check-files", forceCheckFiles)
	api.GET("/get-bk-buckets", getBackupBuckets)
	api.POST("/add-bk-bucket", addBackupBucket)
	api.POST("/delete-bk-bucket", deleteBackupBucket)
	api.POST("/get-buckets-info", bucketsInfo)
	api.POST("/sync-backup", syncBackup)
	api.POST("/repair-files", repairFiles)

	api.GET("/waitingFolder", waitingFolder)
	api.GET("/waiting-files", waitingFiles)
	api.GET("/all-files", allFiles) // file.Deleted == false
	api.GET("/deleted-files", deletedFiles)
	api.GET("/download/:id", downloadFile)
	api.POST("/add-files", addFiles, autoCheck)
	api.POST("/delete-file", deleteFile)
	api.POST("/undelete-file", undeleteFile)
	api.POST("/really-delete-file", reallyDeleteFile)
	api.POST("/update-tags", updateTags)
	api.POST("/rename-file", renameFile)

	api.POST("/get-groups-by-tag", getGroupsByTag)

	api.GET("/tags-by-date", allTagsByDate)
	api.GET("/tags-by-name", allTagsByName)
	api.POST("/rename-tag", renameTag)
	api.POST("/is-tag-exist", isTagExist)
	api.POST("/delete-tag", deleteTag)
	api.POST("/add-taggroup", addTagGroup)
	api.GET("/tag-groups", getTagGroups)
	api.GET("/protect-taggroup/:id", protectTagGroup)
	api.GET("/unprotect-taggroup/:id", unprotectTagGroup)
	api.GET("/delete-taggroup/:id", deleteTagGroup)

	api.POST("/search-tags", searchTags)
	api.POST("/search-title", searchTitle)
	api.POST("/search-by-id", searchByID)

	log.Print("localtags database path: ", dbPath)
	e.Logger.Fatal(e.Start(cfg.Address))
}
