package database

import "database/sql"

// DB 数据库
type DB struct {
	DB *sql.DB
}
