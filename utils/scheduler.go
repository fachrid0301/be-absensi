package utils

import (
	"fmt"
	"log"
	"time"
)

// RunAbsenceScheduler menjalankan goroutine yang setiap menit memeriksa apakah
// jam pulang sudah lewat, lalu memanggil MarkAbsentPeserta sekali per hari.
// Parameter getJadwal adalah fungsi untuk membaca jadwal dari DB (menghindari
// import siklik antara utils dan config/models).
func RunAbsenceScheduler(markFn func(tanggal time.Time, jamPulang string) (int, error), getJadwal func() (string, string, error)) {
	go func() {
		markedDate := ""

		for {
			now := time.Now()

			jamMasuk, jamPulang, err := getJadwal()
			if err != nil {
				// DB belum siap, coba lagi 1 menit kemudian
				log.Printf("[scheduler] gagal membaca jadwal: %v", err)
				time.Sleep(1 * time.Minute)
				continue
			}
			_ = jamMasuk

			// Parse jam pulang
			var pHour, pMinute int
			if _, scanErr := fmt.Sscanf(jamPulang, "%d:%d", &pHour, &pMinute); scanErr != nil {
				pHour = 17
				pMinute = 0
			}

			pulangTime := time.Date(now.Year(), now.Month(), now.Day(), pHour, pMinute, 0, 0, now.Location())
			todayStr := now.Format("2006-01-02")

			// Tandai absen hanya sekali per hari, setelah jam pulang lewat
			if now.After(pulangTime) && markedDate != todayStr {
				tanggal := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				count, markErr := markFn(tanggal, jamPulang)
				if markErr != nil {
					log.Printf("[scheduler] gagal menandai tidak hadir: %v", markErr)
				} else {
					if count > 0 {
						log.Printf("[scheduler] %d peserta ditandai 'tidak hadir' untuk tanggal %s", count, todayStr)
					}
					markedDate = todayStr
				}
			}

			time.Sleep(1 * time.Minute)
		}
	}()
}
