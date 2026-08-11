package main

// BuildUniversalCities converts every non-home national location into an
// international location. This produces one globe containing the complete
// Forecast Channel location database while preserving each city's real country
// and province labels.
func BuildUniversalCities(list *WeatherList, homeCountry string) []InternationalCity {
	cities := make([]InternationalCity, 0, len(list.International.Cities)+4000)
	seen := make(map[string]struct{}, len(list.International.Cities)+4000)

	add := func(city InternationalCity) {
		if city.Country.English == homeCountry {
			return
		}
		key := city.Country.English + "\x00" + city.Province.English + "\x00" + city.Name.English
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		cities = append(cities, city)
	}

	for _, city := range list.International.Cities {
		add(city)
	}

	for _, country := range list.National {
		if country.Name.English == homeCountry {
			continue
		}
		for _, city := range country.Cities {
			add(InternationalCity{
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
			})
		}
	}

	return cities
}
