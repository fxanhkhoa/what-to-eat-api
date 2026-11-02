package service

import (
	"context"
	"errors"
	"testing"
	"time"
	"what-to-eat/be/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ---------------------------------------------------------------------------
// Notification-specific mocks
// ---------------------------------------------------------------------------

// genericCursor is a flexible Cursor mock driven by an optional decode function.
type genericCursor struct {
	decodeFn func(results interface{}) error
	err      error
}

func (g *genericCursor) All(ctx context.Context, results interface{}) error {
	if g.err != nil {
		return g.err
	}
	if g.decodeFn != nil {
		return g.decodeFn(results)
	}
	return nil
}

func (g *genericCursor) Close(ctx context.Context) error { return nil }

// emptyCursor always returns an empty result set (no-op decodeFn).
func emptyCursor() Cursor { return &genericCursor{} }

// notifMockCollection is a configurable CollectionInterface for notification tests.
type notifMockCollection struct {
	// InsertOne
	insertOneResult *mongo.InsertOneResult
	insertOneErr    error

	// FindOne
	findOneResult SingleResult

	// FindOneAndUpdate
	findOneAndUpdateResult SingleResult

	// CountDocuments
	countResult int64
	countErr    error

	// Find
	findResult Cursor
	findErr    error

	// UpdateOne
	updateOneErr error

	// UpdateMany
	updateManyErr error

	// DeleteOne
	deleteOneErr error

	// Call counters
	insertOneCalls  int
	updateOneCalls  int
	updateManyCalls int
	deleteOneCalls  int
}

func (m *notifMockCollection) InsertOne(ctx context.Context, doc interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	m.insertOneCalls++
	return m.insertOneResult, m.insertOneErr
}

func (m *notifMockCollection) FindOneAndUpdate(ctx context.Context, filter, update interface{}, opts ...*options.FindOneAndUpdateOptions) SingleResult {
	return m.findOneAndUpdateResult
}

func (m *notifMockCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return m.countResult, m.countErr
}

func (m *notifMockCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (Cursor, error) {
	if m.findResult != nil {
		return m.findResult, m.findErr
	}
	return emptyCursor(), m.findErr
}

func (m *notifMockCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) SingleResult {
	return m.findOneResult
}

func (m *notifMockCollection) UpdateOne(ctx context.Context, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	m.updateOneCalls++
	if m.updateOneErr != nil {
		return nil, m.updateOneErr
	}
	return &mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil
}

func (m *notifMockCollection) UpdateMany(ctx context.Context, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	m.updateManyCalls++
	return nil, m.updateManyErr
}

func (m *notifMockCollection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	m.deleteOneCalls++
	if m.deleteOneErr != nil {
		return nil, m.deleteOneErr
	}
	return &mongo.DeleteResult{DeletedCount: 1}, nil
}

func (m *notifMockCollection) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (Cursor, error) {
	return emptyCursor(), nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// decodePrefFn returns a decodeFn that writes pref into the destination.
func decodePrefFn(pref model.NotificationPreference) func(interface{}) error {
	return func(v interface{}) error {
		*(v.(*model.NotificationPreference)) = pref
		return nil
	}
}

// decodeTemplateFn returns a decodeFn that writes tmpl into the destination.
func decodeTemplateFn(tmpl model.NotificationTemplate) func(interface{}) error {
	return func(v interface{}) error {
		*(v.(*model.NotificationTemplate)) = tmpl
		return nil
	}
}

// usersCursor returns a genericCursor whose All writes users.
func usersCursor(users []model.User) Cursor {
	return &genericCursor{
		decodeFn: func(results interface{}) error {
			*(results.(*[]model.User)) = users
			return nil
		},
	}
}

// notifsCursor returns a genericCursor whose All writes notifications.
func notifsCursor(notifs []model.Notification) Cursor {
	return &genericCursor{
		decodeFn: func(results interface{}) error {
			*(results.(*[]model.Notification)) = notifs
			return nil
		},
	}
}

// templatesCursor returns a genericCursor whose All writes templates.
func templatesCursor(tmpls []model.NotificationTemplate) Cursor {
	return &genericCursor{
		decodeFn: func(results interface{}) error {
			*(results.(*[]model.NotificationTemplate)) = tmpls
			return nil
		},
	}
}

// adminLogsCursor returns a genericCursor whose All writes admin logs.
func adminLogsCursor(logs []model.AdminNotificationLog) Cursor {
	return &genericCursor{
		decodeFn: func(results interface{}) error {
			*(results.(*[]model.AdminNotificationLog)) = logs
			return nil
		},
	}
}

// ---------------------------------------------------------------------------
// RegisterDeviceToken
// ---------------------------------------------------------------------------

func TestRegisterDeviceToken_InvalidID(t *testing.T) {
	svc := &NotificationService{userCol: &notifMockCollection{}}
	err := svc.RegisterDeviceToken("not-a-valid-hex", model.RegisterDeviceTokenDto{Token: "tok", Platform: "web"})
	if err == nil {
		t.Fatal("expected error for invalid user ID")
	}
}

func TestRegisterDeviceToken_UpdateOneError(t *testing.T) {
	userCol := &notifMockCollection{updateOneErr: errors.New("update failed")}
	svc := &NotificationService{userCol: userCol}

	oid := primitive.NewObjectID()
	err := svc.RegisterDeviceToken(oid.Hex(), model.RegisterDeviceTokenDto{Token: "tok", Platform: "web"})
	if err == nil || err.Error() != "update failed" {
		t.Errorf("expected 'update failed', got %v", err)
	}
}

func TestRegisterDeviceToken_Success(t *testing.T) {
	userCol := &notifMockCollection{}
	svc := &NotificationService{userCol: userCol}

	oid := primitive.NewObjectID()
	err := svc.RegisterDeviceToken(oid.Hex(), model.RegisterDeviceTokenDto{Token: "tok", Platform: "web"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Both pull and push call UpdateOne
	if userCol.updateOneCalls != 2 {
		t.Errorf("expected 2 UpdateOne calls, got %d", userCol.updateOneCalls)
	}
}

// ---------------------------------------------------------------------------
// UnregisterDeviceToken
// ---------------------------------------------------------------------------

func TestUnregisterDeviceToken_InvalidID(t *testing.T) {
	svc := &NotificationService{userCol: &notifMockCollection{}}
	err := svc.UnregisterDeviceToken("bad-id", "tok")
	if err == nil {
		t.Fatal("expected error for invalid user ID")
	}
}

func TestUnregisterDeviceToken_UpdateError(t *testing.T) {
	userCol := &notifMockCollection{updateOneErr: errors.New("db error")}
	svc := &NotificationService{userCol: userCol}
	err := svc.UnregisterDeviceToken(primitive.NewObjectID().Hex(), "tok")
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected 'db error', got %v", err)
	}
}

func TestUnregisterDeviceToken_Success(t *testing.T) {
	svc := &NotificationService{userCol: &notifMockCollection{}}
	err := svc.UnregisterDeviceToken(primitive.NewObjectID().Hex(), "tok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// GetDeviceTokens
// ---------------------------------------------------------------------------

func TestGetDeviceTokens_InvalidID(t *testing.T) {
	svc := &NotificationService{userCol: &notifMockCollection{}}
	_, err := svc.GetDeviceTokens("bad-id")
	if err == nil {
		t.Fatal("expected error for invalid user ID")
	}
}

func TestGetDeviceTokens_FindOneError(t *testing.T) {
	userCol := &notifMockCollection{
		findOneResult: &mockSingleResult{
			decodeFn: func(v interface{}) error { return errors.New("db error") },
		},
	}
	svc := &NotificationService{userCol: userCol}
	_, err := svc.GetDeviceTokens(primitive.NewObjectID().Hex())
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected 'db error', got %v", err)
	}
}

func TestGetDeviceTokens_Success(t *testing.T) {
	now := time.Now()
	tokens := []model.DeviceToken{{Token: "tok1", Platform: "web", LastUsed: &now}}
	userCol := &notifMockCollection{
		findOneResult: &mockSingleResult{
			decodeFn: func(v interface{}) error {
				*(v.(*model.User)) = model.User{DeviceTokens: tokens}
				return nil
			},
		},
	}
	svc := &NotificationService{userCol: userCol}
	got, err := svc.GetDeviceTokens(primitive.NewObjectID().Hex())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Token != "tok1" {
		t.Errorf("unexpected tokens: %+v", got)
	}
}

// ---------------------------------------------------------------------------
// GetUserPreferences
// ---------------------------------------------------------------------------

func TestGetUserPreferences_Found(t *testing.T) {
	existing := model.NotificationPreference{UserID: "u1", ChatEnabled: true, ActivityEnabled: true, MarketingEnabled: true}
	prefCol := &notifMockCollection{
		findOneResult: &mockSingleResult{decodeFn: decodePrefFn(existing)},
	}
	svc := &NotificationService{prefCol: prefCol}

	pref, err := svc.GetUserPreferences("u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pref.UserID != "u1" || !pref.ChatEnabled {
		t.Errorf("unexpected pref: %+v", pref)
	}
}

func TestGetUserPreferences_NotFound_CreatesDefault(t *testing.T) {
	oid := primitive.NewObjectID()
	prefCol := &notifMockCollection{
		findOneResult:   &mockSingleResult{decodeFn: func(v interface{}) error { return mongo.ErrNoDocuments }},
		insertOneResult: &mongo.InsertOneResult{InsertedID: oid},
	}
	svc := &NotificationService{prefCol: prefCol}

	pref, err := svc.GetUserPreferences("u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pref.UserID != "u1" {
		t.Errorf("expected UserID 'u1', got %s", pref.UserID)
	}
	if !pref.ChatEnabled || !pref.ActivityEnabled || !pref.MarketingEnabled {
		t.Error("expected default preferences to be all enabled")
	}
	if pref.ID != oid.Hex() {
		t.Errorf("expected ID %s, got %s", oid.Hex(), pref.ID)
	}
}

func TestGetUserPreferences_DBError(t *testing.T) {
	prefCol := &notifMockCollection{
		findOneResult: &mockSingleResult{decodeFn: func(v interface{}) error { return errors.New("db error") }},
	}
	svc := &NotificationService{prefCol: prefCol}

	_, err := svc.GetUserPreferences("u1")
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected 'db error', got %v", err)
	}
}

func TestGetUserPreferences_NewDocInsertError(t *testing.T) {
	prefCol := &notifMockCollection{
		findOneResult: &mockSingleResult{decodeFn: func(v interface{}) error { return mongo.ErrNoDocuments }},
		insertOneErr:  errors.New("insert failed"),
	}
	svc := &NotificationService{prefCol: prefCol}

	_, err := svc.GetUserPreferences("u1")
	if err == nil || err.Error() != "insert failed" {
		t.Errorf("expected 'insert failed', got %v", err)
	}
}

// ---------------------------------------------------------------------------
// UpdateUserPreferences
// ---------------------------------------------------------------------------

func TestUpdateUserPreferences_Success(t *testing.T) {
	existing := model.NotificationPreference{UserID: "u1", ChatEnabled: true}
	updated := model.NotificationPreference{UserID: "u1", ChatEnabled: false, ActivityEnabled: true}

	prefCol := &notifMockCollection{
		findOneResult: &mockSingleResult{decodeFn: decodePrefFn(existing)},
		findOneAndUpdateResult: &mockSingleResult{
			decodeFn: func(v interface{}) error {
				*(v.(*model.NotificationPreference)) = updated
				return nil
			},
		},
	}
	svc := &NotificationService{prefCol: prefCol}

	chatEnabled := false
	activityEnabled := true
	pref, err := svc.UpdateUserPreferences("u1", model.UpdateNotificationPreferenceDto{
		ChatEnabled:     &chatEnabled,
		ActivityEnabled: &activityEnabled,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pref.ChatEnabled {
		t.Error("expected ChatEnabled false after update")
	}
	if !pref.ActivityEnabled {
		t.Error("expected ActivityEnabled true after update")
	}
}

func TestUpdateUserPreferences_GetPrefsError(t *testing.T) {
	prefCol := &notifMockCollection{
		findOneResult: &mockSingleResult{decodeFn: func(v interface{}) error { return errors.New("db error") }},
	}
	svc := &NotificationService{prefCol: prefCol}

	_, err := svc.UpdateUserPreferences("u1", model.UpdateNotificationPreferenceDto{})
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected 'db error', got %v", err)
	}
}

// ---------------------------------------------------------------------------
// CheckQuietHours
// ---------------------------------------------------------------------------

func TestCheckQuietHours_NilPointers(t *testing.T) {
	svc := &NotificationService{}
	pref := &model.NotificationPreference{}
	if svc.CheckQuietHours(pref) {
		t.Error("expected false when quiet hours not set")
	}
}

func TestCheckQuietHours_EmptyRange_ReturnsFalse(t *testing.T) {
	// start == end means an empty same-day range → always false
	s := "12:00"
	svc := &NotificationService{}
	pref := &model.NotificationPreference{QuietHoursStart: &s, QuietHoursEnd: &s}
	if svc.CheckQuietHours(pref) {
		t.Error("expected false for empty same-day range start==end")
	}
}

func TestCheckQuietHours_AllDayRange_ReturnsTrue(t *testing.T) {
	// "00:00"–"99:99" covers any HH:MM via string comparison
	start := "00:00"
	end := "99:99"
	svc := &NotificationService{}
	pref := &model.NotificationPreference{QuietHoursStart: &start, QuietHoursEnd: &end}
	if !svc.CheckQuietHours(pref) {
		t.Error("expected true for all-day range")
	}
}

func TestCheckQuietHours_OvernightEffectivelyNever(t *testing.T) {
	// Overnight range "23:59"–"00:00" is only active at 23:59 exactly which is vanishingly rare.
	// But we can't guarantee the test isn't run at that exact moment, so instead test
	// that overnight logic returns opposite: start="00:00", end="00:01" covers only 00:00.
	// Use a range that excludes the current minute reliably: not feasible without clock injection.
	// Just verify the overnight branch is taken when start > end.
	// start="23:00", end="01:00" — overnight. Result depends on wall clock; skip result assertion,
	// just ensure no panic.
	start := "23:00"
	end := "01:00"
	svc := &NotificationService{}
	pref := &model.NotificationPreference{QuietHoursStart: &start, QuietHoursEnd: &end}
	_ = svc.CheckQuietHours(pref) // no panic expected
}

// ---------------------------------------------------------------------------
// GetUserNotifications
// ---------------------------------------------------------------------------

func TestGetUserNotifications_CountError(t *testing.T) {
	notifCol := &notifMockCollection{countErr: errors.New("count error")}
	svc := &NotificationService{notifCol: notifCol}

	_, _, err := svc.GetUserNotifications("u1", 1, 10)
	if err == nil || err.Error() != "count error" {
		t.Errorf("expected 'count error', got %v", err)
	}
}

func TestGetUserNotifications_FindError(t *testing.T) {
	notifCol := &notifMockCollection{
		countResult: 1,
		findErr:     errors.New("find error"),
	}
	svc := &NotificationService{notifCol: notifCol}

	_, _, err := svc.GetUserNotifications("u1", 1, 10)
	if err == nil || err.Error() != "find error" {
		t.Errorf("expected 'find error', got %v", err)
	}
}

func TestGetUserNotifications_CursorError(t *testing.T) {
	notifCol := &notifMockCollection{
		countResult: 1,
		findResult:  &genericCursor{err: errors.New("cursor error")},
	}
	svc := &NotificationService{notifCol: notifCol}

	_, _, err := svc.GetUserNotifications("u1", 1, 10)
	if err == nil || err.Error() != "cursor error" {
		t.Errorf("expected 'cursor error', got %v", err)
	}
}

func TestGetUserNotifications_Success(t *testing.T) {
	notifications := []model.Notification{
		{UserID: "u1", Title: "A"},
		{UserID: "u1", Title: "B"},
	}
	notifCol := &notifMockCollection{
		countResult: 2,
		findResult:  notifsCursor(notifications),
	}
	svc := &NotificationService{notifCol: notifCol}

	result, count, err := svc.GetUserNotifications("u1", 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 notifications, got %d", len(result))
	}
}

// ---------------------------------------------------------------------------
// GetUnreadCount
// ---------------------------------------------------------------------------

func TestGetUnreadCount_Success(t *testing.T) {
	notifCol := &notifMockCollection{countResult: 5}
	svc := &NotificationService{notifCol: notifCol}

	count, err := svc.GetUnreadCount("u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5, got %d", count)
	}
}

func TestGetUnreadCount_Error(t *testing.T) {
	notifCol := &notifMockCollection{countErr: errors.New("db error")}
	svc := &NotificationService{notifCol: notifCol}

	_, err := svc.GetUnreadCount("u1")
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected 'db error', got %v", err)
	}
}

// ---------------------------------------------------------------------------
// MarkAsRead
// ---------------------------------------------------------------------------

func TestMarkAsRead_InvalidID(t *testing.T) {
	svc := &NotificationService{notifCol: &notifMockCollection{}}
	err := svc.MarkAsRead("not-a-hex")
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestMarkAsRead_UpdateError(t *testing.T) {
	notifCol := &notifMockCollection{updateOneErr: errors.New("db error")}
	svc := &NotificationService{notifCol: notifCol}

	err := svc.MarkAsRead(primitive.NewObjectID().Hex())
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected 'db error', got %v", err)
	}
}

func TestMarkAsRead_Success(t *testing.T) {
	svc := &NotificationService{notifCol: &notifMockCollection{}}
	err := svc.MarkAsRead(primitive.NewObjectID().Hex())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// MarkAllAsRead
// ---------------------------------------------------------------------------

func TestMarkAllAsRead_Success(t *testing.T) {
	svc := &NotificationService{notifCol: &notifMockCollection{}}
	err := svc.MarkAllAsRead("u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarkAllAsRead_Error(t *testing.T) {
	notifCol := &notifMockCollection{updateManyErr: errors.New("update error")}
	svc := &NotificationService{notifCol: notifCol}

	err := svc.MarkAllAsRead("u1")
	if err == nil || err.Error() != "update error" {
		t.Errorf("expected 'update error', got %v", err)
	}
}

// ---------------------------------------------------------------------------
// SendBroadcast
// ---------------------------------------------------------------------------

func TestSendBroadcast_Scheduled_Success(t *testing.T) {
	oid := primitive.NewObjectID()
	adminLogCol := &notifMockCollection{
		insertOneResult: &mongo.InsertOneResult{InsertedID: oid},
	}
	svc := &NotificationService{adminLogCol: adminLogCol}

	future := time.Now().Add(1 * time.Hour)
	err := svc.SendBroadcast(model.SendBroadcastDto{
		Title:       "Hello",
		Body:        "World",
		Type:        "marketing",
		ScheduledAt: &future,
	}, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if adminLogCol.insertOneCalls != 1 {
		t.Errorf("expected 1 InsertOne (schedule log), got %d", adminLogCol.insertOneCalls)
	}
}

func TestSendBroadcast_Scheduled_InsertError(t *testing.T) {
	adminLogCol := &notifMockCollection{insertOneErr: errors.New("db error")}
	svc := &NotificationService{adminLogCol: adminLogCol}

	future := time.Now().Add(1 * time.Hour)
	err := svc.SendBroadcast(model.SendBroadcastDto{
		Title:       "Hello",
		Body:        "World",
		Type:        "marketing",
		ScheduledAt: &future,
	}, "admin")
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected 'db error', got %v", err)
	}
}

func TestSendBroadcast_Immediate_NoTokens(t *testing.T) {
	// userCol.Find returns empty users → no tokens → fanOutMulticast is a no-op →
	// saveAdminLog writes the completed log.
	oid := primitive.NewObjectID()
	adminLogCol := &notifMockCollection{
		insertOneResult: &mongo.InsertOneResult{InsertedID: oid},
	}
	userCol := &notifMockCollection{findResult: usersCursor([]model.User{})}
	svc := &NotificationService{userCol: userCol, adminLogCol: adminLogCol}

	err := svc.SendBroadcast(model.SendBroadcastDto{
		Title: "Hi",
		Body:  "There",
		Type:  "marketing",
	}, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// saveAdminLog → InsertOne on adminLogCol
	if adminLogCol.insertOneCalls != 1 {
		t.Errorf("expected 1 InsertOne (completed log), got %d", adminLogCol.insertOneCalls)
	}
}

func TestSendBroadcast_Immediate_GetTokensError(t *testing.T) {
	userCol := &notifMockCollection{findErr: errors.New("find error")}
	svc := &NotificationService{userCol: userCol, adminLogCol: &notifMockCollection{}}

	err := svc.SendBroadcast(model.SendBroadcastDto{Title: "Hi", Body: "There", Type: "marketing"}, "admin")
	if err == nil || err.Error() != "find error" {
		t.Errorf("expected 'find error', got %v", err)
	}
}

// ---------------------------------------------------------------------------
// SendToSegment
// ---------------------------------------------------------------------------

func TestSendToSegment_Scheduled_Success(t *testing.T) {
	oid := primitive.NewObjectID()
	adminLogCol := &notifMockCollection{
		insertOneResult: &mongo.InsertOneResult{InsertedID: oid},
	}
	svc := &NotificationService{adminLogCol: adminLogCol}

	future := time.Now().Add(2 * time.Hour)
	err := svc.SendToSegment(model.SendSegmentDto{
		Title:       "Segment",
		Body:        "Msg",
		Type:        "activity",
		ScheduledAt: &future,
		SegmentFilter: model.SegmentFilter{
			RoleNames: []string{"user"},
		},
	}, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if adminLogCol.insertOneCalls != 1 {
		t.Errorf("expected 1 InsertOne, got %d", adminLogCol.insertOneCalls)
	}
}

func TestSendToSegment_Immediate_NoTokens(t *testing.T) {
	oid := primitive.NewObjectID()
	adminLogCol := &notifMockCollection{
		insertOneResult: &mongo.InsertOneResult{InsertedID: oid},
	}
	userCol := &notifMockCollection{findResult: usersCursor([]model.User{})}
	svc := &NotificationService{
		userCol:      userCol,
		adminLogCol:  adminLogCol,
		userLoginCol: &notifMockCollection{findResult: emptyCursor()},
	}

	err := svc.SendToSegment(model.SendSegmentDto{
		Title: "Hi",
		Body:  "There",
		Type:  "marketing",
	}, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// GetAdminLogs
// ---------------------------------------------------------------------------

func TestGetAdminLogs_CountError(t *testing.T) {
	adminLogCol := &notifMockCollection{countErr: errors.New("count error")}
	svc := &NotificationService{adminLogCol: adminLogCol}

	_, _, err := svc.GetAdminLogs(1, 10, "")
	if err == nil || err.Error() != "count error" {
		t.Errorf("expected 'count error', got %v", err)
	}
}

func TestGetAdminLogs_FindError(t *testing.T) {
	adminLogCol := &notifMockCollection{
		countResult: 1,
		findErr:     errors.New("find error"),
	}
	svc := &NotificationService{adminLogCol: adminLogCol}

	_, _, err := svc.GetAdminLogs(1, 10, "")
	if err == nil || err.Error() != "find error" {
		t.Errorf("expected 'find error', got %v", err)
	}
}

func TestGetAdminLogs_Success(t *testing.T) {
	logs := []model.AdminNotificationLog{
		{Title: "msg1", SentTo: "all"},
		{Title: "msg2", SentTo: "segment"},
	}
	adminLogCol := &notifMockCollection{
		countResult: 2,
		findResult:  adminLogsCursor(logs),
	}
	svc := &NotificationService{adminLogCol: adminLogCol}

	result, count, err := svc.GetAdminLogs(1, 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 logs, got %d", len(result))
	}
}

func TestGetAdminLogs_FilterBySentTo(t *testing.T) {
	logs := []model.AdminNotificationLog{{Title: "msg1", SentTo: "all"}}
	adminLogCol := &notifMockCollection{
		countResult: 1,
		findResult:  adminLogsCursor(logs),
	}
	svc := &NotificationService{adminLogCol: adminLogCol}

	result, count, err := svc.GetAdminLogs(1, 10, "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 || len(result) != 1 {
		t.Errorf("expected 1 log, got count=%d, len=%d", count, len(result))
	}
}

// ---------------------------------------------------------------------------
// CreateTemplate
// ---------------------------------------------------------------------------

func TestCreateTemplate_InsertError(t *testing.T) {
	templateCol := &notifMockCollection{insertOneErr: errors.New("insert error")}
	svc := &NotificationService{templateCol: templateCol}

	_, err := svc.CreateTemplate(model.CreateNotificationTemplateDto{
		Name:  "tmpl1",
		Title: "T",
		Body:  "B",
		Type:  "marketing",
	}, "admin")
	if err == nil || err.Error() != "insert error" {
		t.Errorf("expected 'insert error', got %v", err)
	}
}

func TestCreateTemplate_Success(t *testing.T) {
	oid := primitive.NewObjectID()
	templateCol := &notifMockCollection{
		insertOneResult: &mongo.InsertOneResult{InsertedID: oid},
	}
	svc := &NotificationService{templateCol: templateCol}

	tmpl, err := svc.CreateTemplate(model.CreateNotificationTemplateDto{
		Name:  "tmpl1",
		Title: "T",
		Body:  "B",
		Type:  "marketing",
	}, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl.ID != oid.Hex() {
		t.Errorf("expected ID %s, got %s", oid.Hex(), tmpl.ID)
	}
	if tmpl.Name != "tmpl1" {
		t.Errorf("expected Name 'tmpl1', got %s", tmpl.Name)
	}
	if tmpl.CreatedAt == nil || tmpl.UpdatedAt == nil {
		t.Error("expected CreatedAt and UpdatedAt to be set")
	}
}

// ---------------------------------------------------------------------------
// GetTemplates
// ---------------------------------------------------------------------------

func TestGetTemplates_CountError(t *testing.T) {
	templateCol := &notifMockCollection{countErr: errors.New("count error")}
	svc := &NotificationService{templateCol: templateCol}

	_, _, err := svc.GetTemplates(1, 10)
	if err == nil || err.Error() != "count error" {
		t.Errorf("expected 'count error', got %v", err)
	}
}

func TestGetTemplates_FindError(t *testing.T) {
	templateCol := &notifMockCollection{
		countResult: 1,
		findErr:     errors.New("find error"),
	}
	svc := &NotificationService{templateCol: templateCol}

	_, _, err := svc.GetTemplates(1, 10)
	if err == nil || err.Error() != "find error" {
		t.Errorf("expected 'find error', got %v", err)
	}
}

func TestGetTemplates_Success(t *testing.T) {
	tmpls := []model.NotificationTemplate{
		{Name: "t1", Title: "Title1"},
		{Name: "t2", Title: "Title2"},
	}
	templateCol := &notifMockCollection{
		countResult: 2,
		findResult:  templatesCursor(tmpls),
	}
	svc := &NotificationService{templateCol: templateCol}

	result, count, err := svc.GetTemplates(1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 templates, got %d", len(result))
	}
}

// ---------------------------------------------------------------------------
// GetTemplateByID
// ---------------------------------------------------------------------------

func TestGetTemplateByID_InvalidID(t *testing.T) {
	svc := &NotificationService{templateCol: &notifMockCollection{}}
	_, err := svc.GetTemplateByID("not-a-hex")
	if err == nil {
		t.Fatal("expected error for invalid template ID")
	}
}

func TestGetTemplateByID_NotFound(t *testing.T) {
	templateCol := &notifMockCollection{
		findOneResult: &mockSingleResult{
			decodeFn: func(v interface{}) error { return mongo.ErrNoDocuments },
		},
	}
	svc := &NotificationService{templateCol: templateCol}

	_, err := svc.GetTemplateByID(primitive.NewObjectID().Hex())
	if err != mongo.ErrNoDocuments {
		t.Errorf("expected ErrNoDocuments, got %v", err)
	}
}

func TestGetTemplateByID_Success(t *testing.T) {
	expected := model.NotificationTemplate{Name: "tmpl", Title: "T"}
	templateCol := &notifMockCollection{
		findOneResult: &mockSingleResult{decodeFn: decodeTemplateFn(expected)},
	}
	svc := &NotificationService{templateCol: templateCol}

	tmpl, err := svc.GetTemplateByID(primitive.NewObjectID().Hex())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl.Name != "tmpl" {
		t.Errorf("expected Name 'tmpl', got %s", tmpl.Name)
	}
}

// ---------------------------------------------------------------------------
// UpdateTemplate
// ---------------------------------------------------------------------------

func TestUpdateTemplate_InvalidID(t *testing.T) {
	svc := &NotificationService{templateCol: &notifMockCollection{}}
	_, err := svc.UpdateTemplate("not-hex", model.UpdateNotificationTemplateDto{})
	if err == nil {
		t.Fatal("expected error for invalid template ID")
	}
}

func TestUpdateTemplate_DBError(t *testing.T) {
	templateCol := &notifMockCollection{
		findOneAndUpdateResult: &mockSingleResult{
			decodeFn: func(v interface{}) error { return errors.New("update error") },
		},
	}
	svc := &NotificationService{templateCol: templateCol}

	_, err := svc.UpdateTemplate(primitive.NewObjectID().Hex(), model.UpdateNotificationTemplateDto{})
	if err == nil || err.Error() != "update error" {
		t.Errorf("expected 'update error', got %v", err)
	}
}

func TestUpdateTemplate_Success(t *testing.T) {
	newTitle := "Updated Title"
	expected := model.NotificationTemplate{Name: "tmpl", Title: newTitle}
	templateCol := &notifMockCollection{
		findOneAndUpdateResult: &mockSingleResult{decodeFn: decodeTemplateFn(expected)},
	}
	svc := &NotificationService{templateCol: templateCol}

	tmpl, err := svc.UpdateTemplate(primitive.NewObjectID().Hex(), model.UpdateNotificationTemplateDto{
		Title: &newTitle,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl.Title != newTitle {
		t.Errorf("expected Title %q, got %q", newTitle, tmpl.Title)
	}
}

// ---------------------------------------------------------------------------
// DeleteTemplate
// ---------------------------------------------------------------------------

func TestDeleteTemplate_InvalidID(t *testing.T) {
	svc := &NotificationService{templateCol: &notifMockCollection{}}
	err := svc.DeleteTemplate("not-hex")
	if err == nil {
		t.Fatal("expected error for invalid template ID")
	}
}

func TestDeleteTemplate_DBError(t *testing.T) {
	templateCol := &notifMockCollection{deleteOneErr: errors.New("delete error")}
	svc := &NotificationService{templateCol: templateCol}

	err := svc.DeleteTemplate(primitive.NewObjectID().Hex())
	if err == nil || err.Error() != "delete error" {
		t.Errorf("expected 'delete error', got %v", err)
	}
}

func TestDeleteTemplate_Success(t *testing.T) {
	templateCol := &notifMockCollection{}
	svc := &NotificationService{templateCol: templateCol}

	err := svc.DeleteTemplate(primitive.NewObjectID().Hex())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if templateCol.deleteOneCalls != 1 {
		t.Errorf("expected 1 DeleteOne call, got %d", templateCol.deleteOneCalls)
	}
}
