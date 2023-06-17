package middleware

import (
	"github.com/gin-gonic/gin"
	"headscale-panel/config"
	"headscale-panel/model"
	"headscale-panel/repository"
	"strings"
	"time"
)

// Operation log channel
var OperationLogChan = make(chan *model.OperationLog, 30)

func OperationLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start time
		startTime := time.Now()

		// Processing requests
		c.Next()

		// End time
		endTime := time.Now()

		// Execution time consuming
		timeCost := endTime.Sub(startTime).Milliseconds()

		// Get the currently logged in user
		var username string
		ctxUser, exists := c.Get("user")
		if !exists {
			username = "Not logged in"
		}
		user, ok := ctxUser.(model.User)
		if !ok {
			username = "Not logged in"
		}
		username = user.Name

		// Get access path
		path := strings.TrimPrefix(c.FullPath(), "/"+config.Conf.System.UrlPathPrefix)

		// Request method
		method := c.Request.Method
		if path == "/system/status" || (path == "/console/machine" && method == "GET") {
			return
		}

		// Get a description of the interface
		apiRepository := repository.NewApiRepository()
		apiDesc, _ := apiRepository.GetApiDescByPath(path, method)

		operationLog := model.OperationLog{
			Username:   username,
			Ip:         c.ClientIP(),
			IpLocation: "",
			Method:     method,
			Path:       path,
			Desc:       apiDesc,
			Status:     c.Writer.Status(),
			StartTime:  startTime,
			TimeCost:   timeCost,
			//UserAgent:  c.Request.UserAgent(),
		}

		// It is best to send the logs to rabbitmq or kafka
		// Here it is sent to the channel and 3 goroutines are opened for processing
		OperationLogChan <- &operationLog
	}
}
