package model

type Headscale struct {
	GRPCServerAddr string `gorm:"type:varchar(50);notnull" json:"grpc_server_addr"`
	Insecure       bool   `gorm:"type:boolean" json:"insecure"`
	//CustomCert     bool   `gorm:"type:boolean" json:"custom_cert"`
	//Cert           string `gorm:"type:varchar(50)" json:"cert"`
	//Key            string `gorm:"type:text" json:"key"`
	//CA             string `gorm:"type:text" json:"ca"`
	//ServerName     string `gorm:"type:varchar(100)" json:"server_name"`
	ApiKey     string `gorm:"type:varchar(100)" json:"api_key"`
	BaseDomain string `gorm:"type:varchar(50);notnull" json:"base_domain"`
}

type HeadscaleConfig struct {
	GRPCListenAddr string    `json:"grpc_listen_addr" mapstructure:"grpc_listen_addr"`
	ApiKey         string    `json:"api_key" mapstructure:"-"`
	Insecure       bool      `json:"grpc_allow_insecure" mapstructure:"grpc_allow_insecure"`
	CustomCert     bool      `json:"custom_cert" mapstructure:"-"`
	Cert           []byte    `json:"tls_cert_path" mapstructure:"-"`
	CA             []byte    `json:"ca_path" mapstructure:"-"`
	Key            []byte    `json:"tls_key_path" mapstructure:"-"`
	ServerName     string    `json:"server_name" mapstructure:"-"`
	AccessControl  string    `json:"acl_policy_path" mapstructure:"acl_policy_path"`
	DNS            DNSConfig `json:"dns_config" mapstructure:"dns_config"`
	OIDC           OIDC      `json:"oidc" mapstructure:"oidc"`

	//CertPath string `mapstructure:"tls_cert_path"`
	//KeyPath  string `mapstraucture:"tls_key_path"`
}

type DNSConfig struct {
	BaseDomain string `gorm:"type:varchar(50)" json:"base_domain" mapstructure:"base_domain"`
}

type OIDC struct {
	OnlyStartIfOIDCIsAvailable bool     `json:"only_start_if_oidc_is_available" mapstructure:"only_start_if_oidc_is_available"`
	Issuer                     string   `json:"issuer" mapstructure:"issuer"`
	Authorization              string   `json:"authorization" mapstructure:"-"`
	ClientID                   string   `json:"client_id" mapstructure:"client_id"`
	ClientSecret               string   `json:"client_secret" mapstructure:"client_secret"`
	Scope                      []string `json:"scope" mapstructure:"scope"`
}
