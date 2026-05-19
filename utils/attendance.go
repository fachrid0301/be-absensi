package utils

import "time"

// AttendanceStatus mengembalikan "telat" jika absen setelah jam 08:00, selain itu "hadir".
func AttendanceStatus(now time.Time) string {
	deadline := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
	if now.After(deadline) {
		return "telat"
	}
	return "hadir"
}

func TodayDate() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func TimeOnly(t time.Time) time.Time {
	return time.Date(1, 1, 1, t.Hour(), t.Minute(), t.Second(), 0, t.Location())
}
