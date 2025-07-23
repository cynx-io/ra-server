package config

import "github.com/cynxees/cynx-core/src/configuration"

var Config *AppConfig

type AppConfig struct {
	Aws      AwsConfig      `mapstructure:"aws"`
	Elastic  ElasticConfig  `mapstructure:"elastic"`
	App      App            `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
}

type App struct {
	Name    string `mapstructure:"name"`
	Address string `mapstructure:"address"`
	Key     string `mapstructure:"key"`
	Port    int    `mapstructure:"port"`
	Debug   bool   `mapstructure:"debug"`
}

type DatabaseConfig struct {
	Host        string `mapstructure:"host"`
	Database    string `mapstructure:"database"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	Dialect     string `mapstructure:"dialect"`
	AutoMigrate bool   `mapstructure:"autoMigrate"`
	Pool        struct {
		Max     int `mapstructure:"max"`
		Min     int `mapstructure:"min"`
		Acquire int `mapstructure:"acquire"`
		Idle    int `mapstructure:"idle"`
	} `mapstructure:"pool"`
	Port int `mapstructure:"port"`
}

type ElasticConfig struct {
	Url   string `json:"url"`
	Level string `json:"level"`
}

type AwsConfig struct {
	S3 struct {
		Region          string `mapstructure:"region"`
		AccessKeyId     string `mapstructure:"accessKeyId"`
		SecretAccessKey string `mapstructure:"secretAccessKey"`
	} `mapstructure:"s3"`
}

func InitConfig() {
	Config = &AppConfig{}
	err := configuration.InitConfig("config.json", Config)
	if err != nil {
		panic("failed to initialize config: " + err.Error())
	}
}
