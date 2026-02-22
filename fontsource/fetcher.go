package fontsource

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type fontMetadata struct {
	ID       string `json:"id"`
	Family   string `json:"family"`
	Variants map[string]map[string]map[string]struct {
		URL struct {
			TTF string `json:"ttf"`
		} `json:"url"`
	} `json:"variants"`
}

type fontSearchItem struct {
	ID string `json:"id"`
}

type DownloadTask struct {
	URL  string
	Path string
}

// GetDownloadTasks discovers all available TTF variants for a family.
func GetDownloadTasks(ctx context.Context, family string) ([]DownloadTask, error) {
	id, err := findFontID(ctx, family)

	if err != nil {
		return nil, fmt.Errorf("finding font id: %w", err)
	}

	apiURL := fmt.Sprintf("https://api.fontsource.org/v1/fonts/%s", id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)

	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var meta fontMetadata
	err = json.NewDecoder(resp.Body).Decode(&meta)

	if err != nil {
		return nil, fmt.Errorf("decoding metadata: %w", err)
	}

	var tasks []DownloadTask
	familyDir := strings.ReplaceAll(meta.Family, " ", "")

	for weight, styles := range meta.Variants {
		for style, subsets := range styles {
			for subsetName, subset := range subsets {
				if subset.URL.TTF == "" {
					continue
				}

				filename := fmt.Sprintf("%s-%s-%s.ttf", weight, style, subsetName)
				tasks = append(tasks, DownloadTask{
					URL:  subset.URL.TTF,
					Path: filepath.Join("kodata", "fonts", familyDir, filename),
				})
			}
		}
	}

	return tasks, nil
}

func findFontID(ctx context.Context, family string) (string, error) {
	query := url.Values{}
	query.Set("family", family)
	query.Set("subsets", "latin")

	apiURL := "https://api.fontsource.org/v1/fonts?" + query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)

	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var results []fontSearchItem
	err = json.NewDecoder(resp.Body).Decode(&results)

	if err != nil {
		return "", fmt.Errorf("decoding search results: %w", err)
	}

	if len(results) == 0 {
		return "", fmt.Errorf("no font found for family %q", family)
	}

	return results[0].ID, nil
}

