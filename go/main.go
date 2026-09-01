package main

import (
	"math/rand"
	// "myproject/DOWNLOAD"
	"myproject/interarrival"
	"time"
)

func main() {
	// Seed the global random number generator to ensure unique results per run
	rand.Seed(time.Now().UnixNano())

	// Run static multi-iteration simulation
	// DOWNLOAD.SimulateDBRDownloadTime(10, 1000, 80.0, 100.0, 1.0, 2.0)

	interarrival.ObserveDynamicInterArrival(10, 40.0, 0.50, 500.0, 50.0, 50.0)
}
