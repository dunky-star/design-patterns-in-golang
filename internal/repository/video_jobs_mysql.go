package repository

import (
	"context"
	"time"

	"go-breeders/models"
)

const videoJobColumns = `id, input_media_key, encoding_type, status,
	coalesce(output_reference, ''), coalesce(error_message, ''),
	created_at, updated_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanVideoJob(scanner rowScanner) (*models.VideoJob, error) {
	var job models.VideoJob
	err := scanner.Scan(
		&job.ID,
		&job.InputMediaKey,
		&job.EncodingType,
		&job.Status,
		&job.OutputReference,
		&job.ErrorMessage,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (m *mysqlRepository) AllVideoJobs() ([]*models.VideoJob, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select ` + videoJobColumns + ` from video_jobs order by id`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.VideoJob
	for rows.Next() {
		job, err := scanVideoJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (m *mysqlRepository) GetVideoJobByID(id int) (*models.VideoJob, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select ` + videoJobColumns + ` from video_jobs where id = ?`
	return scanVideoJob(m.DB.QueryRowContext(ctx, query, id))
}

func (m *mysqlRepository) ClaimVideoJob(id int, encodingType string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(
		ctx,
		`update video_jobs
		 set encoding_type = ?, status = 'processing',
		     output_reference = null, error_message = null
		 where id = ? and status <> 'processing'`,
		encodingType,
		id,
	)
	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected == 1, nil
}

func (m *mysqlRepository) CompleteVideoJob(id int, outputReference string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(
		ctx,
		`update video_jobs
		 set status = 'completed', output_reference = ?, error_message = null
		 where id = ?`,
		outputReference,
		id,
	)
	return err
}

func (m *mysqlRepository) FailVideoJob(id int, errorMessage string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(
		ctx,
		`update video_jobs
		 set status = 'failed', output_reference = null, error_message = ?
		 where id = ?`,
		errorMessage,
		id,
	)
	return err
}

func (m *mysqlRepository) ResetProcessingVideoJobs() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(
		ctx,
		`update video_jobs
		 set status = 'pending', error_message = 'Processing interrupted by application restart'
		 where status = 'processing'`,
	)
	return err
}
