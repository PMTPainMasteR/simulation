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
	rand.Seed(time.Now().UnixNano())

	// Define global simulation parameters
	totalB := 300.0
	fileSizeMB := 187.5
	ET0 := 439.0
	ET1 := 1679.0

	// Define command-line flags
	modeFlag := flag.String("mode", "SDBR", "Simulation mode: SDBR, DBR, Static, or Compare")
	nFlag := flag.Int("n", 10, "Number of users")
	iterFlag := flag.Int("i", 10, "Number of iterations to run")
	alphaFlag := flag.Float64("a", 0.50, "Alpha reallocation rate")
	flag.Parse()

	mode := *modeFlag
	N := *nFlag
	iterations := *iterFlag
	alpha := *alphaFlag

	fmt.Println("=========================================================================================")
	fmt.Printf("INITIALIZING BATCH SIMULATION - MODE: %s | ITERATIONS: %d\n", strings.ToUpper(mode), iterations)
	fmt.Println("=========================================================================================")

	switch strings.ToUpper(mode) {
	case "DBR":
		InterArrival.RunMultipleIterations("DBR", iterations, N, totalB, alpha, fileSizeMB, ET0, ET1)
	case "SDBR":
		InterArrival.RunMultipleIterations("SDBR", iterations, N, totalB, alpha, fileSizeMB, ET0, ET1)
	case "COMPARE":
		InterArrival.CompareMultipleIterations(iterations, N, totalB, alpha, fileSizeMB, ET0, ET1)
	default:
		fmt.Printf("Batch mode currently configured for DBR, SDBR, and COMPARE evaluations.\n")
	}
}
