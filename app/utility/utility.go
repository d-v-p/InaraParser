package utility

import (
	"github.com/grokify/html-strip-tags-go"
	log "github.com/sirupsen/logrus"
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
		log.Warnln(err)
		return 0
	}

	return int(math.Round(res))
}
