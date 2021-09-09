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


type RequesterMethods struct {
	Get httpRequester.GetMessage
	Post httpRequester.PostMessage
}

var requester = RequesterMethods{httpRequester.Get, httpRequester.Post}

var commodityNameToId map[string]int

type CommodityData struct {
	System string
	Station string
	Pad string
	Distance int
	Quantity int
	MaxQuantity int
	Price int
	Updated string
}

var commodityPrices = []CommodityData{}

// TODO: error handling

func InitRequestor(get httpRequester.GetMessage, post httpRequester.PostMessage) {
	requester.Get = get
	requester.Post = post
}

func getCommodityListFromHtml(html string) {
	commodityNameToId = make(map[string]int)

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
	for _, option := range optionList {

		commodityNameToId[strings.ToLower(utility.ParseString(option[2]))] = utility.ParseInteger(option[1])
	}

}

func getCommodityIdByName(commodityName string) int {
	if commodityNameToId == nil {
		commodityHtml := requester.Get("https://inara.cz/commodity/")
		getCommodityListFromHtml(commodityHtml)
	}

	return commodityNameToId[strings.ToLower(commodityName)]
}

func getRefferenceSystemId(commodityId int, systemName string) int {
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

func GetCommodityPrices(commodityName string, systemName string) {
	commodityId := getCommodityIdByName(commodityName)
	if commodityId == 0 {
		log.Panic("Can't find commodity id by name ", commodityName)
	}
	refSystemId := getRefferenceSystemId(commodityId, systemName)
	if refSystemId != 0 {
		url := fmt.Sprintf("https://inara.cz/ajaxaction.php?act=goodsdata&refname=sellmax&refid=%d&refid2=%d" , commodityId, refSystemId)
		systemListHtml := requester.Get(url)
		r := regexp.MustCompile(`<tr.*?>(.+?)</tr>`)
		res := r.FindAllStringSubmatch(systemListHtml, -1)

		for _, systemLine := range res {
			r = regexp.MustCompile(`<td.*?>(.*?)</td>`)
			res = r.FindAllStringSubmatch(systemLine[1], -1)
			if res != nil {
				systemStation := strings.Split(utility.ParseString(res[0][1]), "|")

				var station CommodityData

				station.System = utility.ParseString(systemStation[0])
				station.Station = utility.ParseString(systemStation[1])
				station.Pad = utility.ParseString(res[1][1])
				station.Distance = utility.ParseInteger(res[3][1])
				station.Quantity = utility.ParseInteger(res[4][1])
				station.Price = utility.ParseInteger(res[5][1])
				station.Updated = utility.ParseString(res[7][1])

				r = regexp.MustCompile(`more than (\d+)`)
				res4 := r.FindStringSubmatch(res[4][1])
				if res4 != nil {
					station.MaxQuantity = utility.ParseInteger(res4[1])
				} else {
					station.MaxQuantity = station.Quantity
				}

				commodityPrices = append(commodityPrices, station)
			}
		}

		//fmt.Println(commodityPrices)
	}
}

func GetBestCommodityPrice(commodityName string, refSystemName string, maxDistance int, landingPad string, itemsQuantity int) CommodityData {
	if len(commodityPrices) == 0 {
		GetCommodityPrices(commodityName, refSystemName)
	}

	var bestStation CommodityData
	for _, commodityPrice := range commodityPrices {

		if commodityPrice.System == "Chambo" {
			fmt.Println(commodityPrice)
		}

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