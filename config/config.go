package config

import (
	"github.com/spf13/viper"
)

// Camer 配置
type Camer struct {
	URL              string
	Name             string
	VideoSegmentTime int
}

// GetVideoOputPath 获取视频输出地址
func (camer Camer) GetVideoOputPath() string {
	return Keys.VideoOputPath + camer.Name + "/"
}

type keys struct {
	Camers             []Camer
	VideoOputPath      string
	ConfigFile         string
	GoogleFolder       string
	ThreadNumber       int
	GoogleConfigFolder string
}

var viperConfig *viper.Viper

// ConfigFile ....
const ConfigFile = "config.json"

// Keys 配置
var Keys = keys{}

// InitConfig ...
func InitConfig(configFile string) {
	var err error
	viperConfig = viper.GetViper()
	// viperConfig.SetTypeByDefaultValue(true)

	viperConfig.SetConfigFile(configFile)
	err = viperConfig.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = viper.Unmarshal(&Keys)
	verifyKeys()
}

func verifyKeys() {
	verifiers := make([]func(), 0)
	verifiers = append(verifiers,
		verifyVideoOputPath)
	for _, v := range verifiers {
		v()
	}
}

func verifyVideoOputPath() {
	if Keys.VideoOputPath == "" {
		panic("VideoOputPath not config !!")
	}
}
