package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Feedback struct {
	ID        primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	UserID    *primitive.ObjectID `json:"userId,omitempty" bson:"userId,omitempty"`
	UserName  string              `json:"userName,omitempty" bson:"userName,omitempty"`
	Email     string              `json:"email" bson:"email" binding:"required,email"`
	Rating    int                 `json:"rating" bson:"rating" binding:"required,min=1,max=5"`
	Comment   string              `json:"comment" bson:"comment" binding:"required"`
	Page      string              `json:"page,omitempty" bson:"page,omitempty"`
	UserAgent string              `json:"userAgent,omitempty" bson:"userAgent,omitempty"`
	CreatedAt time.Time           `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time           `json:"updatedAt" bson:"updatedAt"`
}

type CreateFeedbackDto struct {
	UserName  string `json:"userName,omitempty" bson:"userName,omitempty"`
	Email     string `json:"email" validate:"required,email"`
	Rating    int    `json:"rating" validate:"required,min=1,max=5"`
	Comment   string `json:"comment" validate:"required"`
	Page      string `json:"page,omitempty"`
	UserAgent string `json:"userAgent,omitempty"`
}

type UpdateFeedbackDto struct {
	Rating  *int    `json:"rating,omitempty" validate:"omitempty,min=1,max=5"`
	Comment *string `json:"comment,omitempty"`
}

type FeedbackListDto struct {
	BaseDto
	Rating *int    `json:"rating,omitempty"`
	Email  *string `json:"email,omitempty"`
}
