package model

import "time"

type Notification struct {
	ID        string            `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID    string            `json:"userId" bson:"userId"`
	Title     string            `json:"title" bson:"title"`
	Body      string            `json:"body" bson:"body"`
	ImageURL  string            `json:"imageUrl,omitempty" bson:"imageUrl,omitempty"`
	Data      map[string]string `json:"data,omitempty" bson:"data,omitempty"`
	Type      string            `json:"type" bson:"type"` // "chat" | "activity" | "marketing"
	SentAt    *time.Time        `json:"sentAt,omitempty" bson:"sentAt,omitempty"`
	ReadAt    *time.Time        `json:"readAt,omitempty" bson:"readAt,omitempty"`
	ClickedAt *time.Time        `json:"clickedAt,omitempty" bson:"clickedAt,omitempty"`
	SendError *string           `json:"resultFromFirebase,omitempty" bson:"resultFromFirebase,omitempty"`
}

type NotificationPreference struct {
	ID               string  `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID           string  `json:"userId" bson:"userId"`
	ChatEnabled      bool    `json:"chatEnabled" bson:"chatEnabled"`
	ActivityEnabled  bool    `json:"activityEnabled" bson:"activityEnabled"`
	MarketingEnabled bool    `json:"marketingEnabled" bson:"marketingEnabled"`
	QuietHoursStart  *string `json:"quietHoursStart,omitempty" bson:"quietHoursStart,omitempty"` // "22:00"
	QuietHoursEnd    *string `json:"quietHoursEnd,omitempty" bson:"quietHoursEnd,omitempty"`     // "08:00"
}

// DTOs
type RegisterDeviceTokenDto struct {
	Token      string `json:"token" validate:"required"`
	Platform   string `json:"platform" validate:"required,oneof=web android"`
	DeviceInfo string `json:"deviceInfo,omitempty"`
}

type UnregisterDeviceTokenDto struct {
	Token string `json:"token" validate:"required"`
}

type SendNotificationDto struct {
	UserID   string            `json:"userId" validate:"required"`
	Title    string            `json:"title" validate:"required"`
	Body     string            `json:"body" validate:"required"`
	ImageURL string            `json:"imageUrl,omitempty"`
	Data     map[string]string `json:"data,omitempty"`
	Type     string            `json:"type" validate:"required,oneof=chat activity marketing"`
}

type UpdateNotificationPreferenceDto struct {
	ChatEnabled      *bool   `json:"chatEnabled,omitempty"`
	ActivityEnabled  *bool   `json:"activityEnabled,omitempty"`
	MarketingEnabled *bool   `json:"marketingEnabled,omitempty"`
	QuietHoursStart  *string `json:"quietHoursStart,omitempty"`
	QuietHoursEnd    *string `json:"quietHoursEnd,omitempty"`
}
