package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ahui2016/localtags/database"
	"github.com/ahui2016/localtags/model"
	"github.com/ahui2016/localtags/stmt"
	"github.com/ahui2016/localtags/thumb"
	"github.com/ahui2016/localtags/util"
	"github.com/labstack/echo/v4"
	"tcw.im/go-disk-usage/du"
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
		return nil, fmt.Errorf(`"waiting" 里面不可存放文件夹`)
	}

	// 填充文件体积、文件名、文件类型
	file = &File{Size: info.Size()}
	file.SetNameType(info.Name())

	// 填充同名文件数
	ids, err := db.GetFileIDsByName(file.Name)
	if err != nil {
		return nil, err
	}
	file.Count = len(ids)

	// 填充文件标签
	if file.Count > 0 {
		tags, err := db.GetTagsByFile(ids[0])
		if err != nil {
			return nil, err
		}
		file.Tags = tags
	}

	// 填充文件哈希值
	file.Hash, err = util.FileSha256Hex(name)
	if err != nil {
		return
	}

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

/*
func indexByName(meta map[string]*File) map[string]*File {
	byName := make(map[string]*File)
	for _, file := range meta {
		name := file.ID + thumbSuffix
		byName[name] = file
	}
	return byName
}
*/

// getFormValue gets the c.FormValue(key), trims its spaces,
// and checks if it is empty or not.
func getFormValue(c echo.Context, key string) (string, error) {
	value := strings.TrimSpace(c.FormValue(key))
	if value == "" {
		return "", fmt.Errorf("form value [%s] is empty", key)
	}
	return value, nil
}

func getNumber(c echo.Context, key string) (int, error) {
	s, err := getFormValue(c, key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(s)
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
	if err := os.WriteFile(fullpath, []byte("abc"), 0666); err != nil {
		return err
	}
	return os.Remove(fullpath)
}

// https://stackoverflow.com/questions/30697324/how-to-check-if-directory-on-path-is-empty
func checkBucketFolder(folder string) error {
	f, err := os.Open(folder)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("[%s] 不是文件夹。", folder)
	}

	// 备份仓库的第一层目录内应该不超过 10 个项目。
	files, err := f.Readdir(10) // Or f.Readdir(1)

	// 如果是空文件夹，则没有问题。
	if err == io.EOF {
		return nil
	}
	// 如果有其他错误，则返回错误。
	if err != nil {
		return err
	}
	// 如果文件夹内有文件，则检查有没有 backupDatabase
	for _, file := range files {
		if file.Name() == backupDBFileName {
			return nil
		}
	}
	// 如果文件夹内有文件，但找不到 backupDatabase
	return fmt.Errorf("[%s] 不是空文件夹，请指定一个空文件夹。", folder)
}

// getBucketsInfo 返回主仓库与备份仓库的状态信息。
func getBucketsInfo(bkFolder string) (map[string]database.Info, error) {
	// 检查备份仓库文件夹的有效性
	if err := checkBucketFolder(bkFolder); err != nil {
		return nil, err
	}

	// 获取 main bucket 的状态信息
	info := make(map[string]database.Info)
	dbInfo, err := db.GetInfo()
	if err != nil {
		return nil, err
	}
	info["main-bucket"] = dbInfo
	info["backup-bucket"] = database.Info{}

	// 如果找不到备份数据库文件，则说明这是一个空文件夹，是一个全新的备份仓库。
	// 此时，生成一个新的备份仓库数据库文件，为后续的备份做准备。
	bkPath := filepath.Join(bkFolder, backupDBFileName)
	if util.PathIsNotExist(bkPath) {
		bk := new(database.DB)
		if err := bk.Open(bkPath); err != nil {
			return nil, err
		}
		bk.Close()
		return info, nil
	}

	// 如果能找到备份数据库文件，则打开备份数据库。
	bk := new(database.DB)
	if err := bk.OpenBackup(bkPath); err != nil {
		return nil, err
	}
	defer bk.Close()

	// 检查备份文件的完整性后获取备份仓库的状态信息。
	bakBucket := filepath.Join(bkFolder, bakBucketName)
	if err := bk.CheckFilesHash(bakBucket); err != nil {
		return nil, err
	}
	info["backup-bucket"], err = bk.GetInfo()
	if err != nil {
		return nil, err
	}
	// 主仓库的备份时间只可能等于或大于备份仓库的备份时间，
	// 不可能备份仓库的备份日期是今天而主仓库的备份日期是昨天。
	// (因为只能单向备份，备份时只能从主仓库复制文件到备份仓库。)
	if info["backup-bucket"].LastBackup > info["main-bucket"].LastBackup {
		return nil, fmt.Errorf("仓库不匹配：备份仓库的日期比主仓库更新")
	}

	// 最后返回主仓库与备份仓库的状态信息
	return info, nil
}

func checkDiskUsage(bkFolder string, bkDB *database.DB) error {
	diskInfo := du.DiskInfo(bkFolder)
	totalOfBackup, e1 := bkDB.TotalSize()
	totalOfMain, e2 := db.TotalSize()
	total := totalOfMain - totalOfBackup // 备份后将会增加的体积
	if err := util.WrapErrors(e1, e2); err != nil {
		return err
	}
	var margin uint64 = 1024 * 512 // 500 MB 余量
	if uint64(total)+margin > diskInfo.Available {
		return fmt.Errorf("备份仓库空间不足")
	}
	return nil
}

func deleteDamagedFiles(bkFolder string) error {
	bkPath := filepath.Join(bkFolder, backupDBFileName)
	bakBucket := filepath.Join(bkFolder, bakBucketName)
	util.MustMkdir(bakBucket)

	bk := new(database.DB)
	if err := bk.OpenBackup(bkPath); err != nil {
		return err
	}
	defer bk.Close()

	fileIDs, err := bk.DamagedFileIDs()
	if err != nil {
		return err
	}

	for _, id := range fileIDs {
		if err := os.Remove(filepath.Join(bakBucket, id)); err != nil {
			return err
		}
		if err := bk.Exec(stmt.DeleteFile, id); err != nil {
			return err
		}
	}
	return nil
}

// syncMainToBackup 同步主仓库与备份仓库，以主仓库为准单向同步，
// 最终效果相当于清空备份仓库后把主仓库的全部文件复制到备份仓库。
func syncMainToBackup(bkFolder string) error {
	bkPath := filepath.Join(bkFolder, backupDBFileName)
	bakBucket := filepath.Join(bkFolder, bakBucketName)
	util.MustMkdir(bakBucket)

	bk := new(database.DB)
	if err := bk.OpenBackup(bkPath); err != nil {
		return err
	}
	defer bk.Close()

	// 检查备份仓库的可用空间
	if err := checkDiskUsage(bkFolder, bk); err != nil {
		return err
	}

	// 如果有损坏文件则拒绝备份
	n, err := damagedIn2Buckets(bk, db)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("发现 %d 个损坏文件, 修复后才能备份", n)
	}

	// 在复制、删除文件之前更新备份时间。
	if err := db.UpdateLastBackupNow(); err != nil {
		return err
	}

	bkFiles, e1 := bk.AllFilesWithoutTags()
	dbFiles, e2 := db.AllFilesWithoutTags()
	if err := util.WrapErrors(e1, e2); err != nil {
		return err
	}

	// 如果一个文件存在于备份仓库中，但不存在于主仓库中，
	// 那么说明该文件已被彻底删除，因此在备份仓库中也需要删除它。
	for _, bkFile := range bkFiles {
		if !db.IsFileExist(bkFile.ID) {
			if err := os.Remove(filepath.Join(bakBucket, bkFile.ID)); err != nil {
				return err
			}
		}
	}

	// 如果一个文件存在于主仓库中，但不存在于备份仓库中，则直接拷贝。
	// 如果一个文件存在于两个仓库中，则进一步对比其日期，按需拷贝覆盖。
	for _, file := range dbFiles {
		bkCTime, err := bk.FileCTime(file.ID)
		if err != nil || file.CTime > bkCTime {
			bkFile := filepath.Join(bakBucket, file.ID)
			if err := util.CopyFile(bkFile, mainBucketFile(file.ID)); err != nil {
				return err
			}
		}
	}

	// 最后复制数据库文件
	bk.Close()
	return util.CopyFile(bkPath, dbPath)
}

func damagedIn2Buckets(db1, db2 *database.DB) (int64, error) {
	info1, e1 := db1.GetInfo()
	info2, e2 := db2.GetInfo()
	return info1.DamagedFilesCount + info2.DamagedFilesCount, util.WrapErrors(e1, e2)
}

// repairDamagedFiles 自动修复文件，对于主仓库里的损坏文件，尝试从备份仓库中获取未损坏版本，
// 对于备份仓库中的损坏文件则尝试从主仓库中获取未损坏版本，如果一个文件在主仓库及备份仓库中都损坏了，
// 则无法修复该文件，后续提醒用户手动修复。
func repairDamagedFiles(bkFolder string) error {
	bkPath := filepath.Join(bkFolder, backupDBFileName)
	bakBucket := filepath.Join(bkFolder, bakBucketName)
	bk := new(database.DB)
	if err := bk.OpenBackup(bkPath); err != nil {
		return err
	}
	defer bk.Close()

	if err := repair(db, bk, mainBucket, bakBucket); err != nil {
		return err
	}
	if err := repair(bk, db, bakBucket, mainBucket); err != nil {
		return err
	}

	// 经自动修复后，再次检查有没有损坏文件。
	n, err := damagedIn2Buckets(bk, db)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("仍有 %d 个损坏文件无法使用该备份仓库进行修复，请使用另一个备份仓库(如有), 或手动修复。", n)
	}
	return nil
}

// 从 badDB 中找出 badFiles, 然后尝试从 goodDB 中获取未损坏版本进行修复。
// 如果修复成功，则将 badFile 标记为未损坏，
// 如果不可修复，则不进行任何操作, badFile 仍然保持 "已损坏" 的标记。
func repair(badDB, goodDB *database.DB, badFolder, goodFolder string) error {
	badFiles, err := badDB.DamagedFiles()
	if err != nil {
		return err
	}
	for _, file := range badFiles {
		// 如果 goodDB 中的文件已损坏或找不到文件，则无法修复，如果未损坏则进行修复。
		damaged, err := goodDB.RecheckFile(goodFolder, file)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return err
		}
		if damaged {
			continue
		}

		// 进行修复
		goodFile := filepath.Join(goodFolder, file.ID)
		badFile := filepath.Join(badFolder, file.ID)
		if err := util.CopyFile(badFile, goodFile); err != nil {
			return err
		}
		// 更新校验日期，标记为未损坏
		if err = badDB.Exec(stmt.SetFileChecked, model.TimeNow(), false, file.ID); err != nil {
			return err
		}
	}
	return nil
}
