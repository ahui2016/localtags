package database

import (
	"database/sql"

	"github.com/ahui2016/localtags/config"
	"github.com/ahui2016/localtags/model"
	"github.com/ahui2016/localtags/stmt"
	"github.com/ahui2016/localtags/util"
	_ "github.com/mattn/go-sqlite3"
)

var cfg = config.Public

type (
	Stmt     = sql.Stmt
	File     = model.File
	ShortID  = model.ShortID
	TagGroup = model.TagGroup
)

type TX interface {
	Exec(string, ...interface{}) (sql.Result, error)
	QueryRow(string, ...interface{}) *sql.Row
	Prepare(string) (*Stmt, error)
}

// DB 数据库
type DB struct {
	DB *sql.DB
}

func (db *DB) Open(dbPath string) (err error) {
	if db.DB, err = sql.Open("sqlite3", dbPath+"?_fk=1"); err != nil {
		return
	}
	if err = db.Exec(stmt.CreateTables); err != nil {
		return
	}
	return db.initMetadata()
}
func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) mustBegin() *sql.Tx {
	tx, err := db.DB.Begin()
	util.Panic(err)
	return tx
}

func (db *DB) NewFile() *File {
	return model.NewFile(db.GetNextFileID())
}

func (db *DB) InsertFiles(files []*File) (err error) {
	tx := db.mustBegin()
	defer tx.Rollback()

	for _, file := range files {
		if err = addFile(tx, file); err != nil {
			return
		}
		group := model.NewTagGroup()
		group.SetTags(file.Tags)
		if err = addTagGroup(tx, group); err != nil {
			return
		}
	}
	return tx.Commit()
}

func addFile(tx TX, file *File) (err error) {
	_, err = tx.Exec(stmt.InsertFile,
		file.ID,
		file.Name,
		file.Size,
		file.Type,
		file.Thumb,
		file.Hash,
		file.Like,
		file.CTime,
		file.UTime,
		file.Deleted,
	)
	return
}

func addTag(tx TX, tag, noteID string) error {
	tagID, err := getText1(tx, stmt.GetTag)
}

func addTagGroup(tx TX, group *TagGroup) error {
	tags := group.String()
	groupID, err := getText1(tx, stmt.GetTagGroupID, tags)

	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == sql.ErrNoRows {
		err = exec(tx, stmt.InsertTagGroup,
			group.ID,
			tags,
			group.Protected,
			group.CTime,
			group.UTime)
	} else {
		// err == nil
		err = updateNow(tx, stmt.UpdateTagGroupNow, groupID)
	}
	if err != nil {
		return err
	}
	return deleteOldTagGroup(tx)
}

func deleteOldTagGroup(tx TX) error {
	count, err := getInt1(tx, stmt.TagGroupCount)
	if err != nil {
		return err
	}
	if count < cfg.TagGroupLimit {
		return nil
	}
	groupID, err := getText1(tx, stmt.LastTagGroup)
	if err != nil {
		return err
	}
	return exec(tx, stmt.DeleteTagGroup, groupID)
}

// getText1 gets one text value from the database.
func getText1(tx TX, st string, args ...interface{}) (text string, err error) {
	row := tx.QueryRow(st, args...)
	err = row.Scan(&text)
	return
}

// getInt1 gets one text value from the database.
func getInt1(tx TX, st string, arg ...interface{}) (n int, err error) {
	row := tx.QueryRow(st, arg...)
	err = row.Scan(&n)
	return
}

func updateNow(tx TX, st, arg string) error {
	return exec(tx, st, model.TimeNow(), arg)
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

func (db *DB) getNextID(key string) (nextID string, err error) {
	currentID, err := getCurrentID(key, db.DB)
	if err != nil {
		return
	}
	nextID = currentID.Next().String()
	err = db.Exec(stmt.UpdateTextValue, nextID, key)
	return
}

func (db *DB) initMetadata() error {
	return initFirstID(file_id_key, file_id_prefix, db.DB)
}

func (db *DB) Exec(query string, args ...interface{}) (err error) {
	_, err = db.DB.Exec(query, args...)
	return
}

func exec(tx TX, st string, args ...interface{}) (err error) {
	_, err = tx.Exec(st, args...)
	return
}
