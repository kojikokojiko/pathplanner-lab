package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pathplanner-lab/backend/internal/domain"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"
const defaultModel = "claude-opus-4-6"

type ReviewResult struct {
	Summary         string   `json:"summary"`
	Recommendations []string `json:"recommendations"`
}

type MapSummary struct {
	Width         int     `json:"width"`
	Height        int     `json:"height"`
	ObstacleRatio float64 `json:"obstacle_ratio"`
}

type Client struct {
	apiKey     string
	httpClient *http.Client
	model      string
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
		model:      defaultModel,
	}
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) call(ctx context.Context, prompt string) (*ReviewResult, error) {
	reqBody, err := json.Marshal(anthropicRequest{
		Model:     c.model,
		MaxTokens: 1024,
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http call: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var apiResp anthropicResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	if apiResp.Error != nil {
		return nil, fmt.Errorf("anthropic error: %s", apiResp.Error.Message)
	}
	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	text := apiResp.Content[0].Text
	var result ReviewResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		// Fallback: treat text as summary
		result = ReviewResult{
			Summary:         text,
			Recommendations: []string{},
		}
	}
	return &result, nil
}

func (c *Client) ReviewRun(ctx context.Context, run *domain.Run, metrics []domain.RunMetric, mapSummary MapSummary, mode string) (*ReviewResult, error) {
	payload := buildRunPayload(run, metrics, mapSummary)
	prompt := buildRunPrompt(payload, mode)
	return c.call(ctx, prompt)
}

func (c *Client) ReviewCompare(ctx context.Context, rows interface{}, mapSummary MapSummary, mode string) (*ReviewResult, error) {
	payload := map[string]interface{}{
		"map":     mapSummary,
		"compare": rows,
	}
	b, _ := json.MarshalIndent(payload, "", "  ")
	prompt := buildComparePrompt(string(b), mode)
	return c.call(ctx, prompt)
}

func buildRunPayload(run *domain.Run, metrics []domain.RunMetric, ms MapSummary) map[string]interface{} {
	mMap := make(map[string]float64)
	for _, m := range metrics {
		mMap[m.Key] = m.Value
	}
	var params interface{}
	_ = json.Unmarshal(run.Params, &params)
	return map[string]interface{}{
		"map":       ms,
		"algorithm": run.AlgorithmKey,
		"params":    params,
		"metrics":   mMap,
	}
}

func buildRunPrompt(payload map[string]interface{}, mode string) string {
	b, _ := json.MarshalIndent(payload, "", "  ")
	style := "as a teacher explaining to a student learning path planning algorithms"
	if mode == "concise" {
		style = "concisely in bullet points"
	}
	return fmt.Sprintf(`You are a path planning expert. Analyze this algorithm run result and explain it %s.

Input data:
%s

Respond with a JSON object with exactly these fields:
{
  "summary": "2-4 sentence explanation of the results",
  "recommendations": ["recommendation 1", "recommendation 2", "recommendation 3"]
}

Only output the JSON, no other text.`, style, string(b))
}

func buildComparePrompt(data string, mode string) string {
	style := "as a teacher explaining to a student"
	if mode == "concise" {
		style = "concisely"
	}
	return fmt.Sprintf(`You are a path planning expert. Compare these algorithm results and explain them %s.

Input data:
%s

Respond with a JSON object with exactly these fields:
{
  "summary": "2-4 sentence comparison summary",
  "recommendations": ["recommendation 1", "recommendation 2", "recommendation 3"]
}

Only output the JSON, no other text.`, style, data)
}
