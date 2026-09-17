// Modify main.go
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"myproject/InterArrival"
	"strings"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano()) // Seed the global random number generator[cite: 4]

	// Define global simulation parameters[cite: 4]
	totalB := 300.0
	fileSizeMB := 187.5
	ET0 := 439.0
	ET1 := 1679.0
	defaultAlpha := 0.50

	// Define command-line flags[cite: 4]
	modeFlag := flag.String("mode", "SDBR", "Simulation mode: SDBR, DBR, Static, or Compare")
	nFlag := flag.Int("n", 10, "Number of users")
	iterFlag := flag.Int("i", 10, "Number of iterations to run")
	flag.Parse()

	mode := *modeFlag
	N := *nFlag
	iterations := *iterFlag

	fmt.Println("=========================================================================================")
	fmt.Printf("INITIALIZING BATCH SIMULATION - MODE: %s | ITERATIONS: %d\n", strings.ToUpper(mode), iterations)
	fmt.Println("=========================================================================================")

	switch strings.ToUpper(mode) {
	case "DBR":
		InterArrival.RunMultipleIterations("DBR", iterations, N, totalB, defaultAlpha, fileSizeMB, ET0, ET1)
	case "SDBR":
		InterArrival.RunMultipleIterations("SDBR", iterations, N, totalB, defaultAlpha, fileSizeMB, ET0, ET1)
	default:
		fmt.Printf("Batch mode currently configured for DBR and SDBR evaluations.\n")
	}
}
