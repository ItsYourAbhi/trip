// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Destination struct {
	ID          pgtype.UUID
	Name        string
	Description string
	Attraction  string
	PicUrl      string
}

type Trip struct {
	ID            pgtype.UUID
	Name          string
	StartDate     pgtype.Date
	EndDate       pgtype.Date
	DestinationID pgtype.UUID
}

type User struct {
	ID       pgtype.UUID
	Email    string
	Name     string
	Password string
	Admin    bool
}
