package server

import (
	"github.com/olegsu/iris/pkg/logger"

	"github.com/gin-gonic/gin"
)

func StartServer() {
	logger.Get().Info("Starting server", nil)
	r := gin.Default()
	r.Run()
}
