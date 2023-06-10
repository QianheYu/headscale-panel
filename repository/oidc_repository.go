package repository

import (
	"fmt"
	"github.com/aidarkhanov/nanoid"
	"github.com/go-jose/go-jose/v3"
	"github.com/golang-jwt/jwt/v4"
	"github.com/patrickmn/go-cache"
	"headscale-panel/common"
	"headscale-panel/config"
	"headscale-panel/log"
	"headscale-panel/model"
	"headscale-panel/util"
	"time"
)

var oidcCache *cache.Cache
var privateKey any
var publicKey any
var key []jose.JSONWebKey
var token = jwt.New(jwt.SigningMethodRS256)
var oidcConfig *model.OIDC

type OIDC interface {
	GenerateCode(user *model.User, redirectURL string) string
	VerifyCode(code, redirectURL string) (interface{}, bool)
	GetAccessTokenAndRefreshToken(user interface{}) (string, string, error)
	GetIDToken(user interface{}, accessToken string) (string, error)
	GetJsonWebKeys() []jose.JSONWebKey
	GetClientId() string
}

type oidc struct{}

func NewOIDC() OIDC {
	o := &oidc{}
	key = append(key, jose.JSONWebKey{Key: config.Conf.System.PublicKey})
	oidcConfig = &common.GetHeadscaleConfig().OIDC
	oidcCache = cache.New(5*time.Minute, 10*time.Minute)
	return o
}

func (o *oidc) GenerateCode(user *model.User, redirectURL string) string {
	code := nanoid.New()
	oidcCache.Set(code+redirectURL, user, cache.DefaultExpiration)
	return code
}

func (o *oidc) VerifyCode(code, redirectURL string) (interface{}, bool) {
	user, ok := oidcCache.Get(code + redirectURL)
	oidcCache.Delete(code)
	return user, ok
}

func (o *oidc) GetUserInfo(code string) (interface{}, error) {
	user, ok := oidcCache.Get(code)
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (o *oidc) GetAccessTokenAndRefreshToken(user interface{}) (string, string, error) {
	// TODO
	accessToken := nanoid.New()
	refreshToken := nanoid.New()
	return accessToken, refreshToken, nil
}

func (o *oidc) GetIDToken(user interface{}, accessToken string) (string, error) {
	// encode access_token
	hash, err := util.RSAEncrypt([]byte(accessToken), config.Conf.System.PublicKey)
	if err != nil {
		return "", fmt.Errorf("encrypt access_token: %w", err)
	}
	u := user.(*model.User)
	token.Claims = o.setClaims(u, util.EncodeStr2Base64(string(hash[:16])))
	id_token, err := token.SignedString(config.Conf.System.PrivateKey)
	if err != nil {
		log.Log.Error(err)
	}
	return id_token, err
}

func (o *oidc) setClaims(user *model.User, accessToken string) jwt.Claims {
	groups := []string{}
	for _, role := range user.Roles {
		groups = append(groups, role.Keyword)
	}
	return jwt.MapClaims{
		"sub":                nanoid.New(),
		"iss":                o.getIssuer(),
		"aud":                o.GetClientId(),
		"iat":                time.Now().Unix(),
		"exp":                time.Now().Add(time.Minute).Unix(),
		"name":               user.Name,
		"preferred_username": user.Name,
		"groups":             groups,
		// As headscale uses mailboxes to identify users rather than usernames, usernames are used here
		"email":   user.Name + "@example.com",
		"at_hash": accessToken,
	}
}

func (o *oidc) GetJsonWebKeys() []jose.JSONWebKey {
	return key
}

func (o *oidc) GetClientId() string {
	return oidcConfig.ClientID
}

func (o *oidc) GetClientSecret() string {
	return oidcConfig.ClientSecret
}

func (o *oidc) getIssuer() string {
	return oidcConfig.Issuer
}
