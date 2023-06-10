package vo

type InstallRequest struct {
	State string `json:"state" form:"state" validate:"required,oneof=version install upgrade"`
	ID    uint   `json:"id" form:"id" validate:"required_if=State install"`
}

type SystemSettingHeadscale struct {
	ServerAddr string `json:"grpc_listen_addr" form:"grpc_listen_addr" validate:"required_without_all=Yaml"`
	ApiKey     string `json:"api_key" form:"api_key" validate:"required_without_all=Yaml"`
	Insecure   bool   `json:"grpc_allow_insecure" form:"grpc_allow_insecure"`
	BaseDomain string `json:"base_domain" form:"base_domain"`
	Yaml       string `json:"yaml" form:"yaml"`
}
