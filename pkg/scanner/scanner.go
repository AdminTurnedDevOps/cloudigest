package scanner

import (
	"context"
	"encoding/json"
	"fmt"

	anthropic "github.com/liushuangls/go-anthropic/v2"
	"github.com/sashabaranov/go-openai"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type ResourceInfo struct {
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata"`
	Specs    map[string]interface{} `json:"specs"`
	Status   map[string]interface{} `json:"status"`
}

type Scanner struct {
	openaiClient *openai.Client
	claudeClient *anthropic.Client
	maxTokens    int
	useOpenAI    bool
}

func NewScanner(apiKey string) *Scanner {
	return &Scanner{
		openaiClient: openai.NewClient(apiKey),
		claudeClient: nil,
		maxTokens:    4000,
		useOpenAI:    true,
	}
}

func NewScannerWithClaude(claudeKey string) *Scanner {
	return &Scanner{
		openaiClient: nil,
		claudeClient: anthropic.NewClient(claudeKey),
		maxTokens:    4000,
		useOpenAI:    false,
	}
}

func (s *Scanner) ScanKubernetesCluster(kubeconfig string) (string, error) {
	// Load kubernetes configuration
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %v", err)
	}

	// Create kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	// Collect cluster information
	var resources []ResourceInfo

	// Get nodes
	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list nodes: %v", err)
	}

	for _, node := range nodes.Items {
		resources = append(resources, ResourceInfo{
			Type: "Node",
			Name: node.Name,
			Metadata: map[string]interface{}{
				"labels":      node.Labels,
				"annotations": node.Annotations,
			},
			Specs: map[string]interface{}{
				"capacity":    node.Status.Capacity,
				"allocatable": node.Status.Allocatable,
			},
			Status: map[string]interface{}{
				"conditions": node.Status.Conditions,
			},
		})
	}

	// Get pods across all namespaces
	pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list pods: %v", err)
	}

	for _, pod := range pods.Items {
		resources = append(resources, ResourceInfo{
			Type: "Pod",
			Name: pod.Name,
			Metadata: map[string]interface{}{
				"namespace":   pod.Namespace,
				"labels":      pod.Labels,
				"annotations": pod.Annotations,
			},
			Specs: map[string]interface{}{
				"containers":   pod.Spec.Containers,
				"nodeSelector": pod.Spec.NodeSelector,
			},
			Status: map[string]interface{}{
				"phase":      pod.Status.Phase,
				"conditions": pod.Status.Conditions,
			},
		})
	}

	// Convert resources to JSON for analysis
	resourcesJSON, err := json.MarshalIndent(resources, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal resources: %v", err)
	}

	// Analyze cluster state with OpenAI
	return s.analyzeClusterState(string(resourcesJSON))
}

func (s *Scanner) analyzeClusterState(clusterInfo string) (string, error) {
	if s.useOpenAI {
		// Use OpenAI
		resp, err := s.openaiClient.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: openai.GPT4o,
				Messages: []openai.ChatCompletionMessage{
					{
						Role: openai.ChatMessageRoleSystem,
						Content: "You are a Kubernetes infrastructure expert. Analyze the cluster state and provide " +
							"detailed recommendations for optimization, focusing on resource utilization, " +
							"scalability, and best practices.",
					},
					{
						Role: openai.ChatMessageRoleUser,
						Content: fmt.Sprintf("Please analyze this Kubernetes cluster state and provide insights about:\n"+
							"1. Resource utilization and allocation\n"+
							"2. Pod distribution and placement\n"+
							"3. Potential bottlenecks or issues\n"+
							"4. Security considerations\n"+
							"5. Optimization recommendations\n\n"+
							"Cluster state:\n%s", clusterInfo),
					},
				},
				MaxTokens: s.maxTokens,
			},
		)

		if err != nil {
			return "", fmt.Errorf("failed to analyze cluster state: %v", err)
		}

		return resp.Choices[0].Message.Content, nil
	} else {
		// Use Claude
		promptText := "You are a Kubernetes infrastructure expert. Analyze the cluster state and provide " +
			"detailed recommendations for optimization, focusing on resource utilization, " +
			"scalability, and best practices.\n\n" +
			fmt.Sprintf("Please analyze this Kubernetes cluster state and provide insights about:\n"+
				"1. Resource utilization and allocation\n"+
				"2. Pod distribution and placement\n"+
				"3. Potential bottlenecks or issues\n"+
				"4. Security considerations\n"+
				"5. Optimization recommendations\n\n"+
				"Cluster state:\n%s", clusterInfo)

		resp, err := s.claudeClient.CreateMessages(
			context.Background(),
			anthropic.MessagesRequest{
				Model: anthropic.ModelClaude3Dot7SonnetLatest,
				Messages: []anthropic.Message{
					anthropic.NewUserTextMessage(promptText),
				},
				MaxTokens: s.maxTokens,
			},
		)

		if err != nil {
			return "", fmt.Errorf("failed to analyze cluster state: %v", err)
		}

		return resp.GetFirstContentText(), nil
	}
}

func (s *Scanner) ScanVirtualMachine(vmInfo map[string]interface{}) (string, error) {
	// Convert VM info to JSON for analysis
	vmInfoJSON, err := json.MarshalIndent(vmInfo, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal VM info: %v", err)
	}

	if s.useOpenAI {
		// Use OpenAI
		resp, err := s.openaiClient.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: openai.GPT4o,
				Messages: []openai.ChatCompletionMessage{
					{
						Role: openai.ChatMessageRoleSystem,
						Content: "You are a virtual infrastructure expert. Analyze the VM configuration and metrics " +
							"to provide optimization recommendations.",
					},
					{
						Role: openai.ChatMessageRoleUser,
						Content: fmt.Sprintf("Please analyze this VM configuration and provide insights about:\n"+
							"1. Resource allocation and utilization\n"+
							"2. Performance metrics\n"+
							"3. Cost optimization opportunities\n"+
							"4. Security considerations\n"+
							"5. Recommendations for improvement\n\n"+
							"VM configuration:\n%s", string(vmInfoJSON)),
					},
				},
				MaxTokens: s.maxTokens,
			},
		)

		if err != nil {
			return "", fmt.Errorf("failed to analyze VM: %v", err)
		}

		return resp.Choices[0].Message.Content, nil
	} else {
		// Use Claude
		promptText := "You are a virtual infrastructure expert. Analyze the VM configuration and metrics " +
			"to provide optimization recommendations.\n\n" +
			fmt.Sprintf("Please analyze this VM configuration and provide insights about:\n"+
				"1. Resource allocation and utilization\n"+
				"2. Performance metrics\n"+
				"3. Cost optimization opportunities\n"+
				"4. Security considerations\n"+
				"5. Recommendations for improvement\n\n"+
				"VM configuration:\n%s", string(vmInfoJSON))

		resp, err := s.claudeClient.CreateMessages(
			context.Background(),
			anthropic.MessagesRequest{
				Model: anthropic.ModelClaude3Dot7SonnetLatest,
				Messages: []anthropic.Message{
					anthropic.NewUserTextMessage(promptText),
				},
				MaxTokens: s.maxTokens,
			},
		)

		if err != nil {
			return "", fmt.Errorf("failed to analyze VM: %v", err)
		}

		return resp.GetFirstContentText(), nil
	}
}

func (s *Scanner) GenerateRecommendations(clusterAnalysis, vmAnalysis string) (string, error) {
	if s.useOpenAI {
		// Use OpenAI
		resp, err := s.openaiClient.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: openai.GPT4o,
				Messages: []openai.ChatCompletionMessage{
					{
						Role: openai.ChatMessageRoleSystem,
						Content: "You are an infrastructure optimization expert. Based on the analysis of cluster " +
							"and VM states, provide comprehensive recommendations for infrastructure improvements.",
					},
					{
						Role: openai.ChatMessageRoleUser,
						Content: fmt.Sprintf("Based on the following analyses, provide detailed recommendations:\n\n"+
							"Cluster Analysis:\n%s\n\n"+
							"VM Analysis:\n%s\n\n"+
							"Please provide specific, actionable recommendations for:\n"+
							"1. Resource optimization\n"+
							"2. Performance improvements\n"+
							"3. Cost reduction\n"+
							"4. Security enhancements\n"+
							"5. Operational efficiency",
							clusterAnalysis, vmAnalysis),
					},
				},
				MaxTokens: s.maxTokens,
			},
		)

		if err != nil {
			return "", fmt.Errorf("failed to generate recommendations: %v", err)
		}

		return resp.Choices[0].Message.Content, nil
	} else {
		// Use Claude
		promptText := "You are an infrastructure optimization expert. Based on the analysis of cluster " +
			"and VM states, provide comprehensive recommendations for infrastructure improvements.\n\n" +
			fmt.Sprintf("Based on the following analyses, provide detailed recommendations:\n\n"+
				"Cluster Analysis:\n%s\n\n"+
				"VM Analysis:\n%s\n\n"+
				"Please provide specific, actionable recommendations for:\n"+
				"1. Resource optimization\n"+
				"2. Performance improvements\n"+
				"3. Cost reduction\n"+
				"4. Security enhancements\n"+
				"5. Operational efficiency",
				clusterAnalysis, vmAnalysis)

		resp, err := s.claudeClient.CreateMessages(
			context.Background(),
			anthropic.MessagesRequest{
				Model: anthropic.ModelClaude3Dot7SonnetLatest,
				Messages: []anthropic.Message{
					anthropic.NewUserTextMessage(promptText),
				},
				MaxTokens: s.maxTokens,
			},
		)

		if err != nil {
			return "", fmt.Errorf("failed to generate recommendations: %v", err)
		}

		return resp.GetFirstContentText(), nil
	}
}
