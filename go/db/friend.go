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
	sqlFunction := newReadSQLFunction("get_friends", []string{"id", "display_order", "name"}, st)
	rows, err := db.Query(sqlFunction.sql(), sqlFunction.args...)
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

	queries := make(chan writeSQLFunction, len(insertFriends)+len(updateFriends)+len(previousFriends))
	quit := make(chan error)
	go executeInTransaction(queries, quit)
	for deleteFriendID := range previousFriends {
		queries <- newWriteSQLFunction("del_friend", deleteFriendID)
	}
	for _, insertFriend := range insertFriends {
		// [friends are added for the active year]
		queries <- newWriteSQLFunction("add_friend", insertFriend.DisplayOrder, insertFriend.Name, st)
	}
	for _, updateFriend := range updateFriends {
		queries <- newWriteSQLFunction("set_friend", updateFriend.DisplayOrder, updateFriend.Name, updateFriend.ID)
	}
	close(queries)
	return <-quit
}