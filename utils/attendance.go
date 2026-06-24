package utils

import "time"

// TodayDate tanggal hari ini tanpa komponen jam (untuk kolom type date).
func TodayDate() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
