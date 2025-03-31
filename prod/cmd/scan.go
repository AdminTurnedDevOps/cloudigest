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
	Short: "Scan infrastructure resources",
	Long: `Scan infrastructure resources such as Kubernetes clusters and virtual machines
to provide optimization recommendations.`,
}

var scanKubernetesCmd = &cobra.Command{
	Use:   "kubernetes",
	Short: "Scan a Kubernetes cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey := viper.GetString("openai.api_key")
		if apiKey == "" {
			return fmt.Errorf("OpenAI API key not found in configuration")
		}

		scanner := scanner.NewScanner(apiKey)

		// Try to get kubeconfig from flag or default location
		kubeconfig := cmd.Flag("kubeconfig").Value.String()
		if kubeconfig == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %v", err)
			}
			kubeconfig = filepath.Join(home, ".kube", "config")
		}

		analysis, err := scanner.ScanKubernetesCluster(kubeconfig)
		if err != nil {
			return fmt.Errorf("failed to scan cluster: %v", err)
		}

		fmt.Println("Kubernetes Cluster Analysis:")
		fmt.Println("===========================")
		fmt.Println(analysis)
		return nil
	},
}

var scanVMCmd = &cobra.Command{
	Use:   "vm",
	Short: "Scan a virtual machine",
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey := viper.GetString("openai.api_key")
		if apiKey == "" {
			return fmt.Errorf("OpenAI API key not found in configuration")
		}

		scanner := scanner.NewScanner(apiKey)

		// This is a placeholder for VM info gathering
		// In a real implementation, you would gather this information
		// from the cloud provider's API or system metrics
		vmInfo := map[string]interface{}{
			"cpu": map[string]interface{}{
				"cores":       4,
				"utilization": 0.75,
			},
			"memory": map[string]interface{}{
				"total": "16Gi",
				"used":  "12Gi",
			},
			"disk": map[string]interface{}{
				"total": "100Gi",
				"used":  "75Gi",
			},
		}

		analysis, err := scanner.ScanVirtualMachine(vmInfo)
		if err != nil {
			return fmt.Errorf("failed to scan VM: %v", err)
		}

		fmt.Println("Virtual Machine Analysis:")
		fmt.Println("========================")
		fmt.Println(analysis)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.AddCommand(scanKubernetesCmd)
	scanCmd.AddCommand(scanVMCmd)

	scanKubernetesCmd.Flags().String("kubeconfig", "", "path to kubeconfig file")
}
