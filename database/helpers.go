package database

import (
	"database/sql"
	"errors"

	"github.com/ahui2016/localtags/model"
	"github.com/ahui2016/localtags/stmt"
	"github.com/ahui2016/localtags/stringset"
	"github.com/ahui2016/localtags/util"
)

func (db *DB) mustBegin() *sql.Tx {
	tx, err := db.DB.Begin()
	util.Panic(err)
	return tx
}

func countFiles(tx TX, name string) (int64, error) {
	return getInt1(tx, stmt.CountFilesByName, name)
}

func addFile(tx TX, file *File) (err error) {
	_, err = tx.Exec(stmt.InsertFile,
		file.ID,
		file.Name,
		file.Count,
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
		&file.Count,
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
	if len(group.Tags) < 2 {
		return errors.New("a tag group needs at least two tags")
	}
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
func getText1(tx TX, query string, args ...interface{}) (text string, err error) {
	row := tx.QueryRow(query, args...)
	err = row.Scan(&text)
	return
}

// getInt1 gets one text value from the database.
func getInt1(tx TX, query string, arg ...interface{}) (n int64, err error) {
	row := tx.QueryRow(query, arg...)
	err = row.Scan(&n)
	return
}

func updateNow(tx TX, query, arg string) error {
	return exec(tx, query, model.TimeNow(), arg)
}

func (db *DB) getNextID(key string) (nextID string, err error) {
	currentID, err := getCurrentID(key, db.DB)
	if err != nil {
		return
	}
	nextID = currentID.Next().String()
	err = db.exec(stmt.UpdateTextValue, nextID, key)
	return
}

func (db *DB) initMetadata() error {
	return initFirstID(file_id_key, file_id_prefix, db.DB)
}

func (db *DB) exec(query string, args ...interface{}) (err error) {
	_, err = db.DB.Exec(query, args...)
	return
}

func exec(tx TX, query string, args ...interface{}) (err error) {
	_, err = tx.Exec(query, args...)
	return
}

func getFiles(tx TX, query string) (files []*File, err error) {
	rows, err := tx.Query(query)
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

func getFileIDs(tx TX, query string, arg string) (fileIDs []string, err error) {
	rows, err := tx.Query(query, arg)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return
		}
		fileIDs = append(fileIDs, id)
	}
	if err = rows.Err(); err != nil {
		return
	}
	if len(fileIDs) == 0 {
		err = errors.New("no files related to " + arg)
	}
	return
}

func fillTags(tx TX, files []*File) error {
	for _, file := range files {
		if err := fillTag(tx, file); err != nil {
			return err
		}
	}
	return nil
}

func fillTag(tx TX, file *File) error {
	tags, err := getTagsByFile(tx, file.ID)
	if err != nil {
		return err
	}
	file.Tags = tags
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

func (db *DB) getFileIDsByTags(tags []string) ([]string, error) {
	var idSets []*Set
	for _, tag := range tags {
		fileIDs, err := getFileIDs(db.DB, stmt.GetFilesByTag, tag)
		if err != nil {
			return nil, err
		}
		idSets = append(idSets, stringset.From(fileIDs))
	}
	return stringset.Intersect(idSets).Slice(), nil
}

func (db *DB) getFilesByIDs(fileIDs []string) (files []*File, err error) {
	for _, id := range fileIDs {
		file, err := db.GetFileByID(id)
		if err != nil {
			return nil, err
		}
		files = append(files, &file)
	}
	return
}

func (db *DB) isFileDeleted(id string) (bool, error) {
	file, err := db.GetFileByID(id)
	if err != nil {
		return false, err
	}
	return file.Deleted, nil
}

func deleteTags(tx TX, toDelete []string, fileID string) error {
	for _, tag := range toDelete {
		if err := exec(tx, stmt.DeleteTag, fileID, tag); err != nil {
			return err
		}
	}
	return nil
}

func getSameNameFiles(tx TX, fileID string) ([]string, error) {
	name, err := getText1(tx, stmt.GetFileName, fileID)
	if err != nil {
		return nil, err
	}
	return getFileIDs(tx, stmt.GetFileIDsByName, name)
}

func updateTagsNow(tx TX, fileID string, toAdd, toDelete []string) error {
	e1 := deleteTags(tx, toDelete, fileID)
	e2 := addTags(tx, toAdd, fileID)
	e3 := exec(tx, stmt.UpdateNow, model.TimeNow(), fileID)
	return util.WrapErrors(e1, e2, e3)
}
