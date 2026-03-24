package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Laiirton/time-control-api-go/internal/models"
	"github.com/Laiirton/time-control-api-go/internal/repository"
	"github.com/Laiirton/time-control-api-go/internal/storage"
	"github.com/gin-gonic/gin"
)

type TimeRecordHandler struct {
	repo    *repository.TimeRecordRepository
	storage *storage.SupabaseStorage
}

const defaultSignedPhotoURLTTLSeconds = 7 * 24 * 60 * 60

func NewTimeRecordHandler(repo *repository.TimeRecordRepository, storage *storage.SupabaseStorage) *TimeRecordHandler {
	return &TimeRecordHandler{repo: repo, storage: storage}
}

func (h *TimeRecordHandler) Clock(c *gin.Context) {
	userID := c.GetInt64("user_id")
	eventType := strings.TrimSpace(c.PostForm("event_type"))
	if eventType == "" {
		eventType = strings.TrimSpace(c.PostForm("event"))
	}
	if eventType == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "event_type is required"})
		return
	}
	if len(eventType) > 100 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "event_type must be at most 100 characters"})
		return
	}

	latitude, err := parseCoordinate(c.PostForm("latitude"), -90, 90)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid latitude"})
		return
	}

	longitude, err := parseCoordinate(c.PostForm("longitude"), -180, 180)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid longitude"})
		return
	}

	fileHeader, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "photo is required"})
		return
	}

	photoPath, photoURL, err := h.storage.UploadClockPhoto(c.Request.Context(), userID, fileHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record := &models.TimeRecord{
		UserID:    userID,
		EventType: eventType,
		Latitude:  latitude,
		Longitude: longitude,
		PhotoPath: photoPath,
		PhotoURL:  photoURL,
	}

	if err := h.repo.Create(record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clock in"})
		return
	}

	h.attachSignedPhotoURL(c.Request.Context(), record)

	c.JSON(http.StatusCreated, record)
}

func (h *TimeRecordHandler) Me(c *gin.Context) {
	userID := c.GetInt64("user_id")

	records, err := h.repo.FindByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch time records"})
		return
	}

	h.attachSignedPhotoURLs(c.Request.Context(), records)

	c.JSON(http.StatusOK, records)
}

func (h *TimeRecordHandler) MeToday(c *gin.Context) {
	userID := c.GetInt64("user_id")
	now := time.Now().UTC()

	records, err := h.repo.FindByUserAndDate(userID, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch today's time records"})
		return
	}

	h.attachSignedPhotoURLs(c.Request.Context(), records)

	response := models.TimeRecordTodayResponse{
		Data:          records,
		WorkedHours:   durationToHours(calculateWorkedDuration(records, now)),
		ExpectedHours: durationToHours(expectedWorkDuration(now)),
		NextEventType: nextEventType(records),
	}

	c.JSON(http.StatusOK, response)
}

func (h *TimeRecordHandler) List(c *gin.Context) {
	records, err := h.repo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch time records"})
		return
	}

	h.attachSignedPhotoURLs(c.Request.Context(), records)

	c.JSON(http.StatusOK, records)
}

func parseCoordinate(value string, min, max float64) (float64, error) {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	if parsed < min || parsed > max {
		return 0, strconv.ErrRange
	}
	return parsed, nil
}

func (h *TimeRecordHandler) attachSignedPhotoURLs(ctx context.Context, records []models.TimeRecord) {
	for i := range records {
		h.attachSignedPhotoURL(ctx, &records[i])
	}
}

func (h *TimeRecordHandler) attachSignedPhotoURL(ctx context.Context, record *models.TimeRecord) {
	if record == nil || strings.TrimSpace(record.PhotoPath) == "" {
		return
	}

	signedURL, err := h.storage.GenerateSignedPhotoURL(ctx, record.PhotoPath, defaultSignedPhotoURLTTLSeconds)
	if err != nil {
		return
	}

	record.PhotoURL = signedURL
}

func calculateWorkedDuration(records []models.TimeRecord, now time.Time) time.Duration {
	var total time.Duration
	var currentEntry *time.Time
	usedTypedEvents := false

	for _, record := range records {
		eventType := normalizeEventType(record.EventType)
		recordedAt := record.RecordedAt.UTC()

		switch {
		case isEntryEventType(eventType):
			usedTypedEvents = true
			t := recordedAt
			currentEntry = &t
		case isExitEventType(eventType):
			usedTypedEvents = true
			if currentEntry != nil && recordedAt.After(*currentEntry) {
				total += recordedAt.Sub(*currentEntry)
				currentEntry = nil
			}
		}
	}

	if usedTypedEvents {
		if currentEntry != nil && now.After(*currentEntry) {
			total += now.Sub(*currentEntry)
		}
		return total
	}

	for i := 0; i+1 < len(records); i += 2 {
		start := records[i].RecordedAt.UTC()
		end := records[i+1].RecordedAt.UTC()
		if end.After(start) {
			total += end.Sub(start)
		}
	}

	if len(records)%2 == 1 {
		last := records[len(records)-1].RecordedAt.UTC()
		if now.After(last) {
			total += now.Sub(last)
		}
	}

	return total
}

func expectedWorkDuration(date time.Time) time.Duration {
	_ = date
	return 8 * time.Hour
}

func nextEventType(records []models.TimeRecord) string {
	if len(records) == 0 {
		return "entrada"
	}

	lastEvent := normalizeEventType(records[len(records)-1].EventType)

	switch {
	case isEntryEventType(lastEvent):
		return "saida"
	case isExitEventType(lastEvent):
		return "entrada"
	default:
		if len(records)%2 == 0 {
			return "entrada"
		}
		return "saida"
	}
}

func durationToHours(duration time.Duration) float64 {
	return duration.Hours()
}

func normalizeEventType(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isEntryEventType(value string) bool {
	return strings.Contains(value, "entrada")
}

func isExitEventType(value string) bool {
	return strings.Contains(value, "saida") || strings.Contains(value, "saída")
}
