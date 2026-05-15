# DocuMind OSS Agent

**Agentic Documentation Infrastructure for Cloud-Native OSS Ecosystems**

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

DocuMind is an intelligent documentation infrastructure platform for OSS ecosystems like **OpenKruise**, **Kubernetes**, and **Argo**. It continuously ingests repositories, maintains semantic documentation context, detects documentation drift, and exposes repository intelligence through MCP-compatible interfaces.

---

## ✨ Features

- **🔄 Repository Ingestion** — Clone and incrementally index GitHub repositories
- **📝 Semantic Chunking** — Parse Markdown, YAML, and Go with heading-aware chunking
- **🧠 Embedding Pipeline** — Local embeddings via Ollama (nomic-embed-text)
- **🔍 Hybrid Search** — Vector similarity + keyword matching with version filtering
- **📋 Quality Evaluation** — Detect broken links, stale docs, invalid YAML, duplicates
- **🕸️ Knowledge Graph** — Cross-repository feature-doc-code linking
- **🔌 MCP Server** — Model Context Protocol for AI IDE integration
- **🤖 Agentic Workflows** — ReAct self-healing loop for automated doc fixes
- **📊 Metrics** — Prometheus + OpenTelemetry observability

## 🏗️ Architecture

```
GitHub Repositories
        │
        ▼
┌─────────────────────┐
│  Ingestion Engine    │ ← Clone/Pull, File Walking, Change Detection
│  (go-git + walker)   │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│  Parser Pipeline     │ ← Markdown AST (goldmark), YAML, Go parsing
│  (goldmark + regex)  │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│  Embedding Pipeline  │ ← Ollama (nomic-embed-text) / OpenAI compatible
│  (Ollama client)     │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│  Vector Store        │ ← chromem-go (embedded, persistent)
│  + SQLite Metadata   │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│  Knowledge Graph     │ ← Feature ↔ Doc ↔ Code ↔ Release linking
│  (in-memory graph)   │
└────────┬────────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌────────┐ ┌────────────┐
│  MCP   │ │  REST API   │
│ Server │ │  + Dashboard│
└────────┘ └────────────┘
    │              │
    ▼              ▼
┌─────────────────────┐
│  AI IDEs / Agents   │ ← Cursor, Windsurf, external agents
│  GitHub Actions     │
└─────────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.22+**
- **Ollama** (for local embeddings): [Install Guide](https://ollama.ai)
- **Git**

### Setup

```bash
# Clone the repository
git clone https://github.com/priya-sharma/documind.git
cd documind

# Pull the embedding model
ollama pull nomic-embed-text

# Build
make build

# Run
./bin/documind version
```

### Usage

```bash
# Ingest a repository
./bin/documind ingest --repo https://github.com/openkruise/kruise

# Ingest all configured repos
./bin/documind ingest --all

# Search documentation
./bin/documind search "How does CloneSet handle scaling?"

# Search with version filter
./bin/documind search --version v1.3 "sidecar injection"

# Run quality evaluation
./bin/documind evaluate --repo kruise

# Start HTTP API server
./bin/documind serve --http :8080

# Start MCP server (for AI IDEs)
./bin/documind serve --mcp
```

## 📁 Project Structure

```
documind/
├── cmd/
│   └── documind/
│       └── main.go              # CLI entry point (cobra)
├── internal/
│   ├── config/                  # Viper-based configuration
│   ├── models/                  # Core domain types
│   ├── storage/                 # SQLite metadata storage
│   ├── version/                 # Build version info
│   ├── ingestion/               # Repository cloning & file walking
│   ├── parser/                  # Markdown/YAML/Go parsing
│   ├── embedding/               # Ollama embedding pipeline
│   ├── vectorstore/             # chromem-go vector database
│   ├── search/                  # Hybrid search engine
│   ├── evaluator/               # Documentation quality checks
│   ├── graph/                   # Knowledge graph
│   ├── mcp/                     # MCP server implementation
│   ├── api/                     # REST API server
│   └── agent/                   # ReAct agentic workflows
├── web/                         # Dashboard (HTML/CSS/JS)
├── docs/                        # Architecture documentation
├── .github/workflows/           # CI/CD pipelines
├── config.yaml                  # Default configuration
├── Makefile                     # Build automation
└── README.md
```

## 🔧 Configuration

DocuMind uses a layered configuration system: `defaults → config.yaml → env vars → CLI flags`

```yaml
# config.yaml
embedding:
  provider: "ollama"
  model: "nomic-embed-text"
  endpoint: "http://localhost:11434"

repositories:
  - name: "kruise"
    url: "https://github.com/openkruise/kruise"
    branches: ["master"]

search:
  top_k: 10
  hybrid_weight: 0.7  # 0=keyword, 1=vector
```

Environment variables use the `DOCUMIND_` prefix:

```bash
export DOCUMIND_EMBEDDING_ENDPOINT=http://localhost:11434
export DOCUMIND_LOGGING_LEVEL=debug
```

## 🔌 MCP Tools

DocuMind exposes the following MCP tools for AI IDE integration:

| Tool | Description |
|------|-------------|
| `docs_lookup` | Semantic documentation retrieval |
| `release_lookup` | Version-aware retrieval |
| `architecture_summary` | Repository architecture overview |
| `feature_context` | Feature-to-code-to-doc mapping |
| `code_reference` | Implementation lookup |
| `evaluate_docs` | Documentation quality check |

## 📈 Roadmap

- [x] Project scaffolding & CLI
- [x] Repository ingestion engine
- [x] Markdown/YAML/Go parsing pipeline
- [x] Embedding pipeline (Ollama + chromem-go)
- [x] Semantic search with hybrid retrieval
- [x] Documentation quality evaluation
- [x] Knowledge graph
- [x] Version-aware retrieval
- [x] MCP server
- [x] REST API
- [ ] Prometheus metrics
- [ ] ReAct agentic workflows
- [ ] GitHub Actions CI/CD
- [ ] Web dashboard
- [ ] Graph RAG
- [ ] Multi-language support

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📄 License

Apache License 2.0 — see [LICENSE](LICENSE) for details.
