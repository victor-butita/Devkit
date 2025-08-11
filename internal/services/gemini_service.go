package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type GeminiService struct { APIKey string; HTTPClient *http.Client }
func NewGeminiService(apiKey string) *GeminiService { return &GeminiService{APIKey: apiKey, HTTPClient: &http.Client{}} }

// --- API Structs ---
type GeminiPart struct { Text string `json:"text"` }
type GeminiContent struct { Parts []GeminiPart `json:"parts"` }
type GeminiCandidate struct { Content GeminiContent `json:"content"`; FinishReason string `json:"finishReason"` }
type GenerationConfig struct { Temperature float32 `json:"temperature"`; MaxOutputTokens int `json:"maxOutputTokens"` }
type SafetySetting struct { Category string `json:"category"`; Threshold string `json:"threshold"` }
type GeminiRequest struct {
	Contents         []GeminiContent   `json:"contents"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings   []SafetySetting   `json:"safetySettings,omitempty"`
}
type GeminiResponse struct { Candidates []GeminiCandidate `json:"candidates"` }

// GenerateRegex uses AI to create a regex pattern and explanation.
func (s *GeminiService) GenerateRegex(description string) (string, error) {
	prompt := fmt.Sprintf(
		"You are a regular expression expert. Based on the following description, generate a valid PCRE-compatible regex pattern. "+
			"Then, on a new line, provide a brief, step-by-step explanation of how the regex works. "+
			"Use '|||' as a separator between the regex and the explanation. "+
			"Respond with ONLY the regex, the separator, and the explanation. Example: `^[a-zA-Z0-9]+$|||Asserts position at the start of the string. Matches one or more alphanumeric characters. Asserts position at the end.` "+
			"Description: \"%s\"",
		description,
	)
	return s.generateContent(prompt)
}

// GenerateSQL uses AI to create an SQL query.
func (s *GeminiService) GenerateSQL(schema, description string) (string, error) {
	prompt := fmt.Sprintf(
		"You are an expert SQL developer. Based on the provided database schema and the user's request, generate a single, valid SQL query. "+
			"The query should be formatted for PostgreSQL. "+
			"Do not add any explanation or markdown formatting like ```sql. Respond with ONLY the raw SQL query. "+
			"Schema:\n%s\n\nUser Request: \"%s\"",
		schema, description,
	)
	return s.generateContent(prompt)
}

// generateContent is the shared helper function for all AI calls.
func (s *GeminiService) generateContent(prompt string) (string, error) {
	config := &GenerationConfig{Temperature: 0.2, MaxOutputTokens: 2048}
	safetySettings := []SafetySetting{
		{Category: "HARM_CATEGORY_HARASSMENT", Threshold: "BLOCK_NONE"},
		{Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "BLOCK_NONE"},
		{Category: "HARM_CATEGORY_SEXUALLY_EXPLICIT", Threshold: "BLOCK_NONE"},
		{Category: "HARM_CATEGORY_DANGEROUS_CONTENT", Threshold: "BLOCK_NONE"},
	}
	reqBody := GeminiRequest{ Contents: []GeminiContent{{Parts: []GeminiPart{{Text: prompt}}}}, GenerationConfig: config, SafetySettings: safetySettings }
	jsonData, err := json.Marshal(reqBody)
	if err != nil { return "", fmt.Errorf("error creating request body: %w", err) }
	apiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key=" + s.APIKey
	var resp *http.Response
	maxRetries := 3
	backoffDuration := 1 * time.Second
	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil { return "", fmt.Errorf("error creating http request: %w", err) }
		req.Header.Set("Content-Type", "application/json")
		resp, err = s.HTTPClient.Do(req)
		if err != nil || resp.StatusCode == http.StatusServiceUnavailable {
			if err == nil { resp.Body.Close() }
			log.Printf("Attempt %d: API call failed. Retrying in %v...", i+1, backoffDuration)
			time.Sleep(backoffDuration)
			backoffDuration *= 2
			continue
		}
		break
	}
	if resp == nil { return "", fmt.Errorf("API did not respond after %d retries", maxRetries) }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil { return "", fmt.Errorf("error reading successful response body: %w", err) }
	var geminiResp GeminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return "", fmt.Errorf("error parsing successful Gemini response: %w", err)
	}
	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}
	return "", fmt.Errorf("no content found in Gemini response")
}