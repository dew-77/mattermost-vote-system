package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Mattermost MattermostConfig
	Tarantool  TarantoolConfig
	Bot        BotConfig
}

type MattermostConfig struct {
	ServerURL  string
	Token      string
	TeamName   string
	BotUserID  string
}

type TarantoolConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Space    string
}

type BotConfig struct {
	LogLevel string
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	
	viper.SetDefault("mattermost.serverURL", "http://host.docker.internal:8065")
	viper.SetDefault("mattermost.token", "")
	viper.SetDefault("mattermost.teamName", "")
	viper.SetDefault("mattermost.botUserID", "")
	
	viper.SetDefault("tarantool.host", "localhost")
	viper.SetDefault("tarantool.port", 3301)
	viper.SetDefault("tarantool.user", "admin")
	viper.SetDefault("tarantool.password", "password")
	viper.SetDefault("tarantool.space", "polls")
	
	viper.SetDefault("bot.logLevel", "info")
	
	viper.AutomaticEnv()
	
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}
	
	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	
	return &config, nil
}