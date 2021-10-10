package models

import "time"

type User struct {
	Username       string    `json:"username"`
	HashedPassword *[]byte   `json:"hashedPassword,omitempty"`
	Created        time.Time `json:"created"`
}
