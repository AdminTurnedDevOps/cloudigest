package cmd

import (
	"fmt"
	"path/filepath"

	"cloudigest/pkg/image"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze architecture diagrams and documentation",
	Long: `Analyze architecture diagrams and documentation to provide infrastructure
optimization recommendations. This command can process both images and documents.`,
}

var analyzeDiagramCmd = &cobra.Command{
	Use:   "diagram [file]",
	Short: "Analyze an architecture diagram",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey := viper.GetString("openai.api_key")
		if apiKey == "" {
			return fmt.Errorf("OpenAI API key not found in configuration")
		}

		analyzer := image.NewAnalyzer(apiKey)

		filePath, err := filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %v", err)
		}

		analysis, err := analyzer.AnalyzeImage(filePath)
		if err != nil {
			return fmt.Errorf("failed to analyze diagram: %v", err)
		}

		fmt.Println("Architecture Diagram Analysis:")
		fmt.Println("==============================")
		fmt.Println(analysis)
		return nil
	},
}

var analyzeDocCmd = &cobra.Command{
	Use:   "doc [file]",
	Short: "Analyze a documentation file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey := viper.GetString("openai.api_key")
		if apiKey == "" {
			return fmt.Errorf("OpenAI API key not found in configuration")
		}

		analyzer := image.NewAnalyzer(apiKey)

		filePath, err := filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %v", err)
		}

		analysis, err := analyzer.AnalyzeDocument(filePath)
		if err != nil {
			return fmt.Errorf("failed to analyze document: %v", err)
		}

		fmt.Println("Documentation Analysis:")
		fmt.Println("======================")
		fmt.Println(analysis)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.AddCommand(analyzeDiagramCmd)
	analyzeCmd.AddCommand(analyzeDocCmd)
}
