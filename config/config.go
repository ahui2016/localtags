package config

import (
	"path/filepath"

	"github.com/ahui2016/localtags/util"
)

const (
	dataFolderName    = "localtags_data_folder"
	waitingFolderName = "waiting"
	TimeUnit          = 60 * 60 * 24 // 1天(24小时)
)

var Public = Default()

// Config 用来设置一些全局变量
type Config struct {

	// 本地 IP 和端口，建议选择一个不常用的端口。
	Address string

	// 数据文件夹的完整路径
	DataFolder string

	// 待上传文件的文件夹, 下载文件也是这个文件夹。
	WaitingFolder string

	// 上传时，单个文件的体积上限。由于在处理文件时需要把整个文件读入内存，
	// 因此需要限制文件体积，避免爆内存。
	FileSizeLimit int64

	// TagGroupLimit 限制标签组数量上限。
	// 当超过上限时，不受保护的标签组会被覆盖。可通过点击 "protect" 按钮保护标签。
	TagGroupLimit int64

	// FileListLimit 限制最近文件（及最近图片）的上限，但不限制搜索结果的上限。
	// 因此，在 “全部文件” 和 “全部图片” 的列表中不会列出超过上限的文件，只能通过搜索来找出被隐藏的文件。
	FileListLimit int64

	// 当 TimeNow - file.Checked > checkInterval 时，该文件需要重新检查。
	CheckInterval int64
}

// Default 默认设定
func Default() Config {
	dataDir := filepath.Join(util.UserHomeDir(), dataFolderName)
	return Config{
		Address:       "127.0.0.1:53549",
		DataFolder:    dataDir,
		WaitingFolder: filepath.Join(dataDir, waitingFolderName),
		FileSizeLimit: 1 << 29, // 512 MB
		TagGroupLimit: 50,
		FileListLimit: 100,
		CheckInterval: TimeUnit * 90, // 90天
	}
}
