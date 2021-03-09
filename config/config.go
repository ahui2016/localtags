package config

// Config 用来设置一些全局变量
type Config struct {

	// 数据文件夹名称
	DataFolderName string

	// 待上传的文件要放进这个文件夹里
	WaitingFolderName string
}

// Default 默认设定
func Default() Config {
	return Config{
		DataFolderName:    "localtags_data_folder",
		WaitingFolderName: "waiting",
	}
}
