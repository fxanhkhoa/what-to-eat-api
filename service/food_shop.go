package service

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"
	"what-to-eat/be/config"
	"what-to-eat/be/constants"
	"what-to-eat/be/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FoodShopService struct{}

func NewFoodShopService() *FoodShopService {
	return &FoodShopService{}
}

func (s *FoodShopService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	return config.GetDBInstance().GetClient().Database(dbName).Collection(constants.FOOD_SHOP_COLLECTION)
}

// ComputeIsOpen determines whether the shop is currently open based on:
// 1. Manual IsOpen override (if set, it takes priority)
// 2. OpeningHours schedule (using Vietnam timezone UTC+7)
func ComputeIsOpen(shop *model.FoodShop) bool {
	if shop.IsOpen != nil {
		return *shop.IsOpen
	}
	if len(shop.OpeningHours) == 0 {
		return false
	}

	// Vietnam is UTC+7
	loc := time.FixedZone("Asia/Ho_Chi_Minh", 7*3600)
	now := time.Now().In(loc)

	dayNames := []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}
	todayName := dayNames[now.Weekday()]

	for _, oh := range shop.OpeningHours {
		if oh == nil || strings.ToLower(oh.Day) != todayName {
			continue
		}
		if oh.Closed {
			return false
		}
		// Parse "HH:MM" open/close
		openH, openM := parseHHMM(oh.Open)
		closeH, closeM := parseHHMM(oh.Close)
		nowMinutes := now.Hour()*60 + now.Minute()
		openMinutes := openH*60 + openM
		closeMinutes := closeH*60 + closeM

		if closeMinutes <= openMinutes {
			// Overnight: e.g. 22:00 - 02:00
			return nowMinutes >= openMinutes || nowMinutes < closeMinutes
		}
		return nowMinutes >= openMinutes && nowMinutes < closeMinutes
	}
	return false
}

func parseHHMM(s string) (int, int) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, 0
	}
	h, m := 0, 0
	for _, ch := range parts[0] {
		if ch >= '0' && ch <= '9' {
			h = h*10 + int(ch-'0')
		}
	}
	for _, ch := range parts[1] {
		if ch >= '0' && ch <= '9' {
			m = m*10 + int(ch-'0')
		}
	}
	return h, m
}

// haversineKm returns the distance in km between two lat/lng points.
func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func (s *FoodShopService) Create(dto model.CreateFoodShopDto, profile *model.JwtCustomClaims) (*model.FoodShop, error) {
	collection := s.Collection()
	now := time.Now()

	shop := model.FoodShop{
		Name:         dto.Name,
		Description:  dto.Description,
		Phone:        dto.Phone,
		Website:      dto.Website,
		Thumbnail:    dto.Thumbnail,
		Images:       dto.Images,
		Address:      dto.Address,
		Latitude:     dto.Latitude,
		Longitude:    dto.Longitude,
		OpeningHours: dto.OpeningHours,
		IsOpen:       dto.IsOpen,
		Deleted:      false,
		CreatedAt:    &now,
		CreatedBy:    &profile.ID,
		UpdatedAt:    &now,
		UpdatedBy:    &profile.ID,
	}

	result, err := collection.InsertOne(context.TODO(), shop)
	if err != nil {
		return nil, err
	}

	shop.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return &shop, nil
}

func (s *FoodShopService) Update(dto model.UpdateFoodShopDto, profile *model.JwtCustomClaims) (*model.FoodShop, error) {
	collection := s.Collection()
	now := time.Now()

	objectID, err := primitive.ObjectIDFromHex(dto.ID)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	update := bson.M{
		"$set": bson.D{
			{Key: "name", Value: dto.Name},
			{Key: "description", Value: dto.Description},
			{Key: "phone", Value: dto.Phone},
			{Key: "website", Value: dto.Website},
			{Key: "thumbnail", Value: dto.Thumbnail},
			{Key: "images", Value: dto.Images},
			{Key: "address", Value: dto.Address},
			{Key: "latitude", Value: dto.Latitude},
			{Key: "longitude", Value: dto.Longitude},
			{Key: "openingHours", Value: dto.OpeningHours},
			{Key: "isOpen", Value: dto.IsOpen},
			{Key: "updatedAt", Value: now},
			{Key: "updatedBy", Value: profile.ID},
		},
	}

	filter := bson.M{"_id": objectID, "deleted": false}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var shop model.FoodShop
	result := collection.FindOneAndUpdate(context.TODO(), filter, update, opts)
	if result.Err() != nil {
		return nil, result.Err()
	}
	if err := result.Decode(&shop); err != nil {
		return nil, err
	}
	return &shop, nil
}

// VerifyStatus updates the open/closed status and stamps lastVerifiedAt.
func (s *FoodShopService) VerifyStatus(id string, isOpen bool, profile *model.JwtCustomClaims) (*model.FoodShop, error) {
	collection := s.Collection()
	now := time.Now()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID, "deleted": false}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	update := bson.M{
		"$set": bson.M{
			"isOpen":         isOpen,
			"lastVerifiedAt": now,
			"verifiedBy":     profile.ID,
			"updatedAt":      now,
			"updatedBy":      profile.ID,
		},
	}

	var shop model.FoodShop
	result := collection.FindOneAndUpdate(context.TODO(), filter, update, opts)
	if result.Err() != nil {
		return nil, result.Err()
	}
	if err := result.Decode(&shop); err != nil {
		return nil, err
	}
	return &shop, nil
}

func (s *FoodShopService) Remove(id string, profile *model.JwtCustomClaims) (*model.FoodShop, error) {
	collection := s.Collection()
	now := time.Now()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID, "deleted": false}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	update := bson.M{
		"$set": bson.M{
			"deleted":   true,
			"deletedAt": now,
			"deletedBy": profile.ID,
		},
	}

	var shop model.FoodShop
	result := collection.FindOneAndUpdate(context.TODO(), filter, update, opts)
	if result.Err() != nil {
		return nil, result.Err()
	}
	if err := result.Decode(&shop); err != nil {
		return nil, err
	}
	return &shop, nil
}

func (s *FoodShopService) FindOne(id string) (*model.FoodShop, error) {
	collection := s.Collection()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	var shop model.FoodShop
	err = collection.FindOne(context.TODO(), bson.M{"_id": objectID, "deleted": false}).Decode(&shop)
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

// FoodShopWithStatus wraps the model with a computed isOpen status.
type FoodShopWithStatus struct {
	model.FoodShop `bson:",inline"`
	ComputedIsOpen bool `json:"computedIsOpen"`
}

func (s *FoodShopService) Find(query model.QueryFoodShopDto) ([]*FoodShopWithStatus, int64, error) {
	collection := s.Collection()

	filter := bson.D{{Key: "deleted", Value: false}}

	if query.Keyword != nil && *query.Keyword != "" {
		regex := primitive.Regex{Pattern: *query.Keyword, Options: "i"}
		filter = append(filter, bson.E{Key: "$or", Value: bson.A{
			bson.M{"name": bson.M{"$regex": regex}},
			bson.M{"address": bson.M{"$regex": regex}},
		}})
	}

	// Bounding box pre-filter for proximity search
	if query.Latitude != nil && query.Longitude != nil && query.Radius != nil && *query.Radius > 0 {
		// 1 degree latitude ≈ 111.32 km
		latDelta := *query.Radius / 111.32
		lonDelta := *query.Radius / (111.32 * math.Cos(*query.Latitude*math.Pi/180))
		filter = append(filter,
			bson.E{Key: "latitude", Value: bson.M{"$gte": *query.Latitude - latDelta, "$lte": *query.Latitude + latDelta}},
			bson.E{Key: "longitude", Value: bson.M{"$gte": *query.Longitude - lonDelta, "$lte": *query.Longitude + lonDelta}},
		)
	}

	page := query.BaseDto.Page
	if page < 1 {
		page = 1
	}
	limit := query.BaseDto.Limit
	if limit < 1 {
		limit = 10
	}
	skip := int64((page - 1) * limit)

	opts := options.Find().SetSkip(skip).SetLimit(int64(limit)).SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.TODO())

	var shops []*model.FoodShop
	if err = cursor.All(context.TODO(), &shops); err != nil {
		return nil, 0, err
	}

	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, 0, err
	}

	// Apply precise Haversine filter and IsOpen filter after bounding box
	result := make([]*FoodShopWithStatus, 0, len(shops))
	for _, shop := range shops {
		if query.Latitude != nil && query.Longitude != nil && query.Radius != nil && *query.Radius > 0 {
			if shop.Latitude == nil || shop.Longitude == nil {
				continue
			}
			dist := haversineKm(*query.Latitude, *query.Longitude, *shop.Latitude, *shop.Longitude)
			if dist > *query.Radius {
				continue
			}
		}

		computed := ComputeIsOpen(shop)

		if query.IsOpen != nil && computed != *query.IsOpen {
			continue
		}

		result = append(result, &FoodShopWithStatus{
			FoodShop:       *shop,
			ComputedIsOpen: computed,
		})
	}

	return result, count, nil
}
