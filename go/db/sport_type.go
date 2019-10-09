package db

import (
	"fmt"
	"sort"
)

type (
	// SportType is an enumeration of types of sports
	SportType int

	sportType struct {
		name         string
		url          string
		displayOrder int
	}
)

// The expected SportTypes
const (
	SportTypeMlb SportType = 1
	SportTypeNfl SportType = 2
)

var sportTypes = make(map[SportType]sportType)

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
	for st := range sportTypes {
		if st.URL() == url {
			return st
		}
	}
	return 0
}

// DisplayOrder gets the display order for a SportType
func (st SportType) DisplayOrder() int {
	return sportTypes[st].displayOrder
}

// SportTypes returns the loaded SportTypes
func SportTypes() []SportType {
	sportTypesList := make([]SportType, 0, len(sportTypes))
	for st := range sportTypes {
		sportTypesList = append(sportTypesList, st)
	}
	sort.Slice(sportTypesList, func(i, j int) bool {
		return sportTypesList[i].DisplayOrder() < sportTypesList[j].DisplayOrder()
	})
	return sportTypesList
}

// LoadSportTypes loads the SportTypes from the database
func LoadSportTypes() error {
	sqlFunction := newReadSQLFunction("get_sport_types", []string{"id", "name", "url"})
	rows, err := db.Query(sqlFunction.sql(), sqlFunction.args...)
	if err != nil {
		return fmt.Errorf("reading sportTypes: %w", err)
	}
	defer rows.Close()

	var (
		id   SportType
		name string
		url  string
	)
	sportTypes = make(map[SportType]sportType)
	displayOrder := 0
	for rows.Next() {
		err = rows.Scan(&id, &name, &url)
		if err != nil {
			return fmt.Errorf("reading SportType: %w", err)
		}
		sportType := sportType{
			name:         name,
			url:          url,
			displayOrder: displayOrder,
		}
		switch id {
		case SportTypeMlb, SportTypeNfl:
			sportTypes[id] = sportType
		default:
			return fmt.Errorf("unknown SportType id: %v", id)
		}
		displayOrder++
	}

	_, hasMlbSportType := sportTypes[SportTypeMlb]
	_, hasNflSportType := sportTypes[SportTypeNfl]
	if len(sportTypes) != 2 ||
		!hasNflSportType ||
		!hasMlbSportType {
		return fmt.Errorf("did not load expected SportTypes.  Loaded: %v", sportTypes)
	}
	return nil
}
