package main

import (
	"encoding/json"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ahui2016/localtags/model"
	"github.com/ahui2016/localtags/thumb"
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

func errorHandler(err error, c echo.Context) {
	if e, ok := err.(*echo.HTTPError); ok {
		c.JSON(e.Code, e.Message)
	}
	c.JSON(500, Text{err.Error()})
}

func waitingFolder(c echo.Context) error {
	return c.JSON(OK, Text{cfg.WaitingFolder})
}

func waitingFiles(c echo.Context) error {
	fileNames, err1 := getTempFiles()
	metadata, err2 := getMetadata()
	if err := util.WrapErrors(err1, err2); err != nil {
		return err
	}

	// 即将在临时文件夹里写数据，因此先确保该文件夹存在。
	util.MustMkdir(tempFolder)
	var files []File
	for _, name := range fileNames {
		info, err := os.Lstat(name)
		if err != nil {
			return err
		}
		if info.IsDir() {
			continue
		}
		file, err1 := infoToFile(name, info, metadata)
		if err1 != nil {
			// 在函数 infoToFile 中可能生成一些缩略图，如果发生错误，要删除这些缩略图。
			err2 := cleanThumbFiles(files)
			return util.WrapErrors(err1, err2)
		}
		files = append(files, file)
	}

	// 更新 metadata, 因为文件有可能已经发生变化。
	metadata = filesToMeta(files)
	util.MarshalWrite(metadata, tempMetadata)
	return c.JSON(OK, files)
}

func cleanThumbFiles(files []File) error {
	for _, file := range files {
		if err := os.Remove(strings.TrimPrefix(file.Thumb, "/")); err != nil {
			return err
		}
	}
	return nil
}

func getTempFiles() ([]string, error) {
	pattern := filepath.Join(cfg.WaitingFolder, "*")
	return filepath.Glob(pattern)
}

func infoToFile(name string, info fs.FileInfo, meta map[string]File) (
	file File, err error) {

	file = File{Size: info.Size()}
	file.SetNameType(info.Name())

	fileBytes, err := os.ReadFile(name)
	if err != nil {
		return
	}
	file.Hash = util.Sha256Hex(fileBytes)

	// 如果文件已经在 metadata 里，则不进行处理，立即返回。
	if metaFile, ok := meta[file.Hash]; ok {
		file.ID = metaFile.ID
		file.Thumb = metaFile.Thumb
		return
	}

	file.ID = model.RandomID()
	thumbPath := filepath.Join(tempFolder, file.ID+".jpg")

	if strings.HasPrefix(file.Type, "image") {
		file.Thumb = "/" + filepath.ToSlash(thumbPath)
		// 注意下面这个 err 是个新变量，不同于函数返回值的那个 err.
		if err := thumb.NailWrite(name, thumbPath); err != nil {
			// 如果生成缩略图失败，可能原图已损坏，或根本不是图片（后缀名错误）。
			file.Thumb = ""
		}
	}

	if hasFFmpeg && strings.HasPrefix(file.Type, "video") {
		file.Thumb = "/" + filepath.ToSlash(thumbPath)
		// 注意下面这个 err 是个新变量，不同于函数返回值的那个 err.
		if err := thumb.FrameNail(name, thumbPath, 10); err != nil {
			// 如果截图失败，可能视频已损坏，或根本不是视频（后缀名错误）。
			file.Thumb = ""
		}
	}
	return
}

func getMetadata() (map[string]File, error) {
	metadata := make(map[string]File)
	metaJSON, err := os.ReadFile(tempMetadata)
	if err != nil {
		// 如果读取文件失败，则反回一个空的 metadata, 不处理错误。
		return metadata, nil
	}
	err = json.Unmarshal(metaJSON, &metadata)
	return metadata, err
}

func filesToMeta(files []File) map[string]File {
	meta := make(map[string]File)
	for _, file := range files {
		meta[file.Hash] = file
	}
	return meta
}

func checkFFmpeg(c echo.Context) error {
	ok := thumb.CheckFFmpeg()
	return c.JSON(OK, ok)
}

func addFiles(c echo.Context) error {
	value := c.FormValue("files-tags")
	var filesTags map[string][]string // hash[tags]
	if err := json.Unmarshal([]byte(value), &filesTags); err != nil {
		return err
	}
	log.Print(filesTags)
	return c.NoContent(OK)
}
