package SDBR

import (
	"math"
)

// UserEquipment holds the state for a single user
type UserEquipment struct {
	Name       string
	B1         float64 // Current Cellular Bandwidth
	B2         float64 // Current WiFi Bandwidth
	IsAdmitted bool
}

// System manages the base station's shared bandwidth pool (S) and connected users
type System struct {
	W     float64 // Total Bandwidth
	Wg    float64 // Guaranteed Bandwidth (w_g)
	NMax  int     // Threshold for standard admission vs shared-pool admission
	S     float64 // Shared bandwidth pool for reallocation
	Users []*UserEquipment
}

// NewSystem initializes the SDBR environment.
func NewSystem(totalB float64, wg float64) *System {
	return &System{
		W:    totalB,
		Wg:   wg,
		NMax: int(totalB / wg),
		S:    0.0,
	}
}

// AddUser represents Algorithm 1: SDBR Algorithm for admission control
func (sys *System) AddUser(u_new *UserEquipment, fileSize float64) bool {
	n := len(sys.Users)

	// {Collect bandwidth when initialize}
	if n <= sys.NMax {
		for _, u_i := range sys.Users {
			if u_i.B2 > 0 {
				sys.SDBRBandwidthAssignment(u_i, fileSize)
			}
		}
	}

	// {Distribute bandwidth & Check Admission}
	if n < sys.NMax {
		// Standard admission: we haven't reached the theoretical limit yet
		sys.Users = append(sys.Users, u_new)
		u_new.B1 = sys.Wg
		u_new.IsAdmitted = true

		if u_new.B2 > 0 {
			sys.SDBRBandwidthAssignment(u_new, fileSize)
		}
		return true
	} else {
		// Shared-pool admission: NMax reached, now we rely on the reclaimed pool S
		if sys.S >= sys.Wg {
			u_new.IsAdmitted = true
			sys.Users = append(sys.Users, u_new)
			u_new.B1 = sys.Wg
			sys.S = sys.S - sys.Wg

			if u_new.B2 > 0 {
				sys.SDBRBandwidthAssignment(u_new, fileSize)
			}
			return true
		} else if sys.S > 0 && sys.S < sys.Wg {
			u_new.IsAdmitted = true
			sys.Users = append(sys.Users, u_new)
			u_new.B1 = sys.S
			sys.S = 0
			return true
		} else {
			// sys.S == 0, completely out of bandwidth
			u_new.IsAdmitted = false
			return false // Block u_new
		}
	}
}

// SDBRBandwidthAssignment represents Algorithm 2: Reclaiming bandwidth
func (sys *System) SDBRBandwidthAssignment(u *UserEquipment, fileSize float64) {
	alpha_i := FindReclaimRatio(fileSize, sys.Wg, u.B2)
	var a float64

	if u.B2 > 0 {
		r := math.Min(sys.Wg, u.B2) * alpha_i
		a = sys.Wg - r
	} else {
		a = sys.Wg
	}

	// {Bandwidth Update}
	if u.B1 >= a {
		sys.S = sys.S + (u.B1 - a)
		u.B1 = a
	} else {
		if sys.S > (a - u.B1) {
			sys.S = sys.S - (a - u.B1)
			u.B1 = a
		}
	}
}

// FindReclaimRatio represents Algorithm 3: Maximizing satisfaction
func FindReclaimRatio(fileSize float64, w_g float64, b2 float64) float64 {
	t_l := fileSize / w_g
	G := Satisfaction(t_l, 500.0, 1.0, 1.2)

	// Evaluate baseline satisfaction before reallocation
	t_l1 := fileSize / (w_g + b2)
	t_l2 := fileSize / w_g
	G_prime := Satisfaction(t_l1, 500.0, 1.0, 1.2) + Satisfaction(t_l2, 500.0, 1.0, 1.2)

	if G > G_prime {
		return 0.0 // Reclaiming Bandwidth is disabled
	}

	bestAlpha := 0.0
	maxSat := -math.MaxFloat64

	for i := 1; i <= 100; i++ {
		alpha := float64(i) / 100.0

		// FIXED: Replaced raw alpha with proper reclaimed bandwidth amount calculation
		reclaimedAmount := math.Min(w_g, b2) * alpha
		lat_1 := fileSize / (w_g*(1-alpha) + b2)
		lat_2 := fileSize / (w_g + reclaimedAmount)

		sat := Satisfaction(lat_1, 500.0, 1.0, 1.2) + Satisfaction(lat_2, 500.0, 1.0, 1.2)

		if sat > maxSat {
			maxSat = sat
			bestAlpha = alpha
		}
	}

	return bestAlpha
}

// Satisfaction maps transmission latency to price
func Satisfaction(t_l float64, t_d float64, P_max float64, b float64) float64 {
	a := P_max / math.Pow(t_d, b)
	if t_l < t_d {
		return P_max - (a * math.Pow(t_l, b))
	}
	return 0.0
}
