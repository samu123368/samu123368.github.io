package main

// BuildCompatibleCities keeps the globe within the size the retail Forecast
// Channel accepts. A feed containing all 4,038 source locations downloads but
// is rejected by the channel at runtime, so use Nintendo's international set
// and insert requested national cities with the highest display priority.
func BuildCompatibleCities(list *WeatherList, homeCountry string) []InternationalCity {
	priorityNames := []string{
		"San Giovanni in Fiore",
		"Maglie",
		"Otranto",
		"Santa Maria di Leuca",
		"Gallipoli",
		"Minervino di Lecce",
		"Lecce",
		"Bari",
		"Brindisi",
		"Crotone",
		"Cosenza",
	}
	cities := make([]InternationalCity, 0, len(list.International.Cities)+len(priorityNames))
	seen := make(map[string]struct{}, len(list.International.Cities)+len(priorityNames))

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

	find := func(name string) (InternationalCity, bool) {
		for _, city := range list.International.Cities {
			if city.Name.English == name {
				return city, true
			}
		}
		for _, country := range list.National {
			for _, city := range country.Cities {
				if city.English != name {
					continue
				}
				return InternationalCity{
					Name: LocalizedNames{
						Japanese: city.Japanese,
						English:  city.English,
						German:   city.German,
						French:   city.French,
						Spanish:  city.Spanish,
						Italian:  city.Italian,
						Dutch:    city.Dutch,
						Russian:  city.Russian,
					},
					Province:  city.Province,
					Country:   country.Name,
					Longitude: city.Longitude,
					Latitude:  city.Latitude,
					Zoom1:     city.Zoom1,
					Zoom2:     city.Zoom2,
				}, true
			}
		}
		return InternationalCity{}, false
	}

	for _, name := range priorityNames {
		if city, ok := find(name); ok {
			add(city, true)
		}
	}
	for _, city := range list.International.Cities {
		add(city, false)
	}

	return cities
}
