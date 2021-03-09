package main

import (
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.File("/", "public/hello.html")
	e.Logger.Fatal(e.Start(":80"))
}
