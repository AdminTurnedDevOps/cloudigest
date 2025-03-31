package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"cloudigest/pkg/rag"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var queryCmd = &cobra.Command{
	Use:   "query [question]",
	Short: "Query the knowledge base with a natural language question",
	Long: `Query the knowledge base with a natural language question about cloud infrastructure,
Kubernetes, or any other cloud-related topics. The system will use retrieval-augmented
generation (RAG) to provide you with relevant, actionable information and best practices.

Example:
  cloudigest query "what is the best way to deploy Kubernetes?"
  cloudigest query "how can I optimize my AWS EC2 costs?"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get OpenAI API key from config
		openAIKey := viper.GetString("openai.api_key")
		if openAIKey == "" {
			return fmt.Errorf("OpenAI API key not found in configuration")
		}

		// Claude is optional - if missing, we'll use OpenAI exclusively
		claudeKey := viper.GetString("claude.api_key")
		useClaudeIfAvailable := claudeKey != "" && claudeKey != "your-claude-api-key-here"

		// Initialize RAG system based on available API keys
		var ragSystem *rag.RAG
		if useClaudeIfAvailable {
			fmt.Println("Using Claude for query processing")
			ragSystem = rag.NewRAGWithClaude(claudeKey)
		} else {
			fmt.Println("Using OpenAI for query processing")
			ragSystem = rag.NewRAG(openAIKey)
		}

		// Configure RAG settings from config
		if chunkSize := viper.GetInt("rag.chunk_size"); chunkSize > 0 {
			ragSystem.SetChunkSize(chunkSize)
		}

		if maxTokens := viper.GetInt("rag.max_tokens"); maxTokens > 0 {
			ragSystem.SetMaxTokens(maxTokens)
		}

		// Get user's question
		question := args[0]
		fmt.Printf("Processing query: %s\n", question)

		// Load knowledge base documents from config
		fmt.Println("Loading knowledge base...")
		err := loadKnowledgeBase(ragSystem)
		if err != nil {
			return fmt.Errorf("failed to load knowledge base: %v", err)
		}

		// Query the RAG system
		fmt.Println("Searching knowledge base and generating answer...")
		answer, err := ragSystem.Query(question)
		if err != nil {
			return fmt.Errorf("failed to process query: %v", err)
		}

		// Display results
		fmt.Println("\nAnswer:")
		fmt.Println("=======")
		fmt.Println(answer)
		return nil
	},
}

// loadKnowledgeBase adds documents to the RAG system
// It loads document sources from the configuration
func loadKnowledgeBase(r *rag.RAG) error {
	// Get sources from config
	var sources []struct {
		URL  string `mapstructure:"url"`
		Type string `mapstructure:"type"`
		Name string `mapstructure:"name"`
	}

	if err := viper.UnmarshalKey("rag.sources", &sources); err != nil {
		return fmt.Errorf("failed to read document sources from config: %v", err)
	}

	if len(sources) == 0 {
		fmt.Println("Warning: No document sources found in configuration. Using example documents.")
		return loadExampleDocuments(r)
	}

	// Load content from each source
	for _, source := range sources {
		fmt.Printf("Loading document from %s...\n", source.URL)

		// Fetch content from URL
		resp, err := http.Get(source.URL)
		if err != nil {
			fmt.Printf("Warning: Failed to fetch document from %s: %v\n", source.URL, err)
			continue
		}
		defer resp.Body.Close()

		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Warning: Failed to read document from %s: %v\n", source.URL, err)
			continue
		}

		// Add document to RAG system
		doc := rag.Document{
			Content: string(content),
			Source:  source.Name,
			Type:    source.Type,
		}

		if err := r.AddDocument(doc); err != nil {
			fmt.Printf("Warning: Failed to add document from %s: %v\n", source.URL, err)
			continue
		}
	}

	return nil
}

// loadExampleDocuments loads example documents when no sources are configured
func loadExampleDocuments(r *rag.RAG) error {
	documents := []rag.Document{
		{
			Content: "Kubernetes is a portable, extensible, open source platform for managing containerized workloads and services. " +
				"The best practices for deploying Kubernetes include: using a managed Kubernetes service like GKE, EKS, or AKS; " +
				"implementing proper resource requests and limits; setting up monitoring and alerting; using Helm for package management; " +
				"and implementing a robust CI/CD pipeline for deployments.",
			Source: "Kubernetes Best Practices",
			Type:   "documentation",
		},
		{
			Content: "When optimizing AWS EC2 costs, consider: using Reserved Instances for predictable workloads; " +
				"implementing auto-scaling groups; choosing the right instance types; regularly reviewing and terminating unused resources; " +
				"using Spot Instances for fault-tolerant workloads; and leveraging AWS Cost Explorer for detailed analysis.",
			Source: "AWS Cost Optimization Guide",
			Type:   "article",
		},
		{
			Content: "Container security best practices include: scanning images for vulnerabilities; using minimal base images; " +
				"implementing network policies; running containers with least privileges; using immutable infrastructure; " +
				"and implementing a zero-trust security model.",
			Source: "Container Security Guidelines",
			Type:   "documentation",
		},
	}

	for _, doc := range documents {
		if err := r.AddDocument(doc); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(queryCmd)
}
