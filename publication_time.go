package main

import (
	"fmt"
	"strconv"
)

// The signed header is a validity interval. Location timestamps continue to
// use currentTime (actual collection time), never the planned publication time.
func parseForecastValidFrom(value string, collectedAt int64) (int64, error) {
	if value == "" {
		return collectedAt, nil
	}
	validFrom, err := strconv.ParseInt(value, 10, 64)
	if err != nil || validFrom < 946684800 || validFrom%3600 != 0 {
		return 0, fmt.Errorf("FORECAST_VALID_FROM must be an integer UTC hour boundary after 2000")
	}
	if validFrom-collectedAt > 15*60 || collectedAt >= validFrom+3600 {
		return 0, fmt.Errorf("forecast validity cycle is too early or already expired")
	}
	return validFrom, nil
}
