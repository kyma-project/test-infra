package gcscleaner

import "regexp"

func ExtractTimestampSuffix(name string, regTimestampSuffix regexp.Regexp) *string {
	return extractTimestampSuffix(name, regTimestampSuffix)
}
