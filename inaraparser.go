package inaraParser

import (
	"github.com/d-v-p/InaraParser/app/commodity"
)

func FindCommodityBestPrice(commodityName string, refSystemName string, maxDistance int, landingPad string, itemsQuantity int) commodity.CommodityData {
	return commodity.GetBestCommodityPrice(commodityName, refSystemName, maxDistance, landingPad, itemsQuantity)
}
