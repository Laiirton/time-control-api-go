package repository

import (
	"database/sql"
	"time"

	"github.com/Laiirton/time-control-api-go/internal/models"
)

type TimeRecordRepository struct {
	db *sql.DB
}

func NewTimeRecordRepository(db *sql.DB) *TimeRecordRepository {
	return &TimeRecordRepository{db: db}
}

func (r *TimeRecordRepository) Create(record *models.TimeRecord) error {
	query := `
		INSERT INTO time_records (user_id, event_type, latitude, longitude, photo_path, photo_url, recorded_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	now := time.Now().UTC()
	if record.RecordedAt.IsZero() {
		record.RecordedAt = now
	}
	record.CreatedAt = now

	return r.db.QueryRow(
		query,
		record.UserID,
		record.EventType,
		record.Latitude,
		record.Longitude,
		record.PhotoPath,
		record.PhotoURL,
		record.RecordedAt,
		record.CreatedAt,
	).Scan(&record.ID)
}

func (r *TimeRecordRepository) CountByUserAndDate(userID int64, date time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM time_records
		WHERE user_id = $1
		  AND recorded_at::date = $2::date`

	var count int
	err := r.db.QueryRow(query, userID, date.UTC()).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *TimeRecordRepository) FindByUserAndDate(userID int64, date time.Time) ([]models.TimeRecord, error) {
	query := `
		SELECT id, user_id, event_type, latitude, longitude, photo_path, photo_url, recorded_at, created_at
		FROM time_records
		WHERE user_id = $1
		  AND recorded_at::date = $2::date
		ORDER BY recorded_at ASC, id ASC`

	rows, err := r.db.Query(query, userID, date.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.TimeRecord
	for rows.Next() {
		var item models.TimeRecord
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.EventType,
			&item.Latitude,
			&item.Longitude,
			&item.PhotoPath,
			&item.PhotoURL,
			&item.RecordedAt,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (r *TimeRecordRepository) FindByUser(userID int64) ([]models.TimeRecord, error) {
	query := `
		SELECT id, user_id, event_type, latitude, longitude, photo_path, photo_url, recorded_at, created_at
		FROM time_records
		WHERE user_id = $1
		ORDER BY recorded_at DESC, id DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.TimeRecord
	for rows.Next() {
		var item models.TimeRecord
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.EventType,
			&item.Latitude,
			&item.Longitude,
			&item.PhotoPath,
			&item.PhotoURL,
			&item.RecordedAt,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (r *TimeRecordRepository) FindAll() ([]models.TimeRecord, error) {
	query := `
		SELECT id, user_id, event_type, latitude, longitude, photo_path, photo_url, recorded_at, created_at
		FROM time_records
		ORDER BY recorded_at DESC, id DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.TimeRecord
	for rows.Next() {
		var item models.TimeRecord
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.EventType,
			&item.Latitude,
			&item.Longitude,
			&item.PhotoPath,
			&item.PhotoURL,
			&item.RecordedAt,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}
