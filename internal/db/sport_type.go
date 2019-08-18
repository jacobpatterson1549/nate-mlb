package db

import "strings"

// SportType is an enumeration of types of sports
type SportType int

// The SportTypes
const (
	SportTypeMlb SportType = 1
	SportTypeNfl SportType = 2
)

var (
	sportTypeNames = map[SportType]string{
		SportTypeMlb: "MLB",
		SportTypeNfl: "NFL",
	}
)

// Name gets the name for a SportType
func (st SportType) Name() string {
	return sportTypeNames[st]
}

// GetSportType retrieves a SportType from it's name
func GetSportType(query string) SportType {
	query = strings.ToUpper(query)
	for st, name := range sportTypeNames {
		if strings.ToUpper(name) == query {
			return st
		}
	}
	return 0
}
