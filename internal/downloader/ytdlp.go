package downloader

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Progress struct {
	Percent float64
	Speed   string
	ETA     string
}

type Result struct {
	Output string
}

type Downloader struct {
	binary      string
	downloadDir string
}

func New(downloadDir string) *Downloader {
	return &Downloader{binary: "yt-dlp", downloadDir: downloadDir}
}

func (d *Downloader) Download(ctx context.Context, url string, mode string, setTitle func(string), onProgress func(Progress)) (Result, error) {
	args := d.buildArgs(mode, url)

	cmd := exec.CommandContext(ctx, d.binary, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Result{}, fmt.Errorf("creating stdout pipe: %w", err)
	}

	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return Result{}, fmt.Errorf("starting yt-dlp: %w", err)
	}

	var result Result
	var video_title string
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 4096), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		switch {
		case strings.HasPrefix(line, "TITLE|"):
			video_title = strings.TrimPrefix(line, "TITLE|")
			setTitle(video_title)
			continue
		case strings.HasPrefix(line, "FILE|"):
			result.Output = strings.TrimPrefix(line, "FILE|")
			continue
		case !strings.HasPrefix(line, "PROGRESS|"):
			continue
		}

		fields := strings.SplitN(strings.TrimPrefix(line, "PROGRESS|"), "|", 3)
		if len(fields) != 3 {
			continue
		}
		p := Progress{
			Speed: strings.TrimSpace(fields[1]),
			ETA:   strings.TrimSpace(fields[2]),
		}

		percent := strings.TrimSpace(strings.TrimSuffix(fields[0], "%"))
		percent = strings.ReplaceAll(percent, ",", "")

		if value, err := strconv.ParseFloat(percent, 64); err == nil {
			p.Percent = value
		}

		if onProgress != nil {
			onProgress(p)
		}
	}
	scanErr := scanner.Err()
	waitErr := cmd.Wait()

	if ctx.Err() != nil {
		return result, ctx.Err()
	}
	if scanErr != nil {
		return result, fmt.Errorf("reading yt-dlp output: %w", scanErr)
	}
	if waitErr != nil {
		return result, fmt.Errorf(
			"yt-dlp failed: %w: %s",
			waitErr,
			strings.TrimSpace(stderr.String()),
		)
	}

	return result, nil

}
