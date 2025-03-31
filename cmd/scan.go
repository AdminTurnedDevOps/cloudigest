package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"cloudigest/pkg/scanner"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan infrastructure and provide optimization recommendations",
	Long: `Scan your infrastructure (Kubernetes clusters, VMs, etc.) and provide detailed
optimization recommendations for resources, performance, cost, and security.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get OpenAI API key from config
		openAIKey := viper.GetString("openai.api_key")
		if openAIKey == "" {
			return fmt.Errorf("OpenAI API key not found in configuration")
		}

		// Claude is optional - if missing, we'll use OpenAI exclusively
		claudeKey := viper.GetString("claude.api_key")
		useClaudeIfAvailable := claudeKey != "" && claudeKey != "your-claude-api-key-here"

		// Initialize scanner based on available API keys
		var infraScanner *scanner.Scanner
		if useClaudeIfAvailable {
			fmt.Println("Using Claude for infrastructure scanning")
			infraScanner = scanner.NewScannerWithClaude(claudeKey)
		} else {
			fmt.Println("Using OpenAI for infrastructure scanning")
			infraScanner = scanner.NewScanner(openAIKey)
		}

		// Scan Kubernetes cluster if enabled
		if viper.GetBool("scanning.kubernetes") {
			kubeconfigPath := os.Getenv("KUBECONFIG")
			if kubeconfigPath == "" {
				kubeconfigPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")
			}

			fmt.Println("Scanning Kubernetes cluster...")
			results, err := infraScanner.ScanKubernetesCluster(kubeconfigPath)
			if err != nil {
				fmt.Printf("Warning: failed to scan Kubernetes cluster: %v\n", err)
			} else {
				fmt.Println("\nKubernetes Scan Results:")
				fmt.Println("========================")
				fmt.Println(results)
			}
		}

		// Additional scanning could be added here for cloud providers

		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
