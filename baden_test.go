package main

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
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
	if header.Filesize != uint32(len(decompressed)) {
		t.Fatalf("header size %d does not match decoded size %d", header.Filesize, len(decompressed))
	}
	if checksum := crc32.ChecksumIEEE(decompressed[12:]); header.CRC32 != checksum {
		t.Fatalf("header CRC32 %08x does not match decoded CRC32 %08x", header.CRC32, checksum)
	}
	if validity := header.CloseTimestamp - header.OpenTimestamp; validity != forecastValidityMinutes {
		t.Fatalf("forecast validity is %d minutes, expected %d", validity, forecastValidityMinutes)
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

	type expectedLocation struct {
		countryCode uint8
		found       int
	}
	expected := map[string]expectedLocation{
		"San Giovanni in Fiore": {countryCode: countryCodes["Italy"]},
		"Maglie":                {countryCode: countryCodes["Italy"]},
		"Otranto":               {countryCode: countryCodes["Italy"]},
		"Santa Maria di Leuca":  {countryCode: countryCodes["Italy"]},
		"Gallipoli":             {countryCode: countryCodes["Italy"]},
		"Minervino di Lecce":    {countryCode: countryCodes["Italy"]},
		"Lecce":                 {countryCode: countryCodes["Italy"]},
		"Bari":                  {countryCode: countryCodes["Italy"]},
		"Brindisi":              {countryCode: countryCodes["Italy"]},
		"Crotone":               {countryCode: countryCodes["Italy"]},
		"Cosenza":               {countryCode: countryCodes["Italy"]},
	}
	priority := map[string]struct{}{
		"San Giovanni in Fiore": {},
		"Maglie":                {},
		"Otranto":               {},
		"Santa Maria di Leuca":  {},
		"Gallipoli":             {},
		"Minervino di Lecce":    {},
		"Lecce":                 {},
		"Bari":                  {},
		"Brindisi":              {},
		"Crotone":               {},
		"Cosenza":               {},
	}
	for _, city := range compatibleHomeCities {
		expected[city.name] = expectedLocation{countryCode: countryCodes["Switzerland"]}
		priority[city.name] = struct{}{}
	}
	countryCodesFound := make(map[uint8]struct{})
	for i := uint32(0); i < header.NumberOfLocations; i++ {
		offset := header.LocationsTableOffset + i*24
		countryCodesFound[decompressed[offset]] = struct{}{}
		cityTextOffset := binary.BigEndian.Uint32(decompressed[offset+4:])
		name := decodeUTF16BE(decompressed, cityTextOffset)
		if location, ok := expected[name]; ok && decompressed[offset] == location.countryCode {
			location.found++
			expected[name] = location
		}
		if location, ok := expected[name]; ok && decompressed[offset] == location.countryCode {
			if _, priorityLocation := priority[name]; priorityLocation && decompressed[offset+20] != 9 {
				t.Errorf("expected %s to have maximum map priority, found %d", name, decompressed[offset+20])
			}
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
	for name, location := range expected {
		if location.found != 1 {
			t.Errorf("expected exactly one %s location for country code %d, found %d", name, location.countryCode, location.found)
		}
	}
	shortData, err := os.ReadFile("files/1/108/short.bin")
	if err != nil {
		t.Fatal(err)
	}
	shortDecompressed, err := lz10.Decompress(shortData[320:])
	if err != nil {
		t.Fatal(err)
	}
	var shortHeader ShortHeader
	if err := binary.Read(bytes.NewReader(shortDecompressed), binary.BigEndian, &shortHeader); err != nil {
		t.Fatal(err)
	}
	if shortHeader.Filesize != uint32(len(shortDecompressed)) {
		t.Fatalf("short header size %d does not match decoded size %d", shortHeader.Filesize, len(shortDecompressed))
	}
	if checksum := crc32.ChecksumIEEE(shortDecompressed[12:]); shortHeader.CRC32 != checksum {
		t.Fatalf("short CRC32 %08x does not match decoded CRC32 %08x", shortHeader.CRC32, checksum)
	}
	if validity := shortHeader.CloseTimestamp - shortHeader.OpenTimestamp; validity != forecastValidityMinutes {
		t.Fatalf("short forecast validity is %d minutes, expected %d", validity, forecastValidityMinutes)
	}
	if shortHeader.NumberOfCurrentForecastTables != header.NumberOfLocations {
		t.Fatalf("short feed has %d current forecasts for %d locations", shortHeader.NumberOfCurrentForecastTables, header.NumberOfLocations)
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
