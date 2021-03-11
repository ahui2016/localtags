package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ahui2016/localtags/config"
	"github.com/ahui2016/localtags/database"
	"github.com/ahui2016/localtags/thumbnail"
	"github.com/ahui2016/localtags/util"
)

const (
	OK = http.StatusOK
)
const (
	dbFileName   = "localtags.db"
	tempFolder   = "public/temp"               // 临时文件夹的完整路径
	tempMetadata = "public/temp/metadata.json" // 临时文件数据
)

var (
	cfgFlag = flag.String("config", "", "config file path")
)

var (
	cfg       config.Config
	dbPath    string // 数据库文件完整路径
	hasFFmpeg bool   // 系统中有没有安装 FFmpeg
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
	hasFFmpeg = thumbnail.CheckFFmpeg()
}

func setConfig() {
	cfg = config.Default()
	configJSON, err := os.ReadFile(configFile)

	// 找不到文件或内容为空
	if err != nil || len(configJSON) == 0 {
		util.MarshalWrite(cfg, configFile)
		return
	}

	// configFile 有内容
	util.Panic(json.Unmarshal(configJSON, &cfg))
}

func setPaths() {
	util.MustMkdir(cfg.DataFolder)
	util.MustMkdir(cfg.WaitingFolder)
	dbPath = filepath.Join(cfg.DataFolder, dbFileName)
}
