package cmd

import (
	"fmt"
	"path/filepath"

	"cloudigest/pkg/image"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [file]",
	Short: "Analyze an infrastructure diagram or documentation",
	Long: `Analyze infrastructure diagrams or documentation to provide optimization
recommendations and best practices. The tool can analyze images (diagrams) or
text documentation.`,
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

		// Get file path
		filePath := args[0]
		fileExt := filepath.Ext(filePath)

		// Initialize analyzer based on available API keys
		var analyzer *image.Analyzer
		if useClaudeIfAvailable {
			fmt.Println("Using Claude for analysis")
			analyzer = image.NewAnalyzerWithClaude(claudeKey)
		} else {
			fmt.Println("Using OpenAI for analysis")
			analyzer = image.NewAnalyzer(openAIKey)
		}

		var analysis string
		var err error

		// Analyze based on file type
		switch fileExt {
		case ".jpg", ".jpeg", ".png":
			fmt.Println("Analyzing infrastructure diagram...")
			if useClaudeIfAvailable {
				fmt.Println("Note: Image analysis is only supported with OpenAI. Switching to OpenAI for this operation.")
				tempAnalyzer := image.NewAnalyzer(openAIKey)
				analysis, err = tempAnalyzer.AnalyzeImage(filePath)
			} else {
				analysis, err = analyzer.AnalyzeImage(filePath)
			}
		case ".txt", ".md", ".yaml", ".yml", ".json", ".tf", ".hcl":
			fmt.Println("Analyzing documentation...")
			analysis, err = analyzer.AnalyzeDocument(filePath)
		default:
			return fmt.Errorf("unsupported file type: %s", fileExt)
		}

		if err != nil {
			return fmt.Errorf("failed to analyze file: %v", err)
		}

		fmt.Println("\nAnalysis Results:")
		fmt.Println("================")
		fmt.Println(analysis)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
