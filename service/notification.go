package service

import (
	"context"
	"fmt"
	"time"
	"what-to-eat/be/config"
	"what-to-eat/be/constants"
	firebasePkg "what-to-eat/be/firebase"
	"what-to-eat/be/model"

	"firebase.google.com/go/v4/messaging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NotificationService struct {
	notifCol CollectionInterface
	prefCol  CollectionInterface
	userCol  CollectionInterface
}

func NewNotificationService(notifCol, prefCol, userCol CollectionInterface) *NotificationService {
	return &NotificationService{
		notifCol: notifCol,
		prefCol:  prefCol,
		userCol:  userCol,
	}
}

func NewNotificationServiceFromDB() *NotificationService {
	dbName := config.GetDBInstance().GetDbName()
	db := config.GetDBInstance().GetClient().Database(dbName)
	return &NotificationService{
		notifCol: NewMongoCollectionAdapter(db.Collection(constants.NOTIFICATION_COLLECTION)),
		prefCol:  NewMongoCollectionAdapter(db.Collection(constants.NOTIFICATION_PREFERENCE_COLLECTION)),
		userCol:  NewMongoCollectionAdapter(db.Collection(constants.USER_COLLECTION)),
	}
}

// RegisterDeviceToken upserts an FCM token for a user
func (ns *NotificationService) RegisterDeviceToken(userID string, dto model.RegisterDeviceTokenDto) error {
	now := time.Now()

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Remove any existing entry for this token first to avoid duplicates
	_, err = ns.userCol.UpdateOne(
		context.TODO(),
		bson.M{"_id": userObjID},
		bson.M{"$pull": bson.M{"deviceTokens": bson.M{"token": dto.Token}}},
	)

	if err != nil {
		return err
	}

	token := model.DeviceToken{
		Token:      dto.Token,
		Platform:   dto.Platform,
		DeviceInfo: dto.DeviceInfo,
		LastUsed:   &now,
		CreatedAt:  &now,
	}

	_, err = ns.userCol.UpdateOne(
		context.TODO(),
		bson.M{"_id": userObjID, "deleted": false},
		bson.M{"$push": bson.M{"deviceTokens": token}},
	)

	return err
}

// UnregisterDeviceToken removes a specific FCM token from a user
func (ns *NotificationService) UnregisterDeviceToken(userID string, token string) error {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	_, err = ns.userCol.UpdateOne(
		context.TODO(),
		bson.M{"_id": userObjID},
		bson.M{"$pull": bson.M{"deviceTokens": bson.M{"token": token}}},
	)
	return err
}

// GetDeviceTokens returns all FCM tokens for a user
func (ns *NotificationService) GetDeviceTokens(userID string) ([]model.DeviceToken, error) {
	var user model.User
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	err = ns.userCol.FindOne(
		context.TODO(),
		bson.M{"_id": userObjID, "deleted": false},
		options.FindOne().SetProjection(bson.M{"deviceTokens": 1}),
	).Decode(&user)
	if err != nil {
		return nil, err
	}
	return user.DeviceTokens, nil
}

// GetUserPreferences fetches or creates default notification preferences for a user
func (ns *NotificationService) GetUserPreferences(userID string) (*model.NotificationPreference, error) {
	var pref model.NotificationPreference
	err := ns.prefCol.FindOne(context.TODO(), bson.M{"userId": userID}).Decode(&pref)
	if err == mongo.ErrNoDocuments {
		// Create defaults
		pref = model.NotificationPreference{
			UserID:           userID,
			ChatEnabled:      true,
			ActivityEnabled:  true,
			MarketingEnabled: true,
		}
		result, insertErr := ns.prefCol.InsertOne(context.TODO(), pref)
		if insertErr != nil {
			return nil, insertErr
		}
		if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
			pref.ID = oid.Hex()
		}
		return &pref, nil
	}
	return &pref, err
}

// UpdateUserPreferences updates notification preferences for a user
func (ns *NotificationService) UpdateUserPreferences(userID string, dto model.UpdateNotificationPreferenceDto) (*model.NotificationPreference, error) {
	// Ensure preferences doc exists
	if _, err := ns.GetUserPreferences(userID); err != nil {
		return nil, err
	}

	update := bson.M{}
	if dto.ChatEnabled != nil {
		update["chatEnabled"] = *dto.ChatEnabled
	}
	if dto.ActivityEnabled != nil {
		update["activityEnabled"] = *dto.ActivityEnabled
	}
	if dto.MarketingEnabled != nil {
		update["marketingEnabled"] = *dto.MarketingEnabled
	}
	if dto.QuietHoursStart != nil {
		update["quietHoursStart"] = *dto.QuietHoursStart
	}
	if dto.QuietHoursEnd != nil {
		update["quietHoursEnd"] = *dto.QuietHoursEnd
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated model.NotificationPreference
	err := ns.prefCol.FindOneAndUpdate(
		context.TODO(),
		bson.M{"userId": userID},
		bson.M{"$set": update},
		opts,
	).Decode(&updated)
	return &updated, err
}

// CheckQuietHours returns true if current time falls within quiet hours
func (ns *NotificationService) CheckQuietHours(pref *model.NotificationPreference) bool {
	if pref.QuietHoursStart == nil || pref.QuietHoursEnd == nil {
		return false
	}
	now := time.Now()
	currentHHMM := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())
	start := *pref.QuietHoursStart
	end := *pref.QuietHoursEnd

	if start <= end {
		return currentHHMM >= start && currentHHMM < end
	}
	// Overnight range e.g. 22:00 - 08:00
	return currentHHMM >= start || currentHHMM < end
}

// SendToUser sends an FCM notification to all devices of a user
func (ns *NotificationService) SendToUser(userID, title, body, imageURL, notifType string, data map[string]string) error {
	// Check preferences
	pref, err := ns.GetUserPreferences(userID)
	if err == nil {
		// Check if this notification type is enabled
		switch notifType {
		case "chat":
			if !pref.ChatEnabled {
				return nil
			}
		case "activity":
			if !pref.ActivityEnabled {
				return nil
			}
		case "marketing":
			if !pref.MarketingEnabled {
				return nil
			}
		}
		// Check quiet hours
		if ns.CheckQuietHours(pref) {
			return nil
		}
	}

	tokens, err := ns.GetDeviceTokens(userID)
	if err != nil || len(tokens) == 0 {
		return err
	}

	tokenStrings := make([]string, 0, len(tokens))
	for _, t := range tokens {
		tokenStrings = append(tokenStrings, t.Token)
	}

	sendErr := ns.SendMulticast(tokenStrings, title, body, imageURL, data)

	// Save notification record regardless of send error
	var sendErrorStr *string
	if sendErr != nil {
		errStr := sendErr.Error()
		sendErrorStr = &errStr
	}

	ns.saveNotification(userID, title, body, imageURL, notifType, data, sendErrorStr)

	return sendErr
}

// SendMulticast sends a notification to multiple FCM tokens
func (ns *NotificationService) SendMulticast(tokens []string, title, body, imageURL string, data map[string]string) error {
	if firebasePkg.FirebaseMessagingClient == nil {
		return fmt.Errorf("firebase messaging client not initialized")
	}

	if len(tokens) == 0 {
		return nil
	}

	msg := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title:    title,
			Body:     body,
			ImageURL: imageURL,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound: "default",
			},
		},
		Webpush: &messaging.WebpushConfig{
			Notification: &messaging.WebpushNotification{
				Title: title,
				Body:  body,
				Icon:  "/assets/icons/logo-128x128.png",
			},
		},
	}

	response, err := firebasePkg.FirebaseMessagingClient.SendEachForMulticast(context.TODO(), msg)
	if err != nil {
		return err
	}

	// Clean up invalid tokens
	if response.FailureCount > 0 {
		for i, result := range response.Responses {
			if !result.Success && i < len(tokens) && result.Error != nil {
				// Only remove tokens that are clearly invalid/unregistered. Other errors may be transient.
				if messaging.IsUnregistered(result.Error) || messaging.IsInvalidArgument(result.Error) {
					ns.cleanupInvalidToken(tokens[i])
				}
			}
		}
	}
	return nil
}

// SendToTopic broadcasts to an FCM topic
func (ns *NotificationService) SendToTopic(topic, title, body, imageURL string, data map[string]string) error {
	if firebasePkg.FirebaseMessagingClient == nil {
		return fmt.Errorf("firebase messaging client not initialized")
	}

	msg := &messaging.Message{
		Topic: topic,
		Notification: &messaging.Notification{
			Title:    title,
			Body:     body,
			ImageURL: imageURL,
		},
		Data: data,
	}

	_, err := firebasePkg.FirebaseMessagingClient.Send(context.TODO(), msg)
	return err
}

func (ns *NotificationService) saveNotification(
	userID,
	title,
	body,
	imageURL,
	notifType string,
	data map[string]string,
	sendError *string,
) {
	now := time.Now()
	notif := model.Notification{
		UserID:    userID,
		Title:     title,
		Body:      body,
		ImageURL:  imageURL,
		Type:      notifType,
		Data:      data,
		SentAt:    &now,
		SendError: sendError,
	}
	_, err := ns.notifCol.InsertOne(context.TODO(), notif)
	if err != nil {
		fmt.Printf("Error inserting notification: %v\n", err)
	}
}

// GetUserNotifications returns paginated notifications for a user
func (ns *NotificationService) GetUserNotifications(userID string, page, limit int) ([]model.Notification, int64, error) {
	filter := bson.M{"userId": userID}

	total, err := ns.notifCol.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().
		SetSort(bson.D{{Key: "sentAt", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(limit))

	cursor, err := ns.notifCol.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.TODO())

	var notifications []model.Notification
	if err := cursor.All(context.TODO(), &notifications); err != nil {
		return nil, 0, err
	}
	return notifications, total, nil
}

// GetUnreadCount counts unread notifications for a user
func (ns *NotificationService) GetUnreadCount(userID string) (int64, error) {
	return ns.notifCol.CountDocuments(context.TODO(), bson.M{"userId": userID, "readAt": nil})
}

// MarkAsRead marks a single notification as read
func (ns *NotificationService) MarkAsRead(notificationID string) error {
	oid, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		return err
	}
	now := time.Now()
	_, err = ns.notifCol.UpdateOne(
		context.TODO(),
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{"readAt": now}},
	)
	return err
}

// MarkAllAsRead marks all notifications for a user as read
func (ns *NotificationService) MarkAllAsRead(userID string) error {
	now := time.Now()
	_, err := ns.notifCol.UpdateMany(
		context.TODO(),
		bson.M{"userId": userID, "readAt": nil},
		bson.M{"$set": bson.M{"readAt": now}},
	)
	return err
}

// cleanupInvalidToken removes a stale FCM token from all users
func (ns *NotificationService) cleanupInvalidToken(token string) {
	ns.userCol.UpdateMany(
		context.TODO(),
		bson.M{"deviceTokens.token": token},
		bson.M{"$pull": bson.M{"deviceTokens": bson.M{"token": token}}},
	)
}
