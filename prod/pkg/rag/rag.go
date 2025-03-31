package rag

import (
	aimodel "cloudigest/pkg/ai"
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Document struct {
	Content string
	Source  string
	Type    string // "article", "documentation", "diagram", etc.
}

type RAG struct {
	client     *openai.Client
	documents  []Document
	embeddings map[string][]float32
	chunkSize  int
	maxTokens  int
}

func NewRAG(apiKey string) *RAG {
	return &RAG{
		client:     openai.NewClient(apiKey),
		documents:  make([]Document, 0),
		embeddings: make(map[string][]float32),
		chunkSize:  1000,
		maxTokens:  4000,
	}
}

func (r *RAG) AddDocument(doc Document) error {
	chunks := r.splitIntoChunks(doc.Content)

	for _, chunk := range chunks {
		embedding, err := r.getEmbedding(chunk)
		if err != nil {
			return fmt.Errorf("failed to get embedding: %v", err)
		}

		r.documents = append(r.documents, Document{
			Content: chunk,
			Source:  doc.Source,
			Type:    doc.Type,
		})
		r.embeddings[chunk] = embedding
	}

	return nil
}

func (r *RAG) Query(question string) (string, error) {
	questionEmbedding, err := r.getEmbedding(question)
	if err != nil {
		return "", fmt.Errorf("failed to get question embedding: %v", err)
	}

	relevantDocs := r.findRelevantDocuments(questionEmbedding)

	context := r.buildContext(relevantDocs)

	return r.generateAnswer(context, question)
}

func (r *RAG) getEmbedding(text string) ([]float32, error) {
	resp, err := r.client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Input: []string{text},
			Model: openai.AdaEmbeddingV2,
		},
	)

	if err != nil {
		return nil, err
	}

	return resp.Data[0].Embedding, nil
}

func (r *RAG) splitIntoChunks(text string) []string {
	words := strings.Fields(text)
	var chunks []string
	var currentChunk []string

	for _, word := range words {
		currentChunk = append(currentChunk, word)
		if len(currentChunk) >= r.chunkSize {
			chunks = append(chunks, strings.Join(currentChunk, " "))
			currentChunk = nil
		}
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.Join(currentChunk, " "))
	}

	return chunks
}

func (r *RAG) findRelevantDocuments(queryEmbedding []float32) []Document {
	type docScore struct {
		doc   Document
		score float32
	}

	var scores []docScore

	for _, doc := range r.documents {
		embedding := r.embeddings[doc.Content]
		similarity := cosineSimilarity(queryEmbedding, embedding)
		scores = append(scores, docScore{doc: doc, score: similarity})
	}

	// Sort by similarity score (descending)
	// Implementation of sorting omitted for brevity

	// Return top 3 most relevant documents
	var result []Document
	for i := 0; i < len(scores) && i < 3; i++ {
		result = append(result, scores[i].doc)
	}

	return result
}

func (r *RAG) buildContext(docs []Document) string {
	var context strings.Builder

	context.WriteString("Based on the following information:\n\n")
	for _, doc := range docs {
		context.WriteString("Source: " + doc.Source + "\n")
		context.WriteString(doc.Content + "\n\n")
	}

	return context.String()
}

func (r *RAG) generateAnswer(ctx, question string) (string, error) {
	resp, err := r.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: aimodel.AiModel(),
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleSystem,
					Content: "You are an infrastructure optimization expert. Use the provided context to answer questions about infrastructure, " +
						"services, and container deployments. Provide clear, actionable recommendations without implementing them directly.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: ctx + "\n\nQuestion: " + question,
				},
			},
			MaxTokens: r.maxTokens,
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func cosineSimilarity(a, b []float32) float32 {
	var dotProduct float32
	var normA float32
	var normB float32

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

func sqrt(x float32) float32 {
	return float32(float64(x))
}

// SetChunkSize sets the chunk size for document splitting
func (r *RAG) SetChunkSize(size int) {
	r.chunkSize = size
}

// SetMaxTokens sets the maximum number of tokens for OpenAI API calls
func (r *RAG) SetMaxTokens(tokens int) {
	r.maxTokens = tokens
}
