package cmd

import (
	"fmt"
	"io"
	"net/http"

	"cloudigest/pkg/ai"
	"cloudigest/pkg/rag"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type RAGSource struct {
	URL  string `mapstructure:"url"`
	Type string `mapstructure:"type"`
	Name string `mapstructure:"name"`
}

type RAGConfig struct {
	ChunkSize int         `mapstructure:"chunk_size"`
	MaxTokens int         `mapstructure:"max_tokens"`
	Sources   []RAGSource `mapstructure:"sources"`
}

var queryCmd = &cobra.Command{
	Use:   "query [question]",
	Short: "Query the knowledge base",
	Long: `Query the knowledge base for infrastructure optimization recommendations.
The system uses RAG (Retrieval Augmented Generation) to provide accurate and
contextual answers based on best practices and documentation.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		openAIKey := viper.GetString("openai.api_key")
		if openAIKey == "" {
			return fmt.Errorf("OpenAI API key not found in configuration")
		}

		serperKey := viper.GetString("serper.api_key")
		if serperKey == "" {
			return fmt.Errorf("Serper API key not found in configuration")
		}

		// Load RAG configuration
		var ragConfig RAGConfig
		if err := viper.UnmarshalKey("rag", &ragConfig); err != nil {
			return fmt.Errorf("failed to load RAG configuration: %v", err)
		}

		// Initialize search engine
		searchEngine := ai.NewSearchEngine(openAIKey, serperKey)

		// Perform web search
		searchResults, err := searchEngine.Search(args[0])
		if err != nil {
			return fmt.Errorf("failed to perform web search: %v", err)
		}

		// Initialize RAG system
		ragSystem := rag.NewRAG(openAIKey)

		// Apply configuration settings
		if ragConfig.ChunkSize > 0 {
			ragSystem.SetChunkSize(ragConfig.ChunkSize)
		}
		if ragConfig.MaxTokens > 0 {
			ragSystem.SetMaxTokens(ragConfig.MaxTokens)
		}

		// Add initial knowledge base document
		content, err := fetchContent("https://controlplane.com/community-blog/post/optimize-kubernetes-workloads")
		if err != nil {
			return fmt.Errorf("failed to fetch initial content: %v", err)
		}

		err = ragSystem.AddDocument(rag.Document{
			Content: content,
			Source:  "Kubernetes Workload Optimization Guide",
			Type:    "article",
		})
		if err != nil {
			return fmt.Errorf("failed to add document: %v", err)
		}

		// Process the query with both RAG and search results
		ragAnswer, err := ragSystem.Query(args[0])
		if err != nil {
			return fmt.Errorf("failed to process RAG query: %v", err)
		}

		searchRecommendations, err := searchEngine.GenerateRecommendations(searchResults)
		if err != nil {
			return fmt.Errorf("failed to generate search recommendations: %v", err)
		}

		fmt.Println("Knowledge Base Answer:")
		fmt.Println("=====================")
		fmt.Println(ragAnswer)
		fmt.Println("\nWeb Search Recommendations:")
		fmt.Println("=========================")
		fmt.Println(searchRecommendations)
		return nil
	},
}

func fetchContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func init() {
	rootCmd.AddCommand(queryCmd)
}
