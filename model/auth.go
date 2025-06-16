package model

type LoginDto struct {
	Token string `json:"token" bson:"token"`
}
