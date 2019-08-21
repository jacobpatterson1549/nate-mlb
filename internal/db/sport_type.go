package db

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
	sportTypeUrls = map[SportType]string{
		SportTypeMlb: "mlb",
		SportTypeNfl: "nfl",
	}
	urlSportTypes = map[string]SportType{
		"mlb": SportTypeMlb,
		"nfl": SportTypeNfl,
	}
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
