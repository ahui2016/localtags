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

func initIntValue(key string, value int64, tx TX) error {
	_, err := getIntValue(key, tx)
	if err == sql.ErrNoRows {
		err = exec(tx, stmt.InsertIntValue, key, value)
	}
	return err
}

func initTextValue(key string, value string, tx TX) error {
	_, err := getTextValue(key, tx)
	if err == sql.ErrNoRows {
		err = exec(tx, stmt.InsertTextValue, key, value)
	}
	return err
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

// CheckFilesHash 只校验长时间未校验的文件，忽略短期内曾校验过的文件。
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
	if err := exec(tx, stmt.UpdateIntValue, model.TimeNow(), last_check_key); err != nil {
		return err
	}
	return tx.Commit()
}

// ForceCheckFilesHash 不检查上次校验日期，强制校验全部文件。
func (db *DB) ForceCheckFilesHash(bucket string) error {
	tx := db.mustBegin()
	defer tx.Rollback()

	files, err := getFiles(tx, stmt.GetAllFiles)
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := checkFile(tx, bucket, file); err != nil {
			return err
		}
	}
	if err := exec(tx, stmt.UpdateIntValue, model.TimeNow(), last_check_key); err != nil {
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

// 注意这里的 folder 与 db 匹配, db 与 file 必须分别属于不同的仓库，进行对比。
// 即，如果 db 是主仓库，则 file 应属于备份仓库; 如果 db 是备份仓库，则 file 应属于主仓库。
func (db *DB) RecheckFile(folder string, file *File) (damaged bool, err error) {
	// 如果在 db 中标记了该文件已损坏，则直接返回结果。
	value, err := getInt1(db.DB, stmt.GetFileDamaged, file.ID)
	if err != nil {
		return
	}
	if value > 0 {
		return true, nil
	}

	// 如果在 db 中标记了该文件未损坏，则再检查一次。
	filePath := filepath.Join(folder, file.ID)
	hash, err := util.FileSha256Hex(filePath)
	if err != nil {
		return
	}
	if file.Hash != hash {
		damaged = true
		err = exec(db.DB, stmt.SetFileChecked, model.TimeNow(), damaged, file.ID)
	}
	return
}

func (db *DB) UpdateLastBackupNow() error {
	return exec(db.DB, stmt.UpdateIntValue, model.TimeNow(), last_backup_key)
}

func (db *DB) GetBackupBuckets() ([]string, error) {
	return getBackupBuckets(db.DB)
}

func getBackupBuckets(tx TX) (buckets []string, err error) {
	s, err := getTextValue(backup_buckets_key, tx)
	if err != nil {
		return
	}
	if s == "" {
		return
	}
	err = json.Unmarshal([]byte(s), &buckets)
	return
}

func saveBackupBuckets(tx TX, buckets []string) error {
	bucketsJSON := util.MustMarshal(buckets)
	return exec(tx, stmt.UpdateTextValue, string(bucketsJSON), backup_buckets_key)
}

func (db *DB) AddBackupBucket(bucket string) error {
	buckets, err := getBackupBuckets(db.DB)
	if err != nil {
		return err
	}
	buckets = append(buckets, bucket)
	return saveBackupBuckets(db.DB, buckets)
}

func (db *DB) DeleteBackupBucket(i int) error {
	buckets, err := getBackupBuckets(db.DB)
	if err != nil {
		return err
	}
	buckets = append(buckets[:i], buckets[i+1:]...)
	return saveBackupBuckets(db.DB, buckets)
}

// TotalSize 返回数据库的总体积，单位是 KB.
func (db *DB) TotalSize() (int64, error) {
	total, err := getInt1(db.DB, stmt.TotalSize)
	if err != nil {
		return 0, err
	}
	return total / 1024, nil
}
