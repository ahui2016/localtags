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
	e.File("/", "public/hello.html")

	light := e.Group("/light")
	light.File("/waiting", "public/waiting.html")

	// api 只使用 GET, POST, 不采用 RESTful.
	api := e.Group("/api")
	api.GET("/waitingFolder", waitingFolder)
	api.GET("/waiting-files", waitingFiles)
	api.GET("/check", checkFFmpeg)
	api.GET("/all-files", allFiles) // file.Deleted == false
	api.POST("/add-files", addFiles)

	log.Print("localtags database path: ", dbPath)
	e.Logger.Fatal(e.Start(":80"))
}
