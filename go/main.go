package main

import (
	// "myproject/DOWNLOAD"
	"myproject/InterArrival"
)

func main() {
	// DOWNLOAD.SimulateDBRDownloadTime(10, 1000, 500, 100.0, 50.0)
	InterArrival.InterArrivalTimeline(10, 0.50, 500.0, 50.0, 50.0)
}
