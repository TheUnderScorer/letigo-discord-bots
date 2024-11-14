package server

import (
	"app/domain/interaction"
	"app/server/responses"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Ctx = gin.Context

func CreateRouter(ctx context.Context, gin *gin.Engine, version string) {
	gin.POST("/interactions/register", func(c *Ctx) {
		go interaction.Init(ctx)

		c.Status(http.StatusAccepted)
	})

	gin.GET("/", func(c *Ctx) {
		c.JSON(http.StatusOK, &responses.VersionInfo{
			Version: version,
			Result:  true,
		})
	})

}
