package main

// BuildCompatibleCities keeps the globe within the size the retail Forecast
// Channel accepts. A feed containing all 4,038 source locations downloads but
// is rejected by the channel at runtime, so use Nintendo's international set
// and put the requested additions first with the highest display priority.
func BuildCompatibleCities(list *WeatherList, homeCountry string) []InternationalCity {
	priorityNames := []string{
		"San Giovanni in Fiore",
		"Maglie",
		"Otranto",
		"Santa Maria di Leuca",
		"Gallipoli",
		"Minervino di Lecce",
	}
	cities := make([]InternationalCity, 0, len(list.International.Cities))
	seen := make(map[string]struct{}, len(list.International.Cities))

	add := func(city InternationalCity, priority bool) {
		if city.Country.English == homeCountry {
			return
		}
		key := city.Country.English + "\x00" + city.Province.English + "\x00" + city.Name.English
		if _, ok := seen[key]; ok {
			return
		}
		if priority {
			city.Zoom1 = 9
			city.Zoom2 = 3
		}
		seen[key] = struct{}{}
		cities = append(cities, city)
	}

	for _, name := range priorityNames {
		for _, city := range list.International.Cities {
			if city.Name.English == name {
				add(city, true)
				break
			}
		}
	}
	for _, city := range list.International.Cities {
		add(city, false)
	}

	return cities
}
