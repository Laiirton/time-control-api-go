package handlers

import (
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

	fileHeader, err := c.FormFile("foto")
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar pontos"})
		return
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar pontos de hoje"})
		return
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar pontos"})
		return
	}

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
