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

	fileHeader, err := c.FormFile("foto")
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 15
	}

	records, total, err := h.repo.FindByUser(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch time records"})
		return
	}

	h.attachSignedPhotoURLs(c.Request.Context(), records)

	c.JSON(http.StatusOK, models.TimeRecordListResponse{
		Data:  records,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

func (h *TimeRecordHandler) MeToday(c *gin.Context) {
	userID := c.GetInt64("user_id")

	records, err := h.repo.FindByUserAndDate(userID, time.Now().UTC())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch today's time records"})
		return
	}

	h.attachSignedPhotoURLs(c.Request.Context(), records)

	c.JSON(http.StatusOK, models.TimeRecordTodayResponse{
		Data: records,
	})
}

func (h *TimeRecordHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 15
	}

	records, total, err := h.repo.FindAll(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch time records"})
		return
	}

	h.attachSignedPhotoURLs(c.Request.Context(), records)

	c.JSON(http.StatusOK, models.TimeRecordListResponse{
		Data:  records,
		Total: total,
		Page:  page,
		Limit: limit,
	})
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
