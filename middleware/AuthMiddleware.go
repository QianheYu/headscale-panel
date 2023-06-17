package middleware

import (
	"fmt"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"headscale-panel/config"
	"headscale-panel/log"
	"headscale-panel/model"
	"headscale-panel/repository"
	"headscale-panel/response"
	"headscale-panel/util"
	"headscale-panel/vo"
	"time"
)

// Initialize jwt middleware
func InitAuth() (*jwt.GinJWTMiddleware, error) {
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:           config.Conf.Jwt.Realm,                                 // jwt realm
		Key:             []byte(config.Conf.Jwt.Key),                           // Server side key
		Timeout:         time.Hour * time.Duration(config.Conf.Jwt.Timeout),    // token expiry time
		MaxRefresh:      time.Hour * time.Duration(config.Conf.Jwt.MaxRefresh), // token max refresh time (RefreshToken expire time=Timeout+MaxRefresh)
		PayloadFunc:     payloadFunc,                                           // Payload processing
		IdentityHandler: identityHandler,                                       // Parsing Claims
		Authenticator:   login,                                                 // Verify the correctness of token and process login logic
		Authorizator:    authorizator,                                          // Processing of successful user login verification
		Unauthorized:    unauthorized,                                          // Processing of failed user login verification
		LoginResponse:   loginResponse,                                         // Response after successful login
		LogoutResponse:  logoutResponse,                                        // Response after logout
		RefreshResponse: refreshResponse,                                       // Response after refreshing token
		TokenLookup:     "header: Authorization, query: token, cookie: jwt",    // Automatically look for the token in the request in these places
		TokenHeadName:   "Bearer",                                              // Header name
		TimeFunc:        time.Now,
	})
	return authMiddleware, err
}

// Payload processing
func payloadFunc(data interface{}) jwt.MapClaims {
	if v, ok := data.(map[string]interface{}); ok {
		var user model.User
		// Converting user json to structs
		util.JsonI2Struct(v["user"], &user)
		return jwt.MapClaims{
			jwt.IdentityKey: user.ID,
			"user":          v["user"],
		}
	}
	return jwt.MapClaims{}
}

// Parsing Claims
func identityHandler(c *gin.Context) interface{} {
	claims := jwt.ExtractClaims(c)
	// Here the return value type map[string]interface{} must match the data type of the payloadFunc and authorizator,
	// otherwise the authorization will fail and the cause will not be easily found
	return map[string]interface{}{
		"IdentityKey": claims[jwt.IdentityKey],
		"user":        claims["user"],
	}
}

// Verify the correctness of token and process login logic
func login(c *gin.Context) (interface{}, error) {
	var req vo.RegisterAndLoginRequest
	// Request json binding
	if err := c.ShouldBind(&req); err != nil {
		return "", err
	}

	// Password decryption via RSA
	decodeData, err := util.RSADecrypt([]byte(req.Password), config.Conf.System.PrivateKey)
	if err != nil {
		return nil, err
	}

	u := &model.User{
		Name:     req.Username,
		Password: string(decodeData),
	}

	// Password verification
	userRepository := repository.NewUserRepository()
	user, err := userRepository.Login(u)
	if err != nil {
		return nil, err
	}
	// Writing the user in json format, payloadFunc/authorizator will use the
	return map[string]interface{}{
		"user": util.Struct2Json(user),
	}, nil
}

// Processing of successful user login verification
func authorizator(data interface{}, c *gin.Context) bool {
	if v, ok := data.(map[string]interface{}); ok {
		userStr := v["user"].(string)
		var user model.User
		// Converting user json to structs
		util.Json2Struct(userStr, &user)
		// Save user to context, easy to fetch data when called by api
		c.Set("user", user)
		return true
	}
	return false
}

// Processing of failed user login verification
func unauthorized(c *gin.Context, code int, message string) {
	log.Log.Debugf("JWT authentication failed, error code: %d, error message: %s", code, message)
	response.Response(c, code, code, nil, fmt.Sprintf("JWT authentication failed, error code: %d, error message: %s", code, message))
}

// Response after successful login
func loginResponse(c *gin.Context, code int, token string, expires time.Time) {
	response.Response(c, code, code,
		gin.H{
			"token":   token,
			"expires": expires.Format("2006-01-02 15:04:05"),
		},
		"Login success")
}

// Response after logout
func logoutResponse(c *gin.Context, code int) {
	response.Success(c, nil, "Logout success")
}

// Response after refreshing token
func refreshResponse(c *gin.Context, code int, token string, expires time.Time) {
	response.Response(c, code, code,
		gin.H{
			"token":   token,
			"expires": expires,
		},
		"Refresh token success")
}
