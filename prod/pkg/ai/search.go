package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sashabaranov/go-openai"
)

var model string = openai.GPT4oMini

type SearchResult struct {
	Title       string
	URL         string
	Description string
	Content     string
}

type SearchEngine struct {
	client    *openai.Client
	serperKey string
	maxTokens int
}

type SerperRequest struct {
	Q string `json:"q"`
}

type SerperResponse struct {
	SearchParameters struct {
		Q string `json:"q"`
	} `json:"searchParameters"`
	Organic []struct {
		Title       string `json:"title"`
		Link        string `json:"link"`
		Snippet     string `json:"snippet"`
		Position    int    `json:"position"`
		DisplayLink string `json:"displayLink"`
	} `json:"organic"`
}

func AiModel() string {
	return model
}

func NewSearchEngine(openAIKey, serperKey string) *SearchEngine {
	return &SearchEngine{
		client:    openai.NewClient(openAIKey),
		serperKey: serperKey,
		maxTokens: 4000,
	}
}

func (s *SearchEngine) Search(query string) ([]SearchResult, error) {
	// Prepare the request to Serper
	reqBody := SerperRequest{Q: query}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://google.serper.dev/search", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add headers
	req.Header.Set("X-API-KEY", s.serperKey)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var serperResp SerperResponse
	if err := json.Unmarshal(body, &serperResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Convert Serper results to SearchResults
	var results []SearchResult
	for _, result := range serperResp.Organic {
		results = append(results, SearchResult{
			Title:       result.Title,
			URL:         result.Link,
			Description: result.Snippet,
			Content:     result.Snippet, // Initially use snippet as content
		})
	}

	// Enhance search results with AI analysis
	enhancedResults, err := s.enhanceResults(results, query)
	if err != nil {
		return nil, err
	}

	return enhancedResults, nil
}

func (s *SearchEngine) enhanceResults(results []SearchResult, query string) ([]SearchResult, error) {
	for i := range results {
		analysis, err := s.analyzeContent(results[i].Content, query)
		if err != nil {
			return nil, err
		}
		results[i].Content = analysis
	}
	return results, nil
}

func (s *SearchEngine) analyzeContent(content, query string) (string, error) {
	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleSystem,
					Content: "You are an infrastructure expert analyzing search results. " +
						"Extract and summarize relevant information about infrastructure best practices " +
						"and optimization recommendations. Focus on actionable insights.",
				},
				{
					Role: openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("Query: %s\n\nContent to analyze: %s\n\n"+
						"Please analyze this content and provide relevant insights related to the query.",
						query, content),
				},
			},
			MaxTokens: s.maxTokens,
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func (s *SearchEngine) GenerateRecommendations(results []SearchResult) (string, error) {
	var combinedContent strings.Builder
	for _, result := range results {
		combinedContent.WriteString(fmt.Sprintf("Source: %s\n%s\n\n", result.URL, result.Content))
	}

	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleSystem,
					Content: "You are an infrastructure optimization expert. Based on the search results, " +
						"provide clear, actionable recommendations for infrastructure improvements. " +
						"Focus on best practices and practical implementation advice.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: combinedContent.String(),
				},
			},
			MaxTokens: s.maxTokens,
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
