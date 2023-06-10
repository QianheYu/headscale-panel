package vo

// Authorize represents the authorization request parameters.
type Authorize struct {
	ClientID     string `json:"client_id" form:"client_id"`
	ClientSecret string `json:"client_secret" form:"client_secret"`
	RedirectURI  string `json:"redirect_uri" form:"redirect_uri" validate:"required,url"`
	ResponseType string `json:"response_type" form:"response_type" validate:"required,eq=code"`
	Scope        string `json:"scope" form:"scope"`
	State        string `json:"state" form:"state"`
	//Code         string `json:"code"`
}

// Token represents the token request parameters.
type Token struct {
	ClientID            string `json:"client_id" form:"client_id" validate:"required_if=GrantType authorization_code"`                                  // used to register a client and get token
	ClientSecret        string `json:"client_secret" form:"client_secret" validate:"required_if=GrantType refresh_token|required_with_all=RedirectURI"` // used to refresh token and get token
	Code                string `json:"code" form:"code" validate:"required_with_all=ClientID"`                                                          // used to register a client and get token
	RedirectURI         string `json:"redirect_uri" form:"redirect_uri" validate:"required_with_all=ClientID ClientSecret,url"`                         // only used in get token
	GrantType           string `json:"grant_type" form:"grant_type" validate:"required"`
	RefreshToken        string `json:"refresh_token" form:"refresh_token" validate:"required_if=GrantType refresh_token"`              // only used in refresh token
	Scope               string `json:"scope" form:"scope" validate:"required_if=GrantType refresh_token"`                              // only used in refresh token
	ClientAssertionType string `json:"client_assertion_type" form:"client_assertion_type" validate:"required_without_all=RedirectURI"` // only used in client authorization
	ClientAssertion     string `json:"client_assertion" form:"client_assertion" validate:"required_without_all=RedirectURI"`           // only used in client authorization
}
