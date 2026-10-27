package downloader

import (
	"path/filepath"
)

func (d *Downloader) buildArgs(mode string, url string) []string {
	args := []string{
		"--encoding", "utf-8",
		"--newline",
		"--no-colors",
		"--no-warnings",
		"--progress",
		"--progress-template",
		"download:PROGRESS|%(progress._percent_str)s|%(progress._speed_str)s|%(progress._eta_str)s",
		"--print",
		"before_dl:TITLE|%(title)s",
		"--print",
		"after_move:FILE|%(filepath)s",
		"--no-overwrites",
		"--windows-filenames",
		"-o", filepath.Join(d.downloadDir, "%(title).180B [%(id)s].%(ext)s")}
	if mode == "audio" {
		args = append(args, "-x", "--audio-format", "best")
	}
	return append(args, "--", url)
}
