package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTExpired        string
	JWTRefreshExpired string
	JWTSecret         string
}

var singleConfigInstance *Config
var lockConfig = &sync.Mutex{}

func GetInstanceConfig() *Config {
	if singleConfigInstance == nil {
		lockConfig.Lock()
		defer lockConfig.Unlock()
		if singleConfigInstance == nil {
			fmt.Println("Creating single instance Config now.")
			singleConfigInstance = &Config{}
			singleConfigInstance.Init()
		} else {
			fmt.Println("Single instance already created.")
		}
	} else {
		// fmt.Println("Single instance already created.")
	}

	return singleConfigInstance
}

func (cf *Config) Init() {
	if err := godotenv.Load(); err != nil {
		fmt.Print(err)
	}

	cf.JWTExpired = os.Getenv("JWT_EXPIRED")
	cf.JWTRefreshExpired = os.Getenv("JWT_REFRESH_EXPIRED")
	cf.JWTSecret = os.Getenv("JWT_SECRET")
}
