package request

import (
	"nate-mlb/internal/db"
)

// Requestors maps PlayerTypes to Requestors for them
var Requestors = map[db.PlayerType]Requestor{
	db.PlayerTypeTeam:    mlbTeamRequestor{},
	db.PlayerTypeHitter:  mlbPlayerRequestor{playerType: db.PlayerTypeHitter},
	db.PlayerTypePitcher: mlbPlayerRequestor{playerType: db.PlayerTypePitcher},
}

// Requestor requests data for and creates a ScoreCategory for the FriendPlayerInfo
// TODO: rename to scoreCategorizer
type Requestor interface {
	RequestScoreCategory(fpi FriendPlayerInfo, pt db.PlayerType) (ScoreCategory, error)
}
