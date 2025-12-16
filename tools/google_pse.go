package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// GooglePSETool represents the Google PSE tool definition
type GooglePSETool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// GetGooglePSETool returns the Google PSE tool definition
func GetGooglePSETool() GooglePSETool {
	return GooglePSETool{
		Name:        "google_pse_search",
		Description: "Search the web using Google Programmable Search Engine",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "The search query",
				},
				"num": map[string]interface{}{
					"type":        "integer",
					"description": "Number of results to return (1-10, default: 10)",
					"default":     10,
					"minimum":     1,
					"maximum":     10,
				},
				"start": map[string]interface{}{
					"type":        "integer",
					"description": "Start index for pagination (default: 1)",
					"default":     1,
					"minimum":     1,
				},
			},
			"required": []string{"query"},
		},
	}
}

// GooglePSESearchResult represents a single search result
type GooglePSESearchResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// GooglePSEResponse represents the Google PSE API response
type GooglePSEResponse struct {
	Items []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"items"`
	SearchInformation struct {
		TotalResults string `json:"totalResults"`
	} `json:"searchInformation"`
}

// GooglePSEConfig holds the configuration for Google PSE
type GooglePSEConfig struct {
	APIKey         string
	SearchEngineID string // CX parameter
}

var googlePSEConfig *GooglePSEConfig

// SetGooglePSEConfig sets the Google PSE configuration
func SetGooglePSEConfig(apiKey, searchEngineID string) {
	googlePSEConfig = &GooglePSEConfig{
		APIKey:         apiKey,
		SearchEngineID: searchEngineID,
	}
}

// GetGooglePSEConfig returns the current configuration
func GetGooglePSEConfig() *GooglePSEConfig {
	return googlePSEConfig
}

// CallGooglePSE executes a Google PSE search
func CallGooglePSE(arguments map[string]interface{}) (string, error) {
	if googlePSEConfig == nil {
		return "", fmt.Errorf("Google PSE not configured. Please set API key and Search Engine ID")
	}

	query, ok := arguments["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("query argument is required and must be a non-empty string")
	}

	// Get optional parameters
	num := 10
	if n, ok := arguments["num"].(float64); ok {
		num = int(n)
		if num < 1 || num > 10 {
			num = 10
		}
	}

	start := 1
	if s, ok := arguments["start"].(float64); ok {
		start = int(s)
		if start < 1 {
			start = 1
		}
	}

	// Build Google Custom Search API URL
	baseURL := "https://www.googleapis.com/customsearch/v1"
	params := url.Values{}
	params.Set("key", googlePSEConfig.APIKey)
	params.Set("cx", googlePSEConfig.SearchEngineID)
	params.Set("q", query)
	params.Set("num", fmt.Sprintf("%d", num))
	params.Set("start", fmt.Sprintf("%d", start))

	searchURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Make HTTP request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Google PSE API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp GooglePSEResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Format results
	if len(apiResp.Items) == 0 {
		return "No results found for your search query.", nil
	}

	result := fmt.Sprintf("Found %s results:\n\n", apiResp.SearchInformation.TotalResults)
	for i, item := range apiResp.Items {
		result += fmt.Sprintf("%d. %s\n", i+1, item.Title)
		result += fmt.Sprintf("   URL: %s\n", item.Link)
		result += fmt.Sprintf("   %s\n\n", item.Snippet)
	}

	return result, nil
}
