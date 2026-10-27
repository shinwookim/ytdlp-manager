package app

import (
	"database/sql"

	"github.com/shinwookim/ytdlp-manager/internal/downloader"
	"github.com/shinwookim/ytdlp-manager/internal/httpapi"
	"github.com/shinwookim/ytdlp-manager/internal/jobs"
	"github.com/shinwookim/ytdlp-manager/internal/store/sqlc"
)

type App struct {
	HTTP   *httpapi.Server
	Worker *jobs.Worker
}

func New(db *sql.DB, downloadPath string) (*App, error) {
	queries := sqlc.New(db)
	jobService := jobs.NewService(queries)

	dl := downloader.New(downloadPath)
	worker := jobs.NewWorker(jobService, dl)

	server, err := httpapi.NewServer(jobService)
	if err != nil {
		return nil, err
	}

	return &App{
		HTTP:   server,
		Worker: worker,
	}, nil
}
