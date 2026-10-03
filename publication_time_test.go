package main

import "testing"

func TestForecastPublicationValidity(t *testing.T) {
	const boundary int64 = 1791057600
	for _, test := range []struct {
		name, input       string
		collectedAt, want int64
		wantError         bool
	}{
		{"normal generator", "", boundary - 530, boundary - 530, false},
		{"early preparation", "1791057600", boundary - 530, boundary, false},
		{"late preparation", "1791057600", boundary + 90, boundary, false},
		{"bad value", "bad", boundary, 0, true},
		{"not hour", "1791057601", boundary, 0, true},
		{"too early", "1791061200", boundary, 0, true},
		{"expired", "1791057600", boundary + 3600, 0, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseForecastValidFrom(test.input, test.collectedAt)
			if (err != nil) != test.wantError || (!test.wantError && got != test.want) {
				t.Fatalf("got (%d, %v), want (%d, error=%v)", got, err, test.want, test.wantError)
			}
		})
	}
}

func TestHeaderValidityDoesNotChangeCollectionTimestamp(t *testing.T) {
	oldCollected, oldValidity := currentTime, forecastValidFrom
	defer func() { currentTime, forecastValidFrom = oldCollected, oldValidity }()
	currentTime = 1791057070
	forecastValidFrom = 1791057600
	f := Forecast{}
	f.MakeHeader()
	if f.Header.OpenTimestamp != fixTime(int(forecastValidFrom)) ||
		f.Header.CloseTimestamp-f.Header.OpenTimestamp != 60 {
		t.Fatal("header does not cover the planned hour")
	}
	if currentTime != 1791057070 {
		t.Fatal("publication plan changed collection time")
	}
}
