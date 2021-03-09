package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"path/filepath"

	localConfig "github.com/ahui2016/localtags/config"
	"github.com/ahui2016/localtags/database"
	"github.com/ahui2016/localtags/util"
)

var (
	cfgFlag   = flag.String("config", "", "run with a config file")
	dbDirFlag = flag.String("dir", "", "database directory")
)

var (
	config  localConfig.Config
	dataDir string // 数据库文件夹
	dbPath  string // 数据库文件完整路径
)

var (
	db         = new(database.DB)
	configFile = "config.json"
)

const dbFileName = "localtags.db"

func init() {
	flag.Parse()
	if *cfgFlag != "" {
		configFile = *cfgFlag
	}
	if *dbDirFlag != "" {
		dataDir = *dbDirFlag
	}

	setConfig()
	setPaths()
}

func setConfig() {
	config = localConfig.Default()
	configJSON, err := ioutil.ReadFile(configFile)

	// 找不到文件或内容为空
	if err != nil || len(configJSON) == 0 {
		configJSON, err = json.MarshalIndent(config, "", "    ")
		util.Panic(err)
		util.Panic(ioutil.WriteFile(configFile, configJSON, 0600))
		return
	}

	// configFile 有内容
	util.Panic(json.Unmarshal(configJSON, &config))
}

func setPaths() {
	if dataDir == "" {
		if config.DataFolderName == "" {
			log.Fatal("config.DataFolderName is empty")
		}
		dataDir = filepath.Join(util.UserHomeDir(), config.DataFolderName)
	}
	dbPath = filepath.Join(dataDir, dbFileName)
}
