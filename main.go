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
	e.File("/", "public/hello.html")

	light := e.Group("/light")
	light.File("/waiting", "public/waiting.html")
	light.File("/files", "public/files.html")
	light.File("/search", "public/search.html")
	light.File("/info", "public/info.html")
	light.File("/backup", "public/backup.html")

	// api 只使用 GET, POST, 不采用 RESTful.
	api := e.Group("/api")
	api.Use(sleep)
	api.GET("/get-db-info", databaseInfo)
	api.GET("/force-check-files", forceCheckFiles)
	api.GET("/get-bk-buckets", backupBuckets)
	api.POST("/add-bk-bucket", addBackupBucket)
	api.POST("/delete-bk-bucket", deleteBackupBucket)
	api.POST("/get-buckets-info", bucketsInfo)
	api.GET("/waitingFolder", waitingFolder)
	api.GET("/waiting-files", waitingFiles)
	api.GET("/all-files", allFiles) // file.Deleted == false
	api.POST("/add-files", addFiles, autoCheck)
	api.POST("/delete-file", deleteFile)
	api.POST("/update-tags", updateTags)
	api.POST("/rename-file", renameFile)

	api.POST("/search-tags", searchTags)
	// api.POST("/search-title", searchTitle)

	log.Print("localtags database path: ", dbPath)
	e.Logger.Fatal(e.Start(":80"))
}
