package commodity

import (
	"fmt"
	"github.com/d-v-p/InaraParser/app/httpRequester"
	"github.com/d-v-p/InaraParser/app/utility"
	log "github.com/sirupsen/logrus"
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

var NameToId map[string]int

type SystemLine struct {
	System string
	Station string
	Pad int
	Distance int
	Quantity int
	MaxQuantity int
	Price int
	Updated string
}

func SetRequesterMethods(get httpRequester.GetMessage, post httpRequester.PostMessage) {
	requester.Get = get
	requester.Post = post
}

func GetSystemList(commodityName string, systemName string) []SystemLine {
	log.Tracef("getting system list for commodity '%s' and reference system '%s'", commodityName, systemName)

	commodityId := getCommodityIdByName(commodityName)
	if commodityId == 0 {
		log.Warnln("can't get system list, empty commodity id")
		return nil
	}
	refSystemId := getReferenceSystemIdByNameFromInara(commodityId, systemName)
	if refSystemId == 0 {
		log.Warnln("can't get system list, empty reference system id")
		return nil
	}

	res := getSystemListFromInara(commodityId, refSystemId)

	return res
}

func getCommodityIdByName(commodityName string) int {
	log.Traceln("getting id for commodity", commodityName)

	fillCommodityNameToIdMap()
	commodityId, ok := NameToId[strings.ToLower(commodityName)]
	if !ok {
		log.Warnln("can't find id for commodity", commodityName)
		// TODO: search by name part
		return 0
	}

	log.Traceln("commodity id is", commodityId)

	return commodityId
}

func fillCommodityNameToIdMap() {
	log.Traceln("fill commodity name hash map")

	// fill map once
	if NameToId != nil {
		log.Traceln("hash map already filled, skipping")
		return
	}

	NameToId = make(map[string]int)
	commodityList := getCommodityListFromInara()
	for _, commodity := range commodityList {
		NameToId[strings.ToLower(utility.ParseString(commodity[2]))] = utility.ParseInteger(commodity[1])
	}
}

func getCommodityListFromInara() [][]string {
	log.Traceln("getting commodity list from inara")

	html := requester.Get("https://inara.cz/commodity/")

	r := regexp.MustCompile(`<select.*?name="searchcommodity".*?>(.*?)</select>`)
	optionsHtml := r.FindStringSubmatch(html)
	if optionsHtml == nil {
		log.Warnln("can't find select with commodity list in html")
		return nil
	}

	r = regexp.MustCompile(`<option.*?value="(\d+)".*?>(.*?)</option>`)
	optionList := r.FindAllStringSubmatch(optionsHtml[1], -1)
	if optionList == nil {
		log.Warnln("can't find option with commodity in html")
		return nil
	}

	return optionList
}

func getReferenceSystemIdByNameFromInara(commodityId int, systemName string) int {
	log.Tracef("getting id for reference system %s, using commodity id %d", systemName, commodityId)

	commodityPage := requester.Post("https://inara.cz/commodity/", url.Values{
		"formact": {"SEARCH_COMMODITIES"},
		"searchcommodity": {strconv.Itoa(commodityId)},
		"searchcommoditysystem": {systemName},
	})
	r := regexp.MustCompile(`refid2=(\d+)`)
	res := r.FindStringSubmatch(commodityPage)
	if res == nil {
		log.Warnln("can't find id for reference system", systemName)
		return 0
	}

	systemId, err := strconv.Atoi(res[1])
	if err != nil {
		log.Warnln(err)
		return 0
	}

	log.Traceln("system id is", systemId)

	return systemId
}

func getSystemListFromInara(commodityId int, refSystemId int) []SystemLine {
	log.Tracef("getting system list for commodity %d and reference system %d", commodityId, refSystemId)

	var SystemList []SystemLine

	systemListUrl := fmt.Sprintf("https://inara.cz/ajaxaction.php?act=goodsdata&refname=sellmax&refid=%d&refid2=%d" , commodityId, refSystemId)
	systemListHtml := requester.Get(systemListUrl)

	r := regexp.MustCompile(`<tr.*?>(.+?)</tr>`)
	systemList := r.FindAllStringSubmatch(systemListHtml, -1)
	if systemList == nil {
		log.Warnln("can't parse system list")
		return nil
	}

	for _, systemLine := range systemList {
		r = regexp.MustCompile(`<td.*?>(.*?)</td>`)
		systemParamList := r.FindAllStringSubmatch(systemLine[1], -1)
		if systemParamList == nil {
			continue
		}

		var system SystemLine

		stationSystem := strings.Split(utility.ParseString(systemParamList[0][1]), "|")

		system.Station = utility.ParseString(stationSystem[0])
		system.System = utility.ParseString(stationSystem[1])
		switch strings.ToLower(utility.ParseString(systemParamList[1][1]))  {
			case "s":
				system.Pad = 1
				break
			case "m":
				system.Pad = 2
				break
			case "l":
				system.Pad = 3
				break
		}
		if system.Pad == 0 {
			continue
		}
		system.Distance = utility.ParseInteger(systemParamList[3][1])
		system.Quantity = utility.ParseInteger(systemParamList[4][1])
		system.Price = utility.ParseInteger(systemParamList[5][1])
		system.Updated = utility.ParseString(systemParamList[7][1])

		r := regexp.MustCompile(`more than (\d+)`)
		res4 := r.FindStringSubmatch(systemParamList[4][1])
		if res4 != nil {
			system.MaxQuantity = utility.ParseInteger(res4[1])
		} else {
			system.MaxQuantity = system.Quantity
		}

		SystemList = append(SystemList, system)
	}

	return SystemList
}

func GetBestPrice(sysytemList []SystemLine, maxDistance int, landingPad int, itemsQuantity int) SystemLine {
	log.Tracef("getting system with best price in max distance %d, landing pad %d, min items quantity %d", maxDistance, landingPad, itemsQuantity)

	var bpSystem SystemLine
	for _, system := range sysytemList {
		if system.Distance <= maxDistance && system.MaxQuantity >= itemsQuantity && system.Pad >= landingPad  {
			if system.Price > bpSystem.Price {
				bpSystem = system
			}
		}
	}

	log.Traceln("best price system", bpSystem)

	return bpSystem
}