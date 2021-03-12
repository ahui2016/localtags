package database

import (
	"database/sql"

	"github.com/ahui2016/localtags/model"
	"github.com/ahui2016/localtags/st"
)

const (
	file_id_key    = "file-id-key"
	file_id_prefix = "F"
)

func getCurrentID(key string, tx TX) (id ShortID, err error) {
	var strID string
	row := tx.QueryRow(st.GetTextValue, key)
	if err = row.Scan(&strID); err != nil {
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
		_, err = tx.Exec(st.InsertTextValue, key, id.String())
	}
	return
}
func setCurrentID(tx TX, key, id string) (err error) {
	_, err = tx.Exec(st.UpdateTextValue, id, key)
	return
}
