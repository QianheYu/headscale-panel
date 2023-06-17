package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// CORS Cross-Domain Middleware
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin") // Request header
		if origin != "" {
			// Receive origin from the client (important!)
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			// Methods for all cross-domain requests supported by the server
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
			// Allow cross-domain settings can return other subfields and can customize fields
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session")
			// Headers that allow browsers (clients) to parse (important)
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			// Set cache time
			c.Header("Access-Control-Max-Age", "172800")
			// Allow clients to pass checks such as cookies (important)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Allow Type Checking
		if method == http.MethodOptions {
			c.AbortWithStatus(200)
			return
		}

		//defer func() {
		//	if err := recover(); err != nil {
		//		log.Printf("Panic info is: %v", err)
		//	}
		//}()

		c.Next()
	}
}
