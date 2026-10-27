package jobs

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/shinwookim/ytdlp-manager/internal/downloader"
	"github.com/shinwookim/ytdlp-manager/internal/store/sqlc"
)

type Worker struct {
	service    *Service
	downloader *downloader.Downloader
	interval   time.Duration
}

func NewWorker(s *Service, d *downloader.Downloader) *Worker {
	return &Worker{
		service:    s,
		downloader: d,
		interval:   time.Second,
	}
}

func (w *Worker) Run(ctx context.Context) {
	// Any downloads interrupted by a previous container process are
	// re-queued. yt-dlp may restart a partial download if supported.
	if err := w.service.queries.RecoverInterruptedJobs(ctx); err != nil {
		log.Printf("recovering interrupted jobs: %v", err)
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		if ctx.Err() != nil {
			return
		}

		job, err := w.service.queries.ClaimNextJob(ctx)
		switch {
		case err == nil:
			w.runJob(ctx, job)
			continue
		case errors.Is(err, sql.ErrNoRows):
			// No queued work yet.
		case ctx.Err() != nil:
			return
		default:
			log.Printf("claiming job: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (w *Worker) runJob(parent context.Context, job sqlc.Job) {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	w.service.RegisterRunning(job.ID, cancel)
	defer w.service.UnregisterRunning(job.ID)

	result, err := w.downloader.Download(
		ctx,
		job.Url,
		job.Mode,
		func(video_title string) {
			if updateErr := w.service.queries.UpdateTitle(context.Background(),
				sqlc.UpdateTitleParams{
					ID:    job.ID,
					Title: video_title,
				},
			); updateErr != nil {
				log.Printf("updating video title for job %d: %v", job.ID, updateErr)
			}
		},
		func(p downloader.Progress) {
			if updateErr := w.service.queries.UpdateProgress(
				context.Background(),
				sqlc.UpdateProgressParams{
					Progress: p.Percent,
					Speed:    p.Speed,
					Eta:      p.ETA,
					ID:       job.ID,
				},
			); updateErr != nil {
				log.Printf("updating progress for job %d: %v", job.ID, updateErr)
			}
		},
	)

	if ctx.Err() != nil {
		if err := w.service.queries.CancelRunningJob(
			context.Background(), job.ID,
		); err != nil {
			log.Printf("marking job %d cancelled: %v", job.ID, err)
		}
		return
	}

	if err != nil {
		if updateErr := w.service.queries.FailJob(
			context.Background(),
			sqlc.FailJobParams{
				Error: err.Error(),
				ID:    job.ID,
			},
		); updateErr != nil {
			log.Printf("marking job %d failed: %v", job.ID, updateErr)
		}
		return
	}

	if err := w.service.queries.CompleteJob(
		context.Background(),
		sqlc.CompleteJobParams{
			Output: result.Output,
			ID:     job.ID,
		},
	); err != nil {
		log.Printf("marking job %d completed: %v", job.ID, err)
	}
}
