package database

import (
	"database/sql"
	"encoding/json"

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
		file.Checked,
		file.Damaged,
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
		&file.Checked,
		&file.Damaged,
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

func addTagGroup(tx TX, group *TagGroup, limit int64) error {
	/*
		if len(group.Tags) < 2 {
			return errors.New("a tag group needs at least two tags")
		}
	*/
	tags := group.Blob()
	groupID, err := getText1(tx, stmt.GetTagGroupID, tags)

	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == sql.ErrNoRows {
		err = exec(tx, stmt.InsertTagGroup,
			group.ID,
			tags,
			group.Protected,
			group.UTime)
	} else {
		// err == nil
		err = updateNow(tx, stmt.UpdateTagGroupNow, groupID)
	}
	if err != nil {
		return err
	}
	return deleteOldTagGroup(tx, limit)
}

func deleteOldTagGroup(tx TX, limit int64) error {
	count, err := getInt1(tx, stmt.TagGroupCount)
	if err != nil {
		return err
	}
	if count < limit {
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
	err = db.Exec(stmt.UpdateTextValue, nextID, key)
	return
}

func (db *DB) initMetadata() error {
	e1 := initFirstID(file_id_key, file_id_prefix, db.DB)
	e2 := initIntValue(last_check_key, 0, db.DB)
	e3 := initIntValue(last_backup_key, 0, db.DB)
	e4 := initTextValue(backup_buckets_key, "", db.DB)
	return util.WrapErrors(e1, e2, e3, e4)
}

func (db *DB) Exec(query string, args ...interface{}) (err error) {
	_, err = db.DB.Exec(query, args...)
	return
}

func exec(tx TX, query string, args ...interface{}) (err error) {
	_, err = tx.Exec(query, args...)
	return
}

func getFiles(tx TX, query string, args ...interface{}) (files []*File, err error) {
	rows, err := tx.Query(query, args...)
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

func getFileIDs(tx TX, query string, args ...interface{}) (fileIDs []string, err error) {
	rows, err := tx.Query(query, args...)
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

func (db *DB) getFileIDsByTags(tags []string, fileType string) ([]string, error) {
	query := stmt.GetFilesByTag
	if fileType == "image" {
		query = stmt.GetImagesByTag
	}
	var idSets []*Set
	for _, tag := range tags {
		fileIDs, err := getFileIDs(db.DB, query, tag)
		if err != nil {
			return nil, err
		}
		idSets = append(idSets, stringset.From(fileIDs))
	}
	return stringset.Intersect(idSets).Slice(), nil
}

func (db *DB) getFilesByIDs(fileIDs []string) (files []*File, err error) {
	for _, id := range fileIDs {
		file, err := getFileByID(db.DB, id)
		if err != nil {
			return nil, err
		}
		files = append(files, &file)
	}
	return
}

func getFileByID(tx TX, id string) (file File, err error) {
	row := tx.QueryRow(stmt.GetFile, id)
	if file, err = scanFile(row); err != nil {
		return
	}
	err = fillTag(tx, &file)
	return
}

func deleteTags(tx TX, toDelete []string, fileID string) error {
	for _, tag := range toDelete {
		if err := exec(tx, stmt.DeleteTags, fileID, tag); err != nil {
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

func updateTags(tx TX, fileID string, newTags []string, limit int64) error {
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
	if err := addTagGroup(tx, group, limit); err != nil {
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

func scanTagGroup(rows *sql.Rows) (g TagGroup, err error) {
	var tagsJSON []byte
	if err = rows.Scan(&g.ID, &tagsJSON, &g.Protected, &g.UTime); err != nil {
		return
	}
	g.Tags = mustGetTags(tagsJSON)
	return
}

func mustGetTags(data []byte) []string {
	var tags []string
	err := json.Unmarshal(data, &tags)
	util.Panic(err)
	return tags
}
