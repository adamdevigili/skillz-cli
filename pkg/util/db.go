package util

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/adamdevigili/skillz-cli/pkg/models"
	"github.com/boltdb/bolt"
)

// TODO: Create DB interface that wraps the BoltDB client, and these functions become methods on "SkillzDB" struct

func GetUsersDB(dbPath, usersBucket, currentUserBucket string) (*bolt.DB, error) {
	// Open DB and check for bucket, create if not-present
	db, err := bolt.Open(dbPath, 0640, nil)
	if err != nil {
		return nil, err
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(usersBucket))
		return err
	}); err != nil {
		return nil, err
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(currentUserBucket))
		return err
	}); err != nil {
		return nil, err
	}

	return db, err
}

// GetUser gets the details of a user by username
func GetUser(db *bolt.DB, usersBucket, username string) (*models.User, error) {
	var currentUser *models.User

	err := db.View(func(tx *bolt.Tx) error {
		currentUserBytes := tx.Bucket([]byte(usersBucket)).Get([]byte(username))
		if currentUserBytes != nil {
			if err := json.Unmarshal(currentUserBytes, &currentUser); err != nil {
				fmt.Println(fmt.Sprintf("%+v", err))
				return err
			}
		}

		return nil
	})

	return currentUser, err
}

func AddNewUser(db *bolt.DB, usersBucket, username string, hashedPassword []byte) (*models.User, error) {
	newUser := &models.User{
		Username:       username,
		HashedPassword: &hashedPassword,
		Created:        time.Now(),
	}

	newUserBytes, _ := json.Marshal(*newUser)

	if err := db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(usersBucket)).Put([]byte(username), newUserBytes)
	}); err != nil {
		return nil, err
	}

	return newUser, nil
}

func UpdateUser(db *bolt.DB, usersBucket, username string, user *models.User) error {
	userBytes, _ := json.Marshal(*user)
	return db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(usersBucket)).Put([]byte(username), userBytes)
	})
}

// GetCurrentUser gets the details of the currently logged-in user
func GetCurrentUser(db *bolt.DB, currentUserBucket, currentUserKey string) (*models.User, error) {
	var currentUser *models.User

	err := db.View(func(tx *bolt.Tx) error {
		currentUserBytes := tx.Bucket([]byte(currentUserBucket)).Get([]byte(currentUserKey))
		if currentUserBytes != nil {
			if err := json.Unmarshal(currentUserBytes, &currentUser); err != nil {
				return err
			}
		}

		return nil
	})

	return currentUser, err
}

// SetCurrentUser sets the current user in the DB as well. This would probably be stored in a more temporary, auto-expiring local cache
func SetCurrentUser(db *bolt.DB, currentUserBucket, currentUserKey string, currentUser *models.User) error {
	currentUserBytes, _ := json.Marshal(*currentUser)
	return db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(currentUserBucket)).Put([]byte(currentUserKey), currentUserBytes)
	})
}

// UnsetCurrentUser unsets the current user in the DB (logout)
func UnsetCurrentUser(db *bolt.DB, currentUserBucket, currentUserKey string) error {
	return db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(currentUserBucket)).Delete([]byte(currentUserKey))
	})
}
