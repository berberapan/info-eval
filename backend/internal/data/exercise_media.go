package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type MediaType string

const (
	ImageType MediaType = "image"
	VideoType MediaType = "video"
	AudioType MediaType = "audio"
)

type ExerciseMedia struct {
	ID        uuid.UUID `json:"id"`
	MediaURL  string    `json:"media_url"`
	MediaType MediaType `json:"media_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ExerciseMediaModel struct {
	DB *sql.DB
}

func (em *ExerciseMediaModel) GetByExerciseID(exerciseID uuid.UUID) ([]ExerciseMedia, error) {
	query := `
	SELECT id, media_url, media_type, created_at, updated_at
	FROM exercise_media
	WHERE exercise_id = $1`
	var mediaLSlice []ExerciseMedia
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := em.DB.QueryContext(ctx, query, exerciseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var media ExerciseMedia
		if err := rows.Scan(&media.ID, &media.MediaURL, &media.MediaType, &media.CreatedAt, &media.UpdatedAt); err != nil {
			return nil, err
		}
		mediaLSlice = append(mediaLSlice, media)
	}
	return mediaLSlice, nil
}
