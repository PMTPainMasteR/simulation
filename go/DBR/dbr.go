package DBR

import (
	"fmt"
)

type UserEquipment struct {
	Name string
	B1   float64
	b1   float64
	B2   float64
	Be   float64 // Simplified to a single float64 for the specific alpha
}

// Run distributes the bandwidth dynamically based on the actual active Wi-Fi states and a specific alpha.
func Run(activeB2s []float64, B float64, alpha float64) []*UserEquipment {
	numUEs := len(activeB2s)
	N := float64(numUEs)

	// Safety check for empty slices
	if N == 0 {
		return nil
	}

	b1_val := B / N

	var ues []*UserEquipment
	for i := 0; i < numUEs; i++ {
		ue := &UserEquipment{
			Name: fmt.Sprintf("UE%d", i+1),
			B1:   b1_val,
			b1:   b1_val,
			B2:   activeB2s[i],
		}
		ues = append(ues, ue)
	}

	w := b1_val / B
	S := 0.00

	for _, ue := range ues {
		ue.b1 = ue.B1
	}

	// Reclaim bandwidth based on the fed alpha
	for _, ue := range ues {
		if ue.b1 <= ue.B2 {
			ue.b1 = ue.B1 * (1 - alpha)
			S += (ue.B1 * alpha)
		} else {
			ue.b1 = ue.B1 - (ue.B2 * alpha)
			S += (ue.B2 * alpha)
		}
	}

	// Redistribute the shared pool
	for _, ue := range ues {
		ue.Be = ue.b1 + (S * w) + ue.B2
	}

	return ues
}
