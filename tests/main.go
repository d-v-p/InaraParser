package main

import (
	"fmt"
	"github.com/d-v-p/InaraParser"
	"net/url"
)

func RequesterGetMock(url string) string {
	return "<tr><td>system | station</td><td>L</td><td>-</td><td>300ly</td><td>300,100</td><td>100Cr</td><td>-</td><td>1 minute ago</td></tr>"
}

func RequesterPostMock(url string, data url.Values) string {
	return "refid2=222"
}

func main() {
	findBestPrice(20)
	findBestPrice(50)
	findBestPrice(100)
	findBestPrice(9999999)
}

func findBestPrice(maxDistance int)  {
	station := inaraParser.FindCommodityBestPrice("Alexandrite", "Veroandi", maxDistance, "S", 0)
	fmt.Printf("BP in %d ly: %s | %s - %d ly - %d Cr\n", maxDistance, station.System, station.Station, station.Distance, station.Price)
}
