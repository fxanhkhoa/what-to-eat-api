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
	notifCol     CollectionInterface
	prefCol      CollectionInterface
	userCol      CollectionInterface
	templateCol  CollectionInterface
	adminLogCol  CollectionInterface
	userLoginCol CollectionInterface
}

func NewNotificationService(notifCol, prefCol, userCol CollectionInterface) *NotificationService {
	return &NotificationService{
		notifCol: notifCol,
		prefCol:  prefCol,
		userCol:  userCol,
		// templateCol, adminLogCol, userLoginCol are lazily wired via getXxxCol()
	}
}

func NewNotificationServiceFromDB() *NotificationService {
	dbName := config.GetDBInstance().GetDbName()
	db := config.GetDBInstance().GetClient().Database(dbName)
	return &NotificationService{
		notifCol:     NewMongoCollectionAdapter(db.Collection(constants.NOTIFICATION_COLLECTION)),
		prefCol:      NewMongoCollectionAdapter(db.Collection(constants.NOTIFICATION_PREFERENCE_COLLECTION)),
		userCol:      NewMongoCollectionAdapter(db.Collection(constants.USER_COLLECTION)),
		templateCol:  NewMongoCollectionAdapter(db.Collection(constants.NOTIFICATION_TEMPLATE_COLLECTION)),
		adminLogCol:  NewMongoCollectionAdapter(db.Collection(constants.ADMIN_NOTIFICATION_LOG_COLLECTION)),
		userLoginCol: NewMongoCollectionAdapter(db.Collection(constants.USER_LOGIN_TRACKS_COLLECTION)),
	}
}

func (ns *NotificationService) getTemplateCol() CollectionInterface {
	if ns.templateCol != nil {
		return ns.templateCol
	}
	dbName := config.GetDBInstance().GetDbName()
	return NewMongoCollectionAdapter(config.GetDBInstance().GetClient().Database(dbName).Collection(constants.NOTIFICATION_TEMPLATE_COLLECTION))
}

func (ns *NotificationService) getAdminLogCol() CollectionInterface {
	if ns.adminLogCol != nil {
		return ns.adminLogCol
	}
	dbName := config.GetDBInstance().GetDbName()
	return NewMongoCollectionAdapter(config.GetDBInstance().GetClient().Database(dbName).Collection(constants.ADMIN_NOTIFICATION_LOG_COLLECTION))
}

func (ns *NotificationService) getUserLoginCol() CollectionInterface {
	if ns.userLoginCol != nil {
		return ns.userLoginCol
	}
	dbName := config.GetDBInstance().GetDbName()
	return NewMongoCollectionAdapter(config.GetDBInstance().GetClient().Database(dbName).Collection(constants.USER_LOGIN_TRACKS_COLLECTION))
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

// ─── Broadcast ───────────────────────────────────────────────────────────────

// SendBroadcast fans out a notification to all users with registered device tokens.
// If dto.ScheduledAt is set and is in the future the notification is stored as
// "pending" and dispatched later by StartScheduler.
func (ns *NotificationService) SendBroadcast(dto model.SendBroadcastDto, createdBy string) error {
	now := time.Now()
	if dto.ScheduledAt != nil && dto.ScheduledAt.After(now) {
		return ns.scheduleAdminLog("all", nil, dto.Title, dto.Body, dto.ImageURL, dto.Type, dto.Data, dto.ScheduledAt, createdBy)
	}
	return ns.executeBroadcast(dto.Title, dto.Body, dto.ImageURL, dto.Type, dto.Data, createdBy)
}

func (ns *NotificationService) executeBroadcast(title, body, imageURL, notifType string, data map[string]string, createdBy string) error {
	tokens, err := ns.getAllDeviceTokens()
	if err != nil {
		return err
	}
	totalSent, totalFailed := ns.fanOutMulticast(tokens, title, body, imageURL, data)
	return ns.saveAdminLog("all", nil, title, body, imageURL, notifType, data, nil, totalSent, totalFailed, createdBy)
}

// getAllDeviceTokens returns every FCM token across all non-deleted users
func (ns *NotificationService) getAllDeviceTokens() ([]string, error) {
	cursor, err := ns.userCol.Find(
		context.TODO(),
		bson.M{"deleted": false, "deviceTokens": bson.M{"$exists": true, "$ne": bson.A{}}},
		options.Find().SetProjection(bson.M{"deviceTokens": 1}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var users []model.User
	if err := cursor.All(context.TODO(), &users); err != nil {
		return nil, err
	}

	var tokens []string
	for _, u := range users {
		for _, dt := range u.DeviceTokens {
			tokens = append(tokens, dt.Token)
		}
	}
	return tokens, nil
}

// fanOutMulticast sends msg to all tokens in batches of 500 (FCM limit).
// Returns (totalSent, totalFailed).
func (ns *NotificationService) fanOutMulticast(tokens []string, title, body, imageURL string, data map[string]string) (int, int) {
	const batchSize = 500
	totalSent, totalFailed := 0, 0
	for i := 0; i < len(tokens); i += batchSize {
		end := i + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}
		batch := tokens[i:end]
		if err := ns.SendMulticast(batch, title, body, imageURL, data); err != nil {
			totalFailed += len(batch)
		} else {
			totalSent += len(batch)
		}
	}
	return totalSent, totalFailed
}

// ─── Segment ─────────────────────────────────────────────────────────────────

// SendToSegment fans out a notification to users matching SegmentFilter.
// Respects ScheduledAt the same way as SendBroadcast.
func (ns *NotificationService) SendToSegment(dto model.SendSegmentDto, createdBy string) error {
	now := time.Now()
	sf := dto.SegmentFilter
	if dto.ScheduledAt != nil && dto.ScheduledAt.After(now) {
		return ns.scheduleAdminLog("segment", &sf, dto.Title, dto.Body, dto.ImageURL, dto.Type, dto.Data, dto.ScheduledAt, createdBy)
	}
	return ns.executeSegment(sf, dto.Title, dto.Body, dto.ImageURL, dto.Type, dto.Data, createdBy)
}

func (ns *NotificationService) executeSegment(sf model.SegmentFilter, title, body, imageURL, notifType string, data map[string]string, createdBy string) error {
	tokens, err := ns.getSegmentTokens(sf)
	if err != nil {
		return err
	}
	totalSent, totalFailed := ns.fanOutMulticast(tokens, title, body, imageURL, data)
	return ns.saveAdminLog("segment", &sf, title, body, imageURL, notifType, data, nil, totalSent, totalFailed, createdBy)
}

// getSegmentTokens builds a MongoDB user query from a SegmentFilter and returns tokens
func (ns *NotificationService) getSegmentTokens(sf model.SegmentFilter) ([]string, error) {
	filter := bson.M{"deleted": false, "deviceTokens": bson.M{"$exists": true, "$ne": bson.A{}}}

	// Role filter
	if len(sf.RoleNames) > 0 {
		filter["roleName"] = bson.M{"$in": sf.RoleNames}
	}

	// Inactive filter: users whose last login is older than InactiveDays
	if sf.InactiveDays != nil && *sf.InactiveDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -*sf.InactiveDays)
		activeCursor, err := ns.getUserLoginCol().Find(
			context.TODO(),
			bson.M{"createdAt": bson.M{"$gte": cutoff}},
			options.Find().SetProjection(bson.M{"userId": 1}),
		)
		if err != nil {
			return nil, err
		}
		defer activeCursor.Close(context.TODO())
		type loginTrack struct {
			UserID string `bson:"userId"`
		}
		var tracks []loginTrack
		if err := activeCursor.All(context.TODO(), &tracks); err != nil {
			return nil, err
		}
		activeIDs := make([]primitive.ObjectID, 0, len(tracks))
		seenActiveIDs := make(map[primitive.ObjectID]struct{}, len(tracks))
		for _, t := range tracks {
			if t.UserID == "" {
				continue
			}
			oid, err := primitive.ObjectIDFromHex(t.UserID)
			if err != nil {
				continue
			}
			if _, seen := seenActiveIDs[oid]; seen {
				continue
			}
			seenActiveIDs[oid] = struct{}{}
			activeIDs = append(activeIDs, oid)
		}
		if len(activeIDs) > 0 {
			filter["_id"] = bson.M{"$nin": activeIDs}
		}
	}

	cursor, err := ns.userCol.Find(
		context.TODO(),
		filter,
		options.Find().SetProjection(bson.M{"deviceTokens": 1}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var users []model.User
	if err := cursor.All(context.TODO(), &users); err != nil {
		return nil, err
	}
	var tokens []string
	for _, u := range users {
		for _, dt := range u.DeviceTokens {
			tokens = append(tokens, dt.Token)
		}
	}
	return tokens, nil
}

// ─── Scheduler ───────────────────────────────────────────────────────────────

// scheduleAdminLog saves a "pending" log entry for future dispatch
func (ns *NotificationService) scheduleAdminLog(
	sentTo string,
	sf *model.SegmentFilter,
	title, body, imageURL, notifType string,
	data map[string]string,
	scheduledAt *time.Time,
	createdBy string,
) error {
	log := model.AdminNotificationLog{
		Title:         title,
		Body:          body,
		ImageURL:      imageURL,
		Type:          notifType,
		Data:          data,
		SentTo:        sentTo,
		SegmentFilter: sf,
		ScheduledAt:   scheduledAt,
		CreatedBy:     createdBy,
	}
	_, err := ns.getAdminLogCol().InsertOne(context.TODO(), log)
	return err
}

// saveAdminLog persists a completed admin send event
func (ns *NotificationService) saveAdminLog(
	sentTo string,
	sf *model.SegmentFilter,
	title, body, imageURL, notifType string,
	data map[string]string,
	scheduledAt *time.Time,
	totalSent, totalFailed int,
	createdBy string,
) error {
	now := time.Now()
	log := model.AdminNotificationLog{
		Title:         title,
		Body:          body,
		ImageURL:      imageURL,
		Type:          notifType,
		Data:          data,
		SentTo:        sentTo,
		SegmentFilter: sf,
		ScheduledAt:   scheduledAt,
		SentAt:        &now,
		TotalSent:     totalSent,
		TotalFailed:   totalFailed,
		CreatedBy:     createdBy,
	}
	_, err := ns.getAdminLogCol().InsertOne(context.TODO(), log)
	return err
}

// StartScheduler runs a background goroutine that dispatches pending scheduled
// admin notifications. It fires immediately on startup (so any notifications
// scheduled while the instance was down are sent right away), then continues
// polling MongoDB every minute via a ticker.
func (ns *NotificationService) StartScheduler(ctx context.Context) {
	// Dispatch immediately on startup — catches overdue jobs after a restart
	ns.dispatchPending()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ns.dispatchPending()
		}
	}
}

func (ns *NotificationService) dispatchPending() {
	for {
		log, err := ns.claimNextPendingLog()
		if err == mongo.ErrNoDocuments {
			return
		}
		if err != nil {
			if isTransientMongoError(err) {
				fmt.Printf("[Scheduler] transient claim pending error: %v\n", err)
				return
			}
			fmt.Printf("[Scheduler] claim pending error: %v\n", err)
			return
		}

		fmt.Println("[Scheduler] dispatching log ID:", log.ID.Hex())

		var sendErr error
		if log.SentTo == "all" {
			sendErr = ns.executeBroadcast(log.Title, log.Body, log.ImageURL, log.Type, log.Data, log.CreatedBy)
		} else if log.SentTo == "segment" && log.SegmentFilter != nil {
			sendErr = ns.executeSegment(*log.SegmentFilter, log.Title, log.Body, log.ImageURL, log.Type, log.Data, log.CreatedBy)
		} else {
			sendErr = fmt.Errorf("unsupported scheduled notification target: %s", log.SentTo)
		}

		if sendErr != nil {
			fmt.Printf("[Scheduler] dispatch error for log %s: %v\n", log.ID, sendErr)
			ns.getAdminLogCol().UpdateOne(
				context.TODO(),
				bson.M{"_id": log.ID, "sentAt": nil},
				bson.M{
					"$set":   bson.M{"lastError": sendErr.Error(), "failedAt": time.Now()},
					"$unset": bson.M{"lockedAt": ""},
				},
			)
			continue
		}

		_, updateErr := ns.getAdminLogCol().UpdateOne(
			context.TODO(),
			bson.M{"_id": log.ID, "sentAt": nil},
			bson.M{
				"$set":   bson.M{"sentAt": time.Now()},
				"$unset": bson.M{"lockedAt": "", "lastError": "", "failedAt": ""},
			},
		)
		if updateErr != nil {
			fmt.Printf("[Scheduler] mark sent error for log %s: %v\n", log.ID, updateErr)
		}
	}
}

// claimNextPendingLog atomically claims one due scheduled notification log.
// The lock expires automatically after lockTTL to recover from crashed workers.
func (ns *NotificationService) claimNextPendingLog() (*model.AdminNotificationLog, error) {
	now := time.Now()
	lockTTL := 10 * time.Minute

	filter := bson.M{
		"sentAt":      nil,
		"scheduledAt": bson.M{"$lte": now},
		"$or": []bson.M{
			{"lockedAt": bson.M{"$exists": false}},
			{"lockedAt": nil},
			{"lockedAt": bson.M{"$lte": now.Add(-lockTTL)}},
		},
	}

	update := bson.M{"$set": bson.M{"lockedAt": now}}
	opt := options.FindOneAndUpdate().
		SetSort(bson.D{{Key: "scheduledAt", Value: 1}, {Key: "_id", Value: 1}}).
		SetReturnDocument(options.After)

	for attempt := 1; attempt <= 2; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		var log model.AdminNotificationLog
		err := ns.getAdminLogCol().FindOneAndUpdate(ctx, filter, update, opt).Decode(&log)
		cancel()
		if err == nil {
			return &log, nil
		}
		if err == mongo.ErrNoDocuments {
			return nil, err
		}
		if !isTransientMongoError(err) || attempt == 2 {
			return nil, err
		}
		time.Sleep(200 * time.Millisecond)
	}

	return nil, mongo.ErrNoDocuments
}

func isTransientMongoError(err error) bool {
	return mongo.IsNetworkError(err) || mongo.IsTimeout(err)
}

// GetAdminLogs returns paginated admin notification logs
func (ns *NotificationService) GetAdminLogs(page, limit int, sentTo string) ([]model.AdminNotificationLog, int64, error) {
	filter := bson.M{}
	if sentTo != "" {
		filter["sentTo"] = sentTo
	}
	total, err := ns.getAdminLogCol().CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, 0, err
	}
	skip := int64((page - 1) * limit)
	opts := options.Find().
		SetSort(bson.D{{Key: "sentAt", Value: -1}, {Key: "scheduledAt", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(limit))
	cursor, err := ns.getAdminLogCol().Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.TODO())
	var logs []model.AdminNotificationLog
	if err := cursor.All(context.TODO(), &logs); err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// ─── Templates ───────────────────────────────────────────────────────────────

func (ns *NotificationService) CreateTemplate(dto model.CreateNotificationTemplateDto, createdBy string) (*model.NotificationTemplate, error) {
	now := time.Now()
	tmpl := model.NotificationTemplate{
		Name:      dto.Name,
		Title:     dto.Title,
		Body:      dto.Body,
		ImageURL:  dto.ImageURL,
		Type:      dto.Type,
		Data:      dto.Data,
		CreatedAt: &now,
		CreatedBy: createdBy,
		UpdatedAt: &now,
	}
	result, err := ns.getTemplateCol().InsertOne(context.TODO(), tmpl)
	if err != nil {
		return nil, err
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		tmpl.ID = oid.Hex()
	}
	return &tmpl, nil
}

func (ns *NotificationService) GetTemplates(page, limit int) ([]model.NotificationTemplate, int64, error) {
	filter := bson.M{}
	total, err := ns.getTemplateCol().CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, 0, err
	}
	skip := int64((page - 1) * limit)
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(limit))
	cursor, err := ns.getTemplateCol().Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.TODO())
	var templates []model.NotificationTemplate
	if err := cursor.All(context.TODO(), &templates); err != nil {
		return nil, 0, err
	}
	return templates, total, nil
}

func (ns *NotificationService) GetTemplateByID(id string) (*model.NotificationTemplate, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid template ID: %w", err)
	}
	var tmpl model.NotificationTemplate
	if err := ns.getTemplateCol().FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&tmpl); err != nil {
		return nil, err
	}
	return &tmpl, nil
}

func (ns *NotificationService) UpdateTemplate(id string, dto model.UpdateNotificationTemplateDto) (*model.NotificationTemplate, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid template ID: %w", err)
	}
	now := time.Now()
	update := bson.M{"updatedAt": now}
	if dto.Name != nil {
		update["name"] = *dto.Name
	}
	if dto.Title != nil {
		update["title"] = *dto.Title
	}
	if dto.Body != nil {
		update["body"] = *dto.Body
	}
	if dto.ImageURL != nil {
		update["imageUrl"] = *dto.ImageURL
	}
	if dto.Type != nil {
		update["type"] = *dto.Type
	}
	if dto.Data != nil {
		update["data"] = dto.Data
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated model.NotificationTemplate
	if err := ns.getTemplateCol().FindOneAndUpdate(
		context.TODO(),
		bson.M{"_id": oid},
		bson.M{"$set": update},
		opts,
	).Decode(&updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (ns *NotificationService) DeleteTemplate(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid template ID: %w", err)
	}
	_, err = ns.getTemplateCol().DeleteOne(context.TODO(), bson.M{"_id": oid})
	return err
}
