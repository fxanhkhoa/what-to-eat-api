package model

import "time"

type DeviceToken struct {
	Token      string     `json:"token" bson:"token"`
	Platform   string     `json:"platform" bson:"platform"` // "web" | "android"
	DeviceInfo string     `json:"deviceInfo,omitempty" bson:"deviceInfo,omitempty"`
	LastUsed   *time.Time `json:"lastUsed,omitempty" bson:"lastUsed,omitempty"`
	CreatedAt  *time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}
