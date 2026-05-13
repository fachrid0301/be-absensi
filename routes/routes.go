package routes

import (
	"be-absensi/backend/controllers"
	"be-absensi/backend/middleware"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)

	protected := r.Group("")
	protected.Use(middleware.JWTAuth())
	protected.GET("/profile", controllers.Profile)
}
