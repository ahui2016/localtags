package database

import (
	"database/sql"
	"encoding/json"
	"path/filepath"

	"github.com/ahui2016/localtags/model"
	"github.com/ahui2016/localtags/stmt"
	"github.com/ahui2016/localtags/util"
)

const (
	file_id_key        = "file-id-key"
	file_id_prefix     = "F"
	last_check_key     = "last-check-key"
	last_backup_key    = "last-backup-key"
	backup_buckets_key = "backup-buckets-key"
)

func getTextValue(key string, tx TX) (value string, err error) {
	row := tx.QueryRow(stmt.GetTextValue, key)
	err = row.Scan(&value)
	return
}

func getIntValue(key string, tx TX) (value int64, err error) {
	row := tx.QueryRow(stmt.GetIntValue, key)
	err = row.Scan(&value)
	return
}

func getCurrentID(key string, tx TX) (id ShortID, err error) {
	strID, err := getTextValue(key, tx)
	if err != nil {
		return
	}
	return model.ParseID(strID)
}

func initFirstID(key, prefix string, tx TX) (err error) {
	_, err = getCurrentID(key, tx)
	if err == sql.ErrNoRows {
		id, err1 := model.FirstID(prefix)
		if err1 != nil {
			return err1
		}
		err = exec(tx, stmt.InsertTextValue, key, id.String())
	}
	return
}

func needToCheck(tx TX) (need bool, err error) {
	lastCheckTime, err := getIntValue(last_check_key, tx)
	if err != nil {
		return
	}
	if model.TimeNow()-lastCheckTime > cfg.CheckInterval {
		need = true
	}
	return
}

func (db *DB) CheckFilesHash(bucket string) error {
	need, err := needToCheck(db.DB)
	if err != nil {
		return err
	}
	if !need {
		return nil
	}

	tx := db.mustBegin()
	defer tx.Rollback()

	// 如果一个文件的上次校验日期小于(早于) needCheckDate, 那么这个文件就需要再次校验。
	needCheckDate := model.TimeNow() - cfg.CheckInterval
	files, err := getFiles(tx, stmt.GetFilesNeedCheck, needCheckDate)
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := checkFile(tx, bucket, file); err != nil {
			return err
		}
	}

	// 最后记录本次校验时间
	if err := exec(tx, stmt.InsertIntValue, last_check_key, model.TimeNow()); err != nil {
		return err
	}
	return tx.Commit()
}

func checkFile(tx TX, folder string, file *File) error {
	if file.Damaged {
		return nil
	}
	filePath := filepath.Join(folder, file.ID)
	hash, err := util.FileSha256Hex(filePath)
	if err != nil {
		return err
	}
	if file.Hash != hash {
		file.Damaged = true
	}
	return exec(tx, stmt.SetFileChecked, model.TimeNow(), file.Damaged, file.ID)
}

func getBackupBuckets(tx TX) (buckets []string, err error) {
	s, err := getTextValue(backup_buckets_key, tx)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(s), &buckets)
	return
}

func saveBackupBuckets(tx TX, buckets []string) error {
	bucketsJSON := util.MustMarshal(buckets)
	return exec(tx, stmt.InsertTextValue, backup_buckets_key, string(bucketsJSON))
}

func addBackupBucket(tx TX, bucket string) error {
	buckets, err := getBackupBuckets(tx)
	if err != nil {
		return err
	}
	buckets = append(buckets, bucket)
	return saveBackupBuckets(tx, buckets)
}

func deleteBackupBucket(tx TX, i int) error {
	buckets, err := getBackupBuckets(tx)
	if err != nil {
		return err
	}
	buckets = append(buckets[:i], buckets[i+1:]...)
	return saveBackupBuckets(tx, buckets)
}
