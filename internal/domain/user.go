package domain

import (
	"time"
)

type User struct {
	Id        int64     `validate:"required"`
	Username  string    `validate:"required,min=3,max=30"`
	Password  string    `validate:"required,len=60"`
	CreatedAt time.Time `validate:"required"`
}
