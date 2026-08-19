package DBR

import (
	"fmt"
	"myproject/utils" // Ensure this matches your go.mod as discussed previously
)

type UserEquipment struct {
	Name string
	B1   float64
	b1   float64
	b2   float64
	be   map[float64]float64
}

// Run now accepts numUEs (N) as an integer parameter
func Run(numUEs int) {
	B := 80.00
	N := float64(numUEs) // Convert to float64 for calculations
	a := []float64{0.00, 0.25, 0.50, 0.75, 1.00}

	b1_val := B / N

	// Dynamically generate 'N' number of UserEquipments
	var ues []*UserEquipment
	for i := 1; i <= numUEs; i++ {
		ue := &UserEquipment{
			Name: fmt.Sprintf("UE%d", i),
			B1:   b1_val,
			b1:   b1_val,
			b2:   utils.InverseTransformWifiUser(),
			be:   make(map[float64]float64),
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
			if ue.b1 <= ue.b2 {
				ue.b1 = ue.B1 * (1 - alpha)
				S += (ue.B1 * alpha)
			} else {
				ue.b1 = ue.B1 - (ue.b2 * alpha)
				S += (ue.b2 * alpha)
			}
		}

		for _, ue := range ues {
			ue.be[alpha] = ue.b1 + (S * w) + ue.b2
		}

		// Dynamically print the results for all UEs
		fmt.Printf("alpha: %.2f | ", alpha)
		for _, ue := range ues {
			fmt.Printf("%s: %.2f | ", ue.Name, ue.be[alpha])
		}
		fmt.Println() // Print a newline at the end of the alpha iteration
	}
}
