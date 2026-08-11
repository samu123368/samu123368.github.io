package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
	"unicode/utf16"

	"github.com/wii-tools/lzx/lz10"
)

func TestBadenInEnglishForecast(t *testing.T) {
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

	found := 0
	for i := uint32(0); i < header.NumberOfLocations; i++ {
		offset := header.LocationsTableOffset + i*24
		cityTextOffset := binary.BigEndian.Uint32(decompressed[offset+4:])
		if decodeUTF16BE(decompressed, cityTextOffset) == "Baden" {
			found++
		}
	}
	if found != 1 {
		t.Fatalf("expected exactly one Baden location, found %d", found)
	}
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
