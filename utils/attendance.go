package utils

import "time"

const (
	JamMasukHour    = 8
	JamMasukMinute  = 0
)

// AttendanceStatus menentukan status absen masuk: "hadir" (≤ 08:00) atau "telat" (> 08:00).
func AttendanceStatus(now time.Time) string {
	if now.After(jamMasukDeadline(now)) {
		return "telat"
	}
	return "hadir"
}

// AttendanceStatusFromJam memvalidasi jam masuk (kolom time di DB).
func AttendanceStatusFromJam(jam time.Time) string {
	t := timeOnRefDay(jam)
	deadline := time.Date(2000, 1, 1, JamMasukHour, JamMasukMinute, 0, 0, t.Location())
	if t.After(deadline) {
		return "telat"
	}
	return "hadir"
}

func KeteranganStatus(status string) string {
	switch status {
	case "telat":
		return "Terlambat — absen masuk melewati jam 08:00"
	case "hadir":
		return "Hadir tepat waktu — absen masuk pada atau sebelum jam 08:00"
	default:
		return ""
	}
}

func timeOnRefDay(jam time.Time) time.Time {
	return time.Date(2000, 1, 1, jam.Hour(), jam.Minute(), jam.Second(), 0, jam.Location())
}

func jamMasukDeadline(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), JamMasukHour, JamMasukMinute, 0, 0, now.Location())
}

func TodayDate() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func TimeOnly(t time.Time) time.Time {
	return time.Date(1, 1, 1, t.Hour(), t.Minute(), t.Second(), 0, t.Location())
}
