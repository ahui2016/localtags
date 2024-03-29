package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ahui2016/localtags/database"
	"github.com/ahui2016/localtags/model"
	"github.com/ahui2016/localtags/stmt"
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

func getWaitingFolder(c echo.Context) error {
	return c.JSON(OK, Text{db.Config.WaitingFolder})
}

func allFiles(c echo.Context) error {
	files, err := db.AllFiles()
	if err != nil {
		return err
	}
	return c.JSON(OK, files)
}

func allImages(c echo.Context) error {
	files, err := db.AllImages()
	if err != nil {
		return err
	}
	return c.JSON(OK, files)
}

func deletedFiles(c echo.Context) error {
	files, err := db.DeletedFiles()
	if err != nil {
		return err
	}
	return c.JSON(OK, files)
}

func waitingFiles(c echo.Context) error {
	fileNames, err1 := getWaitingFiles()
	metadata, err2 := getMetadata()
	if err := util.WrapErrors(err1, err2); err != nil {
		return err
	}

	// 逐一处理 waiting 文件夹里的文件，将合格的（不是文件夹、不与数据库中的文件重复人）文件
	// 保存到 files 中。同时把全部错误合并为 allErr, 在 for 循环之后再统一处理 allErr。
	var files []*File
	var allErr error
	for _, name := range fileNames {
		file, err := infoToFile(name, metadata)
		if err != nil {
			allErr = util.WrapErrors(allErr, err)
			continue
		}
		files = append(files, file)
	}

	// 至此，不管在上面的 for 循环中有没有发生错误，我们都得到一个 files,
	// 该 files 反映了 waiting 文件夹的最新状态.

	// 更新 metadata, 因为文件有可能已经发生变化。
	// 在 filesToMeta 里还会检查有没有重复的文件。
	// 注意这里要先把 newMeta 写入硬盘，之后再处理错误。
	newMeta, err := filesToMeta(files)
	util.MustMarshalWrite(newMeta, tempMetadata)

	allErr = util.WrapErrors(allErr, err)
	if allErr != nil {
		return allErr
	}
	return c.JSON(OK, files)
}

func setWaitingTags(c echo.Context) error {
	tags, err := getTags(c)
	if err != nil {
		return err
	}
	metadata, err := getMetadata()
	if err != nil {
		return err
	}
	if len(metadata) == 0 {
		return fmt.Errorf("there is no file waiting to upload")
	}
	for i := range metadata {
		metadata[i].Tags = tags
	}
	return util.MarshalWrite(metadata, tempMetadata)
}

func setWaitingTag(c echo.Context) error {
	tags, e1 := getTags(c)
	hash, e2 := getFormValue(c, "hash")
	if err := util.WrapErrors(e1, e2); err != nil {
		return err
	}
	metadata, err := getMetadata()
	if err != nil {
		return err
	}
	if len(metadata) == 0 {
		return fmt.Errorf("there is no file waiting to upload")
	}
	_, ok := metadata[hash]
	if !ok {
		return fmt.Errorf("not found: " + hash)
	}
	metadata[hash].Tags = tags
	return util.MarshalWrite(metadata, tempMetadata)
}

// replaceFile 替换同名文件，如果有多个同名文件，则替换最新那个（而不是最旧那个）。
func replaceFile(c echo.Context) error {
	id, e1 := getID(c)
	hash, e2 := getFormValue(c, "hash")
	if err := util.WrapErrors(e1, e2); err != nil {
		return err
	}
	metadata, err := getMetadata()
	if err != nil {
		return err
	}
	srcFile, ok := metadata[hash]
	if !ok {
		return fmt.Errorf("not found: %s", hash)
	}

	// 如果系统中没有同名文件，则无法替换。
	if !db.IsFileExist(id) {
		return fmt.Errorf("can not replace, it's a new file name: %s", srcFile.Name)
	}

	var copiedFiles []string
	dstFile := &File{ID: id}

	// 先尝试移动文件，不行再复制文件。
	if err := moveTempFile(srcFile, dstFile); err != nil {
		if err = copyTempFile(srcFile, dstFile, &copiedFiles); err != nil {
			return err
		}
	}
	// thumb 文件总是在同一个硬盘分区，因此总能移动成功，不需要复制。
	if err := moveTempThumb(srcFile, dstFile); err != nil {
		return err
	}

	// 需要用到 ID, Size, Hash, UTime.
	// 其中 Size, Hash 已经在 srcFile 里了, ID 由前端提供, UTime 需要更新。
	srcFile.ID = id
	srcFile.UTime = model.TimeNow()
	if err := db.ReplaceFile(srcFile); err != nil {
		return util.WrapErrors(err, util.DeleteFiles(copiedFiles))
	}

	// 如果一切正常，就删除临时文件。
	if len(copiedFiles) > 0 {
		if err := os.Remove(waitingFile(srcFile.Name)); err != nil {
			return err
		}
	}

	// 这里不更新 metadata, 但要注意必须通过刷新前端页面来更新 metadata.
	return nil
}

func addFiles(c echo.Context) error {
	metadata, err := getMetadata()
	if err != nil {
		return err
	}
	for _, file := range metadata {
		if len(file.Tags) < 2 {
			return fmt.Errorf("every file needs at least two tags [%s]", file.Name)
		}
	}
	var (
		copiedFiles []string
		files       []*File // files to be insert
	)
	for _, file := range metadata {
		f := db.NewFile()
		// 先尝试移动文件，不行再复制文件。
		if err := moveTempFile(file, f); err != nil {
			if err = copyTempFile(file, f, &copiedFiles); err != nil {
				return err
			}
		}
		// thumb 文件总是在同一个硬盘分区，因此总能移动成功，不需要复制。
		if err := moveTempThumb(file, f); err != nil {
			return err
		}
		file.ID = f.ID
		file.CTime = f.CTime
		file.UTime = f.UTime
		files = append(files, file)
	}

	// insert files to the database.
	if err := db.InsertFiles(files); err != nil {
		return util.WrapErrors(err, util.DeleteFiles(copiedFiles))
	}

	// 如果一切正常，就清空全部临时文件。
	if len(copiedFiles) > 0 {
		if err := deleteTempFiles(files); err != nil {
			return err
		}
	}
	return os.Remove(tempMetadata)
}

func newNote(c echo.Context) error {
	contents := c.FormValue("contents")
	limit := 100 // title length limit
	firstLine := util.FirstLineLimit(strings.TrimSpace(contents), limit)
	title := util.GetMarkdownTitle(firstLine)
	filename := waitingFile(title) + ".md"
	if err := os.WriteFile(filename, []byte(contents), 0666); err != nil {
		return err
	}
	return c.JSON(OK, Text{filename})
}

func searchTags(c echo.Context) error {
	tags, err := getTags(c)
	if err != nil {
		return err
	}
	fileType := getFileType(c)
	files, err := db.SearchTags(tags, fileType)
	if err != nil {
		return err
	}
	return c.JSON(OK, files)
}

func searchTitle(c echo.Context) error {
	pattern, err := getFormValue(c, "pattern")
	if err != nil {
		return err
	}
	fileType := getFileType(c)
	files, err := db.SearchFileName(pattern, fileType)
	if err != nil {
		return err
	}
	return c.JSON(OK, files)
}

func searchByID(c echo.Context) error {
	id, err := getID(c)
	if err != nil {
		return err
	}
	files, err := db.SearchSameNameFiles(id)
	if err != nil {
		return err
	}
	return c.JSON(OK, files)
}

func searchDamaged(c echo.Context) error {
	files, err := db.SearchDamagedFiles()
	if err != nil {
		return err
	}
	return c.JSON(OK, files)
}

func searchBackupDamaged(c echo.Context) error {
	bkFolder, err := getFormValue(c, "bucket")
	if err != nil {
		return err
	}

	bkPath := filepath.Join(bkFolder, bakDBFileName)
	bk := new(database.DB)
	if err := bk.OpenBackup(bkPath, db.Config); err != nil {
		return err
	}
	defer bk.Close()

	files, err := bk.SearchDamagedFiles()
	if err != nil {
		return err
	}
	return c.JSON(OK, files)
}

func downloadFile(c echo.Context) error {
	id := c.Param("id")
	name, err := db.GetFileName(id)
	if err != nil {
		return err
	}
	filename := waitingFile(name)
	ok, err := util.PathIsExist(filename)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("文件已存在: %s", filename)
	}
	if err := util.CopyFile(filename, mainBucketFile(id)); err != nil {
		return err
	}
	return c.JSON(OK, Text{fmt.Sprintf("下载成功: %s", filename)})
}

func deleteFile(c echo.Context) error {
	id, err := getID(c)
	if err != nil {
		return err
	}
	if err = checkFileExist(id); err != nil {
		return err
	}
	return db.Exec(stmt.SetFileDeletedNow, true, model.TimeNow(), id)
}

func undeleteFile(c echo.Context) error {
	id, err := getID(c)
	if err != nil {
		return err
	}
	if err = checkFileExist(id); err != nil {
		return err
	}
	return db.Exec(stmt.SetFileDeletedNow, false, model.TimeNow(), id)
}

func reallyDeleteFile(c echo.Context) error {
	id, err := getID(c)
	if err != nil {
		return err
	}
	if err = checkFileExist(id); err != nil {
		return err
	}
	if err := os.Remove(mainBucketFile(id)); err != nil {
		return err
	}
	return db.DeleteFile(id)
}

func updateTags(c echo.Context) error {
	id, e1 := getID(c)
	tags, e2 := getTags(c)
	if err := util.WrapErrors(e1, e2); err != nil {
		return err
	}
	return db.UpdateTags(id, tags)
}

func renameFile(c echo.Context) error {
	id, e1 := getID(c)
	name, e2 := getFormValue(c, "name")
	if err := util.WrapErrors(e1, e2); err != nil {
		return err
	}
	if err := tryFileName(name); err != nil {
		return err
	}
	return db.RenameFiles(id, name)
}

func databaseInfo(c echo.Context) error {
	info, err := db.GetInfo()
	if err != nil {
		return err
	}
	return c.JSON(OK, info)
}

func getConfigHandler(c echo.Context) error {
	return c.JSON(OK, db.Config)
}

func updateConfig(c echo.Context) error {
	cfg2, err := getConfig(c)
	if err != nil {
		return err
	}
	return util.MarshalWrite(cfg2, configFile)
}

func forceCheckFiles(c echo.Context) error {
	return db.ForceCheckFilesHash(mainBucket)
}

func checkNow(c echo.Context) error {
	return db.CheckFilesHash(mainBucket)
}

func getBackupBuckets(c echo.Context) error {
	buckets, err := db.GetBackupBuckets()
	if err != nil {
		return err
	}
	return c.JSON(OK, buckets)
}

func addBackupBucket(c echo.Context) error {
	bucket, err := getFormValue(c, "bucket")
	if err != nil {
		return err
	}
	if err = checkBucketFolder(bucket); err != nil {
		return err
	}
	return db.AddBackupBucket(bucket)
}

func deleteBackupBucket(c echo.Context) error {
	i, err := getNumber(c, "index")
	if err != nil {
		return err
	}
	return db.DeleteBackupBucket(i)
}

func checkBackupNow(c echo.Context) error {
	i, err := getNumber(c, "index")
	if err != nil {
		return err
	}
	bkFolder, err := db.GetBackupFolder(i)
	if err != nil {
		return err
	}
	info, err := checkBackupGetInfo(bkFolder, true)
	if err != nil {
		return err
	}
	return c.JSON(OK, info)
}

func bucketsInfo(c echo.Context) error {
	bkFolder, err := getFormValue(c, "bucket")
	if err != nil {
		return err
	}
	info, err := getBucketsInfo(bkFolder)
	if err != nil {
		return err
	}
	return c.JSON(OK, info)
}

func syncBackup(c echo.Context) error {
	bkFolder, err := getFormValue(c, "bucket")
	if err != nil {
		return err
	}
	return syncMainToBackup(bkFolder)
}

func repairFiles(c echo.Context) error {
	bkFolder, err := getFormValue(c, "bucket")
	if err != nil {
		return err
	}
	return repairDamagedFiles(bkFolder)
}

func deleteBackupDamagedFiles(c echo.Context) error {
	bkFolder, err := getFormValue(c, "bucket")
	if err != nil {
		return err
	}
	return deleteDamagedFiles(bkFolder)
}

func addTagGroup(c echo.Context) error {
	tags, err := getTags(c)
	if err != nil {
		return err
	}
	group := model.NewTagGroup()
	group.SetTags(tags)
	if err := db.AddTagGroup(group); err != nil {
		return err
	}
	return c.JSON(OK, group)
}

func getTagGroups(c echo.Context) error {
	groups, err := db.TagGroups()
	if err != nil {
		return err
	}
	return c.JSON(OK, groups)
}

func protectTagGroup(c echo.Context) error {
	return db.Exec(stmt.SetTagGroupProtected, true, c.Param("id"))
}

func unprotectTagGroup(c echo.Context) error {
	return db.Exec(stmt.SetTagGroupProtected, false, c.Param("id"))
}

func deleteTagGroup(c echo.Context) error {
	return db.Exec(stmt.DeleteTagGroup, c.Param("id"))
}

func allTagsByDate(c echo.Context) error {
	tags, err := db.GetAllTags(stmt.AllTagsByDate)
	if err != nil {
		return err
	}
	return c.JSON(OK, tags)
}

func allTagsByName(c echo.Context) error {
	tags, err := db.GetAllTags(stmt.AllTagsByName)
	if err != nil {
		return err
	}
	return c.JSON(OK, tags)
}

func getGroupsByTag(c echo.Context) error {
	name, err := getFormValue(c, "name")
	if err != nil {
		return err
	}
	ok, err := db.IsTagExist(name)
	if err != nil {
		return err
	}
	if !ok {
		return c.String(404, fmt.Sprintf("#%s does not exist", name))
	}
	groups, err := db.GetGroupsByTag(name)
	if err != nil {
		return err
	}
	return c.JSON(OK, groups)
}

func renameTag(c echo.Context) error {
	oldName, e1 := getFormValue(c, "old-name")
	newName, e2 := getFormValue(c, "new-name")
	if err := util.WrapErrors(e1, e2); err != nil {
		return err
	}
	return db.RenameTag(oldName, newName)
}

func isTagExist(c echo.Context) error {
	newName, err := getFormValue(c, "new-name")
	if err != nil {
		return err
	}
	ok, err := db.IsTagExist(newName)
	if err != nil {
		return err
	}
	return c.JSON(OK, ok)
}

func deleteTag(c echo.Context) error {
	tagName, err := getFormValue(c, "tag-name")
	if err != nil {
		return err
	}
	ok, err := db.IsTagExist(tagName)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("系统中不存在该标签 #%s", tagName)
	}
	if err := db.CheckBeforeDeleteTag(tagName); err != nil {
		return err
	}
	return db.Exec(stmt.DeleteTag, tagName)
}
