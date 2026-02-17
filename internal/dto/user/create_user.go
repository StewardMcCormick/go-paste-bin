package user

import "time"

type CreateUserRequest struct {
	Id       int64  `json:"id" validate:"required"`
	Username string `json:"username" validate:"required,min=3,max=30"`
	Password string `json:"password" validate:"required,len=60"`
}

type CreateUserResponse struct {
	Id        int64     `json:"id" validate:"required"`
	Username  string    `json:"username" validate:"required,min=3,max=30"`
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
}
