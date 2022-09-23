package db

import (
	"fmt"
)

type (
	// SportType is an enumeration of types of sports
	SportType int

	// SportTypeInfo contains supplementary information about a SportType
	SportTypeInfo struct {
		Name         string
		URL          string
		DisplayOrder int
	}

	// SportTypeMap contains information about multiple SportTypes and their SportTypeInfos
	SportTypeMap map[SportType]SportTypeInfo
)

// The expected SportTypes
const (
	SportTypeMlb SportType = 1
	SportTypeNfl SportType = 2
)

// GetSportTypes returns the SportTypes from the database
func (ds Datastore) GetSportTypes() (SportTypeMap, error) {
	sportTypes, err := ds.db.GetSportTypes()
	if err != nil {
		return nil, err
	}
	if len(sportTypes) != 2 {
		return nil, fmt.Errorf("did not load expected SportTypes: %v", sportTypes)
	}
	return sportTypes, nil
}

func (ds sqlDB) GetSportTypes() (SportTypeMap, error) {
	sqlFunction := newReadSQLFunction("get_sport_types", []string{"id", "name", "url"})
	rows, err := ds.db.Query(sqlFunction.sql(), sqlFunction.args...)
	if err != nil {
		return nil, fmt.Errorf("reading sportTypes: %w", err)
	}
	defer rows.Close()

	var (
		id   SportType
		name string
		url  string
	)
	sportTypes := make(map[SportType]SportTypeInfo)
	displayOrder := 0
	for rows.Next() {
		err = rows.Scan(&id, &name, &url)
		if err != nil {
			return nil, fmt.Errorf("reading SportType: %w", err)
		}
		switch id {
		case SportTypeMlb, SportTypeNfl:
			sportTypes[id] = SportTypeInfo{
				Name:         name,
				URL:          url,
				DisplayOrder: displayOrder,
			}
		default:
			return nil, fmt.Errorf("unknown SportType id: %v", id)
		}
		displayOrder++
	}
	return sportTypes, nil
}
