package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/ahui2016/localtags/config"
	"github.com/ahui2016/localtags/model"
	"github.com/ahui2016/localtags/stmt"
	"github.com/ahui2016/localtags/stringset"
	"github.com/ahui2016/localtags/util"
	_ "github.com/mattn/go-sqlite3"
)

var cfg = config.Public

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
}

func (db *DB) Open(dbPath string) (err error) {
	if db.DB, err = sql.Open("sqlite3", dbPath+"?_fk=1"); err != nil {
		return
	}
	db.Folder = filepath.Dir(dbPath)
	if err = db.Exec(stmt.CreateTables); err != nil {
		return
	}
	return db.initMetadata()
}

// OpenBackup opens a backup database.
func (db *DB) OpenBackup(dbPath string) (err error) {
	if util.PathIsNotExist(dbPath) {
		return fmt.Errorf("not found: %s", dbPath)
	}
	db.Folder = filepath.Dir(dbPath)
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
	return getFileIDsByName(db.DB, name)
}

func (db *DB) GetTagsByFile(id string) ([]string, error) {
	return getTagsByFile(db.DB, id)
}

func (db *DB) InsertFiles(files []*File) error {
	tx := db.mustBegin()
	defer tx.Rollback()

	for _, file := range files {
		ids, err := getFileIDsByName(tx, file.Name)
		if err != nil {
			return err
		}
		count := int64(len(ids))
		file.Count = count + 1

		// 如果系统中有同名文件，要先统一全部同名文件的标签。
		// 必须在插入新文件之前更新同名文件的标签。
		if count > 0 {
			if err := exec(tx, stmt.SetFilesCount, file.Count, file.Name); err != nil {
				return err
			}
			if err := updateTags(tx, ids[0], file.Tags); err != nil {
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
		if err = addTagGroup(tx, group); err != nil {
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
	return getFiles(db.DB, stmt.GetDamagedFiles)
}

func (db *DB) AllFiles() (files []*File, err error) {
	files, err = getFiles(db.DB, stmt.GetFiles)
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
	_, err := getText1(db.DB, stmt.GetFileName, id)
	return err == nil
}

func (db *DB) FileUTime(id string) (int64, error) {
	return getInt1(db.DB, stmt.GetFileUTime, id)
}

func (db *DB) GetFileByID(id string) (file File, err error) {
	row := db.DB.QueryRow(stmt.GetFile, id)
	if file, err = scanFile(row); err != nil {
		return
	}
	err = fillTag(db.DB, &file)
	return
}

func (db *DB) SearchTags(tags []string) ([]*File, error) {
	fileIDs, err := db.getFileIDsByTags(tags)
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

	if err := updateTags(tx, fileID, newTags); err != nil {
		return err
	}

	return tx.Commit()
}

// RenameFiles 统一修改全部同名文件的文件名。
func (db *DB) RenameFiles(id, name string) error {
	// 1.如果新文件等于旧文件名，不需要改名，直接返回。
	oldName, err := getText1(db.DB, stmt.GetFileName, id)
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
	err := util.WrapErrors(e1, e2, e3, e4)
	info := Info{
		BucketLocation:    db.Folder,
		LastChecked:       lastChecked,
		LastBackup:        lastBackup,
		AllFilesCount:     allFiles,
		DamagedFilesCount: damagedFiles,
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
	return addTagGroup(db.DB, group)
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
	files, err := db.SearchTags([]string{name})
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

// RenameTag .
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
