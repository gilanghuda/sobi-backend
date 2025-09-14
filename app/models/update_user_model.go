package models

type UpdateUserRequest struct {
	Username    *string `json:"username"`
	PhoneNumber *string `json:"phone_number"`
	Gender      *string `json:"gender"`
	Avatar      *int    `json:"avatar"`
}
