package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RefreshTokenBlackList represents a blacklisted refresh token in MongoDB
// Collection name: refresh_token_black_list
// This is used to store tokens that are no longer valid (e.g., after logout).
type RefreshTokenBlackList struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Token     string             `bson:"token" json:"token"`
	UserID    primitive.ObjectID `bson:"userId" json:"user_id"`
	ExpiredAt time.Time          `bson:"expiredAt" json:"expired_at"`
	CreatedAt time.Time          `bson:"createdAt" json:"created_at"`
}
