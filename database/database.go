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
	Stmt    = sql.Stmt
	File    = model.File
	ShortID = model.ShortID
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

func mustPrepare(tx TX, query string) *Stmt {
	stmt, err := tx.Prepare(query)
	util.Panic(err)
	return stmt
}

func (db *DB) NewFile() *File {
	return model.NewFile(db.GetNextFileID())
}

func (db *DB) InsertFiles(files []*File) error {
	tx := db.mustBegin()
	defer tx.Rollback()

	for _, file := range files {
		if err := insertFile(db.DB, file); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func insertFile(tx TX, file *File) (err error) {
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

func exec(st *Stmt, args ...interface{}) (err error) {
	_, err = st.Exec(args...)
	return
}
