package server

import (
	"github.com/gin-gonic/gin"
	"lib/server/responses"
	"net/http"
)

type Ctx = gin.Context

func CreateRouter(gin *gin.Engine, version string) {
	gin.GET("/", func(c *Ctx) {
		c.JSON(http.StatusOK, &responses.VersionInfo{
			Version: version,
			Result:  true,
		})
	})

}
