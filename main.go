package main

import (
	"log"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.Static("/public", "public")
	e.File("/", "public/hello.html")

	light := e.Group("/light")
	light.File("/waiting", "public/waiting.html")

	api := e.Group("/api")
	api.GET("/waitingFolder", waitingFolder)
	api.GET("/waitingFiles", waitingFiles)
	api.GET("/check", checkFFmpeg)

	log.Print("localtags database path: ", dbPath)
	e.Logger.Fatal(e.Start(":80"))
}
