package jobs

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/shinwookim/ytdlp-manager/internal/store/sqlc"
)

type Service struct {
	queries *sqlc.Queries

	mu      sync.Mutex
	running map[int64]context.CancelFunc
}

func NewService(q *sqlc.Queries) *Service {
	return &Service{
		queries: q,
		running: make(map[int64]context.CancelFunc),
	}
}

func (s *Service) Create(ctx context.Context, rawURL, mode string) (sqlc.Job, error) {

	u, err := url.ParseRequestURI(strings.TrimSpace(rawURL))
	if err != nil || u.Hostname() == "" {
		return sqlc.Job{}, fmt.Errorf("invalid URL")
	}

	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	if u.Scheme != "https" && u.Scheme != "http" {
		return sqlc.Job{}, fmt.Errorf("only HTTP and HTTPS URLs are supported")
	}

	if !isAllowedHost(host) {
		return sqlc.Job{}, fmt.Errorf("unsupported URL host")
	}

	if mode != "audio" && mode != "video" {
		return sqlc.Job{}, fmt.Errorf("mode must be audio or video")
	}

	return s.queries.CreateJob(ctx, sqlc.CreateJobParams{
		Url:  rawURL,
		Mode: mode,
	})
}

func (s *Service) List(ctx context.Context) ([]sqlc.Job, error) {
	return s.queries.ListJobs(ctx)
}

func (s *Service) Get(ctx context.Context, id int64) (sqlc.Job, error) {
	return s.queries.GetJob(ctx, id)
}

func (s *Service) RegisterRunning(id int64, cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running[id] = cancel
}

func (s *Service) UnregisterRunning(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.running, id)
}

func (s *Service) Cancel(ctx context.Context, id int64) error {
	s.mu.Lock()
	cancel, active := s.running[id]
	s.mu.Unlock()

	if active {
		cancel()
		return nil
	}

	_, err := s.queries.CancelQueuedJob(ctx, id)
	return err
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	_, err := s.queries.DeleteJob(ctx, id)
	return err
}

func (s *Service) ClearFinishedJobs(ctx context.Context) error {
	_, err := s.queries.ClearFinishedJobs(ctx)
	return err
}
