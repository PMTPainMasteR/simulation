package utils

import (
	"math/rand"
)

func InverseTransformWifiUser() float64 {
	u := rand.Float64()
	wifi := 0.0
	if u <= 0.5 {
		wifi = 50.0
		return wifi
	}
	return wifi
}

func InitState(ET0 float64, ET1 float64) float64 {
	u := rand.Float64()
	P0 := ET0 / (ET0 + ET1)

	if u <= P0 {
		return 0
	}
	return 1
}
