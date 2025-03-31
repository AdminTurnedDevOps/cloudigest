# CloudDigest

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
git clone https://github.com/yourusername/cloudigest.git

# Change to the project directory
cd cloudigest

# Install dependencies
go mod download

# Build the binary
go build -o cloudigest
```

## Usage
The CLI interface provides the following commands:

```
cloudigest analyze diagram [file] - Analyze architecture diagrams
```

```
cloudigest analyze doc [file] - Analyze documentation
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

# Analyze an architecture diagram
cloudigest analyze diagram --file path/to/diagram.png

# Upload documentation
cloudigest analyze doc --file path/to/spec.pdf

# Get recommendations for existing infrastructure
cloudigest recommend --target kubernetes
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
  api_key: your-api-key-here

scanning:
  kubernetes: true
  cloud_providers:
    - aws
    - azure
```

## License

MIT License 