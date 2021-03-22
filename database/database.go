package database

import (
	"database/sql"
	"errors"
	"fmt"

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
	TagGroup = model.TagGroup
	Set      = stringset.Set
)

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
	DB *sql.DB
}

func (db *DB) Open(dbPath string) (err error) {
	if db.DB, err = sql.Open("sqlite3", dbPath+"?_fk=1"); err != nil {
		return
	}
	if err = db.exec(stmt.CreateTables); err != nil {
		return
	}
	return db.initMetadata()
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

func (db *DB) CountFiles(name string) (int64, error) {
	return countFiles(db.DB, name)
}

func (db *DB) InsertFiles(files []*File) error {
	tx := db.mustBegin()
	defer tx.Rollback()

	for _, file := range files {
		ids, err := getFileIDs(tx, stmt.GetFileIDsByName, file.Name)
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

func (db *DB) AllFiles() (files []*File, err error) {
	files, err = getFiles(db.DB, stmt.GetFiles)
	if err != nil {
		return
	}
	err = fillTags(db.DB, files)
	return
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

func (db *DB) SetFileDeleted(id string, deleted bool) error {
	ok, err := db.isFileDeleted(id)
	if err != nil {
		return err
	}
	if !ok {
		return db.exec(stmt.SetFileDeletedNow, deleted, model.TimeNow(), id)
	}
	return nil
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

func updateTags(tx TX, fileID string, newTags []string) error {
	oldTags, err := getTagsByFile(tx, fileID)
	if err != nil {
		return err
	}
	toAdd, toDelete := util.StrSliceDiff(newTags, oldTags)
	if len(toAdd)+len(toDelete) == 0 {
		return nil
	}

	group := model.NewTagGroup()
	group.Tags = newTags
	if err := addTagGroup(tx, group); err != nil {
		return err
	}

	ids, err := getSameNameFiles(tx, fileID)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err = updateTagsNow(tx, id, toAdd, toDelete); err != nil {
			return err
		}
	}
	return nil
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
	return db.exec(stmt.RenameFilesNow,
		file.Name, file.Type, file.UTime, oldName)
}
