package db

import (
	"fmt"
)

// Friend contains the name of the person in the pool.
type Friend struct {
	ID           ID
	DisplayOrder int
	Name         string
}

// GetFriends gets the friends for the active year
func GetFriends(st SportType) ([]Friend, error) {
	rows, err := db.Query(
		`SELECT f.id, f.display_order, f.name
		FROM stats AS s
		JOIN friends AS f ON s.year = f.year AND s.sport_type_id = f.sport_type_id
		WHERE s.sport_type_id = $1
		AND s.active`,
		st,
	)
	if err != nil {
		return nil, fmt.Errorf("problem reading friends: %v", err)
	}
	defer rows.Close()

	var friends []Friend
	i := 0
	for rows.Next() {
		friends = append(friends, Friend{})
		err = rows.Scan(&friends[i].ID, &friends[i].DisplayOrder, &friends[i].Name)
		if err != nil {
			return nil, fmt.Errorf("problem reading friend: %v", err)
		}
		i++
	}
	return friends, nil
}

// SaveFriends saves the specified friends for the active year
func SaveFriends(st SportType, futureFriends []Friend) error {
	friends, err := GetFriends(st)
	if err != nil {
		return err
	}
	previousFriends := make(map[ID]Friend, len(friends))
	for _, friend := range friends {
		previousFriends[friend.ID] = friend
	}

	var insertFriends []Friend
	var updateFriends []Friend
	for _, friend := range futureFriends {
		previousFriend, ok := previousFriends[friend.ID]
		if !ok {
			insertFriends = append(insertFriends, friend)
		} else if friend.DisplayOrder != previousFriend.DisplayOrder || friend.Name != previousFriend.Name {
			updateFriends = append(updateFriends, friend)
		}
		delete(previousFriends, friend.ID)
	}

	queries := make(chan query, len(insertFriends)+len(updateFriends)+len(previousFriends))
	quit := make(chan error)
	go exececuteInTransaction(queries, quit)
	for deleteFriendID := range previousFriends {
		queries <- newQuery(
			`DELETE FROM friends
			WHERE id = $1`,
			deleteFriendID,
		)
	}
	for _, insertFriend := range insertFriends {
		// [friends are added for the active year]
		queries <- newQuery(
			`INSERT INTO friends
			(display_order, name, sport_type_id, year)
			SELECT $1, $2, $3, year
			FROM stats
			WHERE sport_type_id = $3
			AND active`,
			insertFriend.DisplayOrder,
			insertFriend.Name,
			st,
		)
	}
	for _, updateFriend := range updateFriends {
		queries <- newQuery(
			`UPDATE friends
			SET display_order = $1
			, name = $2
			WHERE id = $3`,
			updateFriend.DisplayOrder,
			updateFriend.Name,
			updateFriend.ID,
		)
	}
	close(queries)
	return <-quit
}
