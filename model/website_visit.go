package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WebsiteVisit struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	IP          string             `bson:"ip" json:"ip"`
	Country     string             `bson:"country,omitempty" json:"country,omitempty"`
	CountryCode string             `bson:"countryCode,omitempty" json:"countryCode,omitempty"`
	City        string             `bson:"city,omitempty" json:"city,omitempty"`
	Region      string             `bson:"region,omitempty" json:"region,omitempty"`
	UserAgent   string             `bson:"userAgent" json:"userAgent"`
	Browser     string             `bson:"browser,omitempty" json:"browser,omitempty"`
	OS          string             `bson:"os,omitempty" json:"os,omitempty"`
	Device      string             `bson:"device,omitempty" json:"device,omitempty"`
	Path        string             `bson:"path,omitempty" json:"path,omitempty"`
	Action      string             `bson:"action,omitempty" json:"action,omitempty"` // "visit" or "click"
	Target      string             `bson:"target,omitempty" json:"target,omitempty"` // target identifier if click
	SessionID   string             `bson:"sessionId,omitempty" json:"sessionId,omitempty"`
	Referrer    string             `bson:"referrer,omitempty" json:"referrer,omitempty"`
	VisitedAt   time.Time          `bson:"visitedAt" json:"visitedAt"`
}

type TrackVisitDto struct {
	IP          string `json:"ip,omitempty"`
	Country     string `json:"country,omitempty"`
	CountryCode string `json:"countryCode,omitempty"`
	City        string `json:"city,omitempty"`
	Region      string `json:"region,omitempty"`
	Browser     string `json:"browser,omitempty"`
	OS          string `json:"os,omitempty"`
	Device      string `json:"device,omitempty"`
	Path        string `json:"path,omitempty"`
	Action      string `json:"action,omitempty"`
	Target      string `json:"target,omitempty"`
	SessionID   string `json:"sessionId,omitempty"`
	Referrer    string `json:"referrer,omitempty"`
}

type WebsiteVisitQueryDto struct {
	BaseDto
	Country string `json:"country,omitempty"`
	Action  string `json:"action,omitempty"`
	Search  string `json:"search,omitempty"`
	Days    int    `json:"days,omitempty"`
}

type ChartDataPoint struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
	Code  string `json:"code,omitempty"`
}

type WebsiteVisitAnalytics struct {
	TotalVisits     int64            `json:"totalVisits"`
	TotalClicks     int64            `json:"totalClicks"`
	UniqueVisitors  int64            `json:"uniqueVisitors"`
	TodayVisits     int64            `json:"todayVisits"`
	TodayClicks     int64            `json:"todayClicks"`
	VisitsByCountry []ChartDataPoint `json:"visitsByCountry"`
	VisitsByCity    []ChartDataPoint `json:"visitsByCity"`
	VisitsByDevice  []ChartDataPoint `json:"visitsByDevice"`
	VisitsByBrowser []ChartDataPoint `json:"visitsByBrowser"`
	VisitsByOS      []ChartDataPoint `json:"visitsByOS"`
	VisitsOverTime  []ChartDataPoint `json:"visitsOverTime"`
	TopPages        []ChartDataPoint `json:"topPages"`
	TopClicks       []ChartDataPoint `json:"topClicks"`
}

