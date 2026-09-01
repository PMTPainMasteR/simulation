package DBR

import (
	"fmt"
	"myproject/utils"
)

type UserEquipment struct {
	Name string
	B1   float64
	b1   float64
	B2   float64
	Be   map[float64]float64
}

// Run now accepts total bandwidth (B) as a dynamic parameter
func Run(numUEs int, B float64) ([]float64, []*UserEquipment) {
	N := float64(numUEs)
	a := []float64{0.00, 0.25, 0.50, 0.75, 1.00}

	b1_val := B / N

	// Dynamically generate 'N' number of UserEquipments
	var ues []*UserEquipment
	for i := 1; i <= numUEs; i++ {
		ue := &UserEquipment{
			Name: fmt.Sprintf("UE%d", i),
			B1:   b1_val,
			b1:   b1_val,
			B2:   utils.InverseTransformWifiUser(),
			Be:   make(map[float64]float64),
		}
		ues = append(ues, ue)
	}

	w := b1_val / B

	for k := 0; k < len(a); k++ {
		alpha := a[k]
		S := 0.00

		for _, ue := range ues {
			ue.b1 = ue.B1
		}

		for _, ue := range ues {
			if ue.b1 <= ue.B2 {
				ue.b1 = ue.B1 * (1 - alpha)
				S += (ue.B1 * alpha)
			} else {
				ue.b1 = ue.B1 - (ue.B2 * alpha)
				S += (ue.B2 * alpha)
			}
		}

		for _, ue := range ues {
			ue.Be[alpha] = ue.b1 + (S * w) + ue.B2
		}
	}

	return a, ues
}
