package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ahui2016/localtags/model"
	"github.com/ahui2016/localtags/thumbnail"
	"github.com/ahui2016/localtags/util"
	"github.com/labstack/echo/v4"
)

type (
	File = model.File
)

// Text 用于向前端返回一个简单的文本消息。
// 为了保持一致性，总是向前端返回 JSON, 因此即使是简单的文本消息也使用 JSON.
type Text struct {
	Message string `json:"message"`
}

func waitingFolder(c echo.Context) error {
	return c.JSON(OK, Text{cfg.WaitingFolder})
}

func waitingFiles(c echo.Context) error {
	pattern := filepath.Join(cfg.WaitingFolder, "*")
	fileNames, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	var files []File
	for _, name := range fileNames {
		info, err := os.Lstat(name)
		if err != nil {
			return err
		}
		if info.IsDir() {
			continue
		}
		file := File{Size: info.Size()}
		file.SetNameType(info.Name())

		if hasFFmpeg && strings.HasPrefix(file.Type, "video") {
			log.Print(tempFolder)
			util.MustMkdir(tempFolder)
			thumb := tempThumb(name)
			if err := thumbnail.OneFrame(name, thumb, 10); err != nil {
				return err
			}
			file.Thumb = thumb
		}

		files = append(files, file)
	}
	return c.JSON(OK, files)
}

func tempThumb(filePath string) (tempThumbPath string) {
	thumbName := filepath.Base(filePath) + ".jpg"
	return filepath.Join(tempFolder, thumbName)
}

func checkFFmpeg(c echo.Context) error {
	ok := thumbnail.CheckFFmpeg()
	return c.JSON(OK, ok)
}
