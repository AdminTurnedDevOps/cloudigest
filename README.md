# CloudDigest

<p align="center">
 <img src="images/logo.png?raw=true" alt="Logo" width="50%" height="50%" />
</p>

CloudDigest is an intelligent infrastructure optimization companion that helps engineers make informed decisions about running their infrastructure, services, and containers. It provides recommendations and best practices without implementing them directly, acting as a knowledgeable advisor for your infrastructure decisions.

## Features

1. **RAG-based Knowledge Base**
   - Incorporates best practices and recommendations from trusted sources
   - Continuously updated knowledge base
   - Context-aware suggestions

2. **AI-Powered Internet Search**
   - Real-time search for current best practices
   - Integration with multiple sources for comprehensive recommendations

3. **Document and Diagram Analysis**
   - Upload and analyze architecture diagrams
   - Process technical documentation
   - Extract insights from design specifications
   - AI-powered visual understanding

4. **Infrastructure Scanner**
   - Kubernetes cluster analysis
   - Virtual machine resource utilization scanning
   - Service deployment inspection
   - Resource optimization recommendations

## Installation

```bash
# Clone the repository
git clone https://github.com/AdminTurnedDevOps/cloudigest

# Install dependencies
go mod download

# Build the binary
go build -o cloudigest
```

## Usage
The CLI interface provides the following commands:

```
cloudigest analyze [file] - Analyze architecture diagrams or documentation
```

```
cloudigest scan kubernetes - Scan Kubernetes clusters
```

```
cloudigest scan vm - Scan virtual machines
```

cloudigest query [question] - Query the knowledge base

To use the tool:
- Copy the config.yaml file to ~/.cloudigest/config.yaml
- Add your OpenAI API key to the configuration
- Run the desired commands

## Usage Examples

```bash
# Get help
cloudigest --help

# Scan a Kubernetes cluster
cloudigest scan kubernetes

# Analyze an architecture diagram or doc
cloudigest analyze file_name
```

## Requirements

- Go 1.21 or higher
- OpenAI API key
- Kubernetes configuration (for cluster scanning)
- Cloud provider credentials (for cloud resource scanning)

## Configuration

Create a configuration file at `~/.cloudigest/config.yaml` with your API keys and preferences:

```yaml
openai:
  api_key: "your-openai-api-key-here"

# Optional
claude:
  api_key: "your-claude-api-key-here"

serper:
  api_key: "your-serper-api-key-here"

scanning:
  kubernetes: true
  cloud_providers:
    - aws
    - azure
    - gcp

rag:
  chunk_size: 500
  max_tokens: 2000
  sources:
    - url: "https://controlplane.com/community-blog/post/optimize-kubernetes-workloads"
      type: "article"
      name: "Kubernetes Workload Optimization Guide" 
```

## License

MIT License 