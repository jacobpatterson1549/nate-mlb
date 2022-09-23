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

// GetFriends gets the friends for the active year for a SportType
func (ds Datastore) GetFriends(st SportType) ([]Friend, error) {
	return ds.db.GetFriends(st)
}

func (d sqlDB) GetFriends(st SportType) ([]Friend, error) {
	sqlFunction := newReadSQLFunction("get_friends", []string{"id", "display_order", "name"}, st)
	rows, err := d.db.Query(sqlFunction.sql(), sqlFunction.args...)
	if err != nil {
		return nil, fmt.Errorf("reading friends: %w", err)
	}
	defer rows.Close()

	var friends []Friend
	i := 0
	for rows.Next() {
		friends = append(friends, Friend{})
		err = rows.Scan(&friends[i].ID, &friends[i].DisplayOrder, &friends[i].Name)
		if err != nil {
			return nil, fmt.Errorf("reading friend: %w", err)
		}
		i++
	}
	return friends, nil
}

// SaveFriends saves the specified friends for the active year for a SportType
func (ds Datastore) SaveFriends(st SportType, futureFriends []Friend) error {
	friends, err := ds.GetFriends(st)
	if err != nil {
		return err
	}
	previousFriends := make(map[ID]Friend, len(friends))
	for _, friend := range friends {
		previousFriends[friend.ID] = friend
	}

	insertFriends := make([]Friend, 0, len(futureFriends))
	updateFriends := make([]Friend, 0, len(futureFriends))
	for _, friend := range futureFriends {
		previousFriend, ok := previousFriends[friend.ID]
		switch {
		case !ok:
			insertFriends = append(insertFriends, friend)
		case friend.DisplayOrder != previousFriend.DisplayOrder,
			friend.Name != previousFriend.Name:
			updateFriends = append(updateFriends, friend)
		}
		delete(previousFriends, friend.ID)
	}

	t, err := ds.db.begin()
	if err != nil {
		return err
	}
	for deleteFriendID := range previousFriends {
		t.DelFriend(st, deleteFriendID)
	}
	for _, insertFriend := range insertFriends {
		t.AddFriend(st, insertFriend.DisplayOrder, insertFriend.Name)
	}
	for _, updateFriend := range updateFriends {
		t.SetFriend(st, updateFriend.ID, updateFriend.DisplayOrder, updateFriend.Name)
	}
	return t.execute()
}

func (t *sqlTX) DelFriend(st SportType, id ID) {
	t.queries = append(t.queries, newWriteSQLFunction("del_friend", id, st))
}

func (t *sqlTX) AddFriend(st SportType, displayOrder int, name string) {
	// [friends are added for the active year]
	t.queries = append(t.queries, newWriteSQLFunction("add_friend", displayOrder, name, st))
}

func (t *sqlTX) SetFriend(st SportType, id ID, displayOrder int, name string) {
	t.queries = append(t.queries, newWriteSQLFunction("set_friend", displayOrder, name, id, st))
}
