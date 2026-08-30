// Package forge talks to a code forge and returns the repositories it hosts,
// with the two fields that matter for documentation: the description and the
// topics. Only the standard library is used, on purpose: a tool that generates
// documentation should not itself become a dependency problem.
package forge

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Repo is the subset of a repository we need. Everything else is noise.
type Repo struct {
	Name        string
	URL         string
	Description string
	Topics      []string
	Archived    bool
}

// Client fetches repositories from a forge.
type Client interface {
	Repos() ([]Repo, error)
}

var httpClient = &http.Client{Timeout: 20 * time.Second}

func get(req *http.Request, out any) (*http.Response, error) {
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("%s: %s", req.URL.Host, resp.Status)
	}
	return resp, json.NewDecoder(resp.Body).Decode(out)
}

// GitLab reads every project of a group, including subgroups.
type GitLab struct {
	Host    string // gitlab.com, or your own instance
	GroupID string
	Token   string // optional: only needed for private groups
}

func (g GitLab) Repos() ([]Repo, error) {
	var all []Repo
	for page := 1; ; page++ {
		endpoint := fmt.Sprintf("https://%s/api/v4/groups/%s/projects", g.Host, url.PathEscape(g.GroupID))
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		q := req.URL.Query()
		q.Set("include_subgroups", "true")
		q.Set("per_page", "100")
		q.Set("page", strconv.Itoa(page))
		req.URL.RawQuery = q.Encode()
		if g.Token != "" {
			req.Header.Set("PRIVATE-TOKEN", g.Token)
		}

		var batch []struct {
			Name        string   `json:"name"`
			WebURL      string   `json:"web_url"`
			Description string   `json:"description"`
			Topics      []string `json:"topics"`
			Archived    bool     `json:"archived"`
		}
		if _, err := get(req, &batch); err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			return all, nil
		}
		for _, p := range batch {
			all = append(all, Repo{p.Name, p.WebURL, p.Description, p.Topics, p.Archived})
		}
	}
}

// GitHub reads every repository of an organisation or a user.
type GitHub struct {
	Owner string
	Token string // optional: only needed for private repositories
}

func (g GitHub) Repos() ([]Repo, error) {
	var all []Repo
	for page := 1; ; page++ {
		endpoint := fmt.Sprintf("https://api.github.com/users/%s/repos?per_page=100&page=%d",
			url.PathEscape(g.Owner), page)
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		if g.Token != "" {
			req.Header.Set("Authorization", "Bearer "+g.Token)
		}

		var batch []struct {
			Name        string   `json:"name"`
			HTMLURL     string   `json:"html_url"`
			Description string   `json:"description"`
			Topics      []string `json:"topics"`
			Archived    bool     `json:"archived"`
		}
		if _, err := get(req, &batch); err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			return all, nil
		}
		for _, p := range batch {
			all = append(all, Repo{p.Name, p.HTMLURL, p.Description, p.Topics, p.Archived})
		}
	}
}
