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
)

// Name gets the name for a SportType
func (st SportType) Name() string {
	return sportTypeNames[st]
}
