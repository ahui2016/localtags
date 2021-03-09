package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/ahui2016/localtags/config"
	"github.com/ahui2016/localtags/database"
	"github.com/ahui2016/localtags/util"
)

const (
	OK = http.StatusOK
)
const (
	dbFileName = "localtags.db"
)

var (
	cfgFlag = flag.String("config", "", "config file path")
)

var (
	cfg    config.Config
	dbPath string // 数据库文件完整路径
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
	dbPath = filepath.Join(cfg.DataFolder, dbFileName)

}

func setConfig() {
	cfg = config.Default()
	configJSON, err := ioutil.ReadFile(configFile)

	// 找不到文件或内容为空
	if err != nil || len(configJSON) == 0 {
		configJSON, err = json.MarshalIndent(cfg, "", "    ")
		util.Panic(err)
		util.Panic(ioutil.WriteFile(configFile, configJSON, 0600))
		return
	}

	// configFile 有内容
	util.Panic(json.Unmarshal(configJSON, &cfg))
}
