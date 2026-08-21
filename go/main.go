package main

import (
	"myproject/download"
)

func main() {
	download.SimulateDBRDownloadTime(10, 100000, 500, 50.0, 50.0)
}
