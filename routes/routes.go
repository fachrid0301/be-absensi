package routes

import (
	"be-absensi/controllers"
	"be-absensi/middleware"
	"be-absensi/utils"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	uploadPath := utils.UploadBasePath()
	r.Static("/uploads", uploadPath)

	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)
	r.GET("/divisi", controllers.ListDivisi)

	auth := r.Group("")
	auth.Use(middleware.JWTAuth())
	{
		auth.GET("/profile", controllers.Profile)
		auth.GET("/jadwal", controllers.GetJadwal)
		auth.GET("/log-aktivitas", controllers.ListLogAktivitas)

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
			absensi.PUT("/:id/status", middleware.AdminOnly(), controllers.UpdateAbsensiStatus)
		}

		admin := auth.Group("/admin")
		admin.Use(middleware.AdminOnly())
		{
			admin.GET("/dashboard", controllers.AdminDashboard)
			admin.PUT("/jadwal", controllers.UpdateJadwal)
		}

		pendaftaran := auth.Group("/pendaftaran")
		{
			pendaftaran.POST("", middleware.PesertaOnly(), controllers.CreatePendaftaran)
			pendaftaran.GET("", controllers.ListPendaftaran)
			pendaftaran.PUT("/:id/verifikasi", middleware.AdminOnly(), controllers.VerifikasiPendaftaran)
		}

		sertifikat := auth.Group("/sertifikat")
		{
			sertifikat.POST("/request", middleware.PesertaOnly(), controllers.RequestSertifikat)
			sertifikat.GET("", controllers.ListSertifikat)
			sertifikat.GET("/:id", controllers.GetSertifikat)
			sertifikat.POST("/:id/verifikasi", middleware.AdminOnly(), controllers.VerifikasiSertifikat)
		}
	}
}
