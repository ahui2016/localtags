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

func (db *DB) GetFileID(hash string) (id string, ok bool) {
	id, err := getText1(db.DB, stmt.GetFileID, hash)
	if err == sql.ErrNoRows {
		return
	}
	util.Panic(err)
	return id, true
}

func (db *DB) InsertFiles(files []*File) (err error) {
	tx := db.mustBegin()
	defer tx.Rollback()

	for _, file := range files {
		// add the file
		if err = addFile(tx, file); err != nil {
			return
		}

		// add the tag group
		group := model.NewTagGroup()
		group.SetTags(file.Tags)
		if err = addTagGroup(tx, group); err != nil {
			return
		}

		// add tags
		if err = addTags(tx, file.Tags, file.ID); err != nil {
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

func scanFile(row Row) (file File, err error) {
	err = row.Scan(
		&file.ID,
		&file.Name,
		&file.Size,
		&file.Type,
		&file.Thumb,
		&file.Hash,
		&file.Like,
		&file.CTime,
		&file.UTime,
		&file.Deleted,
	)
	return
}

func addTags(tx TX, tags []string, fileID string) (err error) {
	for _, name := range tags {
		if err = addTag(tx, name, fileID); err != nil {
			return err
		}
	}
	return nil
}

func addTag(tx TX, tagID, fileID string) error {
	tagExist, err := isTagExist(tx, tagID)
	if err != nil {
		return err
	}
	// 如果在数据库中还没有这个标签, 则添加。
	if !tagExist {
		tag := model.NewTag(tagID)
		if err := exec(tx, stmt.InsertTag, tagID, tag.CTime); err != nil {
			return err
		}
	}
	// 最后，不管有没有添加新标签，都与文件关联。
	return exec(tx, stmt.InsertFileTag, fileID, tagID)
}

func isTagExist(tx TX, tagID string) (bool, error) {
	_, err := getInt1(tx, stmt.GetTagCTime, tagID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
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

func (db *DB) AllFiles() (files []*File, err error) {
	files, err = getFiles(db.DB, stmt.GetFiles)
	err = fillTags(db.DB, files)
	return
}

func getFiles(tx TX, st string) (files []*File, err error) {
	rows, err := tx.Query(st)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		file, err := scanFile(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, &file)
	}
	err = rows.Err()
	return
}

func fillTags(tx TX, files []*File) error {
	for _, file := range files {
		tags, err := getTagsByFile(tx, file.ID)
		if err != nil {
			return err
		}
		file.Tags = tags
	}
	return nil
}

func getTagsByFile(tx TX, id string) ([]string, error) {
	rows, err := tx.Query(stmt.GetTagsByFile, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTags(rows)
}

func scanTags(rows *sql.Rows) (tags []string, err error) {
	for rows.Next() {
		var tag string
		if err = rows.Scan(&tag); err != nil {
			return
		}
		tags = append(tags, tag)
	}
	err = rows.Err()
	return
}
