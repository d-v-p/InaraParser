package commodity

import (
	"fmt"
	"github.com/d-v-p/InaraParser/app/httpRequester"
	"github.com/d-v-p/InaraParser/app/utility"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// methods to send http requests
type RequesterMethods struct {
	Get httpRequester.GetMessage
	Post httpRequester.PostMessage
}

// default requester
var requester = RequesterMethods{httpRequester.Get, httpRequester.Post}

// map commodity name to internal inara id
var NameToId map[string]int

// station with commodity price
type StationLine struct {
	System string
	Station string
	Pad string
	Distance int
	Quantity int
	MaxQuantity int
	Price int
	Updated string
}

// TODO: error handling

func SetRequesterMethods(get httpRequester.GetMessage, post httpRequester.PostMessage) {
	requester.Get = get
	requester.Post = post
}

func GetStationList(commodityName string, systemName string) []StationLine {
	commodityId := getCommodityIdByName(commodityName)
	refSystemId := getRefferenceSystemIdByNameFromInara(commodityId, systemName)
	if commodityId == 0 || refSystemId == 0 {
		return nil
	}

	res := getStationListFromInara(commodityId, refSystemId)

	return res
}

func getCommodityIdByName(commodityName string) int {
	fillCommodityNameToIdMap()
	return NameToId[strings.ToLower(commodityName)]
}

func fillCommodityNameToIdMap() {
	// fill list once
	if NameToId != nil {
		return
	}

	commodityList := getCommodityListFromInara()
	for _, commodity := range commodityList {
		NameToId[strings.ToLower(utility.ParseString(commodity[2]))] = utility.ParseInteger(commodity[1])
	}
}

func getCommodityListFromInara() [][]string {
	html := requester.Get("https://inara.cz/commodity/")
	NameToId = make(map[string]int)

	r := regexp.MustCompile(`<select.*?name="searchcommodity".*?>(.*?)</select>`)
	optionsHtml := r.FindStringSubmatch(html)
	if optionsHtml == nil {
		log.Panic("Can't find select with commodity list in html")
	}

	r = regexp.MustCompile(`<option.*?value="(\d+)".*?>(.*?)</option>`)
	optionList := r.FindAllStringSubmatch(optionsHtml[1], -1)
	if optionList == nil {
		log.Panic("Can't find option with commodity in html")
	}

	return optionList
}

func getRefferenceSystemIdByNameFromInara(commodityId int, systemName string) int {
	commodityPage := requester.Post("https://inara.cz/commodity/", url.Values{
		"formact": {"SEARCH_COMMODITIES"},
		"searchcommodity": {string(commodityId)},
		"searchcommoditysystem": {systemName},
	})
	r := regexp.MustCompile(`refid2=(\d+)`)
	res := r.FindStringSubmatch(commodityPage)
	if res == nil {
		log.Panic("Can't find reference system id")
	}

	systemId, err := strconv.Atoi(res[1])
	if err != nil {
		log.Panic(err)
	}

	return systemId
}

func getStationListFromInara(commodityId int, refSystemId int) []StationLine {
	var StationList []StationLine

	url := fmt.Sprintf("https://inara.cz/ajaxaction.php?act=goodsdata&refname=sellmax&refid=%d&refid2=%d" , commodityId, refSystemId)
	systemListHtml := requester.Get(url)

	r := regexp.MustCompile(`<tr.*?>(.+?)</tr>`)
	systemList := r.FindAllStringSubmatch(systemListHtml, -1)
	if systemList == nil {
		return nil
	}

	for _, systemLine := range systemList {
		r = regexp.MustCompile(`<td.*?>(.*?)</td>`)
		systemParamList := r.FindAllStringSubmatch(systemLine[1], -1)
		if systemParamList != nil {
			var stationPrice StationLine

			systemStation := strings.Split(utility.ParseString(systemParamList[0][1]), "|")

			stationPrice.Station = utility.ParseString(systemStation[0])
			stationPrice.System = utility.ParseString(systemStation[1])
			stationPrice.Pad = utility.ParseString(systemParamList[1][1]) // TODO: change to integer
			stationPrice.Distance = utility.ParseInteger(systemParamList[3][1])
			stationPrice.Quantity = utility.ParseInteger(systemParamList[4][1])
			stationPrice.Price = utility.ParseInteger(systemParamList[5][1])
			stationPrice.Updated = utility.ParseString(systemParamList[7][1])

			r := regexp.MustCompile(`more than (\d+)`)
			res4 := r.FindStringSubmatch(systemParamList[4][1])
			if res4 != nil {
				stationPrice.MaxQuantity = utility.ParseInteger(res4[1])
			} else {
				stationPrice.MaxQuantity = stationPrice.Quantity
			}

			StationList = append(StationList, stationPrice)
		}
	}

	return StationList
}

func GetBestPrice(stationList []StationLine, maxDistance int, landingPad string, itemsQuantity int) StationLine {
	var bestStation StationLine
	for _, commodityPrice := range stationList {
		if landingPad == "M" && commodityPrice.Pad == "S" {
			continue
		}
		if landingPad == "L" && commodityPrice.Pad != "L" {
			continue
		}

		if commodityPrice.Distance <= maxDistance && commodityPrice.MaxQuantity >= itemsQuantity {
			if commodityPrice.Price > bestStation.Price {
				bestStation = commodityPrice
			}
		}
	}

	return bestStation
}