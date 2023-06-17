package middleware

import (
	"github.com/gin-gonic/gin"
	"headscale-panel/common"
	"headscale-panel/config"
	"headscale-panel/repository"
	"headscale-panel/response"
	"strings"
	"sync"
)

var checkLock sync.Mutex

// Casbin Middleware, RBAC-based Access Control Model
func CasbinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ur := repository.NewUserRepository()
		user, err := ur.GetCurrentUser(c)
		if err != nil {
			response.Response(c, 401, 401, nil, "User not logged in")
			c.Abort()
			return
		}
		if user.Status != 1 {
			response.Response(c, 401, 401, nil, "The current user has been disabled")
			c.Abort()
			return
		}
		// Get the user's full role
		roles := user.Roles
		// Get the Keyword of all the user's roles that have not been disabled
		var subs []string
		for _, role := range roles {
			if role.Status == 1 {
				subs = append(subs, role.Keyword)
			}
		}

		// Get the request path URL
		//obj := strings.Replace(c.Request.URL.Path, "/"+config.Conf.System.UrlPathPrefix, "", 1)
		obj := strings.TrimPrefix(c.FullPath(), "/"+config.Conf.System.UrlPathPrefix)
		// Get request method
		act := c.Request.Method

		isPass := check(subs, obj, act)
		if !isPass {
			response.Response(c, 401, 401, nil, "No permission")
			c.Abort()
			return
		}

		c.Next()
	}
}

func check(subs []string, obj string, act string) bool {
	// Only one request can be validated at any one time, otherwise the validation may fail
	checkLock.Lock()
	defer checkLock.Unlock()
	isPass := false
	for _, sub := range subs {
		pass, _ := common.CasbinEnforcer.Enforce(sub, obj, act)
		if pass {
			isPass = true
			break
		}
	}
	return isPass
}
