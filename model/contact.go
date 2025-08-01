package model

import "time"

type Contact struct {
	Email     string     `json:"email" bson:"email"`
	Name      string     `json:"name" bson:"name"`
	Message   string     `json:"message" bson:"message"`
	Deleted   bool       `json:"deleted" bson:"deleted"`
	DeletedAt *time.Time `json:"deletedAt,omitempty" bson:"deletedAt"`
	DeletedBy *string    `json:"deletedBy,omitempty" bson:"deletedBy"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty" bson:"updatedAt"`
	UpdatedBy *string    `json:"updatedBy,omitempty" bson:"updatedBy"`
	CreatedAt *time.Time `json:"createdAt,omitempty" bson:"createdAt"`
	CreatedBy *string    `json:"createdBy,omitempty" bson:"createdBy"`
	ID        string     `json:"_id" bson:"_id,omitempty"`
}

type CreateContactDto struct {
	Email   string `json:"email" bson:"email"`
	Name    string `json:"name,omitempty" bson:"name"`
	Message string `json:"message,omitempty" bson:"message"` // Note: fixed field name from "address" to "message"
}

type UpdateContactDto struct {
	ID      string `json:"_id" bson:"_id,omitempty"`
	Email   string `json:"email" bson:"email"`
	Name    string `json:"name" bson:"name"`
	Message string `json:"message" bson:"message"`
}

type QueryContactDto struct {
	BaseDto
	Keyword *string `json:"keyword"`
}
