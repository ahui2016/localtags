package config

import (
	"path/filepath"

	"github.com/ahui2016/localtags/util"
)

const (
	dataFolderName = "localtags_data_folder"
	waitingDirName = "waiting"
)

var Public = Default()

// Config 用来设置一些全局变量
type Config struct {

	// true 表示对文件进行转码 (base64)
	EncodeFile bool

	// 数据文件夹的完整路径
	DataFolder string

	// 待上传文件的文件夹的完整路径
	WaitingFolder string
}

// Default 默认设定
func Default() Config {
	dataDir := filepath.Join(util.UserHomeDir(), dataFolderName)
	return Config{
		EncodeFile:    false,
		DataFolder:    dataDir,
		WaitingFolder: filepath.Join(dataDir, waitingDirName),
	}
}
