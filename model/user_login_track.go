package model

import "time"

type UserLoginTrack struct {
	UserID    string    `bson:"userId" json:"userId"`
	LoginAt   time.Time `bson:"loginAt" json:"loginAt"`
	IP        string    `bson:"ip" json:"ip"`
	UserAgent string    `bson:"userAgent" json:"userAgent"`
}
