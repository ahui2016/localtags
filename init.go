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
	mainBucketName   = "mainbucket"    // 主仓库文件夹名
	thumbsFolderName = "thumbs"        // 仓库里的缩略图的文件夹名
	tempFolderName   = "temp"          // 临时文件夹名称
	thumbSuffix      = ".small.jpg"    // 缩略图的后缀名
	tempMetadataName = "metadata.json" // 临时文件数据
)

var (
	cfgFlag = flag.String("config", "", "config file path")
)

var (
	cfg          config.Config
	dbPath       string // 数据库文件完整路径
	mainBucket   string // 主仓库文件夹完整路径
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
	setConfig()
	setPaths()
	hasFFmpeg = thumb.CheckFFmpeg()

	// open the db here, close the db in main().
	util.Panic(db.Open(dbPath))
}

func setConfig() {
	defer func() { cfg = config.Public }()

	configJSON, err := os.ReadFile(configFile)
	// 找不到文件或内容为空
	if err != nil || len(configJSON) == 0 {
		util.MarshalWrite(config.Public, configFile)
		return
	}
	// configFile 有内容
	util.Panic(json.Unmarshal(configJSON, &config.Public))
}

func setPaths() {
	dbPath = filepath.Join(cfg.DataFolder, dbFileName)
	mainBucket = filepath.Join(cfg.DataFolder, mainBucketName)
	thumbsFolder = filepath.Join(mainBucket, thumbsFolderName)
	tempFolder = filepath.Join(cfg.DataFolder, tempFolderName)
	tempMetadata = filepath.Join(tempFolder, tempMetadataName)
	util.MustMkdir(cfg.DataFolder)
	util.MustMkdir(cfg.WaitingFolder)
	util.MustMkdir(tempFolder)
	util.MustMkdir(mainBucket)
	util.MustMkdir(thumbsFolder)
}
