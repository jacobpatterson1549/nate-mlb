package request

import (
	"fmt"
	"nate-mlb/internal/db"
	"sort"
	"strings"
)

// ScoreCategorizers maps PlayerTypes to ScoreCategorizers for them
var ScoreCategorizers = map[db.PlayerType]ScoreCategorizer{
	db.PlayerTypeTeam:    mlbTeamRequestor{},
	db.PlayerTypeHitter:  mlbPlayerRequestor{playerType: db.PlayerTypeHitter},
	db.PlayerTypePitcher: mlbPlayerRequestor{playerType: db.PlayerTypePitcher},
	db.PlayerTypeNflTeam: nflTeamRequestor{},
	db.PlayerTypeNflQB:   nflPlayerRequestor{},
	db.PlayerTypeNflRBWR: nflPlayerRequestor{},
}

// ScoreCategorizer requests data for and creates a ScoreCategory for the FriendPlayerInfo
type ScoreCategorizer interface {
	RequestScoreCategory(fpi FriendPlayerInfo, pt db.PlayerType) (ScoreCategory, error)
}

// ScoreCategory contain the FriendScores for each PlayerType
type ScoreCategory struct {
	Name         string
	Description  string
	PlayerTypeID db.PlayerType
	FriendScores []FriendScore
}

// FriendScore contain the scores for a Friend for a PlayerType
type FriendScore struct {
	FriendName   string
	FriendID     int
	PlayerScores []PlayerScore
	Score        int
}

// PlayerScore is the score for a particular Player
type PlayerScore struct {
	PlayerName string
	PlayerID   int
	ID         int
	Score      int
}

// FriendPlayerInfo is a helper pojo of information about what is in a ScoreCategory
type FriendPlayerInfo struct {
	Friends []db.Friend
	Players []db.Player
	Year    int
}

func (sc *ScoreCategory) populate(friends []db.Friend, players []db.Player, playerType db.PlayerType, playerScores map[int]*PlayerScore, onlySumTopTwoPlayerScores bool) error {
	sc.Name = playerType.Name()
	sc.Description = playerType.Description()
	sc.PlayerTypeID = playerType
	sc.FriendScores = make([]FriendScore, len(friends))
	for i, friend := range friends {
		friendScore, err := newFriendScore(friend, players, playerType, playerScores, onlySumTopTwoPlayerScores)
		if err != nil {
			return err
		}
		sc.FriendScores[i] = friendScore
	}
	return nil
}

func newFriendScore(friend db.Friend, players []db.Player, playerType db.PlayerType, playerScores map[int]*PlayerScore, onlySumTopTwoPlayerScores bool) (FriendScore, error) {
	var friendScore FriendScore
	friendScore.FriendName = friend.Name
	friendScore.FriendID = friend.ID // must be done before player scores are populated
	err := friendScore.populatePlayerScores(players, playerType, playerScores)
	if err != nil {
		return friendScore, err
	}
	friendScore.populateScore(onlySumTopTwoPlayerScores)
	return friendScore, nil
}

// GetName implements the server.Tab interface for ScoreCategory
func (sc ScoreCategory) GetName() string {
	return sc.Name
}

// GetID implements the server.Tab interface for ScoreCategory
func (sc ScoreCategory) GetID() string {
	return strings.ReplaceAll(sc.GetName(), " ", "-")
}

func (friendScore *FriendScore) populatePlayerScores(players []db.Player, playerType db.PlayerType, playerScores map[int]*PlayerScore) error {
	for _, player := range players {
		if friendScore.FriendID == player.FriendID && playerType == player.PlayerType {
			playerScore, ok := playerScores[player.PlayerID]
			if !ok {
				return fmt.Errorf("no PlayerScore for player id = %v, type = %v", player.PlayerID, playerType)
			}
			playerScoreWithID := PlayerScore{
				PlayerName: playerScore.PlayerName,
				PlayerID:   playerScore.PlayerID,
				ID:         player.ID,
				Score:      playerScore.Score,
			}
			friendScore.PlayerScores = append(friendScore.PlayerScores, playerScoreWithID)
		}
	}
	return nil
}

func (friendScore *FriendScore) populateScore(onlySumTopTwoPlayerScores bool) {
	scores := make([]int, len(friendScore.PlayerScores))
	for i, playerScore := range friendScore.PlayerScores {
		scores[i] = playerScore.Score
	}
	if onlySumTopTwoPlayerScores && len(scores) > 2 {
		sort.Ints(scores) // ex: 1 2 3 4 5
		scores = scores[len(scores)-2:]
	}
	friendScore.Score = 0
	for _, score := range scores {
		friendScore.Score += score
	}
}
