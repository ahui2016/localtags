package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ahui2016/localtags/config"
	"github.com/ahui2016/localtags/database"
	"github.com/ahui2016/localtags/thumb"
	"github.com/ahui2016/localtags/util"
)

const (
	OK = http.StatusOK
)
const (
	dbFileName       = "localtags.db"
	backupDBFileName = "localtags.bak.db"
	mainBucketName   = "mainbucket"    // 主仓库文件夹名
	bakBucketName    = "backup_bucket" // 备份仓库文件夹名
	tempFolderName   = "temp"          // 临时文件夹
	thumbsFolderName = "thumbs"        // 仓库里的缩略图的文件夹名
	thumbSuffix      = ".small.jpg"    // 缩略图的后缀名
	tempMetadataName = "metadata.json" // 临时文件数据
)

var (
	cfgFlag = flag.String("config", "", "config file path")
)

var (
	dbPath       string // 数据库文件完整路径
	mainBucket   string // 主仓库
	thumbsFolder string // 主仓库缩略图文件夹的完整路径
	tempFolder   string // 临时文件夹的完整路径
	tempMetadata string // 临时文件数据完整路径
	hasFFmpeg    bool   // 系统中有没有安装 FFmpeg
)

var (
	db         = new(database.DB)
	configFile = "config.json" // 设定文件的路径
)

func init() {
	flag.Parse()
	if *cfgFlag != "" {
		configFile = *cfgFlag
	}
	cfg := getCfg()
	setPaths(cfg)
	hasFFmpeg = thumb.CheckFFmpeg()

	// open the db here, close the db in main().
	util.Panic(db.Open(dbPath, cfg))
}

func getCfg() config.Config {
	configJSON, err := os.ReadFile(configFile)
	// 找不到文件或内容为空
	if err != nil || len(configJSON) == 0 {
		util.MustMarshalWrite(config.Public, configFile)
		return config.Public
	}
	// configFile 有内容
	var cfg config.Config
	util.Panic(json.Unmarshal(configJSON, &cfg))
	return cfg
}

func setPaths(cfg config.Config) {
	dbPath = filepath.Join(cfg.DataFolder, dbFileName)
	mainBucket = filepath.Join(cfg.DataFolder, mainBucketName)
	tempFolder = filepath.Join(cfg.DataFolder, tempFolderName)
	thumbsFolder = filepath.Join(mainBucket, thumbsFolderName)
	tempMetadata = filepath.Join(tempFolder, tempMetadataName)
	util.MustMkdir(cfg.DataFolder)
	util.MustMkdir(cfg.WaitingFolder)
	util.MustMkdir(tempFolder)
	util.MustMkdir(mainBucket)
	util.MustMkdir(thumbsFolder)
}
