package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"headscale-panel/common"
	"headscale-panel/log"
	"headscale-panel/repository"
	"headscale-panel/response"
	"headscale-panel/vo"
	"net/http"
	"net/url"
)

type IOIDCController interface {
	Authorize(c *gin.Context)
	Token(c *gin.Context)
	JWKs(c *gin.Context)
	GetUserInfo(c *gin.Context)
}

type OIDCController struct {
	repo     repository.OIDC
	userRepo repository.IUserRepository
}

func NewOIDCController() IOIDCController {
	repo := repository.NewOIDC()
	if repo == nil {
		return nil
	}
	return &OIDCController{repo: repo, userRepo: repository.NewUserRepository()}
}

// GetOpenIDConfiguration headscale Access this interface at startup to get the address of the oidc function interface
// Returns a json with issuer, authorization endpoint, token endpoint, jwks uri, userinfo endpoint, list of signature encryption algorithms
func GetOpenIDConfiguration(c *gin.Context) {
	conf := common.GetHeadscaleConfig()
	token, err := url.JoinPath(conf.OIDC.Issuer, "/api/oidc/token")
	if err != nil {
		response.Fail(c, nil, "Unknown error")
		log.Log.Errorf("url join path error: %v", err)
		return
	}
	jwks, err := url.JoinPath(conf.OIDC.Issuer, "/api/oidc/jwk")
	if err != nil {
		response.Fail(c, nil, "Unknown error")
		log.Log.Errorf("url join path error: %v", err)
		return
	}
	userInfo, err := url.JoinPath(conf.OIDC.Issuer, "/api/oidc/user_info")
	if err != nil {
		response.Fail(c, nil, "Unknown error")
		log.Log.Errorf("url join path error: %v", err)
		return
	}

	//authorization, err := url.JoinPath(conf.OIDC.Authorization, "/#/connect")
	//if err != nil {
	//	response.Fail(c, nil, "Unknown error")
	//	log.Log.Errorf("url join path error: %v", err)
	//	return
	//}

	c.JSON(http.StatusOK, gin.H{
		"issuer":                                conf.OIDC.Issuer,
		"authorization_endpoint":                conf.OIDC.Authorization,
		"token_endpoint":                        token,
		"jwks_uri":                              jwks,
		"userinfo_endpoint":                     userInfo,
		"id_token_signing_alg_values_supported": []string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512", "PS256", "PS384", "PS512"},
	})
}

// Authorize is used to get authorization code
// The Headscale oidc client will send the authorization code to the authorization endpoint
func (o *OIDCController) Authorize(c *gin.Context) {
	req := &vo.Authorize{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}

	user, err := o.userRepo.GetCurrentUser(c)
	if err != nil {
		response.Fail(c, nil, "get user info error")
		return
	}
	// Code used to exchange the access token.
	// It needs to randomly generate and storage the code
	code := o.repo.GenerateCode(&user, req.RedirectURI)
	response.Response(c, http.StatusOK, http.StatusFound, fmt.Sprintf("%s?code=%s&state=%s", req.RedirectURI, code, req.State), "")
}

// Token token endpoint
// used to get new by refresh token or verify access token
func (o *OIDCController) Token(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")

	req := &vo.Token{}
	if err := c.ShouldBind(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	// validate data
	if err := common.Validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	switch req.GrantType {
	case "authorization_code":
		// The client id and client secret are only the same as config file of headscale.
		// Not use the RFC6749 compliant client id and secret.
		// It could be realised in the future.
		oidc := common.GetHeadscaleConfig().OIDC
		if req.ClientID != oidc.ClientID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}
		// Determining whether to get a token or register a client.
		if req.ClientSecret == "" {
			// TODO register client id. It could be realised in the future.
			c.JSON(http.StatusOK, nil)
		} else {
			if req.ClientSecret != oidc.ClientSecret {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
				return
			}
			accessToken, refreshToken, err := o.repo.GetAccessTokenAndRefreshToken(req)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
				return
			}
			// search the code from the storage, which has the code response tokens
			user, ok := o.repo.VerifyCode(req.Code, req.RedirectURI)
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
				log.Log.Errorf("code error: code is %s", req.Code)
				return
			}
			idToken, err := o.repo.GetIDToken(user, accessToken)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
				log.Log.Errorf("Signing token error:%s", err)
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"id_token":      idToken,
				"access_token":  accessToken,
				"refresh_token": refreshToken,
				"token_type":    "Bearer",
				"expires_in":    3600,
			})
		}
	case "refresh_token":
		// TODO Depends on whether headscale is required.
		accessToken, refreshToken, err := o.repo.GetAccessTokenAndRefreshToken(nil)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"token_type":    "Bearer",
			"expires_in":    3600,
		})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
	}
}

// JWKs return the cert to verify key signed the jwt
func (o *OIDCController) JWKs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"keys": o.repo.GetJsonWebKeys(),
	})
}

// GetUserInfo using AccessToken
// It Will Return UserInfo
// Depends on whether headscale is required.
func (o *OIDCController) GetUserInfo(c *gin.Context) {
	user, err := o.userRepo.GetCurrentUser(c)
	if err != nil {
		response.Fail(c, nil, "Failed to get user")
		log.Log.Errorf("get current user error: %v", err)
		return
	}
	c.JSON(http.StatusOK, user)
}
