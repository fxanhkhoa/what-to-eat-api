package model

import "time"

type RolePermission struct {
	Name        string     `json:"name" bson:"name"`
	Permission  []string   `json:"permission,omitempty" bson:"permission"`
	Description *string    `json:"description,omitempty" bson:"description"`
	Deleted     bool       `json:"deleted" bson:"deleted"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty" bson:"deletedAt"`
	DeletedBy   *string    `json:"deletedBy,omitempty" bson:"deletedBy"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty" bson:"updatedAt"`
	UpdatedBy   *string    `json:"updatedBy,omitempty" bson:"updatedBy"`
	CreatedAt   *time.Time `json:"createdAt,omitempty" bson:"createdAt"`
	CreatedBy   *string    `json:"createdBy,omitempty" bson:"createdBy"`
	ID          string     `json:"_id" bson:"_id,omitempty"`
}

type CreateRolePermissionDto struct {
	Name        string   `json:"name" bson:"name"`
	Permission  []string `json:"permission,omitempty" bson:"permission"`
	Description *string  `json:"description,omitempty" bson:"description"`
}

type UpdateRolePermissionDto struct {
	ID          string   `json:"id" bson:"id"`
	Name        string   `json:"name" bson:"name"`
	Permission  []string `json:"permission,omitempty" bson:"permission"`
	Description *string  `json:"description,omitempty" bson:"description"`
}
