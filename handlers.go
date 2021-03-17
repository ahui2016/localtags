package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
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

func allFiles(c echo.Context) error {
	files, err := db.AllFiles()
	if err != nil {
		return err
	}
	return c.JSON(OK, files)
}

func waitingFiles(c echo.Context) error {
	fileNames, err1 := getTempFiles()
	metadata, err2 := getMetadata()
	if err := util.WrapErrors(err1, err2); err != nil {
		return err
	}

	var files []*File
	for _, name := range fileNames {
		info, err := os.Lstat(name)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return errors.New(`"waiting" 里面不可存放文件夹`)
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
	// 在 filesToMeta 里还会检查有没有重复的文件。
	if metadata, err1 = filesToMeta(files); err1 != nil {
		return err1
	}
	util.MarshalWrite(metadata, tempMetadata)
	return c.JSON(OK, files)
}

func cleanThumbFiles(thumbFiles []*File) error {
	var files []string
	for _, file := range thumbFiles {
		if file.Thumb {
			files = append(files, tempThumb(file.ID))
		}
	}
	err := util.DeleteFiles(files)
	if util.ErrorContains(err, "cannot find") {
		err = nil
	}
	return err
}

func getTempFiles() ([]string, error) {
	pattern := filepath.Join(cfg.WaitingFolder, "*")
	return filepath.Glob(pattern)
}

func infoToFile(name string, info fs.FileInfo, meta map[string]*File) (
	file *File, err error) {

	file = &File{Size: info.Size()}
	file.SetNameType(info.Name())

	fileBytes, err := os.ReadFile(name)
	if err != nil {
		return
	}
	file.Hash = util.Sha256Hex(fileBytes)

	id, ok := db.GetFileID(file.Hash)
	if ok {
		return nil, fmt.Errorf("文件 [%s] 已存在于数据库中: id[%s]", file.Name, id)
	}

	// 如果文件已经在 metadata 里，则不进行处理，立即返回。
	if metaFile, ok := meta[file.Hash]; ok {
		file.ID = metaFile.ID
		file.Thumb = metaFile.Thumb
		return
	}

	file.ID = model.RandomID()
	thumbPath := tempThumb(file.ID)

	if strings.HasPrefix(file.Type, "image") {
		file.Thumb = true
		// 注意下面这个 err 是个新变量，不同于函数返回值的那个 err.
		if err := thumb.NailWrite(name, thumbPath); err != nil {
			// 如果生成缩略图失败，可能原图已损坏，或根本不是图片（后缀名错误）。
			file.Thumb = false
		}
	}

	if hasFFmpeg && strings.HasPrefix(file.Type, "video") {
		file.Thumb = true
		// 注意下面这个 err 是个新变量，不同于函数返回值的那个 err.
		if err := thumb.FrameNail(name, thumbPath, 10); err != nil {
			// 如果截图失败，可能视频已损坏，或根本不是视频（后缀名错误）。
			file.Thumb = false
		}
	}
	return
}

func getMetadata() (map[string]*File, error) {
	metadata := make(map[string]*File)
	metaJSON, err := os.ReadFile(tempMetadata)
	if err != nil {
		// 如果读取文件失败，则反回一个空的 metadata, 不处理错误。
		return metadata, nil
	}
	err = json.Unmarshal(metaJSON, &metadata)
	return metadata, err
}

func filesToMeta(files []*File) (map[string]*File, error) {
	meta := make(map[string]*File)
	for _, file := range files {
		if f, ok := meta[file.Hash]; ok {
			return nil, fmt.Errorf("[%s] 与 [%s] 重复了（两个文件内容相同）", file.Name, f.Name)
		}
		meta[file.Hash] = file
	}
	return meta, nil
}

func checkFFmpeg(c echo.Context) error {
	ok := thumb.CheckFFmpeg()
	return c.JSON(OK, ok)
}

func addFiles(c echo.Context) error {
	value := c.FormValue("hash-tags")
	var hashTags map[string][]string
	if err := json.Unmarshal([]byte(value), &hashTags); err != nil {
		return err
	}
	metadata, err := getMetadata()
	if err != nil {
		return err
	}
	var (
		copiedFiles []string
		files       []*File // files to be insert
	)
	for _, file := range metadata {
		f := db.NewFile()
		if err := copyTempThumb(file, f, &copiedFiles); err != nil {
			return err
		}
		if err := copyTempFile(file, f, &copiedFiles); err != nil {
			return err
		}
		file.ID = f.ID
		file.CTime = f.CTime
		file.UTime = f.UTime
		file.SetTags(hashTags[file.Hash])
		files = append(files, file)
	}

	// insert files to the database.
	if err := db.InsertFiles(files); err != nil {
		return util.WrapErrors(err, util.DeleteFiles(copiedFiles))
	}

	// 如果一切正常，就清空全部临时文件。
	return cleanTempFolders()
}

func copyFile(dstPath, srcPath string, copied *[]string) error {
	if err := util.CopyFile(dstPath, srcPath); err != nil {
		return util.WrapErrors(err, util.DeleteFiles(*copied))
	}
	*copied = append(*copied, dstPath)
	return nil
}

func copyTempFile(tempFile, newFile *File, copied *[]string) error {
	srcPath := waitingFile(tempFile.Name)
	dstPath := mainBucketFile(newFile.ID)
	return copyFile(dstPath, srcPath, copied)
}

func copyTempThumb(tempFile, newFile *File, copied *[]string) error {
	if !tempFile.Thumb {
		return nil
	}
	srcPath := tempThumb(tempFile.ID)
	dstPath := mainBucketThumb(newFile.ID)
	return copyFile(dstPath, srcPath, copied)
}

// tempThumb 使用 id 组成临时缩略图的位置。
func tempThumb(tempID string) string {
	return filepath.Join(tempFolder, tempID+thumbSuffix)
}
func waitingFile(name string) string {
	return filepath.Join(cfg.WaitingFolder, name)
}
func mainBucketFile(id string) string {
	return filepath.Join(mainBucket, id)
}
func mainBucketThumb(id string) string {
	return filepath.Join(thumbsFolder, id)
}
func cleanTempFolders() error {
	err1 := os.RemoveAll(cfg.WaitingFolder)
	err2 := os.RemoveAll(tempFolder)
	util.MustMkdir(cfg.WaitingFolder)
	util.MustMkdir(tempFolder)
	return util.WrapErrors(err1, err2)
}
