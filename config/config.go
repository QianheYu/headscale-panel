package config

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	"headscale-panel/util"
	"headscale-panel/version"
	"os"
)

const (
	STANDALONE_IN_DOCKER int = iota
	STANDALONE
	//STANDALONE_DISABLE_UPDATE
	MULTI
)

var mode int

func GetMode() int {
	return mode
}

// System configuration, corresponding to yml
// viper has a built-in mapstructure, yml files use "-" to distinguish words, turn them into humps for convenience

// Global configuration variables
var Conf = new(config)

type config struct {
	System    *SystemConfig    `mapstructure:"system" json:"system"`
	Logs      *LogsConfig      `mapstructure:"logs" json:"logs"`
	Database  *DatabaseConfig  `mapstructure:"database" json:"database"`
	Casbin    *CasbinConfig    `mapstructure:"casbin" json:"casbin"`
	Jwt       *JwtConfig       `mapstructure:"jwt" json:"jwt"`
	RateLimit *RateLimitConfig `mapstructure:"rate-limit" json:"rateLimit"`
	Headscale *Headscale       `mapstructure:"headscale" json:"headscale"`
}

// Set to read configuration information
func InitConfig() {
	pflag.BoolP("version", "v", false, "print version")
	pflag.StringP("config", "c", "", "set config")
	pflag.Parse()
	viper.RegisterAlias("c", "config")
	viper.RegisterAlias("v", "version")
	viper.SetDefault("config", "/etc/headscale-panel/config.yaml")

	viper.SetEnvPrefix("HEADSCALE_PANEL")
	if err := viper.BindEnv("config"); err != nil {
		fmt.Printf("bind config file env error: %s\n", err.Error())
	}
	if err := viper.BindEnv("KEY_DECRYPTION_PWD"); err != nil {
		fmt.Printf("bind key decryption password error: %s\n", err.Error())
	}

	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		fmt.Printf("bind param flags error: %s\n", err.Error())
	}

	if viper.GetBool("version") {
		fmt.Println(version.Version)
		os.Exit(0)
	}

	configFile := viper.GetString("config")
	KeyDecryptionPwd := viper.GetString("KEY_DECRYPTION_PWD")

	if len(configFile) <= 0 {
		panic(fmt.Errorf("config path is empty"))
	}

	if fileInfo, err := os.Stat(configFile); err != nil || fileInfo.IsDir() {
		if errors.Is(err, os.ErrNotExist) {
			panic(fmt.Errorf("config path not exist"))
		}
		panic(fmt.Errorf("config file is error or not dir"))
	}

	viper.SetConfigFile(configFile)

	// Read configuration information
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Failed to read configuration file:%s \n", err))
	}

	// Hot update configuration
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		// Save the read configuration information to the global variable Conf
		err := viper.Unmarshal(Conf)
		if err != nil {
			panic(fmt.Errorf("Failed to initialise profile:%s \n", err))
		}
		// Read a pair of rsa keys
		Conf.System.PublicKey, err = util.LoadPublicKey(Conf.System.RSAPublicKey)
		if err != nil {
			panic(fmt.Errorf("init config file failed: load public key err:%s \n", err))
		}
		Conf.System.PrivateKey, err = util.LoadPrivateKey(Conf.System.RSAPrivateKey, Conf.System.KeyDecryptionPwd)
		if err != nil {
			panic(fmt.Errorf("init config file failed: load private key err:%s \n", err))
		}
		Conf.checkMode()
	})

	// Save the read configuration information to the global variable Conf
	err := viper.Unmarshal(Conf)
	if err != nil {
		panic(fmt.Errorf("Failed to initialise profile:%s \n", err))
	}
	// Read a pair of rsa keys
	Conf.System.PublicKey, err = util.LoadPublicKey(Conf.System.RSAPublicKey)
	if err != nil {
		panic(fmt.Errorf("init config file failed: load public key err:%s \n", err))
	}
	Conf.System.PrivateKey, err = util.LoadPrivateKey(Conf.System.RSAPrivateKey, Conf.System.KeyDecryptionPwd)
	if err != nil {
		panic(fmt.Errorf("init config file failed: load private key err:%s \n", err))
	}

	if len(Conf.System.KeyDecryptionPwd) <= 0 {
		Conf.System.KeyDecryptionPwd = KeyDecryptionPwd
	}

	if len(Conf.Headscale.CA) > 0 {
		ca, err := os.ReadFile(Conf.Headscale.CA)
		if err != nil {
			panic(fmt.Errorf("init config file failed: load headscale ca err: %s\n", err))
		}
		Conf.Headscale.CA = string(ca)
	}

	if len(Conf.Headscale.Cert) > 0 {
		cert, err := os.ReadFile(Conf.Headscale.Cert)
		if err != nil {
			panic(fmt.Errorf("init config file failed: load headscale cert err: %s\n", err))
		}
		Conf.Headscale.Cert = string(cert)
	}

	if len(Conf.Headscale.Key) > 0 {
		key, err := os.ReadFile(Conf.Headscale.Key)
		if err != nil {
			panic(fmt.Errorf("init config file failed: load headscale key err: %s\n", err))
		}
		Conf.Headscale.Key = string(key)
	}

	Conf.checkMode()
}

func (i *config) checkMode() {
	container := false
	if _, err := os.Stat("/.dockerenv"); err == nil {
		container = true
	}

	if Conf.Headscale == nil {
		panic("headscale can not empty in config.yaml. You must set it.")
	}

	switch Conf.Headscale.Mode {
	case "multi":
		mode = MULTI
	case "standalone":
		// deployment in docker
		if container {
			Conf.Headscale.App = "/bin/headscale"
			Conf.Headscale.Config = "/etc/headscale/config.yaml"
			Conf.Headscale.ACL = "/etc/headscale/acl.yaml"
			Conf.Headscale.Controller = &Controller{Inside: true}
			mode = STANDALONE_IN_DOCKER
		}

		// deployment in
		if Conf.Headscale.Config == "" || Conf.Headscale.ACL == "" {
			panic("Headscale Config and ACL is empty, you need set the Multi mode. Please set the headscale -> made of the yaml to multi.")
		}

		if Conf.Headscale.Controller == nil {
			panic("If you not need control headscale please set mode to multi in config.yaml.")
		}

		mode = STANDALONE

		if Conf.Headscale.Controller.Inside {
			// using the inside process controller
			if Conf.Headscale.App == "" {
				panic("Not set headscale -> controller -> app")
			}
		} else {
			// using outside process controller. example: systemctl
			cmd := Conf.Headscale.Controller.Command
			if cmd == nil || cmd.Start == "" || cmd.Stop == "" {
				panic("Not set headscale -> controller -> command")
			}
			// No app path set, update will not support
			if Conf.Headscale.App == "" {
				panic("Not set headscale -> controller -> app")
				//mode = STANDALONE_DISABLE_UPDATE
				//fmt.Println("Warning: Upgrade headscale function is not available.")
			}
		}
	default:
		panic("mode set error")
	}
}

type SystemConfig struct {
	Mode          string `mapstructure:"mode" json:"mode"`
	UrlPathPrefix string `mapstructure:"url-path-prefix" json:"urlPathPrefix"`
	//ServerURL     string `mapstructure:"server_url" json:"server_url"`
	ListenAddr       string          `mapstructure:"listen_addr" json:"listen_addr"`
	InitData         bool            `mapstructure:"init-data" json:"initData"`
	RSAPublicKey     string          `mapstructure:"rsa-public-key" json:"rsaPublicKey"`
	RSAPrivateKey    string          `mapstructure:"rsa-private-key" json:"rsaPrivateKey"`
	KeyDecryptionPwd string          `mapstructure:"key-decryption-password" json:"-"`
	PrivateKey       *rsa.PrivateKey `mapstructure:"-" json:"-"`
	PublicKey        *rsa.PublicKey  `mapstructure:"-" json:"-"`
}

type LogsConfig struct {
	Level      zapcore.Level `mapstructure:"level" json:"level"`
	Path       string        `mapstructure:"path" json:"path"`
	MaxSize    int           `mapstructure:"max-size" json:"maxSize"`
	MaxBackups int           `mapstructure:"max-backups" json:"maxBackups"`
	MaxAge     int           `mapstructure:"max-age" json:"maxAge"`
	Compress   bool          `mapstructure:"compress" json:"compress"`
}

type DatabaseConfig struct {
	Driver  string `mapstructure:"driver" json:"driver"`
	Dsn     string `mapstructure:"dsn" json:"dsn"`
	LogMode bool   `mapstructure:"log-mode" json:"logMode"`
}

type CasbinConfig struct {
	ModelPath string `mapstructure:"model-path" json:"modelPath"`
}

type JwtConfig struct {
	Realm      string `mapstructure:"realm" json:"realm"`
	Key        string `mapstructure:"key" json:"key"`
	Timeout    int    `mapstructure:"timeout" json:"timeout"`
	MaxRefresh int    `mapstructure:"max-refresh" json:"maxRefresh"`
}

type RateLimitConfig struct {
	FillInterval int64 `mapstructure:"fill-interval" json:"fillInterval"`
	Capacity     int64 `mapstructure:"capacity" json:"capacity"`
}

type Headscale struct {
	OIDC       *OIDC       `mapstructure:"oidc" json:"oidc"`
	Mode       string      `mapstructure:"mode" json:"mode"`
	App        string      `mapstructure:"app" json:"app"`
	Config     string      `mapstructure:"config" json:"config"`
	ACL        string      `mapstructure:"acl" json:"acl"`
	Controller *Controller `mapstructure:"controller" json:"controller"`
	Cert       string      `mapstructure:"cert" json:"cert"`
	Key        string      `mapstructure:"key" json:"key"`
	CA         string      `mapstructure:"ca" json:"ca"`
	ServerName string      `mapstructure:"server_name" json:"server_name"`
}

type OIDC struct {
	Issuer       string `mapstructure:"issuer" json:"issuer"`
	Authorize    string `mapstructure:"authorize" json:"authorize"`
	ClientID     string `mapstructure:"client_id" json:"client_id"`
	ClientSecret string `mapstructure:"client_secret" json:"client_secret"`
}

type Controller struct {
	Inside  bool     `mapstructure:"inside" json:"inside"`
	Command *Command `mapstructure:"command" json:"command"`
}

type Command struct {
	Start   string `mapstructure:"start" json:"start"`
	Stop    string `mapstructure:"stop" json:"stop"`
	Restart string `mapstructure:"restart" json:"restart"`
}
