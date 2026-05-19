package routes

import (
	"be-absensi/backend/controllers"
	"be-absensi/backend/middleware"
	"be-absensi/backend/utils"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	uploadPath := utils.UploadBasePath()
	r.Static("/uploads", uploadPath)

	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)

	auth := r.Group("")
	auth.Use(middleware.JWTAuth())
	{
		auth.GET("/profile", controllers.Profile)

		peserta := auth.Group("/peserta")
		{
			peserta.GET("", controllers.ListPeserta)
			peserta.GET("/:id", controllers.GetPeserta)
			peserta.POST("", controllers.CreatePeserta)
			peserta.PUT("/:id", controllers.UpdatePeserta)
			peserta.DELETE("/:id", middleware.AdminOnly(), controllers.DeletePeserta)
		}

		absensi := auth.Group("/absensi")
		{
			absensi.POST("/masuk", middleware.PesertaOnly(), controllers.AbsensiMasuk)
			absensi.POST("/pulang", middleware.PesertaOnly(), controllers.AbsensiPulang)
			absensi.GET("/history", controllers.AbsensiHistory)
		}

		admin := auth.Group("/admin")
		admin.Use(middleware.AdminOnly())
		{
			admin.GET("/dashboard", controllers.AdminDashboard)
		}

		pendaftaran := auth.Group("/pendaftaran")
		{
			pendaftaran.POST("", middleware.PesertaOnly(), controllers.CreatePendaftaran)
			pendaftaran.GET("", controllers.ListPendaftaran)
			pendaftaran.PUT("/:id/verifikasi", middleware.AdminOnly(), controllers.VerifikasiPendaftaran)
		}
	}
}
