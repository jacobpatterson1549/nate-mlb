package request

import (
	"nate-mlb/internal/db"
)

// ScoreCategorizers maps PlayerTypes to ScoreCategorizers for them
var ScoreCategorizers = map[db.PlayerType]ScoreCategorizer{
	db.PlayerTypeTeam:    mlbTeamRequestor{},
	db.PlayerTypeHitter:  mlbPlayerRequestor{playerType: db.PlayerTypeHitter},
	db.PlayerTypePitcher: mlbPlayerRequestor{playerType: db.PlayerTypePitcher},
}

// ScoreCategorizer requests data for and creates a ScoreCategory for the FriendPlayerInfo
// TODO: move to stats.go
type ScoreCategorizer interface {
	RequestScoreCategory(fpi FriendPlayerInfo, pt db.PlayerType) (ScoreCategory, error)
}
