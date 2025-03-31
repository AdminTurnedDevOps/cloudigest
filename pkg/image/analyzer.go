package image

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	anthropic "github.com/liushuangls/go-anthropic/v2"
	"github.com/sashabaranov/go-openai"
)

type Analyzer struct {
	openaiClient *openai.Client
	claudeClient *anthropic.Client
	maxTokens    int
	useOpenAI    bool
}

func NewAnalyzer(apiKey string) *Analyzer {
	return &Analyzer{
		openaiClient: openai.NewClient(apiKey),
		claudeClient: nil,
		maxTokens:    4000,
		useOpenAI:    true,
	}
}

func NewAnalyzerWithClaude(claudeKey string) *Analyzer {
	return &Analyzer{
		openaiClient: nil,
		claudeClient: anthropic.NewClient(claudeKey),
		maxTokens:    4000,
		useOpenAI:    false,
	}
}

func (a *Analyzer) AnalyzeImage(imagePath string) (string, error) {
	// Note: Claude doesn't support image analysis through the Go SDK yet
	if !a.useOpenAI {
		return "", fmt.Errorf("image analysis is only supported with OpenAI")
	}

	// Read and encode the image
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %v", err)
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)

	resp, err := a.openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleSystem,
					Content: "You are an infrastructure expert analyzing architecture diagrams. " +
						"Provide detailed insights about the infrastructure design, potential optimizations, " +
						"and best practices recommendations.",
				},
				{
					Role: openai.ChatMessageRoleUser,
					Content: "Please analyze this architecture diagram and provide insights about:\n" +
						"1. The overall architecture design\n" +
						"2. Potential bottlenecks or scalability concerns\n" +
						"3. Security considerations\n" +
						"4. Cost optimization opportunities\n" +
						"5. Recommendations for improvement",
					MultiContent: []openai.ChatMessagePart{
						{
							Type: openai.ChatMessagePartTypeText,
							Text: "Please analyze this architecture diagram",
						},
						{
							Type:     openai.ChatMessagePartTypeImageURL,
							ImageURL: &openai.ChatMessageImageURL{URL: "data:image/png;base64," + base64Image},
						},
					},
				},
			},
			MaxTokens: a.maxTokens,
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to analyze image: %v", err)
	}

	return resp.Choices[0].Message.Content, nil
}

func (a *Analyzer) AnalyzeDocument(docPath string) (string, error) {
	// Read the document
	content, err := os.ReadFile(docPath)
	if err != nil {
		return "", fmt.Errorf("failed to read document: %v", err)
	}

	if a.useOpenAI {
		// Use OpenAI
		resp, err := a.openaiClient.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: openai.GPT4o,
				Messages: []openai.ChatCompletionMessage{
					{
						Role: openai.ChatMessageRoleSystem,
						Content: "You are an infrastructure expert analyzing technical documentation. " +
							"Extract key information about infrastructure requirements, design decisions, " +
							"and provide optimization recommendations.",
					},
					{
						Role: openai.ChatMessageRoleUser,
						Content: fmt.Sprintf("Please analyze this technical document and provide insights about:\n"+
							"1. Infrastructure requirements and dependencies\n"+
							"2. Scalability and performance considerations\n"+
							"3. Security requirements\n"+
							"4. Operational considerations\n"+
							"5. Recommendations for optimal deployment\n\n"+
							"Document content:\n%s", string(content)),
					},
				},
				MaxTokens: a.maxTokens,
			},
		)

		if err != nil {
			return "", fmt.Errorf("failed to analyze document: %v", err)
		}

		return resp.Choices[0].Message.Content, nil
	} else {
		// Use Claude
		promptText := "You are an infrastructure expert analyzing technical documentation. " +
			"Extract key information about infrastructure requirements, design decisions, " +
			"and provide optimization recommendations.\n\n" +
			fmt.Sprintf("Please analyze this technical document and provide insights about:\n"+
				"1. Infrastructure requirements and dependencies\n"+
				"2. Scalability and performance considerations\n"+
				"3. Security requirements\n"+
				"4. Operational considerations\n"+
				"5. Recommendations for optimal deployment\n\n"+
				"Document content:\n%s", string(content))

		resp, err := a.claudeClient.CreateMessages(
			context.Background(),
			anthropic.MessagesRequest{
				Model: anthropic.ModelClaude3Dot7SonnetLatest,
				Messages: []anthropic.Message{
					anthropic.NewUserTextMessage(promptText),
				},
				MaxTokens: a.maxTokens,
			},
		)

		if err != nil {
			return "", fmt.Errorf("failed to analyze document: %v", err)
		}

		return resp.GetFirstContentText(), nil
	}
}

func (a *Analyzer) GenerateRecommendations(imageAnalysis, docAnalysis string) (string, error) {
	if a.useOpenAI {
		// Use OpenAI
		resp, err := a.openaiClient.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: openai.GPT4o,
				Messages: []openai.ChatCompletionMessage{
					{
						Role: openai.ChatMessageRoleSystem,
						Content: "You are an infrastructure optimization expert. Based on the analysis of architecture diagrams " +
							"and technical documentation, provide comprehensive recommendations for infrastructure improvements.",
					},
					{
						Role: openai.ChatMessageRoleUser,
						Content: fmt.Sprintf("Based on the following analyses, provide detailed recommendations:\n\n"+
							"Architecture Analysis:\n%s\n\n"+
							"Documentation Analysis:\n%s\n\n"+
							"Please provide specific, actionable recommendations for:\n"+
							"1. Infrastructure optimization\n"+
							"2. Scalability improvements\n"+
							"3. Security enhancements\n"+
							"4. Cost optimization\n"+
							"5. Operational efficiency",
							imageAnalysis, docAnalysis),
					},
				},
				MaxTokens: a.maxTokens,
			},
		)

		if err != nil {
			return "", fmt.Errorf("failed to generate recommendations: %v", err)
		}

		return resp.Choices[0].Message.Content, nil
	} else {
		// Use Claude
		promptText := "You are an infrastructure optimization expert. Based on the analysis of architecture diagrams " +
			"and technical documentation, provide comprehensive recommendations for infrastructure improvements.\n\n" +
			fmt.Sprintf("Based on the following analyses, provide detailed recommendations:\n\n"+
				"Architecture Analysis:\n%s\n\n"+
				"Documentation Analysis:\n%s\n\n"+
				"Please provide specific, actionable recommendations for:\n"+
				"1. Infrastructure optimization\n"+
				"2. Scalability improvements\n"+
				"3. Security enhancements\n"+
				"4. Cost optimization\n"+
				"5. Operational efficiency",
				imageAnalysis, docAnalysis)

		resp, err := a.claudeClient.CreateMessages(
			context.Background(),
			anthropic.MessagesRequest{
				Model: anthropic.ModelClaude3Dot7SonnetLatest,
				Messages: []anthropic.Message{
					anthropic.NewUserTextMessage(promptText),
				},
				MaxTokens: a.maxTokens,
			},
		)

		if err != nil {
			return "", fmt.Errorf("failed to generate recommendations: %v", err)
		}

		return resp.GetFirstContentText(), nil
	}
}
