package db

import (
	"fmt"
)

type (
	// SportType is an enumeration of types of sports
	SportType int

	sportType struct {
		name string
		url  string
	}
)

// The expected SportTypes
const (
	SportTypeMlb SportType = 1
	SportTypeNfl SportType = 2
)

var (
	sportTypes       = make(map[SportType]sportType)
	urlSportTypes    = make(map[string]SportType)
	loadedSportTypes []SportType
)

// Name gets the name for a SportType
func (st SportType) Name() string {
	return sportTypes[st].name
}

// URL retrieves the url for the SportType
func (st SportType) URL() string {
	return sportTypes[st].url
}

// SportTypeFromURL retrieves the SportType for a url
func SportTypeFromURL(url string) SportType {
	return urlSportTypes[url]
}

func SportTypes() []SportType {
	return loadedSportTypes
}

// LoadSportTypes loads the SportTypes from the database
func LoadSportTypes() error {
	rows, err := db.Query("SELECT id, name, url FROM get_sport_types()")
	if err != nil {
		return fmt.Errorf("reading PlayerTypes for SportTypes: %w", err)
	}
	defer rows.Close()

	var (
		id   SportType
		name string
		url  string
	)
	loadedSportTypes = make([]SportType, 0, 2)
	for rows.Next() {
		err = rows.Scan(&id, &name, &url)
		if err != nil {
			return fmt.Errorf("reading SportType: %w", err)
		}
		sportType := sportType{
			name: name,
			url:  url,
		}
		switch id {
		case SportTypeMlb, SportTypeNfl:
			sportTypes[id] = sportType
			urlSportTypes[url] = id
		default:
			return fmt.Errorf("unknown SportType id: %v", id)
		}
		loadedSportTypes = append(loadedSportTypes, id)
	}

	if len(sportTypes) == 0 {
		return fmt.Errorf("did not load any SportTypes")
	}
	return nil
}
