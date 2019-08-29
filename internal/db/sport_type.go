package db

import "fmt"

// SportType is an enumeration of types of sports
type SportType int

// The expected SportTypes
const (
	SportTypeMlb SportType = 1
	SportTypeNfl SportType = 2
)

var (
	sportTypeNames = make(map[SportType]string)
	sportTypeUrls  = make(map[SportType]string)
	urlSportTypes  = make(map[string]SportType)
)

// Name gets the name for a SportType
func (st SportType) Name() string {
	return sportTypeNames[st]
}

// URL retrieves the url for the SportType
func (st SportType) URL() string {
	return sportTypeUrls[st]
}

// SportTypeFromURL retrieves the SportType for a url
func SportTypeFromURL(url string) SportType {
	return urlSportTypes[url]
}

// LoadSportTypes loads the SportTypes from the database
func LoadSportTypes() ([]SportType, error) {
	rows, err := db.Query(
		`SELECT id, name, url
		FROM sport_types
		ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("problem reading sportTypes: %v", err)
	}
	defer rows.Close()

	var (
		sportTypes []SportType
		sportType  SportType
		name       string
		url        string
	)
	for rows.Next() {
		err = rows.Scan(&sportType, &name, &url)
		if err != nil {
			return nil, fmt.Errorf("problem reading sport type: %v", err)
		}
		sportTypeNames[sportType] = name
		sportTypeUrls[sportType] = url
		urlSportTypes[url] = sportType
		sportTypes = append(sportTypes, sportType)
	}

	_, hasMlbSportType := sportTypeNames[SportTypeMlb]
	_, hasNflSportType := sportTypeNames[SportTypeNfl]
	if len(sportTypes) == 2 && len(sportTypes) == len(sportTypeNames) && len(sportTypeUrls) == len(urlSportTypes) && hasMlbSportType && hasNflSportType {
		return sportTypes, nil
	}
	return nil, fmt.Errorf("Did not load expected SportTypes.  Loaded: %v", sportTypes)
}
