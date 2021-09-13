package main

import (
	"fmt"
	"github.com/d-v-p/InaraParser/app/commodity"
	log "github.com/sirupsen/logrus"
	"net/url"
)

func RequesterGetMock(url string) string {
	return "<tr><td>system | station</td><td>L</td><td>-</td><td>300ly</td><td>300,100</td><td>100Cr</td><td>-</td><td>1 minute ago</td></tr>"
}

func RequesterPostMock(url string, data url.Values) string {
	return "refid2=222"
}

func main() {
	log.SetLevel(log.TraceLevel)

	stationList := commodity.GetSystemList("Void opal", "Veroandi")
	if stationList != nil {
		findBestPrice(stationList, 100)
	}
}

func findBestPrice(stationList []commodity.SystemLine, maxDistance int)  {
	station := commodity.GetBestPrice(stationList, maxDistance, 1, 0)
	fmt.Printf("BP in %d ly: %s | %s - %d ly - %d Cr\n", maxDistance, station.System, station.Station, station.Distance, station.Price)
}
