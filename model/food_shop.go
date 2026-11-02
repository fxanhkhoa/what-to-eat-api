package model

import "time"

// OpeningHour represents the opening hours for a specific day of the week
type OpeningHour struct {
	Day    string `json:"day" bson:"day"`       // "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"
	Open   string `json:"open" bson:"open"`     // "08:00"
	Close  string `json:"close" bson:"close"`   // "22:00"
	Closed bool   `json:"closed" bson:"closed"` // true if closed on this day
}

// FoodShop represents a food shop or restaurant that serves dishes
type FoodShop struct {
	ID             string         `json:"_id" bson:"_id,omitempty"`
	Name           string         `json:"name" bson:"name"`
	Description    *string        `json:"description,omitempty" bson:"description"`
	Phone          *string        `json:"phone,omitempty" bson:"phone"`
	Website        *string        `json:"website,omitempty" bson:"website"`
	Thumbnail      *string        `json:"thumbnail,omitempty" bson:"thumbnail"`
	Images         []*string      `json:"images,omitempty" bson:"images"`
	Address        *string        `json:"address,omitempty" bson:"address"`
	Latitude       *float64       `json:"latitude,omitempty" bson:"latitude"`
	Longitude      *float64       `json:"longitude,omitempty" bson:"longitude"`
	OpeningHours   []*OpeningHour `json:"openingHours,omitempty" bson:"openingHours"`
	IsOpen         *bool          `json:"isOpen,omitempty" bson:"isOpen"`         // manual open/closed override
	LastVerifiedAt *time.Time     `json:"lastVerifiedAt,omitempty" bson:"lastVerifiedAt"` // when status was last confirmed
	VerifiedBy     *string        `json:"verifiedBy,omitempty" bson:"verifiedBy"` // userID who last verified
	Deleted        bool           `json:"deleted" bson:"deleted"`
	DeletedAt      *time.Time     `json:"deletedAt,omitempty" bson:"deletedAt"`
	DeletedBy      *string        `json:"deletedBy,omitempty" bson:"deletedBy"`
	UpdatedAt      *time.Time     `json:"updatedAt,omitempty" bson:"updatedAt"`
	UpdatedBy      *string        `json:"updatedBy,omitempty" bson:"updatedBy"`
	CreatedAt      *time.Time     `json:"createdAt,omitempty" bson:"createdAt"`
	CreatedBy      *string        `json:"createdBy,omitempty" bson:"createdBy"`
}

// CreateFoodShopDto is used for creating a new food shop
type CreateFoodShopDto struct {
	Name         string         `json:"name" bson:"name"`
	Description  *string        `json:"description,omitempty" bson:"description"`
	Phone        *string        `json:"phone,omitempty" bson:"phone"`
	Website      *string        `json:"website,omitempty" bson:"website"`
	Thumbnail    *string        `json:"thumbnail,omitempty" bson:"thumbnail"`
	Images       []*string      `json:"images,omitempty" bson:"images"`
	Address      *string        `json:"address,omitempty" bson:"address"`
	Latitude     *float64       `json:"latitude,omitempty" bson:"latitude"`
	Longitude    *float64       `json:"longitude,omitempty" bson:"longitude"`
	OpeningHours []*OpeningHour `json:"openingHours,omitempty" bson:"openingHours"`
	IsOpen       *bool          `json:"isOpen,omitempty" bson:"isOpen"`
}

// UpdateFoodShopDto is used for updating an existing food shop
type UpdateFoodShopDto struct {
	ID           string         `json:"_id" bson:"_id,omitempty"`
	Name         *string        `json:"name,omitempty" bson:"name"`
	Description  *string        `json:"description,omitempty" bson:"description"`
	Phone        *string        `json:"phone,omitempty" bson:"phone"`
	Website      *string        `json:"website,omitempty" bson:"website"`
	Thumbnail    *string        `json:"thumbnail,omitempty" bson:"thumbnail"`
	Images       []*string      `json:"images,omitempty" bson:"images"`
	Address      *string        `json:"address,omitempty" bson:"address"`
	Latitude     *float64       `json:"latitude,omitempty" bson:"latitude"`
	Longitude    *float64       `json:"longitude,omitempty" bson:"longitude"`
	OpeningHours []*OpeningHour `json:"openingHours,omitempty" bson:"openingHours"`
	IsOpen       *bool          `json:"isOpen,omitempty" bson:"isOpen"`
}

// QueryFoodShopDto is used for querying food shops
type QueryFoodShopDto struct {
	BaseDto
	Keyword   *string  `json:"keyword"`
	IsOpen    *bool    `json:"isOpen"`
	Latitude  *float64 `json:"latitude"`  // center point for proximity search
	Longitude *float64 `json:"longitude"` // center point for proximity search
	Radius    *float64 `json:"radius"`    // radius in kilometers
}
