package InterArrival

import (
	"fmt"
	"math"
	"math/rand"
	"myproject/DBR"
	"myproject/SDBR"
	"myproject/utils"
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

// CompareDBRandSDBR runs the event-driven simulation for both DBR and SDBR using identical arrival parameters.
func CompareDBRandSDBR(N int, totalB float64, alpha float64, fileSizeMB float64, ET0 float64, ET1 float64) {
	fmt.Println("=========================================================================================")
	fmt.Printf("Comparing DBR vs SDBR Timeline: N = %d | Total BW = %.2f | Alpha = %.2f\n", N, totalB, alpha)
	fmt.Println("=========================================================================================")

	var arrivalTimes []float64
	currentTime := 0.0
	for i := 0; i < N; i++ {
		currentTime += utils.InverseCDFExp(rand.Float64(), 1.0/float64(N))
		arrivalTimes = append(arrivalTimes, currentTime)
	}

	var baseUsers []*DynamicUser
	for i := 0; i < N; i++ {
		initialState := strings.ToLower(utils.InitState(ET0, ET1))
		var initialDuration float64
		if initialState == "disconnect" {
			initialDuration = utils.InverseCDFExp(rand.Float64(), ET0)
		} else {
			initialDuration = utils.InverseCDFExp(rand.Float64(), ET1)
		}

		baseUsers = append(baseUsers, &DynamicUser{
			Name:          fmt.Sprintf("U%d", i+1),
			Ti:            arrivalTimes[i],
			RemainingMb:   fileSizeMB * 8,
			State:         initialState,
			DurationLeft:  initialDuration,
			WifiBandwidth: utils.InverseTransformWifiUser(),
		})
	}

	fmt.Println("\n--- RUNNING DBR SIMULATION ---")
	dbrBlocked, dbrAvgTime := simulateInterArrival("DBR", cloneUsers(baseUsers), N, totalB, alpha, ET0, ET1, fileSizeMB)

	fmt.Println("\n--- RUNNING SDBR SIMULATION ---")
	sdbrBlocked, sdbrAvgTime := simulateInterArrival("SDBR", cloneUsers(baseUsers), N, totalB, alpha, ET0, ET1, fileSizeMB)

	dbrProb := float64(dbrBlocked) / float64(N) * 100
	sdbrProb := float64(sdbrBlocked) / float64(N) * 100

	fmt.Println("\n=========================================================================================")
	fmt.Printf("%-20s | %-20s | %-20s\n", "Metric", "Standard DBR", "Stateful SDBR")
	fmt.Println("-----------------------------------------------------------------------------------------")
	fmt.Printf("%-20s | %-20d | %-20d\n", "Total Users", N, N)
	fmt.Printf("%-20s | %-20d | %-20d\n", "Blocked Users", dbrBlocked, sdbrBlocked)
	fmt.Printf("%-20s | %-19.2f%% | %-19.2f%%\n", "Blocking Probability", dbrProb, sdbrProb)
	fmt.Printf("%-20s | %-17.4fs | %-17.4fs\n", "Avg Completion Time", dbrAvgTime, sdbrAvgTime)
	fmt.Println("=========================================================================================")
}

func RunSingleMode(mode string, N int, totalB float64, alpha float64, fileSizeMB float64, ET0 float64, ET1 float64) {

	var arrivalTimes []float64
	currentTime := 0.0
	for i := 0; i < N; i++ {
		currentTime += utils.InverseCDFExp(rand.Float64(), 1.0/float64(N))
		arrivalTimes = append(arrivalTimes, currentTime)
	}

	var baseUsers []*DynamicUser
	for i := 0; i < N; i++ {
		initialState := strings.ToLower(utils.InitState(ET0, ET1))
		var initialDuration float64
		if initialState == "disconnect" {
			initialDuration = utils.InverseCDFExp(rand.Float64(), ET0)
		} else {
			initialDuration = utils.InverseCDFExp(rand.Float64(), ET1)
		}

		baseUsers = append(baseUsers, &DynamicUser{
			Name:          fmt.Sprintf("U%d", i+1),
			Ti:            arrivalTimes[i],
			RemainingMb:   fileSizeMB * 8,
			State:         initialState,
			DurationLeft:  initialDuration,
			WifiBandwidth: utils.InverseTransformWifiUser(),
		})
	}

	blocked, avgTime := simulateInterArrival(mode, baseUsers, N, totalB, alpha, ET0, ET1, fileSizeMB)
	prob := float64(blocked) / float64(N) * 100

	fmt.Println("\n=========================================================================================")
	fmt.Printf("%s SIMULATION COMPLETE\n", mode)
	fmt.Println("-----------------------------------------------------------------------------------------")
	fmt.Printf("Total Users          : %d\n", N)
	fmt.Printf("Successfully Finished: %d\n", N-blocked)
	fmt.Printf("Blocked Users        : %d\n", blocked)
	fmt.Printf("Blocking Probability : %.2f%%\n", prob)
	fmt.Printf("Avg Completion Time  : %.4fs\n", avgTime)
	fmt.Println("=========================================================================================")
}

func simulateInterArrival(mode string, users []*DynamicUser, N int, totalB float64, alpha float64, ET0 float64, ET1 float64, fileSizeMB float64) (int, float64) {
	const minB1 = 40.0
	maxActiveUsers := int(totalB / minB1)

	currentTime := 0.0
	completedUsers := 0
	blockedCount := 0
	arrivalIndex := 0
	var activePool []*DynamicUser
	var totalCompletionTime float64

	for completedUsers < N {
		dt := math.MaxFloat64

		if arrivalIndex < N {
			timeToArrival := users[arrivalIndex].Ti - currentTime
			if timeToArrival > 0 && timeToArrival < dt {
				dt = timeToArrival
			}
		}

		for _, u := range activePool {
			if u.DurationLeft > 0 && u.DurationLeft < dt {
				dt = u.DurationLeft
			}

			currentSpeed := u.TotalBandwidth

			if currentSpeed > 0 {
				timeToFinish := u.RemainingMb / currentSpeed
				if timeToFinish > 0 && timeToFinish < dt {
					dt = timeToFinish
				}
			}
		}

		if dt <= 0 {
			dt = 1e-6
		}

		currentTime += dt

		for _, u := range activePool {
			currentSpeed := u.TotalBandwidth

			if currentSpeed > 0 {
				u.RemainingMb -= currentSpeed * dt
			}
			u.DurationLeft -= dt
		}

		needDbrRecalc := false

		for i := len(activePool) - 1; i >= 0; i-- {
			u := activePool[i]
			if u.RemainingMb <= 1e-6 {
				u.RemainingMb = 0
				u.IsFinished = true
				u.IsActive = false
				u.FinishTime = currentTime

				completionDuration := currentTime - u.Ti
				totalCompletionTime += completionDuration
				completedUsers++
				needDbrRecalc = true

				activePool = append(activePool[:i], activePool[i+1:]...)
			}
		}

		for _, u := range activePool {
			if u.DurationLeft <= 1e-6 {
				if u.State == "connect" {
					u.State = "disconnect"
					u.DurationLeft = utils.InverseCDFExp(rand.Float64(), ET0)
					needDbrRecalc = true
				} else {
					u.State = "connect"
					u.DurationLeft = utils.InverseCDFExp(rand.Float64(), ET1)
					needDbrRecalc = true
				}
			}
		}

		for arrivalIndex < N && users[arrivalIndex].Ti <= currentTime+1e-6 {
			u := users[arrivalIndex]
			isAdmitted := false

			if mode == "DBR" {
				if len(activePool) < maxActiveUsers {
					isAdmitted = true
				}
			} else if mode == "SDBR" {
				sys := SDBR.NewSystem(totalB, minB1)

				for _, au := range activePool {
					b2 := 0.0
					if au.State == "connect" {
						b2 = au.WifiBandwidth
					}
					ue := &SDBR.UserEquipment{Name: au.Name, B2: b2, IsAdmitted: true}
					sys.Users = append(sys.Users, ue)
				}

				for i, ue := range sys.Users {
					if i < sys.NMax {
						ue.B1 = sys.Wg
						if ue.B2 > 0 {
							sys.SDBRBandwidthAssignment(ue, fileSizeMB)
						}
					} else {
						if sys.S >= sys.Wg {
							ue.B1 = sys.Wg
							sys.S -= sys.Wg
							if ue.B2 > 0 {
								sys.SDBRBandwidthAssignment(ue, fileSizeMB)
							}
						} else if sys.S > 0 {
							ue.B1 = sys.S
							sys.S = 0
						}
					}
				}

				u_b2 := 0.0
				if u.State == "connect" {
					u_b2 = u.WifiBandwidth
				}
				isAdmitted = sys.AddUser(&SDBR.UserEquipment{Name: u.Name, B2: u_b2}, fileSizeMB)
			}

			if !isAdmitted {
				u.IsBlocked = true
				u.IsFinished = true
				blockedCount++
				completedUsers++
			} else {
				u.IsActive = true
				activePool = append(activePool, u)
				needDbrRecalc = true
			}
			arrivalIndex++
		}

		if needDbrRecalc {
			currentActiveCount := len(activePool)
			if currentActiveCount > 0 {

				if mode == "DBR" {
					activeB2s := make([]float64, currentActiveCount)
					for i, u := range activePool {
						if u.State == "connect" {
							activeB2s[i] = u.WifiBandwidth
						} else {
							activeB2s[i] = 0.0
						}
					}

					ues := DBR.Run(activeB2s, totalB, alpha)

					for i, u := range activePool {
						u.TotalBandwidth = ues[i].Be
					}

				} else if mode == "SDBR" {
					sys := SDBR.NewSystem(totalB, minB1)

					for _, au := range activePool {
						b2 := 0.0
						if au.State == "connect" {
							b2 = au.WifiBandwidth
						}
						ue := &SDBR.UserEquipment{Name: au.Name, B2: b2, IsAdmitted: true}
						sys.Users = append(sys.Users, ue)
					}

					for i, ue := range sys.Users {
						if i < sys.NMax {
							ue.B1 = sys.Wg
							if ue.B2 > 0 {
								sys.SDBRBandwidthAssignment(ue, fileSizeMB)
							}
						} else {
							if sys.S >= sys.Wg {
								ue.B1 = sys.Wg
								sys.S -= sys.Wg
								if ue.B2 > 0 {
									sys.SDBRBandwidthAssignment(ue, fileSizeMB)
								}
							} else if sys.S > 0 {
								ue.B1 = sys.S
								sys.S = 0
							} else {
								ue.B1 = 0.0
							}
						}
					}

					for i, u := range activePool {
						u.TotalBandwidth = sys.Users[i].B1 + sys.Users[i].B2
					}
				}
			}
		}
	}

	avgTime := 0.0
	successfulUsers := N - blockedCount
	if successfulUsers > 0 {
		avgTime = totalCompletionTime / float64(successfulUsers)
	}

	return blockedCount, avgTime
}

func RunMultipleIterations(mode string, iterations int, N int, totalB float64, alpha float64, fileSizeMB float64, ET0 float64, ET1 float64) {
	totalBlocked := 0
	totalAvgTime := 0.0

	for iter := 0; iter < iterations; iter++ {

		var arrivalTimes []float64
		currentTime := 0.0
		for i := 0; i < N; i++ {
			currentTime += utils.InverseCDFExp(rand.Float64(), 1.0/float64(N))
			arrivalTimes = append(arrivalTimes, currentTime)
		}

		var baseUsers []*DynamicUser
		for i := 0; i < N; i++ {
			initialState := strings.ToLower(utils.InitState(ET0, ET1))
			var initialDuration float64
			if initialState == "disconnect" {
				initialDuration = utils.InverseCDFExp(rand.Float64(), ET0)
			} else {
				initialDuration = utils.InverseCDFExp(rand.Float64(), ET1)
			}

			baseUsers = append(baseUsers, &DynamicUser{
				Name:          fmt.Sprintf("U%d", i+1),
				Ti:            arrivalTimes[i],
				RemainingMb:   fileSizeMB * 8,
				State:         initialState,
				DurationLeft:  initialDuration,
				WifiBandwidth: utils.InverseTransformWifiUser(),
			})
		}

		blocked, avgTime := simulateInterArrival(mode, baseUsers, N, totalB, alpha, ET0, ET1, fileSizeMB)
		totalBlocked += blocked
		totalAvgTime += avgTime
	}

	avgBlockedUsers := float64(totalBlocked) / float64(iterations)
	blockingProbability := (avgBlockedUsers / float64(N)) * 100
	avgTimeConsumption := totalAvgTime / float64(iterations)

	fmt.Println("\n=========================================================================================")
	fmt.Printf("%s SIMULATION COMPLETE (%d ITERATIONS)\n", strings.ToUpper(mode), iterations)
	fmt.Println("-----------------------------------------------------------------------------------------")
	fmt.Printf("Total Users (N)      : %d\n", N)
	fmt.Printf("Avg Blocked Users    : %.2f\n", avgBlockedUsers)
	fmt.Printf("Blocking Probability : %.2f%%\n", blockingProbability)
	fmt.Printf("Avg Time Consumption : %.4fs\n", avgTimeConsumption)
	fmt.Println("=========================================================================================")
}

func cloneUsers(base []*DynamicUser) []*DynamicUser {
	cloned := make([]*DynamicUser, len(base))
	for i, u := range base {
		c := *u
		cloned[i] = &c
	}
	return cloned
}
