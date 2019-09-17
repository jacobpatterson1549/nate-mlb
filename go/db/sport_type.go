package db

import (
	"database/sql/driver"
	"fmt"
)

// SportType is an enumeration of types of sports
type SportType interface {
	ID() int
	Name() string
	URL() string
}

type sportType struct {
	id   int
	name string
	url  string
}

// The expected SportTypes
var (
	SportTypeMlb  SportType
	SportTypeNfl  SportType
	idSportTypes  = make(map[int]sportType)
	urlSportTypes = make(map[string]sportType)
)

// ID gets the id for a SportType
func (st sportType) ID() int {
	return st.id
}

// Name gets the name for a SportType
func (st sportType) Name() string {
	return st.name
}

// URL retrieves the url for the SportType
func (st sportType) URL() string {
	return st.url
}

// Scan implements the sql.Scanner interface
func (st *sportType) Scan(src interface{}) error {
	id, ok := src.(int)
	if !ok {
		return fmt.Errorf("could not scan SportType from %v (type %T) - it is not an int", src, src)
	}
	*st, ok = idSportTypes[id]
	if !ok {
		return fmt.Errorf("no sport type with id = %v", id)
	}
	return nil
}

// Value implements the driver.Valuer interface
func (st sportType) Value() (driver.Value, error) {
	return int64(st.id), nil
}

// SportTypeFromURL retrieves the SportType for a url
func SportTypeFromURL(url string) SportType {
	st, ok := urlSportTypes[url]
	if !ok {
		return nil
	}
	return st
}

// LoadSportTypes loads the SportTypes from the database
func LoadSportTypes() error {
	rows, err := db.Query("SELECT id, name, url FROM get_sport_types()")
	if err != nil {
		return fmt.Errorf("reading PlayerTypes for SportTypes: %w", err)
	}
	defer rows.Close()

	var (
		id   int
		name string
		url  string
	)
	for rows.Next() {
		err = rows.Scan(&id, &name, &url)
		if err != nil {
			return fmt.Errorf("reading SportType: %w", err)
		}
		st := sportType{
			id:   id,
			name: name,
			url:  url,
		}
		urlSportTypes[url] = st
		idSportTypes[id] = st
		switch id {
		case 1:
			SportTypeMlb = st
		case 2:
			SportTypeNfl = st
		default:
			return fmt.Errorf("unknown SportType id: %v", id)
		}
	}

	if len(idSportTypes) != 2 ||
		SportTypeNfl == nil ||
		SportTypeMlb == nil ||
		len(idSportTypes) != len(urlSportTypes) {
		return fmt.Errorf("did not load expected SportTypes.  Loaded: %v", idSportTypes)
	}
	return nil
}
