package model

import "time"

type User struct {
	Email       string     `json:"email" bson:"email"`
	Password    *string    `json:"password,omitempty" bson:"password"`
	Name        *string    `json:"name,omitempty" bson:"name"`
	DateOfBirth *time.Time `json:"dateOfBirth,omitempty" bson:"dateOfBirth"`
	Address     *string    `json:"address,omitempty" bson:"address"`
	Phone       *string    `json:"phone,omitempty" bson:"phone"`
	GoogleID    *string    `json:"googleID,omitempty" bson:"googleID"`
	FacebookID  *string    `json:"facebookID,omitempty" bson:"facebookID"`
	GithubID    *string    `json:"githubID,omitempty" bson:"githubID"`
	Avatar      *string    `json:"avatar,omitempty" bson:"avatar"`
	Deleted     bool       `json:"deleted" bson:"deleted"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty" bson:"deletedAt"`
	DeletedBy   *string    `json:"deletedBy,omitempty" bson:"deletedBy"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty" bson:"updatedAt"`
	UpdatedBy   *string    `json:"updatedBy,omitempty" bson:"updatedBy"`
	CreatedAt   *time.Time `json:"createdAt,omitempty" bson:"createdAt"`
	CreatedBy   *string    `json:"createdBy,omitempty" bson:"createdBy"`
	ID          string     `json:"_id" bson:"_id,omitempty"`
	RoleName    string     `json:"roleName" bson:"roleName"`
}

type QueryUserDto struct {
	BaseDto
	Keyword     string   `json:"keyword"`
	Email       string   `json:"email"`
	PhoneNumber string   `json:"phoneNumber"`
	RoleName    []string `json:"roleName"`
}

type CreateUserDto struct {
	Email       string     `json:"email" bson:"email"`
	Password    *string    `json:"password,omitempty" bson:"password"`
	Name        *string    `json:"name,omitempty" bson:"name"`
	DateOfBirth *time.Time `json:"dateOfBirth,omitempty" bson:"dateOfBirth"`
	Address     *string    `json:"address,omitempty" bson:"address"`
	Phone       *string    `json:"phone,omitempty" bson:"phone"`
	GoogleID    *string    `json:"googleID,omitempty" bson:"googleID"`
	FacebookID  *string    `json:"facebookID,omitempty" bson:"facebookID"`
	GithubID    *string    `json:"githubID,omitempty" bson:"githubID"`
	Avatar      *string    `json:"avatar,omitempty" bson:"avatar"`
}

type UpdateUserDto struct {
	ID          string     `json:"id" bson:"id"`
	Email       string     `json:"email" bson:"email"`
	Name        *string    `json:"name,omitempty" bson:"name"`
	DateOfBirth *time.Time `json:"dateOfBirth,omitempty" bson:"dateOfBirth"`
	Address     *string    `json:"address,omitempty" bson:"address"`
	Phone       *string    `json:"phone,omitempty" bson:"phone"`
	GoogleID    *string    `json:"googleID,omitempty" bson:"googleID"`
	FacebookID  *string    `json:"facebookID,omitempty" bson:"facebookID"`
	GithubID    *string    `json:"githubID,omitempty" bson:"githubID"`
	Avatar      *string    `json:"avatar,omitempty" bson:"avatar"`
}
