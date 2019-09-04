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
	sportTypes     = []SportType{}
	sportTypeNames = make(map[SportType]string)
	sportTypeUrls  = make(map[SportType]string)
	urlSportTypes  = make(map[string]SportType)
)

// GetSportTypes gets the SportTypes
func GetSportTypes() []SportType {
	return sportTypes
}

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
func LoadSportTypes() error {
	rows, err := db.Query(`SELECT id, name, url FROM get_sport_types()`)
	if err != nil {
		return fmt.Errorf("problem reading sportTypes: %v", err)
	}
	defer rows.Close()

	var (
		sportType SportType
		name      string
		url       string
	)
	for rows.Next() {
		err = rows.Scan(&sportType, &name, &url)
		if err != nil {
			return fmt.Errorf("problem reading sport type: %v", err)
		}
		sportTypes = append(sportTypes, sportType)
		sportTypeNames[sportType] = name
		sportTypeUrls[sportType] = url
		urlSportTypes[url] = sportType
	}

	_, hasMlbSportType := sportTypeNames[SportTypeMlb]
	_, hasNflSportType := sportTypeNames[SportTypeNfl]
	if len(sportTypes) != 2 ||
		!hasMlbSportType ||
		!hasNflSportType ||
		len(sportTypes) != len(sportTypeNames) ||
		len(sportTypes) != len(sportTypeUrls) ||
		len(sportTypes) != len(urlSportTypes) {
		return fmt.Errorf("Did not load expected SportTypes.  Loaded: %v", sportTypes)
	}
	return nil
}
