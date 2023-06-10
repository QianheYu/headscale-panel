package repository

import (
	"errors"
	"github.com/patrickmn/go-cache"
	"headscale-panel/common"
	"headscale-panel/config"
	"headscale-panel/model"
	"headscale-panel/util"
	"headscale-panel/vo"
	"os"
	"time"
)

var systemCache = cache.New(168*time.Hour, 24*time.Hour)

// Get or Save grpc or headscale config to different place.
// In Stand-alone mode, save headscale config to config file.
// In Multi mode, save grpc client config to database.

type HeadscaleConfigRepository interface {
	GetHeadscaleConfigFromFile(file string) (string, error)
	GetHeadscaleConfigFromDB() (*model.HeadscaleConfig, error)
	SetHeadscaleConfigFromFile(file string, headscale *vo.SystemSettingHeadscale) error
	SetHeadscaleConfigFromDB(headscale *vo.SystemSettingHeadscale) error
	//SetHeadscaleCert(reader io.Reader) error
	//SetHeadscaleKey(reader io.Reader) error
	//SetHeadscaleCA(reader io.Reader) error
}

type headscaleConfigRepository struct{}

// NewHeadscaleConfigRepository creates a new instance of headscale configuration repository.
func NewHeadscaleConfigRepository() HeadscaleConfigRepository {
	return &headscaleConfigRepository{}
}

// GetHeadscaleConfigFromFile retrieves the content of headscale config file.
func (s headscaleConfigRepository) GetHeadscaleConfigFromFile(file string) (string, error) {
	data, ok := systemCache.Get("headscale-config")
	if ok {
		return data.(string), nil
	}

	bytesData, err := util.ReadFile(file)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		return "", nil
	}
	strData := string(bytesData)
	systemCache.Set("headscale-config", strData, cache.DefaultExpiration)
	return strData, nil
}

// GetHeadscaleConfigFromDB retrieves the grpc configuration to connect headscale from the database.
func (s headscaleConfigRepository) GetHeadscaleConfigFromDB() (*model.HeadscaleConfig, error) {
	data, ok := systemCache.Get("headscale-config")
	if ok {
		return data.(*model.HeadscaleConfig), nil
	}
	modelData := common.GetHeadscaleConfig()
	systemCache.Set("headscale-config", modelData, cache.DefaultExpiration)
	return modelData, nil
}

// SetHeadscaleConfigFromFile sets the content of headscale config file.
func (s headscaleConfigRepository) SetHeadscaleConfigFromFile(file string, h *vo.SystemSettingHeadscale) error {
	if err := util.SaveFile(file, []byte(h.Yaml)); err != nil {
		return err
	}
	systemCache.Delete("headscale-config")
	return nil
}

// SetHeadscaleConfigFromDB sets the grpc configuration to connect headscale in the database.
func (s headscaleConfigRepository) SetHeadscaleConfigFromDB(h *vo.SystemSettingHeadscale) error {
	data := &model.Headscale{
		GRPCServerAddr: h.ServerAddr,
		ApiKey:         h.ApiKey,
		Insecure:       h.Insecure,
		BaseDomain:     h.BaseDomain,
	}
	if err := common.DB.Model(data).Select("g_rpc_server_addr", "api_key", "insecure", "base_domain").Where("insecure in (true, false)").Updates(data).Error; err != nil {
		return err
	}
	//common.SetHeadscaleConfig(data)
	common.SetHeadscale(data)
	return nil
}

//func (s headscaleConfigRepository) SetHeadscaleCert(reader io.Reader) error {
//	cert, err := io.ReadAll(reader)
//	if err != nil {
//		return err
//	}
//	if err = common.DB.Model(&model.Headscale{}).Where("insecure in (true, false)").Update("cert", string(cert)).Error; err != nil {
//		return err
//	}
//	return nil
//}
//
//func (s headscaleConfigRepository) SetHeadscaleKey(reader io.Reader) error {
//	key, err := io.ReadAll(reader)
//	if err != nil {
//		return err
//	}
//	if err = common.DB.Model(&model.Headscale{}).Where("insecure in (true, false)").Update("key", string(key)).Error; err != nil {
//		return err
//	}
//	return nil
//}
//
//func (s headscaleConfigRepository) SetHeadscaleCA(reader io.Reader) error {
//	ca, err := io.ReadAll(reader)
//	if err != nil {
//		return err
//	}
//	if err = common.DB.Model(&model.Headscale{}).Where("insecure in (true, false)").Update("ca", string(ca)).Error; err != nil {
//		return err
//	}
//	return nil
//}

/* AccessControlRepository -------------------------------------- */

var aclFile string

type AccessControlRepository interface {
	GetAccessControl() (string, error)
	SetAccessControl(aclContent string) error
}

type accessControlRepository struct{}

// NewAccessControlRepository new a repository with acl file. Not available at Stand-alone mode
func NewAccessControlRepository() AccessControlRepository {
	aclFile = common.GetHeadscaleConfig().AccessControl
	return &accessControlRepository{}
}

// GetAccessControl get content of access_control.yaml
func (s accessControlRepository) GetAccessControl() (string, error) {
	if config.GetMode() >= config.MULTI {
		return "", errors.New("multi mode not support")
	}

	if aclFile == "" {
		return "", errors.New("acl config not set")
	}

	if data, ok := systemCache.Get("access_control"); ok {
		return data.(string), nil
	}
	content, err := util.ReadFile(aclFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		return "", nil
	}
	systemCache.Set("access_control", string(content), cache.DefaultExpiration)
	return string(content), nil
}

// SetAccessControl set content of access_control.yaml
func (s accessControlRepository) SetAccessControl(acl string) error {
	if config.GetMode() >= config.MULTI {
		return errors.New("multi mode not support")
	}
	if aclFile == "" {
		return errors.New("acl config not set")
	}
	if err := util.SaveFile(aclFile, []byte(acl)); err != nil {
		return err
	}
	// todo restart headscale server if using oidc
	systemCache.Delete("access_control")
	return nil
}
