package handler

import (
	"log"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func NewReverseProxy(targetHost string) gin.HandlerFunc {
	target, err := url.Parse(targetHost)
	if err != nil {
		log.Fatalf("FATAL: Invalid target URL for reverse proxy: %s. Error: %v", targetHost, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	return func(c *gin.Context) {
		log.Printf("API Gateway: Proxying request for %s to %s", c.Request.URL.Path, targetHost)
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
