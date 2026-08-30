package InterArrival

import (
	"fmt"
	"math/rand"
	"myproject/DBR"
	"myproject/DOWNLOAD"
	"myproject/utils"
	"sort"
)

// Event helps track chronological actions (Arrivals and Finishes)
type Event struct {
	Time    float64
	User    string
	Type    string
	Details string
}

func InterArrivalTimeline(N int, alpha float64, fileSizeMB float64, ET0 float64, ET1 float64) {
	fmt.Println("=========================================================================================")
	fmt.Printf("Inter-Arrival Timeline: N = %d | Alpha = %.2f\n", N, alpha)
	fmt.Println("=========================================================================================")

	_, ues := DBR.Run(N)

	var events []Event

	for i, ue := range ues {
		userName := fmt.Sprintf("U%d", i+1)

		totalBandwidth := ue.Be[alpha]
		wifiBandwidth := ue.B2

		t_L := DOWNLOAD.SimulateSingleLinkDownloadTime(fileSizeMB, totalBandwidth, wifiBandwidth, ET0, ET1)
		t_i := utils.InverseCDFExp(rand.Float64(), float64(N))
		endTime := t_i + t_L

		events = append(events, Event{
			Time:    t_i,
			User:    userName,
			Type:    "ARRIVED",
			Details: fmt.Sprintf("Requires %.2fs to download", t_L),
		})

		// Create Finish Event
		events = append(events, Event{
			Time:    endTime,
			User:    userName,
			Type:    "FINISHED",
			Details: "Session complete",
		})
	}

	// 3. Sort all events chronologically by Time
	sort.Slice(events, func(i, j int) bool {
		return events[i].Time < events[j].Time
	})

	// 4. Print the chronological timeline
	activeUsers := 0
	fmt.Printf("%-12s | %-6s | %-10s | %-12s | %s\n", "Time (s)", "User", "Action", "Active Users", "Details")
	fmt.Println("-----------------------------------------------------------------------------------------")

	for _, ev := range events {
		if ev.Type == "ARRIVED" {
			activeUsers++
		} else {
			activeUsers--
		}
		fmt.Printf("%-12.4f | %-6s | %-10s | %-12d | %s\n", ev.Time, ev.User, ev.Type, activeUsers, ev.Details)
	}
	fmt.Println("=========================================================================================")
}
