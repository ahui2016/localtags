package database

import (
	"database/sql"

	"github.com/ahui2016/localtags/config"
	"github.com/ahui2016/localtags/model"
	"github.com/ahui2016/localtags/st"
)

var cfg = config.Public

type (
	Stmt    = sql.Stmt
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
	if err = db.Exec(st.CreateTables); err != nil {
		return
	}
	return db.initMetadata()
}
func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) initMetadata() error {
	return initFirstID(file_id_key, file_id_prefix, db.DB)
}

func (db *DB) Exec(query string, args ...interface{}) (err error) {
	_, err = db.DB.Exec(query, args...)
	return
}

func exec(stmt *Stmt, args ...interface{}) (err error) {
	_, err = stmt.Exec(args...)
	return
}
