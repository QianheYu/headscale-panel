package config

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"headscale-panel/model"
	"headscale-panel/util"
	"os"
	"path"
	"sync/atomic"
)

// 用于保存Headcale Config 确保在操作时不产生竞争问题
var value = atomic.Value{}

// 用于或当前的Headscale Config
func GetHeadscaleConfig() *model.HeadscaleConfig {
	return value.Load().(*model.HeadscaleConfig)
}

// SetApiKey 设置用于grpc连接所使用的APIKEY到Headscale Config中
func SetApiKey(apikey string) {
	conf := value.Load().(*model.HeadscaleConfig)
	conf.ApiKey = apikey
	value.Swap(conf)
}

// SetCert 用于设置grpc连接所使用的公钥、私钥、CA证书和到Headscale Config
//func SetCert(ca, cert, key []byte, serverName string) {
//	conf := value.Load().(*model.HeadscaleConfig)
//	conf.Cert = cert
//	conf.Key = key
//	conf.CA = ca
//	conf.CustomCert = len(ca) > 0
//	conf.ServerName = serverName
//	value.Swap(conf)
//}

// 在单机部署时加载Headscale配置文件并监听配置文件变更，用于grpc连接
func InitHeadscaleConfig() {
	viper.SetConfigType("yaml")
	if mode >= MULTI {
		return
	}

	// create config file dir if not exist
	dir, _ := path.Split(Conf.Headscale.Config)
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(fmt.Errorf("create dir %s error %s", dir, err))
		}
	}

	// check the config file exist
	if _, err := os.Stat(Conf.Headscale.Config); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			panic(fmt.Errorf("stat %s error %s", Conf.Headscale.Config, err))
		}
		// download the config file
		tmp, err := util.Download("https://raw.githubusercontent.com/juanfont/headscale/master/config-example.yaml",
			"/tmp/headscale/config.yaml", 0644)
		if err != nil {
			panic(fmt.Errorf("download config file %s error %s", Conf.Headscale.Config, err))
		}
		if err := util.Update(tmp, Conf.Headscale.Config); err != nil {
			panic(fmt.Errorf("update config file %s error %s", Conf.Headscale.Config, err))
		}
	}

	// 设置监听的配置文件
	viper.SetConfigFile(Conf.Headscale.Config)

	// 读取配置
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("read config file %s error %s", Conf.Headscale.Config, err))
	}

	// 解析配置
	HeadscaleConf := &model.HeadscaleConfig{}
	if err := viper.Unmarshal(HeadscaleConf); err != nil {
		panic(fmt.Errorf("unable to decode into headscale config struct, %v", err))
	}

	// 补充Headscale配置文件中没有的内容
	HeadscaleConf.Cert = []byte(Conf.Headscale.Cert)
	HeadscaleConf.Key = []byte(Conf.Headscale.Key)
	HeadscaleConf.CA = []byte(Conf.Headscale.CA)
	HeadscaleConf.CustomCert = len(Conf.Headscale.CA) > 0
	HeadscaleConf.ServerName = Conf.Headscale.ServerName
	HeadscaleConf.OIDC.Authorization = Conf.Headscale.OIDC.Authorize
	value.Store(HeadscaleConf)

	// 监听配置文件
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		conf := value.Load().(*model.HeadscaleConfig)
		if conf != nil {
			// 临时保存配置文件中不存在的内容
			apikey := conf.ApiKey
			//ca := conf.CA
			//cert := conf.Cert
			//key := conf.Key
			//serverName := conf.ServerName

			if err := viper.Unmarshal(conf); err != nil {
				fmt.Printf("unable to decode into headscale config struct, %v", err)
				panic(fmt.Errorf("unable to decode into headscale config struct, %v", err))
			}

			// 重新赋值
			conf.ApiKey = apikey
			conf.Cert = []byte(Conf.Headscale.Cert)
			conf.Key = []byte(Conf.Headscale.Key)
			conf.CA = []byte(Conf.Headscale.CA)
			conf.CustomCert = len(Conf.Headscale.CA) > 0
			conf.ServerName = Conf.Headscale.ServerName
			conf.OIDC.Authorization = Conf.Headscale.OIDC.Authorize
			value.Swap(conf)
		}
	})
}
