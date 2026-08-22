package DOWNLOAD

import (
	"fmt"
	"math/rand"
	"myproject/DBR"
	"myproject/utils"
	"strings"
)

func SimulateSingleLinkDownloadTime(fileSizeMB float64, totalBandwidth float64, wifiBandwidth float64, ET0 float64, ET1 float64) float64 {
	connectBandwidth := totalBandwidth
	disconnectBandwidth := totalBandwidth - wifiBandwidth

	if connectBandwidth <= 0 {
		return -1
	}

	// Initial Value
	remainingSize := fileSizeMB * 8 // Convert MB to Megabits (Mb)
	totalTime := 0.0
	currentState := strings.ToLower(utils.InitState(ET0, ET1))
	var duration float64

	if currentState == "disconnect" {
		currentState = "Disconnect"
		duration = utils.InverseCDFExp(rand.Float64(), ET0)
	} else {
		currentState = "Connect"
		duration = utils.InverseCDFExp(rand.Float64(), ET1)
	}

	for remainingSize > 0 {
		if currentState == "Connect" {
			timeToFinish := remainingSize / connectBandwidth

			if timeToFinish <= duration {
				totalTime += timeToFinish
				remainingSize = 0
				break
			}
			totalTime += duration
			remainingSize -= connectBandwidth * duration
			currentState = "Disconnect"
			duration = utils.InverseCDFExp(rand.Float64(), ET0)

		} else if currentState == "Disconnect" {
			if disconnectBandwidth > 0 {
				timeToFinish := remainingSize / disconnectBandwidth

				if timeToFinish <= duration {
					totalTime += timeToFinish
					remainingSize = 0
					break
				}
				remainingSize -= disconnectBandwidth * duration
			}

			totalTime += duration
			currentState = "Connect"
			duration = utils.InverseCDFExp(rand.Float64(), ET1)
		}
	}

	return totalTime
}

func SimulateDBRDownloadTime(maxUEs int, iteration int, fileSizeMB float64, ET0 float64, ET1 float64) {
	alphas := []float64{0.00, 0.25, 0.50, 0.75, 1.00}

	fmt.Println("=========================================================================================")
	fmt.Printf("Simulation Settings: Filesize = %.2f MB | ET0 = %.2fs | ET1 = %.2fs | Iterations = %d\n",
		fileSizeMB, ET0, ET1, iteration)
	fmt.Println("=========================================================================================")

	fmt.Printf("%-10s", "Group Size")
	for _, a := range alphas {
		fmt.Printf(" | Alpha: %-4.2f (Avg Sec)", a)
	}
	fmt.Println()
	fmt.Println("-----------------------------------------------------------------------------------------")

	for numUEs := 1; numUEs <= maxUEs; numUEs++ {
		totalTimePerAlpha := make(map[float64]float64)
		totalUserCountPerAlpha := make(map[float64]int)

		for i := 0; i < iteration; i++ {
			alphaList, ues := DBR.Run(numUEs)

			for _, alpha := range alphaList {
				for _, ue := range ues {
					totalBandwidth := ue.Be[alpha]
					wifiBandwidth := ue.B2
					ts := SimulateSingleLinkDownloadTime(fileSizeMB, totalBandwidth, wifiBandwidth, ET0, ET1)

					totalTimePerAlpha[alpha] += ts
					totalUserCountPerAlpha[alpha]++
				}
			}
		}

		fmt.Printf("N = %-7d", numUEs)
		for _, alpha := range alphas {
			totalTime := totalTimePerAlpha[alpha]
			totalCount := totalUserCountPerAlpha[alpha]

			avgTime := 0.0
			if totalCount > 0 {
				avgTime = totalTime / float64(totalCount)
			}
			fmt.Printf(" | %-20.2f", avgTime)
		}
		fmt.Println()
	}
	fmt.Println("=========================================================================================")
}
