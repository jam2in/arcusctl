package internal

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var (
	Flags = struct {
		Verbose    bool
		ConfigFile string
	}{}

	Config = struct {
		ZooKeeper string
		Home      string
	}{}
)

func InitConfig() {
	v := viper.New()

	if Flags.ConfigFile != "" {
		v.SetConfigFile(Flags.ConfigFile)
	} else {
		exePath, err := os.Executable()
		if err != nil {
			panic(err)
		}
		v.AddConfigPath(filepath.Dir(exePath))
		v.AddConfigPath(".")
		v.SetConfigName("config")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	v.SetDefault("home", filepath.Join(homeDir, ".arcusctl"))

	v.SetEnvPrefix("ARCUSCTL")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if Flags.Verbose {
				log.Println(err)
			}
		} else {
			panic(err)
		}
	}

	if err := v.Unmarshal(&Config); err != nil {
		panic(err)
	}
}
