package model

import "time"

// UserDishCollection represents a user's custom collection of dishes for specific occasions
type UserDishCollection struct {
	UserID      string     `json:"userId" bson:"userId"`
	Name        string     `json:"name" bson:"name"`                               // e.g., "My Birthday", "Wife's Birthday", "Christmas Dinner"
	Description *string    `json:"description,omitempty" bson:"description"`       // Optional description
	Occasion    *string    `json:"occasion,omitempty" bson:"occasion"`             // e.g., "birthday", "anniversary", "holiday", "party"
	EventDate   *time.Time `json:"eventDate,omitempty" bson:"eventDate"`           // Optional date for the event
	DishSlugs   []string   `json:"dishSlugs" bson:"dishSlugs"`                     // Array of dish slugs in this collection
	Tags        []*string  `json:"tags,omitempty" bson:"tags"`                     // Custom tags for organization
	IsPublic    bool       `json:"isPublic" bson:"isPublic"`                       // Whether this collection is public/shareable
	Color       *string    `json:"color,omitempty" bson:"color"`                   // Optional color for UI display
	Icon        *string    `json:"icon,omitempty" bson:"icon"`                     // Optional icon name for UI display
	SortOrder   int        `json:"sortOrder" bson:"sortOrder"`                     // User-defined sort order
	Deleted     bool       `json:"deleted" bson:"deleted"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty" bson:"deletedAt"`
	DeletedBy   *string    `json:"deletedBy,omitempty" bson:"deletedBy"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty" bson:"updatedAt"`
	UpdatedBy   *string    `json:"updatedBy,omitempty" bson:"updatedBy"`
	CreatedAt   *time.Time `json:"createdAt,omitempty" bson:"createdAt"`
	CreatedBy   *string    `json:"createdBy,omitempty" bson:"createdBy"`
	ID          string     `json:"_id" bson:"_id,omitempty"`
}

// QueryUserDishCollectionDto is used for querying user dish collections
type QueryUserDishCollectionDto struct {
	BaseDto
	UserID   *string   `json:"userId"`
	Occasion *string   `json:"occasion"`
	Tags     *[]string `json:"tags"`
	IsPublic *bool     `json:"isPublic"`
	Keyword  *string   `json:"keyword"` // Search in name and description
}

// CreateUserDishCollectionDto is used for creating a new dish collection
type CreateUserDishCollectionDto struct {
	UserID      string     `json:"userId" bson:"userId"`
	Name        string     `json:"name" bson:"name"`
	Description *string    `json:"description,omitempty" bson:"description"`
	Occasion    *string    `json:"occasion,omitempty" bson:"occasion"`
	EventDate   *time.Time `json:"eventDate,omitempty" bson:"eventDate"`
	DishSlugs   []string   `json:"dishSlugs" bson:"dishSlugs"`
	Tags        []*string  `json:"tags,omitempty" bson:"tags"`
	IsPublic    bool       `json:"isPublic" bson:"isPublic"`
	Color       *string    `json:"color,omitempty" bson:"color"`
	Icon        *string    `json:"icon,omitempty" bson:"icon"`
	SortOrder   int        `json:"sortOrder" bson:"sortOrder"`
}

// UpdateUserDishCollectionDto is used for updating a dish collection
type UpdateUserDishCollectionDto struct {
	ID          string     `json:"_id" bson:"_id,omitempty"`
	UserID      string     `json:"userId" bson:"userId"`
	Name        string     `json:"name" bson:"name"`
	Description *string    `json:"description,omitempty" bson:"description"`
	Occasion    *string    `json:"occasion,omitempty" bson:"occasion"`
	EventDate   *time.Time `json:"eventDate,omitempty" bson:"eventDate"`
	DishSlugs   []string   `json:"dishSlugs" bson:"dishSlugs"`
	Tags        []*string  `json:"tags,omitempty" bson:"tags"`
	IsPublic    bool       `json:"isPublic" bson:"isPublic"`
	Color       *string    `json:"color,omitempty" bson:"color"`
	Icon        *string    `json:"icon,omitempty" bson:"icon"`
	SortOrder   int        `json:"sortOrder" bson:"sortOrder"`
}

// AddDishToCollectionDto is used for adding a dish to a collection
type AddDishToCollectionDto struct {
	UserID       string `json:"userId" bson:"userId"`
	CollectionID string `json:"collectionId" bson:"collectionId"`
	DishSlug     string `json:"dishSlug" bson:"dishSlug"`
}

// RemoveDishFromCollectionDto is used for removing a dish from a collection
type RemoveDishFromCollectionDto struct {
	UserID       string `json:"userId" bson:"userId"`
	CollectionID string `json:"collectionId" bson:"collectionId"`
	DishSlug     string `json:"dishSlug" bson:"dishSlug"`
}

// ReorderDishesInCollectionDto is used for reordering dishes within a collection
type ReorderDishesInCollectionDto struct {
	UserID       string   `json:"userId" bson:"userId"`
	CollectionID string   `json:"collectionId" bson:"collectionId"`
	DishSlugs    []string `json:"dishSlugs" bson:"dishSlugs"` // New order of dish slugs
}

// DuplicateCollectionDto is used for duplicating an existing collection
type DuplicateCollectionDto struct {
	UserID       string  `json:"userId" bson:"userId"`
	CollectionID string  `json:"collectionId" bson:"collectionId"`
	NewName      string  `json:"newName" bson:"newName"`
	CopyPublic   bool    `json:"copyPublic" bson:"copyPublic"` // Whether to copy the public setting
}

// ShareCollectionDto is used for sharing a collection
type ShareCollectionDto struct {
	UserID       string `json:"userId" bson:"userId"`
	CollectionID string `json:"collectionId" bson:"collectionId"`
	ShareWithUserIDs []string `json:"shareWithUserIds" bson:"shareWithUserIds"` // Optional: specific users to share with
}
