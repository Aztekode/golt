package updatecheck

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Result struct {
	LatestVersion    string
	URL              string
	UpdateAvailable  bool
	ReleaseVerified  bool
	VerifyError      error
	UpdateCheckError error
}

type Options struct {
	Disable bool
	Now     time.Time
}

type ReleaseVerifier func(ctx context.Context, client *http.Client, rel GitHubRelease) (bool, error)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func CheckLatest(ctx context.Context, client *http.Client, currentVersion string, opt Options, verify ReleaseVerifier) Result {
	if opt.Now.IsZero() {
		opt.Now = time.Now()
	}

	if opt.Disable || os.Getenv("GOLT_NO_UPDATE_CHECK") == "1" {
		return Result{}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/Aztekode/golt/releases/latest", nil)
	if err != nil {
		return Result{UpdateCheckError: err}
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "golt")

	resp, err := client.Do(req)
	if err != nil {
		return Result{UpdateCheckError: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{UpdateCheckError: fmt.Errorf("unexpected status: %s", resp.Status)}
	}

	var rel GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return Result{UpdateCheckError: err}
	}

	res := Result{
		LatestVersion: rel.TagName,
		URL:           rel.HTMLURL,
	}

	current, err := ParseSemver(currentVersion)
	if err != nil {
		return res
	}
	latest, err := ParseSemver(rel.TagName)
	if err != nil {
		return res
	}

	if verify != nil {
		ok, verr := verify(ctx, client, rel)
		res.ReleaseVerified = ok
		res.VerifyError = verr
	}

	if current.LessThan(latest) {
		res.UpdateAvailable = true
	}

	return res
}
