package model

import "time"

type Notification struct {
	ID          string            `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID      string            `json:"userId" bson:"userId"`
	Title       string            `json:"title" bson:"title"`
	Body        string            `json:"body" bson:"body"`
	ImageURL    string            `json:"imageUrl,omitempty" bson:"imageUrl,omitempty"`
	Data        map[string]string `json:"data,omitempty" bson:"data,omitempty"`
	Type        string            `json:"type" bson:"type"` // "chat" | "activity" | "marketing"
	Status      string            `json:"status" bson:"status"` // "pending" | "sent" | "failed"
	ScheduledAt *time.Time        `json:"scheduledAt,omitempty" bson:"scheduledAt,omitempty"`
	SentAt      *time.Time        `json:"sentAt,omitempty" bson:"sentAt,omitempty"`
	ReadAt      *time.Time        `json:"readAt,omitempty" bson:"readAt,omitempty"`
	ClickedAt   *time.Time        `json:"clickedAt,omitempty" bson:"clickedAt,omitempty"`
	SendError   *string           `json:"sendError,omitempty" bson:"sendError,omitempty"`
}

// NotificationTemplate is a reusable admin notification template
type NotificationTemplate struct {
	ID        string            `json:"_id,omitempty" bson:"_id,omitempty"`
	Name      string            `json:"name" bson:"name"`
	Title     string            `json:"title" bson:"title"`
	Body      string            `json:"body" bson:"body"`
	ImageURL  string            `json:"imageUrl,omitempty" bson:"imageUrl,omitempty"`
	Type      string            `json:"type" bson:"type"`
	Data      map[string]string `json:"data,omitempty" bson:"data,omitempty"`
	CreatedAt *time.Time        `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	CreatedBy string            `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	UpdatedAt *time.Time        `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

// SegmentFilter defines criteria for targeting a user segment
type SegmentFilter struct {
	RoleNames    []string `json:"roleNames,omitempty" bson:"roleNames,omitempty"`
	InactiveDays *int     `json:"inactiveDays,omitempty" bson:"inactiveDays,omitempty"`
}

// AdminNotificationLog records an admin broadcast or segment send event
type AdminNotificationLog struct {
	ID            string            `json:"_id,omitempty" bson:"_id,omitempty"`
	Title         string            `json:"title" bson:"title"`
	Body          string            `json:"body" bson:"body"`
	ImageURL      string            `json:"imageUrl,omitempty" bson:"imageUrl,omitempty"`
	Type          string            `json:"type" bson:"type"`
	Data          map[string]string `json:"data,omitempty" bson:"data,omitempty"`
	SentTo        string            `json:"sentTo" bson:"sentTo"` // "all" | "segment"
	SegmentFilter *SegmentFilter    `json:"segmentFilter,omitempty" bson:"segmentFilter,omitempty"`
	ScheduledAt   *time.Time        `json:"scheduledAt,omitempty" bson:"scheduledAt,omitempty"`
	SentAt        *time.Time        `json:"sentAt,omitempty" bson:"sentAt,omitempty"`
	TotalSent     int               `json:"totalSent" bson:"totalSent"`
	TotalFailed   int               `json:"totalFailed" bson:"totalFailed"`
	CreatedBy     string            `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
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

// SendBroadcastDto is used by admin to broadcast to all users
type SendBroadcastDto struct {
	Title       string            `json:"title" validate:"required"`
	Body        string            `json:"body" validate:"required"`
	ImageURL    string            `json:"imageUrl,omitempty"`
	Data        map[string]string `json:"data,omitempty"`
	Type        string            `json:"type" validate:"required,oneof=chat activity marketing"`
	ScheduledAt *time.Time        `json:"scheduledAt,omitempty"`
}

// SendSegmentDto is used by admin to send to a filtered user segment
type SendSegmentDto struct {
	Title         string            `json:"title" validate:"required"`
	Body          string            `json:"body" validate:"required"`
	ImageURL      string            `json:"imageUrl,omitempty"`
	Data          map[string]string `json:"data,omitempty"`
	Type          string            `json:"type" validate:"required,oneof=chat activity marketing"`
	ScheduledAt   *time.Time        `json:"scheduledAt,omitempty"`
	SegmentFilter SegmentFilter     `json:"segmentFilter"`
}

// CreateNotificationTemplateDto is used to create a new template
type CreateNotificationTemplateDto struct {
	Name     string            `json:"name" validate:"required"`
	Title    string            `json:"title" validate:"required"`
	Body     string            `json:"body" validate:"required"`
	ImageURL string            `json:"imageUrl,omitempty"`
	Type     string            `json:"type" validate:"required,oneof=chat activity marketing"`
	Data     map[string]string `json:"data,omitempty"`
}

// UpdateNotificationTemplateDto is used to update an existing template
type UpdateNotificationTemplateDto struct {
	Name     *string           `json:"name,omitempty"`
	Title    *string           `json:"title,omitempty"`
	Body     *string           `json:"body,omitempty"`
	ImageURL *string           `json:"imageUrl,omitempty"`
	Type     *string           `json:"type,omitempty"`
	Data     map[string]string `json:"data,omitempty"`
}

// QueryAdminLogDto provides pagination for admin logs
type QueryAdminLogDto struct {
	Page  int    `query:"page"`
	Limit int    `query:"limit"`
	SentTo string `query:"sentTo"`
}
