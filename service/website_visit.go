package service

import (
	"context"
	"strings"
	"time"
	"what-to-eat/be/config"
	"what-to-eat/be/constants"
	"what-to-eat/be/helper"
	"what-to-eat/be/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WebsiteVisitService struct{}

func (s *WebsiteVisitService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	return config.GetDBInstance().GetClient().Database(dbName).Collection(constants.WEBSITE_VISIT_TRACKS_COLLECTION)
}

func (s *WebsiteVisitService) TrackVisit(dto model.TrackVisitDto, rawIP, userAgent, headerCountry, headerCountryCode string) error {
	collection := s.Collection()

	ip := strings.TrimSpace(dto.IP)
	if ip == "" {
		ip = strings.TrimSpace(rawIP)
	}

	browser := dto.Browser
	osName := dto.OS
	device := dto.Device
	if browser == "" || osName == "" || device == "" {
		pBrowser, pOS, pDevice := helper.ParseUserAgent(userAgent)
		if browser == "" {
			browser = pBrowser
		}
		if osName == "" {
			osName = pOS
		}
		if device == "" {
			device = pDevice
		}
	}

	country := dto.Country
	countryCode := dto.CountryCode
	city := dto.City
	region := dto.Region

	if country == "" && headerCountry != "" {
		country = headerCountry
		countryCode = headerCountryCode
	}

	if country == "" && ip != "" {
		geo := helper.ResolveGeoIP(ip)
		if geo != nil && geo.Country != "" && geo.Country != "Unknown" {
			country = geo.Country
			countryCode = geo.CountryCode
			if city == "" {
				city = geo.City
			}
			if region == "" {
				region = geo.RegionName
			}
		}
	}

	if country == "" {
		country = "Unknown"
		countryCode = "UN"
	}

	action := dto.Action
	if action == "" {
		action = "visit"
	}

	visit := model.WebsiteVisit{
		ID:          primitive.NewObjectID(),
		IP:          ip,
		Country:     country,
		CountryCode: countryCode,
		City:        city,
		Region:      region,
		UserAgent:   userAgent,
		Browser:     browser,
		OS:          osName,
		Device:      device,
		Path:        dto.Path,
		Action:      action,
		Target:      dto.Target,
		SessionID:   dto.SessionID,
		Referrer:    dto.Referrer,
		VisitedAt:   time.Now().UTC(),
	}

	_, err := collection.InsertOne(context.TODO(), visit)
	return err
}

func (s *WebsiteVisitService) CountVisits() (int64, error) {
	collection := s.Collection()
	return collection.CountDocuments(context.TODO(), bson.M{})
}

func (s *WebsiteVisitService) Find(query model.WebsiteVisitQueryDto) ([]*model.WebsiteVisit, int64, error) {
	collection := s.Collection()

	filter := bson.M{}

	if query.Country != "" {
		filter["country"] = bson.M{"$regex": "^" + query.Country + "$", "$options": "i"}
	}

	if query.Action != "" {
		filter["action"] = query.Action
	}

	if query.Days > 0 {
		cutoff := time.Now().UTC().AddDate(0, 0, -query.Days)
		filter["visitedAt"] = bson.M{"$gte": cutoff}
	}

	if query.Search != "" {
		pattern := primitive.Regex{Pattern: query.Search, Options: "i"}
		filter["$or"] = bson.A{
			bson.M{"ip": pattern},
			bson.M{"country": pattern},
			bson.M{"city": pattern},
			bson.M{"path": pattern},
			bson.M{"target": pattern},
			bson.M{"browser": pattern},
		}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "visitedAt", Value: -1}}).
		SetSkip(int64((query.Page - 1) * query.Limit)).
		SetLimit(int64(query.Limit))

	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.TODO())

	visits := make([]*model.WebsiteVisit, 0)
	if err = cursor.All(context.TODO(), &visits); err != nil {
		return nil, 0, err
	}

	for _, v := range visits {
		if v.Country == "" {
			v.Country = "Unknown"
			v.CountryCode = "UN"
		}
		if v.Action == "" {
			v.Action = "visit"
		}
	}

	return visits, count, nil
}

func (s *WebsiteVisitService) GetAnalytics(days int) (*model.WebsiteVisitAnalytics, error) {
	collection := s.Collection()

	if days <= 0 {
		days = 30
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -days)

	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "visitedAt", Value: bson.D{{Key: "$gte", Value: cutoff}}},
	}}}

	facetStage := bson.D{{Key: "$facet", Value: bson.D{
		// 1. Visits by Country
		{Key: "visitsByCountry", Value: bson.A{
			bson.D{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$country", "Unknown"}}}},
				{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "code", Value: bson.D{{Key: "$first", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$countryCode", "UN"}}}}}},
			}}},
			bson.D{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
			bson.D{{Key: "$limit", Value: 15}},
		}},

		// 2. Visits by City
		{Key: "visitsByCity", Value: bson.A{
			bson.D{{Key: "$match", Value: bson.D{{Key: "city", Value: bson.D{{Key: "$nin", Value: bson.A{"", nil, "Unknown", "Local"}}}}}}},
			bson.D{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$city"},
				{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			}}},
			bson.D{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
			bson.D{{Key: "$limit", Value: 10}},
		}},

		// 3. Visits by Device
		{Key: "visitsByDevice", Value: bson.A{
			bson.D{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$device", "Desktop"}}}},
				{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			}}},
			bson.D{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		}},

		// 4. Visits by Browser
		{Key: "visitsByBrowser", Value: bson.A{
			bson.D{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$browser", "Other"}}}},
				{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			}}},
			bson.D{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
			bson.D{{Key: "$limit", Value: 8}},
		}},

		// 5. Visits by OS
		{Key: "visitsByOS", Value: bson.A{
			bson.D{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$os", "Other"}}}},
				{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			}}},
			bson.D{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
			bson.D{{Key: "$limit", Value: 8}},
		}},

		// 6. Visits Over Time (Daily)
		{Key: "visitsOverTime", Value: bson.A{
			bson.D{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: bson.D{{Key: "$dateToString", Value: bson.D{
					{Key: "format", Value: "%Y-%m-%d"},
					{Key: "date", Value: "$visitedAt"},
				}}}},
				{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			}}},
			bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
		}},

		// 7. Top Visited Pages
		{Key: "topPages", Value: bson.A{
			bson.D{{Key: "$match", Value: bson.D{
				{Key: "path", Value: bson.D{{Key: "$nin", Value: bson.A{"", nil}}}},
			}}},
			bson.D{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$path"},
				{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			}}},
			bson.D{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
			bson.D{{Key: "$limit", Value: 10}},
		}},

		// 8. Top Click Targets
		{Key: "topClicks", Value: bson.A{
			bson.D{{Key: "$match", Value: bson.D{
				{Key: "action", Value: "click"},
				{Key: "target", Value: bson.D{{Key: "$nin", Value: bson.A{"", nil}}}},
			}}},
			bson.D{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$target"},
				{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			}}},
			bson.D{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
			bson.D{{Key: "$limit", Value: 10}},
		}},

		// 9. Totals and uniques
		{Key: "totals", Value: bson.A{
			bson.D{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil},
				{Key: "totalVisits", Value: bson.D{{Key: "$sum", Value: bson.D{{Key: "$cond", Value: bson.A{
					bson.D{{Key: "$ne", Value: bson.A{"$action", "click"}}},
					1,
					0,
				}}}}}},
				{Key: "totalClicks", Value: bson.D{{Key: "$sum", Value: bson.D{{Key: "$cond", Value: bson.A{
					bson.D{{Key: "$eq", Value: bson.A{"$action", "click"}}},
					1,
					0,
				}}}}}},
				{Key: "uniqueIPs", Value: bson.D{{Key: "$addToSet", Value: "$ip"}}},
			}}},
		}},
	}}}

	pipeline := mongo.Pipeline{matchStage, facetStage}

	cursor, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	type rawFacetResult struct {
		VisitsByCountry []struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
			Code  string `bson:"code"`
		} `bson:"visitsByCountry"`
		VisitsByCity []struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		} `bson:"visitsByCity"`
		VisitsByDevice []struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		} `bson:"visitsByDevice"`
		VisitsByBrowser []struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		} `bson:"visitsByBrowser"`
		VisitsByOS []struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		} `bson:"visitsByOS"`
		VisitsOverTime []struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		} `bson:"visitsOverTime"`
		TopPages []struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		} `bson:"topPages"`
		TopClicks []struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		} `bson:"topClicks"`
		Totals []struct {
			TotalVisits int64    `bson:"totalVisits"`
			TotalClicks int64    `bson:"totalClicks"`
			UniqueIPs   []string `bson:"uniqueIPs"`
		} `bson:"totals"`
	}

	var results []rawFacetResult
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	analytics := &model.WebsiteVisitAnalytics{
		VisitsByCountry: make([]model.ChartDataPoint, 0),
		VisitsByCity:    make([]model.ChartDataPoint, 0),
		VisitsByDevice:  make([]model.ChartDataPoint, 0),
		VisitsByBrowser: make([]model.ChartDataPoint, 0),
		VisitsByOS:      make([]model.ChartDataPoint, 0),
		VisitsOverTime:  make([]model.ChartDataPoint, 0),
		TopPages:        make([]model.ChartDataPoint, 0),
		TopClicks:       make([]model.ChartDataPoint, 0),
	}

	if len(results) == 0 {
		return analytics, nil
	}

	raw := results[0]

	for _, item := range raw.VisitsByCountry {
		name := item.ID
		if name == "" {
			name = "Unknown"
		}
		analytics.VisitsByCountry = append(analytics.VisitsByCountry, model.ChartDataPoint{
			Name:  name,
			Value: item.Count,
			Code:  item.Code,
		})
	}

	for _, item := range raw.VisitsByCity {
		analytics.VisitsByCity = append(analytics.VisitsByCity, model.ChartDataPoint{
			Name:  item.ID,
			Value: item.Count,
		})
	}

	for _, item := range raw.VisitsByDevice {
		analytics.VisitsByDevice = append(analytics.VisitsByDevice, model.ChartDataPoint{
			Name:  item.ID,
			Value: item.Count,
		})
	}

	for _, item := range raw.VisitsByBrowser {
		analytics.VisitsByBrowser = append(analytics.VisitsByBrowser, model.ChartDataPoint{
			Name:  item.ID,
			Value: item.Count,
		})
	}

	for _, item := range raw.VisitsByOS {
		analytics.VisitsByOS = append(analytics.VisitsByOS, model.ChartDataPoint{
			Name:  item.ID,
			Value: item.Count,
		})
	}

	for _, item := range raw.VisitsOverTime {
		analytics.VisitsOverTime = append(analytics.VisitsOverTime, model.ChartDataPoint{
			Name:  item.ID,
			Value: item.Count,
		})
	}

	for _, item := range raw.TopPages {
		analytics.TopPages = append(analytics.TopPages, model.ChartDataPoint{
			Name:  item.ID,
			Value: item.Count,
		})
	}

	for _, item := range raw.TopClicks {
		analytics.TopClicks = append(analytics.TopClicks, model.ChartDataPoint{
			Name:  item.ID,
			Value: item.Count,
		})
	}

	if len(raw.Totals) > 0 {
		analytics.TotalVisits = raw.Totals[0].TotalVisits
		analytics.TotalClicks = raw.Totals[0].TotalClicks
		analytics.UniqueVisitors = int64(len(raw.Totals[0].UniqueIPs))
	}

	// Calculate today's visits and clicks
	startOfToday := time.Now().UTC().Truncate(24 * time.Hour)
	todayVisits, _ := collection.CountDocuments(context.TODO(), bson.M{
		"visitedAt": bson.M{"$gte": startOfToday},
		"action":    bson.M{"$ne": "click"},
	})
	todayClicks, _ := collection.CountDocuments(context.TODO(), bson.M{
		"visitedAt": bson.M{"$gte": startOfToday},
		"action":    "click",
	})
	analytics.TodayVisits = todayVisits
	analytics.TodayClicks = todayClicks

	return analytics, nil
}
