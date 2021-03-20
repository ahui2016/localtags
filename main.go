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

	// api 只使用 GET, POST, 不采用 RESTful.
	api := e.Group("/api")
	api.Use(sleep)
	api.GET("/waitingFolder", waitingFolder)
	api.GET("/waiting-files", waitingFiles)
	api.GET("/check", checkFFmpeg)
	api.GET("/all-files", allFiles) // file.Deleted == false
	api.POST("/add-files", addFiles)
	api.POST("/delete-file", deleteFile)

	api.GET("/search-tags/:tags", searchTags)
	// api.GET("/search-title/:pattern", searchTitle)

	log.Print("localtags database path: ", dbPath)
	e.Logger.Fatal(e.Start(":80"))
}
