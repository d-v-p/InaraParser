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
	MaxQuantity   int
	LimitedDemand bool
	Price         int
	Updated string
	UpdatedSecAgo int
}

func SetRequesterMethods(get httpRequester.GetMessage, post httpRequester.PostMessage) {
	requester.Get = get
	requester.Post = post
}

func GetBestPrice(sysytemList []SystemLine, maxDistance int, landingPad int, itemsQuantity int) SystemLine {
	log.Tracef("getting system with best price in max distance %d, landing pad %d, min items quantity %d", maxDistance, landingPad, itemsQuantity)

	var bpSystem SystemLine
	for _, system := range sysytemList {
		// skip systems with limited items demand updated more than 6 hours ago,
		// because they often contain not actual data
		if system.LimitedDemand && system.UpdatedSecAgo > 3600*6 {
			continue
		}

		if system.Distance <= maxDistance && system.MaxQuantity >= itemsQuantity && system.Pad >= landingPad  {
			if system.Price > bpSystem.Price {
				bpSystem = system
			}
		}
	}

	log.Traceln("best price system", bpSystem)

	return bpSystem
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

	systemList := getSystemListFromInara(commodityId, refSystemId)
	if systemList == nil {
		log.Warnln("can't get system list")
		return nil
	}

	log.Traceln("getting system list done")

	return systemList
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

		stationSystem := strings.Split(utility.ParseString(systemParamList[0][1]), "|") // TODO: move to func, add check | presence

		system.Station = utility.ParseString(stationSystem[0])
		system.System = utility.ParseString(stationSystem[1])
		system.Pad = systemLandingPadToInt(utility.ParseString(systemParamList[1][1]))
		system.Distance = utility.ParseInteger(systemParamList[3][1])
		system.Quantity = utility.ParseInteger(systemParamList[4][1])
		system.MaxQuantity = system.Quantity
		system.Price = utility.ParseInteger(systemParamList[5][1])
		system.Updated = utility.ParseString(systemParamList[7][1])
		system.UpdatedSecAgo = systemUpdatedStrToSec(system.Updated)
		limitedQuantity, limited := systemGetLimitedDemandCount(systemParamList[4][1])
		if limited {
			system.MaxQuantity = limitedQuantity
			system.LimitedDemand = true
		}

		SystemList = append(SystemList, system)
	}

	return SystemList
}

func systemLandingPadToInt(pad string) int {
	padInt := 0

	switch strings.ToLower(pad)  {
		case "s":
			padInt = 1
			break
		case "m":
			padInt = 2
			break
		case "l":
			padInt = 3
			break
	}

	return padInt
}

func systemUpdatedStrToSec(updated string) int {
	updatedSec := 0

	r := regexp.MustCompile(`(\d+) (.*) ago`)
	res := r.FindStringSubmatch(updated)
	if res != nil {
		multiplier := 1
		if strings.Contains(res[2], "minute") {
			multiplier = 60
		} else if strings.Contains(res[2], "hour") {
			multiplier = 60*60
		} else if strings.Contains(res[2], "day") {
			multiplier = 60*60*24
		}

		updatedInt, err := strconv.Atoi(res[1])
		if err != nil {
			log.Warnln(err)
		}
		updatedSec = updatedInt*multiplier
	}

	return updatedSec
}

func systemGetLimitedDemandCount(quantityStr string) (int, bool) {
	quantity := 0
	limited := false

	r := regexp.MustCompile(`more than (\d+)`)
	res := r.FindStringSubmatch(quantityStr)
	if res != nil {
		quantity = utility.ParseInteger(res[1])
		limited = true
	}

	return quantity, limited
}