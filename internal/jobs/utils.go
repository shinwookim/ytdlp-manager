package jobs

func isAllowedHost(host string) bool {
	allowed := []string{
		"youtube.com",
		"www.youtube.com",
		"m.youtube.com",
		"youtu.be",
		"music.youtube.com",
		"youtube-nocookie.com",
		"www.youtube-nocookie.com",
	}

	for _, domain := range allowed {
		if host == domain {
			return true
		}
	}
	return false
}
