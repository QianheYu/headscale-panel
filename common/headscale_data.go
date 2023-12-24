package common

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"headscale-panel/config"
	"headscale-panel/log"
	"headscale-panel/model"
	"sync/atomic"
)

var headscaleConfigValue = &atomic.Value{}

func GetHeadscaleConfig() *model.HeadscaleConfig {
	if config.GetMode() < config.MULTI {
		return config.GetHeadscaleConfig()
	}
	return headscaleConfigValue.Load().(*model.HeadscaleConfig)
}

func SetHeadscaleConfig(headscale *model.HeadscaleConfig) {
	headscaleConfigValue.Swap(headscale)
}

func SetHeadscale(headscale *model.Headscale) {
	conf := headscaleConfigValue.Load().(*model.HeadscaleConfig)
	conf.GRPCListenAddr = headscale.GRPCServerAddr
	conf.Insecure = headscale.Insecure
	//conf.CustomCert = len(config.Conf.Headscale.CA) > 0
	//conf.Cert = []byte(config.Conf.Headscale.Cert)
	//conf.Key = []byte(config.Conf.Headscale.Key)
	//conf.CA = []byte(config.Conf.Headscale.CA)
	//conf.ServerName = config.Conf.Headscale.ServerName
	conf.ApiKey = headscale.ApiKey
	conf.DNS.BaseDomain = headscale.BaseDomain
	headscaleConfigValue.Swap(conf)
}

func InitHeadscale() {
	//HeadscaleConfig := &model.HeadscaleConfig{}
	Headscale := &model.Headscale{}
	if err := DB.First(&Headscale).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		panic(fmt.Errorf("init headscale get config from DB error: %v", err))
	}

	if config.GetMode() < config.MULTI {
		// overwrite headscale config using conifg file, not apikey
		config.SetApiKey(Headscale.ApiKey)
		//config.SetCert([]byte(config.Conf.Headscale.CA), []byte(config.Conf.Headscale.Cert), []byte(config.Conf.Headscale.Key), config.Conf.Headscale.ServerName)
	} else {
		headscaleConfig := &model.HeadscaleConfig{
			GRPCListenAddr: Headscale.GRPCServerAddr,
			ApiKey:         Headscale.ApiKey,
			Insecure:       Headscale.Insecure,
			CustomCert:     len(config.Conf.Headscale.CA) > 0,
			Cert:           []byte(config.Conf.Headscale.Cert),
			Key:            []byte(config.Conf.Headscale.Key),
			CA:             []byte(config.Conf.Headscale.CA),
			ServerName:     config.Conf.Headscale.ServerName,
			DNS: model.DNSConfig{
				BaseDomain: Headscale.BaseDomain,
			},
			OIDC: model.OIDC{
				Issuer:        config.Conf.Headscale.OIDC.Issuer,
				Authorization: config.Conf.Headscale.OIDC.Authorize,
				ClientID:      config.Conf.Headscale.OIDC.ClientID,
				ClientSecret:  config.Conf.Headscale.OIDC.ClientSecret,
			},
		}
		headscaleConfigValue.Store(headscaleConfig)
	}
	log.Log.Info("init headscale finished")
}
