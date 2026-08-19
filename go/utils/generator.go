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
