package interarrival

import (
	"fmt"
	"math"
	"math/rand"
	"myproject/DBR"
	"myproject/utils"
	"sort"
	"strings"
)

type DynamicUser struct {
	Name           string
	Ti             float64
	RemainingMb    float64
	State          string
	DurationLeft   float64
	TotalBandwidth float64
	WifiBandwidth  float64
	FinishTime     float64
	IsActive       bool
	IsFinished     bool
	IsBlocked      bool
}

// ObserveDynamicInterArrival tracks dynamic bandwidth assignment instantly using Discrete Event Simulation
func ObserveDynamicInterArrival(N int, totalB float64, alpha float64, fileSizeMB float64, ET0 float64, ET1 float64) {
	fmt.Println("=========================================================================================")
	fmt.Printf("Event-Driven Inter-Arrival Timeline: N = %d | Total BW = %.2f | Alpha = %.2f\n", N, totalB, alpha)
	fmt.Println("=========================================================================================")

	const minB1 = 5.0
	maxActiveUsers := int(totalB / minB1)

	// 1. Generate and sort arrival times FIRST
	var arrivalTimes []float64
	for i := 0; i < N; i++ {
		t_i := utils.InverseCDFExpLambda(rand.Float64(), float64(N))
		arrivalTimes = append(arrivalTimes, t_i)
	}
	sort.Float64s(arrivalTimes)

	// 2. Initialize Users
	var users []*DynamicUser
	for i := 0; i < N; i++ {
		initialState := strings.ToLower(utils.InitState(ET0, ET1))
		var initialDuration float64
		if initialState == "disconnect" {
			initialDuration = utils.InverseCDFExp(rand.Float64(), ET0)
		} else {
			initialDuration = utils.InverseCDFExp(rand.Float64(), ET1)
		}

		users = append(users, &DynamicUser{
			Name:         fmt.Sprintf("U%d", i+1),
			Ti:           arrivalTimes[i],
			RemainingMb:  fileSizeMB * 8,
			State:        initialState,
			DurationLeft: initialDuration,
		})
	}

	// 3. Event-Driven Simulation tracking variables
	currentTime := 0.0
	completedUsers := 0
	blockedCount := 0
	arrivalIndex := 0
	var activePool []*DynamicUser

	// 4. Run until all N users have finished or been blocked
	for completedUsers < N {
		// --- STEP A: Calculate the time jump (dt) to the very next event ---
		dt := math.MaxFloat64

		// Event 1: Next user arrival
		if arrivalIndex < N {
			timeToArrival := users[arrivalIndex].Ti - currentTime
			if timeToArrival > 0 && timeToArrival < dt {
				dt = timeToArrival
			}
		}

		// Event 2 & 3: Next State Change or Next Download Finish for active users
		for _, u := range activePool {
			if u.DurationLeft > 0 && u.DurationLeft < dt {
				dt = u.DurationLeft
			}

			currentSpeed := u.TotalBandwidth
			if u.State == "disconnect" {
				currentSpeed = u.TotalBandwidth - u.WifiBandwidth
				if currentSpeed < 0 {
					currentSpeed = 0
				}
			}

			if currentSpeed > 0 {
				timeToFinish := u.RemainingMb / currentSpeed
				if timeToFinish > 0 && timeToFinish < dt {
					dt = timeToFinish
				}
			}
		}

		// Prevent floating-point stalls
		if dt <= 0 {
			dt = 1e-6
		}

		// --- STEP B: Fast-forward time and update progress ---
		currentTime += dt

		for _, u := range activePool {
			currentSpeed := u.TotalBandwidth
			if u.State == "disconnect" {
				currentSpeed = u.TotalBandwidth - u.WifiBandwidth
				if currentSpeed < 0 {
					currentSpeed = 0
				}
			}

			if currentSpeed > 0 {
				u.RemainingMb -= currentSpeed * dt
			}
			u.DurationLeft -= dt
		}

		// --- STEP C: Process events triggered at this new exact time ---
		needDbrRecalc := false

		// 1. Check for finished downloads (iterate backwards to safely remove items from slice)
		for i := len(activePool) - 1; i >= 0; i-- {
			u := activePool[i]
			if u.RemainingMb <= 1e-6 { // floating point tolerance
				u.RemainingMb = 0
				u.IsFinished = true
				u.IsActive = false
				u.FinishTime = currentTime
				completedUsers++
				needDbrRecalc = true

				fmt.Printf("\n[%.4fs] %s FINISHED. (Session Duration: %.4fs)\n", currentTime, u.Name, currentTime-u.Ti)

				// Remove user from active pool
				activePool = append(activePool[:i], activePool[i+1:]...)
			}
		}

		// 2. Check for state changes (Connect/Disconnect toggle)
		for _, u := range activePool {
			if u.DurationLeft <= 1e-6 {
				if u.State == "connect" {
					u.State = "disconnect"
					u.DurationLeft = utils.InverseCDFExp(rand.Float64(), ET0)
				} else {
					u.State = "connect"
					u.DurationLeft = utils.InverseCDFExp(rand.Float64(), ET1)
				}
			}
		}

		// 3. Check for new arrivals
		for arrivalIndex < N && users[arrivalIndex].Ti <= currentTime+1e-6 {
			u := users[arrivalIndex]
			if len(activePool) >= maxActiveUsers {
				// Block the user
				u.IsBlocked = true
				u.IsFinished = true
				blockedCount++
				completedUsers++
				fmt.Printf("\n[%.4fs] %s ARRIVED but was BLOCKED (System at max capacity of %d).\n", currentTime, u.Name, maxActiveUsers)
			} else {
				// Admit the user
				u.IsActive = true
				activePool = append(activePool, u)
				needDbrRecalc = true
				fmt.Printf("\n[%.4fs] %s ARRIVED.\n", currentTime, u.Name)
			}
			arrivalIndex++
		}

		// --- STEP D: Recalculate DBR if the active user count changed ---
		if needDbrRecalc {
			currentActiveCount := len(activePool)
			if currentActiveCount > 0 {
				_, ues := DBR.Run(currentActiveCount, totalB)

				fmt.Printf("[%.4fs] DBR Recalculated for %d active users (B1 = %.2f):\n", currentTime, currentActiveCount, totalB/float64(currentActiveCount))
				for i, u := range activePool {
					u.TotalBandwidth = ues[i].Be[alpha]
					u.WifiBandwidth = ues[i].B2
					fmt.Printf("   -> %s allocated Total BW: %.2f | WiFi BW: %.2f\n", u.Name, u.TotalBandwidth, u.WifiBandwidth)
				}
			}
		}
	}

	// 5. Calculate and print Blocking Probability
	blockingProb := float64(blockedCount) / float64(N)
	fmt.Println("=========================================================================================")
	fmt.Printf("SIMULATION COMPLETE\n")
	fmt.Printf("Total Users Processed  : %d\n", N)
	fmt.Printf("Successfully Completed : %d\n", N-blockedCount)
	fmt.Printf("Blocked Users          : %d\n", blockedCount)
	fmt.Printf("Blocking Probability   : %.4f (%.2f%%)\n", blockingProb, blockingProb*100)
	fmt.Println("=========================================================================================")
}
