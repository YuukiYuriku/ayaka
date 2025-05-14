package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      string
	AppVer   string
	Env      string
	Http     HttpConfig
	Log      LogConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Toggle   ToggleConfig
	JWT      JwtConfig
	Email    EmailConfig
	Domain   DomainConfig
}

type HttpConfig struct {
	Port         int
	WriteTimeout int
	ReadTimeout  int
}

type LogConfig struct {
	FileLocation    string
	FileTDRLocation string
	FileMaxSize     int
	FileMaxBackup   int
	FileMaxAge      int
	Stdout          bool
}

type DatabaseConfig struct {
	Host            string
	User            string
	Password        string
	DBName          string
	Port            string
	SSLMode         string
	MaxIdleConn     int
	ConnMaxLifetime int
	MaxOpenConn     int
}

type RedisConfig struct {
	Mode     string
	Address  string
	Port     int
	Password string
}

type ToggleConfig struct {
	AppName string
	URL     string
	Token   string
}

type JwtConfig struct {
	JWTKEY             string
	JWTDuration        int
	ChangePassKey      string
	ChangePassDuration int
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	User         string
	Pass         string
	EmailFrom    string
	SendDuration int
}

type DomainConfig struct {
	FrontendDomain     string
	ForgotPass         string
	ChangePassDuration int
}

func (c *Config) LoadConfig(path string) {
	viper.AddConfigPath(".")
	viper.SetConfigName(path)

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("fatal error config file: default \n", err)
		os.Exit(1)
	}

	for _, k := range viper.AllKeys() {
		value := viper.GetString(k)
		if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
			viper.Set(k, getEnvOrPanic(strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")))
		}
	}

	err = viper.Unmarshal(c)
	if err != nil {
		fmt.Println("fatal error config file: default \n", err)
		os.Exit(1)
	}
}

func getEnvOrPanic(env string) string {
	split := strings.Split(env, ":")
	res := os.Getenv(split[0])
	if len(res) == 0 {
		if len(split) > 1 {
			res = strings.Join(split[1:], ":")
		}
		if len(res) == 0 {
			panic("Mandatory env variable not found:" + env)
		}
	}
	return res
}
