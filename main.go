package main

import (
	"log"
	"os"

	"be-absensi/backend/config"
	"be-absensi/backend/routes"
	"be-absensi/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

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

	config.ConnectDatabase()
	utils.InitJWT()

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
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
