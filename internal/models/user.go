package models

import "time"

type User struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Type       *string   `json:"type,omitempty"`
	Email      string    `json:"email"`
	Password   string    `json:"-"`
	Role       *string   `json:"role,omitempty"`
	Department *string   `json:"department,omitempty"`
	Phone      *string   `json:"phone,omitempty"`
	Location   *string   `json:"location,omitempty"`
	Shift      *string   `json:"shift,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type RegisterRequest struct {
	Name                 string  `json:"name" binding:"required,max=255"`
	Type                 *string `json:"type,omitempty" binding:"omitempty,max=255"`
	Email                string  `json:"email" binding:"required,email,max=255"`
	Password             string  `json:"password" binding:"required,min=6"`
	PasswordConfirmation string  `json:"password_confirmation" binding:"required,eqfield=Password"`
	Role                 *string `json:"role,omitempty" binding:"omitempty,max=255"`
	Department           *string `json:"department,omitempty" binding:"omitempty,max=255"`
	Phone                *string `json:"phone,omitempty" binding:"omitempty,max=50"`
	Location             *string `json:"location,omitempty" binding:"omitempty,max=255"`
	Shift                *string `json:"shift,omitempty" binding:"omitempty,max=50"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateUserRequest struct {
	Name       *string `json:"name,omitempty" binding:"omitempty,max=255"`
	Type       *string `json:"type,omitempty" binding:"omitempty,max=255"`
	Email      *string `json:"email,omitempty" binding:"omitempty,email,max=255"`
	Password   *string `json:"password,omitempty" binding:"omitempty,min=6"`
	Role       *string `json:"role,omitempty" binding:"omitempty,max=255"`
	Department *string `json:"department,omitempty" binding:"omitempty,max=255"`
	Phone      *string `json:"phone,omitempty" binding:"omitempty,max=50"`
	Location   *string `json:"location,omitempty" binding:"omitempty,max=255"`
	Shift      *string `json:"shift,omitempty" binding:"omitempty,max=50"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
