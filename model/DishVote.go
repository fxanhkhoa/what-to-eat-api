package model

import "time"

type DishVoteItem struct {
	Slug          string    `json:"slug" bson:"slug"`
	VoteUser      []*string `json:"voteUser" bson:"voteUser"`
	VoteAnonymous []*string `json:"voteAnonymous" bson:"voteAnonymous"`
	IsCustom      bool      `json:"isCustom" bson:"isCustom"`
	CustomTitle   string    `json:"customTitle" bson:"customTitle"`
}

type QueryDishVoteDto struct {
	BaseDto
	Keyword *string `json:"keyword"`
}

type DishVote struct {
	Title         *string         `json:"title,omitempty" bson:"title"`
	Description   *string         `json:"description,omitempty" bson:"description"`
	DishVoteItems []*DishVoteItem `json:"dishVoteItems" bson:"dishVoteItems"`
	Deleted       bool            `json:"deleted" bson:"deleted"`
	DeletedAt     *time.Time      `json:"deletedAt,omitempty" bson:"deletedAt"`
	DeletedBy     *string         `json:"deletedBy,omitempty" bson:"deletedBy"`
	UpdatedAt     *time.Time      `json:"updatedAt,omitempty" bson:"updatedAt"`
	UpdatedBy     *string         `json:"updatedBy,omitempty" bson:"updatedBy"`
	CreatedAt     *time.Time      `json:"createdAt,omitempty" bson:"createdAt"`
	CreatedBy     *string         `json:"createdBy,omitempty" bson:"createdBy"`
	ID            string          `json:"_id" bson:"_id,omitempty"`
}

type CreateDishVoteDto struct {
	Title         *string         `json:"title,omitempty" bson:"title"`
	Description   *string         `json:"description,omitempty" bson:"description"`
	DishVoteItems []*DishVoteItem `json:"dishVoteItems" bson:"dishVoteItems"`
}

type UpdateDishVoteDto struct {
	ID            string          `json:"_id" bson:"_id,omitempty"`
	Title         *string         `json:"title,omitempty" bson:"title"`
	Description   *string         `json:"description,omitempty" bson:"description"`
	DishVoteItems []*DishVoteItem `json:"dishVoteItems" bson:"dishVoteItems"`
}
