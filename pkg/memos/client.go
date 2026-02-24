package memos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Client is a lightweight HTTP client for Memos API
type Client struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

// NewClient creates a new Memos API client
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		Client:  &http.Client{},
	}
}

// Memo represents a memo structure from the API
type Memo struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Content    string   `json:"content"`
	CreatorID  int      `json:"creatorId"`
	CreatedTs  int64    `json:"createdTs"`
	UpdatedTs  int64    `json:"updatedTs"`
	CreateTime string   `json:"createTime"`
	UpdateTime string   `json:"updateTime"`
	Visibility string   `json:"visibility"`
	Tags       []string `json:"tags"`
}

// CreateMemoRequest represents the payload for creating a memo
type CreateMemoRequest struct {
	Content    string  `json:"content"`
	Visibility *string `json:"visibility,omitempty"`
}

func (c *Client) CreateMemo(content string, visibility *string) (*Memo, error) {
	url := fmt.Sprintf("%s/api/v1/memos", c.BaseURL)
	payload := CreateMemoRequest{Content: content, Visibility: visibility}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.addHeaders(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(body))
	}

	var memo Memo
	if err := json.NewDecoder(resp.Body).Decode(&memo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &memo, nil
}

type UpdateMemoRequest struct {
	Content    *string `json:"content,omitempty"`
	Visibility *string `json:"visibility,omitempty"`
}

func (c *Client) GetMemo(id string) (*Memo, error) {
	normalizedID := normalizeMemoID(id)
	url := fmt.Sprintf("%s/api/v1/memos/%s", c.BaseURL, normalizedID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.addHeaders(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == 404 || resp.StatusCode == 405 {
			if memo := c.findMemoFromList(id); memo != nil {
				return memo, nil
			}
		}
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(body))
	}

	var memo Memo
	if err := json.NewDecoder(resp.Body).Decode(&memo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &memo, nil
}

func (c *Client) UpdateMemo(id string, content *string, visibility *string) (*Memo, error) {
	normalizedID := normalizeMemoID(id)
	url := fmt.Sprintf("%s/api/v1/memos/%s", c.BaseURL, normalizedID)
	payload := UpdateMemoRequest{Content: content, Visibility: visibility}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.addHeaders(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(body))
	}

	var memo Memo
	if err := json.NewDecoder(resp.Body).Decode(&memo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &memo, nil
}

func (c *Client) DeleteMemo(id string) error {
	normalizedID := normalizeMemoID(id)
	url := fmt.Sprintf("%s/api/v1/memos/%s", c.BaseURL, normalizedID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.addHeaders(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) ListMemos() ([]Memo, error) {
	url := fmt.Sprintf("%s/api/v1/memos", c.BaseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.addHeaders(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	type ListResponse struct {
		Memos []Memo `json:"memos"`
	}

	var listResp ListResponse
	if err := json.Unmarshal(data, &listResp); err == nil && listResp.Memos != nil {
		return listResp.Memos, nil
	}

	var memos []Memo
	if err := json.Unmarshal(data, &memos); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return memos, nil
}

func (c *Client) SearchMemosSmart(query string, tags []string, dateFrom *time.Time, dateTo *time.Time) ([]Memo, error) {
	memos, err := c.ListMemos()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(strings.TrimSpace(query))
	lowerTags := make([]string, 0, len(tags))
	for _, t := range tags {
		trimmed := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(t, "#")))
		if trimmed != "" {
			lowerTags = append(lowerTags, trimmed)
		}
	}

	var results []Memo
	for _, memo := range memos {
		contentLower := strings.ToLower(memo.Content)
		if query != "" && !strings.Contains(contentLower, query) {
			continue
		}

		if len(lowerTags) > 0 {
			memoTags := mergeTags(memo.Tags, extractTags(contentLower))
			if !containsAllTags(memoTags, lowerTags) {
				continue
			}
		}

		if dateFrom != nil || dateTo != nil {
			created := memoCreatedTime(memo)
			if dateFrom != nil && created.Before(*dateFrom) {
				continue
			}
			if dateTo != nil && created.After(*dateTo) {
				continue
			}
		}

		results = append(results, memo)
	}

	sort.Slice(results, func(i, j int) bool {
		return memoCreatedTime(results[i]).After(memoCreatedTime(results[j]))
	})

	return results, nil
}

func (c *Client) findMemoFromList(id string) *Memo {
	memos, err := c.ListMemos()
	if err != nil {
		return nil
	}
	normalizedID := normalizeMemoID(id)
	for _, memo := range memos {
		if memo.Name == id || memo.Name == "memos/"+normalizedID || normalizeMemoID(memo.Name) == normalizedID {
			copyMemo := memo
			return &copyMemo
		}
	}
	return nil
}

func extractTags(content string) []string {
	re := regexp.MustCompile(`#([a-zA-Z0-9_-]+)`)
	matches := re.FindAllStringSubmatch(content, -1)
	tags := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			tags = append(tags, strings.ToLower(match[1]))
		}
	}
	return tags
}

func containsAllTags(memoTags []string, required []string) bool {
	if len(required) == 0 {
		return true
	}
	lookup := map[string]struct{}{}
	for _, t := range memoTags {
		lookup[t] = struct{}{}
	}
	for _, r := range required {
		if _, ok := lookup[r]; !ok {
			return false
		}
	}
	return true
}

func mergeTags(fromAPI []string, fromContent []string) []string {
	if len(fromAPI) == 0 && len(fromContent) == 0 {
		return nil
	}
	lookup := map[string]struct{}{}
	for _, t := range fromAPI {
		trimmed := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(t, "#")))
		if trimmed != "" {
			lookup[trimmed] = struct{}{}
		}
	}
	for _, t := range fromContent {
		trimmed := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(t, "#")))
		if trimmed != "" {
			lookup[trimmed] = struct{}{}
		}
	}
	merged := make([]string, 0, len(lookup))
	for t := range lookup {
		merged = append(merged, t)
	}
	sort.Strings(merged)
	return merged
}

func normalizeMemoID(id string) string {
	trimmed := strings.TrimSpace(id)
	if strings.HasPrefix(trimmed, "memos/") {
		return strings.TrimPrefix(trimmed, "memos/")
	}
	return trimmed
}

func memoCreatedTime(memo Memo) time.Time {
	if memo.CreatedTs > 0 {
		return time.Unix(memo.CreatedTs, 0)
	}
	if memo.CreateTime != "" {
		if parsed, err := time.Parse(time.RFC3339, memo.CreateTime); err == nil {
			return parsed
		}
	}
	return time.Unix(0, 0)
}

func memoUpdatedTime(memo Memo) time.Time {
	if memo.UpdatedTs > 0 {
		return time.Unix(memo.UpdatedTs, 0)
	}
	if memo.UpdateTime != "" {
		if parsed, err := time.Parse(time.RFC3339, memo.UpdateTime); err == nil {
			return parsed
		}
	}
	return time.Unix(0, 0)
}

func MemoCreatedTs(memo Memo) int64 {
	return memoCreatedTime(memo).Unix()
}

func MemoUpdatedTs(memo Memo) int64 {
	return memoUpdatedTime(memo).Unix()
}

func (c *Client) addHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
}
