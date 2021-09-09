package utility

import (
	"github.com/grokify/html-strip-tags-go"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
)

func ParseString(s string) string {
	re := regexp.MustCompile("[[:^ascii:]]")
	t := re.ReplaceAllLiteralString(strip.StripTags(s), "")

	return strings.Trim(t, " ")
}

func ParseInteger(s string) int {
	re := regexp.MustCompile("[^0-9.]")
	t := re.ReplaceAllLiteralString(strip.StripTags(s), "")

	res, err := strconv.ParseFloat(t, 32)
	if err != nil {
		log.Panic(err)
	}

	return int(math.Round(res))
}
