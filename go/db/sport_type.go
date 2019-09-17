package db

import (
	"fmt"
)

// SportType is an enumeration of types of sports
type SportType int

type sportType struct {
	name string
	url  string
}

// The expected SportTypes
const (
	SportTypeMlb SportType = 1
	SportTypeNfl SportType = 2
)

var (
	sportTypes    = make(map[SportType]sportType)
	urlSportTypes = make(map[string]SportType)
)

// Name gets the name for a SportType
func (st SportType) Name() string {
	return sportTypes[st].name
}

// URL retrieves the url for the SportType
func (st SportType) URL() string {
	return sportTypes[st].url
}

// TODO: DELETEME
// // Scan implements the sql.Scanner interface
// func (st *sportType) Scan(src interface{}) error {
// 	id, ok := src.(int)
// 	if !ok {
// 		return fmt.Errorf("could not scan SportType from %v (type %T) - it is not an int", src, src)
// 	}
// 	*st, ok = idSportTypes[id]
// 	if !ok {
// 		return fmt.Errorf("no sport type with id = %v", id)
// 	}
// 	return nil
// }

// // Value implements the driver.Valuer interface
// func (st sportType) Value() (driver.Value, error) {
// 	return int64(st.id), nil
// }

// SportTypeFromURL retrieves the SportType for a url
func SportTypeFromURL(url string) SportType {
	return urlSportTypes[url]
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
	}

	_, hasMlbSportType := sportTypes[SportTypeMlb]
	_, hasNflSportType := sportTypes[SportTypeNfl]
	if len(sportTypes) != 2 ||
		!hasMlbSportType ||
		!hasNflSportType ||
		len(sportTypes) != len(urlSportTypes) {
		return fmt.Errorf("did not load expected SportTypes.  Loaded: %v", sportTypes)
	}
	return nil
}
