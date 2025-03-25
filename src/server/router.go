package server

import (
	"app/bots"
	"app/domain/interaction"
	"app/server/responses"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Ctx = gin.Context

type RouterContainer struct {
	Bots []*bots.Bot
}

func CreateRouter(container *RouterContainer, gin *gin.Engine, version string) {
	gin.POST("/interactions/register", func(c *Ctx) {
		go interaction.Init(container.Bots)

		c.Status(http.StatusAccepted)
	})

	gin.GET("/", func(c *Ctx) {
		c.JSON(http.StatusOK, &responses.VersionInfo{
			Version: version,
			Result:  true,
		})
	})

}
