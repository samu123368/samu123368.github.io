package main

type compatibleHomeCity struct {
	name      string
	province  string
	longitude float64
	latitude  float64
}

// Keep the requested Swiss map set deliberately compact. Existing Nintendo
// locations are promoted to maximum priority; missing places are inserted.
var compatibleHomeCities = []compatibleHomeCity{
	// Baden, lower Aargau and the Limmattal.
	{name: "Untersiggenthal", province: "Aargau", longitude: 8.25554, latitude: 47.50213},
	{name: "Gebenstorf", province: "Aargau", longitude: 8.23949, latitude: 47.48136},
	{name: "Baden", province: "Aargau", longitude: 8.30592, latitude: 47.47333},
	{name: "Wettingen", province: "Aargau", longitude: 8.32663, latitude: 47.46606},
	{name: "Spreitenbach", province: "Aargau", longitude: 8.36792, latitude: 47.42285},
	{name: "Dietikon", province: "Zürich", longitude: 8.40015, latitude: 47.40165},
	{name: "Schlieren", province: "Zürich", longitude: 8.44763, latitude: 47.39668},
	{name: "Zürich", province: "Zürich", longitude: 8.55, latitude: 47.36667},
	{name: "Brugg", province: "Aargau", longitude: 8.20869, latitude: 47.48096},
	{name: "Birmenstorf", province: "Aargau", longitude: 8.250005, latitude: 47.463884},

	// Important Reusstal centres and transport hubs, from south to north.
	{name: "Andermatt", province: "Uri", longitude: 8.59388, latitude: 46.63565},
	{name: "Göschenen", province: "Uri", longitude: 8.58709, latitude: 46.66816},
	{name: "Erstfeld", province: "Uri", longitude: 8.65052, latitude: 46.81885},
	{name: "Altdorf", province: "Uri", longitude: 8.64441, latitude: 46.88042},
	{name: "Lucerne", province: "Luzern", longitude: 8.30635, latitude: 47.05048},
	{name: "Emmen", province: "Luzern", longitude: 8.27331, latitude: 47.07819},
	{name: "Gisikon", province: "Luzern", longitude: 8.40356, latitude: 47.12701},
	{name: "Rotkreuz", province: "Zug", longitude: 8.43140, latitude: 47.14283},
	{name: "Muri", province: "Aargau", longitude: 8.33854, latitude: 47.27428},
	{name: "Bremgarten", province: "Aargau", longitude: 8.34214, latitude: 47.35109},
	{name: "Mellingen", province: "Aargau", longitude: 8.27331, latitude: 47.41903},
}

func addCompatibleHomeCities(list *WeatherList, homeCountry string) {
	var home *NationalList
	for i := range list.National {
		if list.National[i].Name.English == homeCountry {
			home = &list.National[i]
			break
		}
	}
	if home == nil {
		return
	}

	provinces := make(map[string]LocalizedNames)
	for _, city := range home.Cities {
		provinces[city.Province.English] = city.Province
	}

	for _, requested := range compatibleHomeCities {
		found := false
		for i := range home.Cities {
			if home.Cities[i].English != requested.name {
				continue
			}
			home.Cities[i].Zoom1 = 9
			home.Cities[i].Zoom2 = 3
			found = true
			break
		}
		if found {
			continue
		}

		province, ok := provinces[requested.province]
		if !ok {
			continue
		}
		home.Cities = append(home.Cities, City{
			Japanese:  requested.name,
			English:   requested.name,
			German:    requested.name,
			French:    requested.name,
			Spanish:   requested.name,
			Italian:   requested.name,
			Dutch:     requested.name,
			Russian:   requested.name,
			Province:  province,
			Longitude: requested.longitude,
			Latitude:  requested.latitude,
			Zoom1:     9,
			Zoom2:     3,
		})
	}
}

// BuildCompatibleCities keeps the globe within the size the retail Forecast
// Channel accepts. A feed containing all 4,038 source locations downloads but
// is rejected by the channel at runtime, so use Nintendo's international set
// and insert requested national cities with the highest display priority.
func BuildCompatibleCities(list *WeatherList, homeCountry string) []InternationalCity {
	addCompatibleHomeCities(list, homeCountry)

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
