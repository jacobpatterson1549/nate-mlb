package db

import (
	"fmt"
)

// Friend contains the name of the person in the pool.
type Friend struct {
	ID           int
	DisplayOrder int
	Name         string
}

// GetFriends gets the friends for the active year
func GetFriends(st SportType) ([]Friend, error) {
	rows, err := db.Query(
		"SELECT f.id, f.display_order, f.name FROM friends AS f JOIN stats AS s ON f.year = s.year WHERE s.sport_type_id = $1 AND s.active ORDER BY f.display_order ASC",
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
	previousFriends := make(map[int]Friend)
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

	queries := make([]query, len(insertFriends)+len(updateFriends)+len(previousFriends))
	i := 0
	for _, insertFriend := range insertFriends {
		queries[i] = query{
			sql:  "INSERT INTO friends (display_order, name, sport_type_id, year) SELECT $1, $2, $3 year FROM stats AS s WHERE s.sport_type_id = %3 AND s.active",
			args: make([]interface{}, 2),
		}
		queries[i].args[0] = insertFriend.DisplayOrder
		queries[i].args[1] = insertFriend.Name
		queries[i].args[2] = st
		i++
	}
	for _, updateFriend := range updateFriends {
		queries[i] = query{
			sql:  "UPDATE friends SET display_order = $1, name = $2 WHERE id = $3",
			args: make([]interface{}, 3),
		}
		queries[i].args[0] = updateFriend.DisplayOrder
		queries[i].args[1] = updateFriend.Name
		queries[i].args[2] = updateFriend.ID
		i++
	}
	for deleteFriendID := range previousFriends {
		queries[i] = query{
			sql:  "DELETE FROM friends WHERE id = $1",
			args: make([]interface{}, 1),
		}
		queries[i].args[0] = deleteFriendID
		i++
	}
	return exececuteInTransaction(&queries)
}
