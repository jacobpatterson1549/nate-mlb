package db

// TODO: audit need for this struct  could stats just call the getters when needed and pass around necessary []item lists?

// FriendPlayerInfo contain all the pool items for each Friend.
type FriendPlayerInfo struct {
	Friends     []Friend
	PlayerTypes []PlayerType
	Players     []Player
	Year        int
}

// GetFriendPlayerInfo retrieves playerTypeas and active friends, players, and year from the database
func GetFriendPlayerInfo() (FriendPlayerInfo, error) {
	fpi := FriendPlayerInfo{}

	friends, err := getFriends()
	if err != nil {
		return fpi, err
	}
	playerTypes, err := getPlayerTypes()
	if err != nil {
		return fpi, err
	}
	players, err := getPlayers()
	if err != nil {
		return fpi, err
	}
	activeYear, err := GetActiveYear()
	if err != nil {
		return fpi, err
	}

	fpi.Friends = friends
	fpi.PlayerTypes = playerTypes
	fpi.Players = players
	fpi.Year = activeYear
	return fpi, nil
}
