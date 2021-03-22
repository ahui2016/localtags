package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ahui2016/localtags/model"
	"github.com/ahui2016/localtags/thumb"
	"github.com/ahui2016/localtags/util"
	"github.com/labstack/echo/v4"
)

/*
func getParamTags(c echo.Context) ([]string, error) {
	tagsStr, err := getParam(c, "tags")
	return strings.Split(tagsStr, " "), err
}

func getParam(c echo.Context, key string) (string, error) {
	return url.QueryUnescape(c.Param(key))
}
*/

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
func getWaitingFiles() ([]string, error) {
	pattern := filepath.Join(cfg.WaitingFolder, "*")
	return filepath.Glob(pattern)
}

func cleanTempFolders() error {
	err1 := os.RemoveAll(cfg.WaitingFolder)
	err2 := os.RemoveAll(tempFolder)
	util.MustMkdir(cfg.WaitingFolder)
	util.MustMkdir(tempFolder)
	return util.WrapErrors(err1, err2)
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

// infoToFile 把 waiting 文件夹里的文件转换为 model.File,
// 如果遇到文件夹则返回错误，如果遇到新文件与数据库中的文件同名，则自动获取标签。
func infoToFile(name string, meta map[string]*File) (
	file *File, err error) {

	info, err := os.Lstat(name)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, errors.New(`"waiting" 里面不可存放文件夹`)
	}

	// 填充文件体积、文件名、文件类型
	file = &File{Size: info.Size()}
	file.SetNameType(info.Name())

	// 填充同名文件数
	ids, err := db.GetFileIDsByName(file.Name)
	if err != nil {
		return nil, err
	}
	file.Count = int64(len(ids))

	// 填充文件标签
	if file.Count > 0 {
		tags, err := db.GetTagsByFile(ids[0])
		if err != nil {
			return nil, err
		}
		file.Tags = tags
	}

	// 填充文件哈希值
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

	// 填充文件 ID
	file.ID = model.RandomID()
	thumbPath := tempThumb(file.ID)

	// 填充文件缩略图
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

	// 全部填充完毕，返回文件
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

// 注意这个函数即使错误也返回一个有用的 map
func filesToMeta(files []*File) (meta map[string]*File, err error) {
	meta = make(map[string]*File)
	for _, file := range files {
		if f, ok := meta[file.Hash]; ok {
			err = util.WrapErrors(err, fmt.Errorf("[%s] 与 [%s] 重复了（两个文件内容相同）", file.Name, f.Name))
			continue
		}
		meta[file.Hash] = file
	}
	return
}

func indexByName(meta map[string]*File) map[string]*File {
	byName := make(map[string]*File)
	for _, file := range meta {
		name := file.ID + thumbSuffix
		byName[name] = file
	}
	return byName
}

// getFormValue gets the c.FormValue(key), trims its spaces,
// and checks if it is empty or not.
func getFormValue(c echo.Context, key string) (string, error) {
	value := strings.TrimSpace(c.FormValue(key))
	if value == "" {
		return "", errors.New(key + " is empty")
	}
	return value, nil
}

func getTags(c echo.Context) ([]string, error) {
	tagsString, err := getFormValue(c, "tags")
	if err != nil {
		return nil, err
	}
	var tags []string
	err = json.Unmarshal([]byte(tagsString), &tags)
	return tags, err
}

// tryFileName 检查文件名是否符合操作系统的要求。
func tryFileName(name string) error {
	fullpath := filepath.Join(tempFolder, name)
	return os.WriteFile(fullpath, []byte("abc"), 0666)
}
