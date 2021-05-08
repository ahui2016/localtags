package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/ahui2016/localtags/config"
	"github.com/ahui2016/localtags/model"
	"github.com/ahui2016/localtags/stmt"
	"github.com/ahui2016/localtags/stringset"
	"github.com/ahui2016/localtags/util"
	_ "github.com/mattn/go-sqlite3"
)

type (
	Stmt     = sql.Stmt
	File     = model.File
	ShortID  = model.ShortID
	Tag      = model.Tag
	TagGroup = model.TagGroup
	Set      = stringset.Set
)

// Info of the database
type Info struct {
	BucketLocation    string
	LastChecked       int64
	LastBackup        int64
	AllFilesCount     int64
	DamagedFilesCount int64
	TotalSize         int64
}

type TX interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	Prepare(string) (*Stmt, error)
}

type Row interface {
	Scan(...interface{}) error
}

// DB 数据库
type DB struct {
	Folder string
	DB     *sql.DB
	Config config.Config
}

func (db *DB) Open(dbPath string, cfg config.Config) (err error) {
	if db.DB, err = sql.Open("sqlite3", dbPath+"?_fk=1"); err != nil {
		return
	}
	db.Folder = filepath.Dir(dbPath)
	if err = db.Exec(stmt.CreateTables); err != nil {
		return
	}
	db.Config = cfg
	return db.initMetadata()
}

// OpenBackup opens a backup database.
func (db *DB) OpenBackup(dbPath string, cfg config.Config) (err error) {
	if util.PathIsNotExist(dbPath) {
		return fmt.Errorf("not found: %s", dbPath)
	}
	db.Folder = filepath.Dir(dbPath)
	db.Config = cfg
	db.DB, err = sql.Open("sqlite3", dbPath+"?_fk=1")
	return
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) NewFile() *File {
	return model.NewFile(db.GetNextFileID())
}

func (db *DB) GetFileID(hash string) (id string, ok bool) {
	id, err := getText1(db.DB, stmt.GetFileID, hash)
	if err == sql.ErrNoRows {
		return
	}
	util.Panic(err)
	return id, true
}

func (db *DB) GetFileIDsByName(name string) ([]string, error) {
	return getFileIDs(db.DB, stmt.GetFileIDsByName, name)
}

func (db *DB) GetTagsByFile(id string) ([]string, error) {
	return getTagsByFile(db.DB, id)
}

func (db *DB) DeleteFile(id string) error {
	tx := db.mustBegin()
	defer tx.Rollback()

	file, err := getFileByID(tx, id)
	if err != nil {
		return err
	}
	// 本来最好应该判断一下 file.Count 是否大于1, 但懒得判断了，反正问题不大。
	e1 := exec(tx, stmt.SetFilesCount, file.Count-1, file.Name)
	e2 := exec(tx, stmt.DeleteFile, id)
	if err := util.WrapErrors(e1, e2); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) InsertFiles(files []*File) error {
	tx := db.mustBegin()
	defer tx.Rollback()

	for _, file := range files {
		ids, err := getFileIDs(tx, stmt.GetFileIDsByName, file.Name)
		if err != nil {
			return err
		}
		count := len(ids)
		file.Count = count + 1

		// 如果系统中有同名文件，要先统一全部同名文件的标签。
		// 必须在插入新文件之前更新同名文件的标签。
		if count > 0 {
			if err := exec(tx, stmt.SetFilesCount, file.Count, file.Name); err != nil {
				return err
			}
			if err := updateTags(tx, ids[0], file.Tags, db.Config.TagGroupLimit); err != nil {
				return err
			}
		}
		// add the file
		if err = addFile(tx, file); err != nil {
			return err
		}

		// add the tag group
		group := model.NewTagGroup()
		group.Tags = file.Tags
		if err = addTagGroup(tx, group, db.Config.TagGroupLimit); err != nil {
			return err
		}

		// add tags
		if err = addTags(tx, file.Tags, file.ID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (db *DB) GetNextFileID() string {
	nextID, err := db.getNextID(file_id_key)
	util.Panic(err)
	return nextID
}

func (db *DB) CurrentFileID() (string, error) {
	currentID, err := getCurrentID(file_id_key, db.DB)
	if err != nil {
		return "", err
	}
	return currentID.String(), nil
}

func (db *DB) AllFilesWithoutTags() ([]*File, error) {
	return getFiles(db.DB, stmt.GetAllFiles)
}

func (db *DB) DamagedFiles() ([]*File, error) {
	return getFiles(db.DB, stmt.DamagedFiles)
}

func (db *DB) DamagedFileIDs() ([]string, error) {
	return getFileIDs(db.DB, stmt.DamagedFileIDs)
}

func (db *DB) SearchDamagedFiles() ([]*File, error) {
	files, err := getFiles(db.DB, stmt.DamagedFiles)
	if err != nil {
		return nil, err
	}
	err = fillTags(db.DB, files)
	return files, err
}

func (db *DB) AllFiles() (files []*File, err error) {
	files, err = getFiles(db.DB, stmt.GetFiles, db.Config.FileListLimit)
	if err != nil {
		return
	}
	err = fillTags(db.DB, files)
	return
}

func (db *DB) AllImages() (files []*File, err error) {
	files, err = getFiles(db.DB, stmt.GetImages, db.Config.FileListLimit)
	if err != nil {
		return
	}
	err = fillTags(db.DB, files)
	return
}

func (db *DB) DeletedFiles() (files []*File, err error) {
	files, err = getFiles(db.DB, stmt.GetDeletedFiles)
	if err != nil {
		return
	}
	err = fillTags(db.DB, files)
	return
}

func (db *DB) IsFileExist(id string) bool {
	_, err := db.GetFileName(id)
	return err == nil
}

func (db *DB) FileCTime(id string) (int64, error) {
	return getInt1(db.DB, stmt.GetFileCTime, id)
}

func (db *DB) SearchTags(tags []string, fileType string) ([]*File, error) {
	fileIDs, err := db.getFileIDsByTags(tags, fileType)
	if err != nil {
		return nil, err
	}
	files, err := db.getFilesByIDs(fileIDs)
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].CTime > files[j].CTime
	})
	return files, nil
}

func (db *DB) SearchFileName(pattern string, fileType string) (files []*File, err error) {
	query := stmt.SearchFileName
	if fileType == "image" {
		query = stmt.SearchImageName
	}

	rows, err := db.DB.Query(query, "%"+pattern+"%")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		file, err := scanFile(rows)
		if err != nil {
			return nil, err
		}
		if err := fillTag(db.DB, &file); err != nil {
			return nil, err
		}
		files = append(files, &file)
	}
	return files, rows.Err()
}

func (db *DB) SearchSameNameFiles(id string) ([]*File, error) {
	fileIDs, err := getSameNameFiles(db.DB, id)
	if err != nil {
		return nil, err
	}
	return db.getFilesByIDs(fileIDs)
}

func (db *DB) UpdateTags(fileID string, tags []string) error {
	newTags := stringset.UniqueSort(tags)
	if len(newTags) < 2 {
		return errors.New("a file needs at least two tags")
	}

	tx := db.mustBegin()
	defer tx.Rollback()

	if err := updateTags(tx, fileID, newTags, db.Config.TagGroupLimit); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *DB) GetFileName(id string) (string, error) {
	return getText1(db.DB, stmt.GetFileName, id)
}

// RenameFiles 统一修改全部同名文件的文件名。
func (db *DB) RenameFiles(id, name string) error {
	// 1.如果新文件等于旧文件名，不需要改名，直接返回。
	oldName, err := db.GetFileName(id)
	if err != nil {
		return err
	}
	if name == oldName {
		return nil
	}
	// 2.如果新文件名发生冲突，返回错误。
	count, err := countFiles(db.DB, name)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("文件名冲突(重名): %s", name)
	}
	// 3.利用 SetNameType 检查新文件名的长度，并根据新文件名更改文件类型
	file := model.NewFile(id)
	if err := file.SetNameType(name); err != nil {
		return err
	}
	// 4.统一改名
	return db.Exec(stmt.RenameFilesNow,
		file.Name, file.Type, file.UTime, oldName)
}

func (db *DB) GetInfo() (Info, error) {
	lastChecked, e1 := getIntValue(last_check_key, db.DB)
	lastBackup, e2 := getIntValue(last_backup_key, db.DB)
	allFiles, e3 := getInt1(db.DB, stmt.CountAllFiles)
	damagedFiles, e4 := getInt1(db.DB, stmt.CountDamagedFiles)
	totalSize, e5 := getInt1(db.DB, stmt.TotalSize)
	err := util.WrapErrors(e1, e2, e3, e4, e5)
	info := Info{
		BucketLocation:    db.Folder,
		LastChecked:       lastChecked,
		LastBackup:        lastBackup,
		AllFilesCount:     allFiles,
		DamagedFilesCount: damagedFiles,
		TotalSize:         totalSize,
	}
	return info, err
}

func (db *DB) TagGroups() (groups []TagGroup, err error) {
	rows, err := db.DB.Query(stmt.AllTagGroups)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var group TagGroup
		group, err = scanTagGroup(rows)
		if err != nil {
			return
		}
		groups = append(groups, group)
	}
	err = rows.Err()
	return
}

func (db *DB) AddTagGroup(group *TagGroup) error {
	return addTagGroup(db.DB, group, db.Config.TagGroupLimit)
}

func (db *DB) GetAllTags(query string) (tags []Tag, err error) {
	rows, err := db.DB.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var tag Tag
		err = rows.Scan(&tag.ID, &tag.CTime, &tag.Count)
		if err != nil {
			return
		}
		tags = append(tags, tag)
	}
	err = rows.Err()
	return
}

func (db *DB) GetGroupsByTag(name string) (groups [][]string, err error) {
	files, err := db.SearchTags([]string{name}, "all")
	if err != nil {
		return nil, err
	}
	set := stringset.NewSet()
	for _, file := range files {
		set.Add(stringset.UniqueSortString(file.Tags))
	}
	for data, ok := range set.Map {
		if ok {
			var group []string
			if err := json.Unmarshal([]byte(data), &group); err != nil {
				return nil, err
			}
			groups = append(groups, group)
		}
	}
	return
}

func (db *DB) IsTagExist(name string) (bool, error) {
	return isTagExist(db.DB, name)
}

func (db *DB) RenameTag(oldName, newName string) error {
	ok, err := isTagExist(db.DB, newName)
	if err != nil {
		return err
	}
	if !ok {
		// 如果新标签名没有冲突，那么，直接改名即可。
		return db.Exec(stmt.RenameTag, newName, oldName)
	}

	// 如果新标签名已存在，则添加新标签，删除旧标签。
	fileIDs, err := getFileIDs(db.DB, stmt.AllFilesByTag, oldName)
	if err != nil {
		return err
	}
	tx := db.mustBegin()
	defer tx.Rollback()

	for _, id := range fileIDs {
		err = exec(tx, stmt.InsertFileTag, id, newName)
		if err != nil && !util.ErrorContains(err, "UNIQUE") {
			return err
		}
	}
	if err := exec(tx, stmt.DeleteTag, oldName); err != nil {
		return err
	}

	return tx.Commit()
}
