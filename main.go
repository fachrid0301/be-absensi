package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"be-absensi/config"
	"be-absensi/controllers"
	"be-absensi/models"
	"be-absensi/routes"
	"be-absensi/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowedOrigin := "*"
		if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "https://localhost:") {
			allowedOrigin = origin
		}

		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func main() {
	loaded := false
	for _, p := range []string{".env", "backend/.env"} {
		if err := godotenv.Load(p); err == nil {
			loaded = true
			break
		}
	}
	if !loaded {
		log.Println("peringatan: .env tidak ditemukan, gunakan variabel lingkungan sistem")
	}

	if os.Getenv("UPLOAD_PATH") == "" {
		if _, err := os.Stat("backend/uploads"); err == nil {
			_ = os.Setenv("UPLOAD_PATH", "backend/uploads")
		}
	}

	config.ConnectDatabase()
	if err := config.DB.AutoMigrate(
		&models.Jadwal{},
		&models.Peserta{},
		&models.Pendaftaran{},
		&models.Sertifikat{},
		&models.LogAktivitas{},
	); err != nil {
		log.Printf("gagal auto migrate: %v", err)
	}
	// Seed default jadwal jika belum ada
	var count int64
	if err := config.DB.Model(&models.Jadwal{}).Count(&count).Error; err == nil && count == 0 {
		defaultJadwal := models.Jadwal{
			ID:        1,
			JamMasuk:  "08:00",
			JamPulang: "17:00",
		}
		if err := config.DB.Create(&defaultJadwal).Error; err != nil {
			log.Printf("gagal seed default jadwal: %v", err)
		} else {
			log.Println("default jadwal seeded")
		}
	}
	utils.InitJWT()
	utils.InitUploadDir()

	// Jalankan scheduler otomatis "tidak hadir"
	// Setiap menit dicek apakah jam pulang sudah lewat;
	// jika iya, peserta yang belum absen hari ini ditandai "tidak hadir".
	utils.RunAbsenceScheduler(
		controllers.MarkAbsentPeserta,
		func() (string, string, error) {
			var jadwal models.Jadwal
			if err := config.DB.First(&jadwal).Error; err != nil {
				return "08:00", "17:00", err
			}
			return jadwal.JamMasuk, jadwal.JamPulang, nil
		},
	)

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(corsMiddleware())
	routes.Register(r)

	addr := os.Getenv("PORT")
	if addr == "" {
		addr = "8080"
	}
	if addr[0] != ':' {
		addr = ":" + addr
	}

	log.Printf("server berjalan di http://localhost%s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
