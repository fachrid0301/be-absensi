package utils

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

type Claims struct {
	IDUser uint   `json:"id_user"`
	Nama   string `json:"nama"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// InitJWT baca JWT_SECRET dari env (wajib saat startup).
func InitJWT() {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		panic("JWT_SECRET wajib diisi di .env")
	}
	jwtSecret = []byte(s)
}

func expireDuration() time.Duration {
	h := os.Getenv("JWT_EXPIRE_HOURS")
	if h == "" {
		return 24 * time.Hour
	}
	n, err := strconv.Atoi(h)
	if err != nil || n <= 0 {
		return 24 * time.Hour
	}
	return time.Duration(n) * time.Hour
}

// GenerateToken buat JWT berisi id_user, nama, role.
func GenerateToken(idUser uint, nama, role string) (string, error) {
	claims := Claims{
		IDUser: idUser,
		Nama:   nama,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireDuration())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(jwtSecret)
}

// ParseToken decode & validasi JWT, kembalikan claims jika valid.
func ParseToken(tokenString string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("metode penandatanganan tidak valid")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, errors.New("token tidak valid")
	}
	return claims, nil
}
