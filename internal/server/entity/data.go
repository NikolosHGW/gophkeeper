package entity

import "time"

type UserData struct {
	ID       int
	UserID   int
	InfoType string
	Info     string
	Meta     string
	Created  time.Time
}
