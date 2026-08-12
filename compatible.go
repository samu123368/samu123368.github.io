package main

type compatibleHomeCity struct {
	name      string
	province  string
	longitude float64
	latitude  float64
}

// The 21-place Limmattal set follows the Regionale 2025 map. Baden and Zürich
// already exist in Nintendo's source list; the remaining places are inserted
// here along with the specifically requested nearby Aargau additions.
var compatibleHomeCities = []compatibleHomeCity{
	{name: "Untersiggenthal", province: "Aargau", longitude: 8.25554, latitude: 47.50213},
	{name: "Turgi", province: "Aargau", longitude: 8.25412, latitude: 47.49201},
	{name: "Gebenstorf", province: "Aargau", longitude: 8.23949, latitude: 47.48136},
	{name: "Baden", province: "Aargau", longitude: 8.30592, latitude: 47.47333},
	{name: "Obersiggenthal", province: "Aargau", longitude: 8.29652, latitude: 47.48750},
	{name: "Ennetbaden", province: "Aargau", longitude: 8.32399, latitude: 47.48055},
	{name: "Wettingen", province: "Aargau", longitude: 8.32663, latitude: 47.46606},
	{name: "Neuenhof", province: "Aargau", longitude: 8.32682, latitude: 47.44985},
	{name: "Killwangen", province: "Aargau", longitude: 8.35097, latitude: 47.43223},
	{name: "Würenlos", province: "Aargau", longitude: 8.36261, latitude: 47.44205},
	{name: "Spreitenbach", province: "Aargau", longitude: 8.36792, latitude: 47.42285},
	{name: "Oetwil an der Limmat", province: "Zürich", longitude: 8.394169, latitude: 47.430834},
	{name: "Geroldswil", province: "Zürich", longitude: 8.41085, latitude: 47.42213},
	{name: "Weiningen", province: "Zürich", longitude: 8.43644, latitude: 47.42022},
	{name: "Unterengstringen", province: "Zürich", longitude: 8.44761, latitude: 47.41396},
	{name: "Dietikon", province: "Zürich", longitude: 8.40015, latitude: 47.40165},
	{name: "Bergdietikon", province: "Aargau", longitude: 8.38624, latitude: 47.38921},
	{name: "Urdorf", province: "Zürich", longitude: 8.42581, latitude: 47.38507},
	{name: "Schlieren", province: "Zürich", longitude: 8.44763, latitude: 47.39668},
	{name: "Oberengstringen", province: "Zürich", longitude: 8.46515, latitude: 47.40841},
	{name: "Zürich", province: "Zürich", longitude: 8.55, latitude: 47.36667},
	{name: "Brugg", province: "Aargau", longitude: 8.20869, latitude: 47.48096},
	{name: "Birmenstorf", province: "Aargau", longitude: 8.250005, latitude: 47.463884},
	{name: "Dättwil", province: "Aargau", longitude: 8.28474, latitude: 47.45506},
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
