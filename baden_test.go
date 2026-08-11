package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
	"unicode/utf16"

	"github.com/wii-tools/lzx/lz10"
)

func TestCustomLocationsInEnglishForecast(t *testing.T) {
	data, err := os.ReadFile("files/1/108/forecast.bin")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) <= 320 {
		t.Fatal("forecast file is too short")
	}

	decompressed, err := lz10.Decompress(data[320:])
	if err != nil {
		t.Fatal(err)
	}

	var header Header
	if err := binary.Read(bytes.NewReader(decompressed), binary.BigEndian, &header); err != nil {
		t.Fatal(err)
	}

	list := ParseWeatherXML()
	list.International.Cities = BuildCompatibleCities(list, "Switzerland")
	PopulateCountryCodes()
	var switzerland *NationalList
	for i := range list.National {
		if list.National[i].Name.English == "Switzerland" {
			switzerland = &list.National[i]
			break
		}
	}
	if switzerland == nil {
		t.Fatal("Switzerland source list is missing")
	}
	expectedForecast := Forecast{
		currentCountryList: switzerland,
		currentCountryCode: countryCodes["Switzerland"],
	}
	expectedForecast.PopulateLocations(list.International.Cities)
	expectedLocations := 0
	expectedCountryCodes := make(map[uint8]struct{})
	for country := expectedForecast.rawLocations.Oldest(); country != nil; country = country.Next() {
		for province := country.Value.Oldest(); province != nil; province = province.Next() {
			expectedLocations += province.Value.Len()
			for city := province.Value.Oldest(); city != nil; city = city.Next() {
				expectedCountryCodes[city.Value.CountryCode] = struct{}{}
			}
		}
	}
	if int(header.NumberOfLocations) != expectedLocations {
		t.Fatalf("generated %d locations, expected the compatible set of %d", header.NumberOfLocations, expectedLocations)
	}
	if header.NumberOfLocations > 1000 {
		t.Fatalf("generated %d locations, exceeding the retail channel compatibility ceiling", header.NumberOfLocations)
	}
	if header.NumberOfLongForecastTables+header.NumberOfShortForecastTables != header.NumberOfLocations {
		t.Fatalf("not every location has a forecast table: %d long + %d short != %d locations", header.NumberOfLongForecastTables, header.NumberOfShortForecastTables, header.NumberOfLocations)
	}

	expected := map[string]int{
		"Baden":                 0,
		"San Giovanni in Fiore": 0,
		"Maglie":                0,
		"Otranto":               0,
		"Santa Maria di Leuca":  0,
		"Gallipoli":             0,
		"Minervino di Lecce":    0,
	}
	priority := map[string]struct{}{
		"San Giovanni in Fiore": {},
		"Maglie":                {},
		"Otranto":               {},
		"Santa Maria di Leuca":  {},
		"Gallipoli":             {},
		"Minervino di Lecce":    {},
	}
	countryCodesFound := make(map[uint8]struct{})
	for i := uint32(0); i < header.NumberOfLocations; i++ {
		offset := header.LocationsTableOffset + i*24
		countryCodesFound[decompressed[offset]] = struct{}{}
		cityTextOffset := binary.BigEndian.Uint32(decompressed[offset+4:])
		name := decodeUTF16BE(decompressed, cityTextOffset)
		if _, ok := expected[name]; ok {
			expected[name]++
		}
		if _, ok := priority[name]; ok && decompressed[offset+20] != 9 {
			t.Errorf("expected %s to have maximum map priority, found %d", name, decompressed[offset+20])
		}
	}
	if len(countryCodesFound) != len(expectedCountryCodes) {
		t.Errorf("country-code coverage mismatch: found %d, expected %d", len(countryCodesFound), len(expectedCountryCodes))
	}

	for i := uint32(0); i < header.NumberOfLongForecastTables; i++ {
		offset := header.LongForecastTableOffset + i*uint32(binary.Size(LongForecastTable{}))
		if icon := binary.BigEndian.Uint16(decompressed[offset+16:]); icon == 0xFFFF {
			t.Errorf("long forecast %d has no weather icon", i)
		}
	}
	for i := uint32(0); i < header.NumberOfShortForecastTables; i++ {
		offset := header.ShortForecastTableOffset + i*uint32(binary.Size(ShortForecastTable{}))
		if icon := binary.BigEndian.Uint16(decompressed[offset+16:]); icon == 0xFFFF {
			t.Errorf("short forecast %d has no weather icon", i)
		}
	}
	for name, found := range expected {
		if found != 1 {
			t.Errorf("expected exactly one %s location, found %d", name, found)
		}
	}
	t.Logf("validated %d compatible locations and all requested additions", header.NumberOfLocations)
}

func decodeUTF16BE(data []byte, offset uint32) string {
	var units []uint16
	for i := int(offset); i+1 < len(data); i += 2 {
		unit := binary.BigEndian.Uint16(data[i:])
		if unit == 0 {
			break
		}
		units = append(units, unit)
	}
	return string(utf16.Decode(units))
}
