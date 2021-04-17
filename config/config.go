package config

import (
	"path/filepath"

	"github.com/ahui2016/localtags/util"
)

const (
	dataFolderName    = "localtags_data_folder"
	waitingFolderName = "waiting"
	timeUnit          = 1000 * 60 * 60 * 24 // 1天(24小时)
)

var Public = Default()

// Config 用来设置一些全局变量
type Config struct {

	// 本地 IP 和端口
	Address string

	// 数据文件夹的完整路径
	DataFolder string

	// 待上传文件的文件夹
	WaitingFolder string

	// TagGroupLimit 限制标签组数量上限。
	// 当超过上限时，不受保护的标签组会被覆盖。可通过点击 "protect" 按钮保护标签。
	TagGroupLimit int64

	// 当 TimeNow - file.Checked > checkInterval 时，该文件需要重新检查。
	CheckInterval int64
}

// Default 默认设定
func Default() Config {
	dataDir := filepath.Join(util.UserHomeDir(), dataFolderName)
	return Config{
		Address:       "127.0.0.1:80",
		DataFolder:    dataDir,
		WaitingFolder: filepath.Join(dataDir, waitingFolderName),
		TagGroupLimit: 100,
		CheckInterval: timeUnit * 30, // 30天
	}
}
