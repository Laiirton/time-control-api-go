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
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "event_type é obrigatório"})
		return
	}
	if len(eventType) > 100 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "event_type deve ter no máximo 100 caracteres"})
		return
	}

	latitude, err := parseCoordinate(c.PostForm("latitude"), -90, 90)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "latitude inválida"})
		return
	}

	longitude, err := parseCoordinate(c.PostForm("longitude"), -180, 180)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "longitude inválida"})
		return
	}

	fileHeader, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "foto é obrigatória"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao registrar ponto"})
		return
	}

	h.attachSignedPhotoURL(c.Request.Context(), record)

	c.JSON(http.StatusCreated, record)
}

func (h *TimeRecordHandler) Me(c *gin.Context) {
	userID := c.GetInt64("user_id")

	records, err := h.repo.FindByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar pontos"})
		return
	}

	h.attachSignedPhotoURLs(c.Request.Context(), records)

	c.JSON(http.StatusOK, records)
}

func (h *TimeRecordHandler) MeToday(c *gin.Context) {
	userID := c.GetInt64("user_id")

	records, err := h.repo.FindByUserAndDate(userID, time.Now().UTC())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar pontos de hoje"})
		return
	}

	h.attachSignedPhotoURLs(c.Request.Context(), records)

	c.JSON(http.StatusOK, records)
}

func (h *TimeRecordHandler) List(c *gin.Context) {
	records, err := h.repo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar pontos"})
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
