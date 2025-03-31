package image

import (
	aimodel "cloudigest/pkg/ai"
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

type Analyzer struct {
	client    *openai.Client
	maxTokens int
}

func NewAnalyzer(apiKey string) *Analyzer {
	return &Analyzer{
		client:    openai.NewClient(apiKey),
		maxTokens: 4000,
	}
}

func (a *Analyzer) AnalyzeImage(imagePath string) (string, error) {
	// Read and encode the image
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %v", err)
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)

	resp, err := a.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: aimodel.AiModel(),
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

	resp, err := a.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: aimodel.AiModel(),
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
}

func (a *Analyzer) GenerateRecommendations(imageAnalysis, docAnalysis string) (string, error) {
	resp, err := a.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: aimodel.AiModel(),
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
}
