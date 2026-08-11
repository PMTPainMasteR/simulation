package DBR

import "fmt"

type UserEquipment struct {
	Name string
	B1   float64
	b1   float64
	b2   float64
	be   map[float64]float64
}

func Run() {
	B := 80.00
	N := 4.00
	wifi := 50.00
	a := []float64{0.00, 0.25, 0.50, 0.75, 1.00}

	b1_val := B / N

	ue1 := &UserEquipment{Name: "UE1", B1: b1_val, b1: b1_val, b2: 0.00, be: make(map[float64]float64)}
	ue2 := &UserEquipment{Name: "UE2", B1: b1_val, b1: b1_val, b2: 0.00, be: make(map[float64]float64)}
	ue3 := &UserEquipment{Name: "UE3", B1: b1_val, b1: b1_val, b2: wifi, be: make(map[float64]float64)}
	ue4 := &UserEquipment{Name: "UE4", B1: b1_val, b1: b1_val, b2: wifi, be: make(map[float64]float64)}

	ues := []*UserEquipment{ue1, ue2, ue3, ue4}

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

		fmt.Printf("alpha: %.2f | UE1: %.2f | UE2: %.2f | UE3: %.2f | UE4: %.2f |\n",
			alpha, ue1.be[alpha], ue2.be[alpha], ue3.be[alpha], ue4.be[alpha])
	}
}
