package db

import (
	"database/sql"
	"fmt"
)

// Friend contains the name of the person in the pool.
type Friend struct {
	ID           int
	DisplayOrder int
	Name         string
}

// TODO: use shared logic to request friends, playerTypes, players (but with helper mapper functions) (look at how NullString scanning works)
func getFriends() ([]Friend, error) {
	rows, err := db.Query("SELECT f.id, f.display_order, f.name FROM friends AS f JOIN stats AS s ON f.year = s.year WHERE s.active ORDER BY f.display_order ASC")
	if err != nil {
		return nil, fmt.Errorf("problem reading friends: %v", err)
	}
	defer rows.Close()

	friends := []Friend{}
	i := 0
	for rows.Next() {
		friends = append(friends, Friend{})
		err = rows.Scan(&friends[i].ID, &friends[i].DisplayOrder, &friends[i].Name)
		if err != nil {
			return nil, fmt.Errorf("problem reading data: %v", err)
		}
		i++
	}
	return friends, nil
}

// SetFriends saves the specied players in for the active year. TODO: rename to SaveFriends
func SetFriends(futureFriends []Friend) error {
	friends, err := getFriends()
	if err != nil {
		return err
	}
	previousFriends := make(map[int]Friend)
	for _, friend := range friends {
		previousFriends[friend.ID] = friend
	}

	insertFriends := []Friend{}
	updateFriends := []Friend{}
	for _, friend := range futureFriends {
		previousFriend, ok := previousFriends[friend.ID]
		if !ok {
			insertFriends = append(insertFriends, friend)
		} else if friend.DisplayOrder != previousFriend.DisplayOrder || friend.Name != previousFriend.Name {
			updateFriends = append(updateFriends, friend)
		}
		delete(previousFriends, friend.ID)
	}
	deleteFriends := previousFriends

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("problem starting transaction: %v", err)
	}
	var result sql.Result
	for _, friend := range insertFriends {
		if err == nil {
			result, err = tx.Exec(
				"INSERT INTO friends (display_order, name, year) SELECT $1, $2, year FROM stats AS s WHERE s.active",
				friend.DisplayOrder,
				friend.Name)
			if err == nil {
				err = expectSingleRowAffected(result)
			}
		}
	}
	for _, friend := range updateFriends {
		if err == nil {
			result, err = tx.Exec(
				"UPDATE friends SET display_order = $1, name = $2 WHERE id = $3",
				friend.DisplayOrder,
				friend.Name,
				friend.ID)
			if err == nil {
				err = expectSingleRowAffected(result)
			}
		}
	}
	for friendID := range deleteFriends {
		if err == nil {
			result, err = tx.Exec(
				"DELETE FROM friends WHERE id = $1",
				friendID)
			if err == nil {
				err = expectSingleRowAffected(result)
			}
		}
	}
	if err != nil {
		if err2 := tx.Rollback(); err2 != nil {
			err = fmt.Errorf("problem: %v, ROLLBACK ERROR: %v", err.Error(), err2.Error())
		}
	} else {
		err = tx.Commit()
	}
	if err != nil {
		return fmt.Errorf("problem saving friends: %v", err)
	}
	return nil
}
