# DocuMind Architecture

This document describes the technical architecture of DocuMind OSS Agent,
including data flow diagrams, component interactions, and design decisions.

## System Overview

DocuMind is designed as a **pipeline architecture** where data flows through
distinct stages: ingestion → parsing → embedding → indexing → retrieval → action.

Each stage is independently testable and can be extended without affecting others.

## Data Flow: Ingestion Pipeline

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   GitHub      │     │   Cloner     │     │   Walker     │
│   Repository  │────▶│  (go-git)    │────▶│  (filepath)  │
│              │     │              │     │              │
└──────────────┘     └──────────────┘     └──────┬───────┘
                                                  │
                                                  ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   SQLite     │◀────│   Differ     │◀────│  File Hashes │
│   State DB   │     │  (SHA-256)   │     │  (current)   │
│              │     │              │     │              │
└──────┬───────┘     └──────┬───────┘     └──────────────┘
       │                    │
       │              ┌─────┴─────┐
       │              │ ChangeSet │
       │              │ (A/M/D)   │
       │              └─────┬─────┘
       │                    │
       ▼                    ▼
 Previous State      Only changed files
 for comparison      proceed to parsing
```

**Key Design Decisions:**
- **Incremental indexing** via content hashing — only re-process changed files
- **SQLite** for metadata state — cleaner querying than JSON, easy joins
- **Shallow clones** by default — faster for large repos

## Data Flow: Parsing Pipeline

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  .md files   │────▶│  Goldmark    │────▶│  Heading-    │
│              │     │  AST Parser  │     │  Aware       │
│              │     │              │     │  Chunker     │
└──────────────┘     └──────────────┘     └──────────────┘

┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  .yaml files │────▶│  YAML        │────▶│  K8s         │
│              │     │  Parser      │     │  Resource     │
│              │     │              │     │  Extractor    │
└──────────────┘     └──────────────┘     └──────────────┘

┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  .go files   │────▶│  Regex-based │────▶│  Struct/Func │
│              │     │  Parser      │     │  Extractor   │
│              │     │              │     │              │
└──────────────┘     └──────────────┘     └──────────────┘
```

**Chunking Strategy:**
- Respect heading boundaries — never split mid-section
- Max chunk size: 512 tokens (configurable)
- 50-token overlap for context continuity
- Metadata inheritance: each chunk carries its heading path (e.g., `CloneSet > Scaling > Partition`)

## Data Flow: Embedding & Retrieval

```
                    ┌──────────────┐
                    │   Chunks     │
                    │  (with meta) │
                    └──────┬───────┘
                           │
                    ┌──────▼───────┐
                    │   Ollama     │
                    │  nomic-embed │
                    │  -text       │
                    └──────┬───────┘
                           │
                    ┌──────▼───────┐
                    │  chromem-go  │
                    │  Vector DB   │
                    │  (embedded)  │
                    └──────┬───────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
       ┌──────▼──┐  ┌──────▼──┐  ┌─────▼───┐
       │ Vector  │  │ Keyword │  │ Graph   │
       │ Search  │  │ Search  │  │ Enhanced│
       └────┬────┘  └────┬────┘  └────┬────┘
            │            │            │
            └────────┬───┘            │
                     │                │
              ┌──────▼───────┐        │
              │   Hybrid     │◀───────┘
              │   Ranker     │
              │ (configurable│
              │  weighting)  │
              └──────┬───────┘
                     │
              ┌──────▼───────┐
              │   Version    │
              │   Filter     │
              │ (semver)     │
              └──────────────┘
```

## Data Flow: MCP Integration

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  AI IDE      │     │  MCP Server  │     │  DocuMind    │
│  (Cursor,    │◀───▶│  (stdio)     │◀───▶│  Core        │
│   Windsurf)  │     │              │     │  Engine      │
└──────────────┘     └──────────────┘     └──────────────┘
                           │
                     ┌─────┴─────┐
                     │ MCP Tools │
                     ├───────────┤
                     │docs_lookup│
                     │release_   │
                     │  lookup   │
                     │feature_   │
                     │  context  │
                     │arch_      │
                     │  summary  │
                     │code_ref   │
                     │eval_docs  │
                     └───────────┘
```

## Data Flow: Agentic Self-Healing

```
┌──────────────┐
│   Quality    │
│   Report     │──────┐
└──────────────┘      │
                      ▼
              ┌───────────────┐
              │   OBSERVE     │
              │ (detect issue)│
              └───────┬───────┘
                      │
              ┌───────▼───────┐
              │    THINK      │
              │ (analyze with │
              │  LLM + graph) │
              └───────┬───────┘
                      │
              ┌───────▼───────┐
              │     ACT       │
              │ (generate fix,│
              │  create diff) │
              └───────┬───────┘
                      │
              ┌───────▼───────┐
              │  Draft PR     │
              │ (GitHub API)  │
              └───────────────┘
```

## Knowledge Graph Schema

```
     ┌──────────┐
     │ Feature  │
     └────┬─────┘
          │ documents
     ┌────▼─────┐        ┌───────────┐
     │ Document │───────▶│ CodeFile  │
     └────┬─────┘ refs    └───────────┘
          │
     ┌────▼─────┐        ┌───────────┐
     │ Example  │───────▶│ CRD       │
     └──────────┘ uses    └─────┬─────┘
                                │ released_in
                          ┌─────▼─────┐
                          │ Release   │
                          └───────────┘
```

**Node Types:** Feature, Document, CodeFile, Release, Example, CRD

**Edge Types:** documents, implements, examples, released_in, depends_on, related_to

## Technology Choices

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| Language | Go 1.22+ | Cloud-native ecosystem, great concurrency |
| Markdown | goldmark | CommonMark-compliant, AST-based, pure Go |
| Vector DB | chromem-go | Zero-dependency, embedded, persistent |
| Embeddings | Ollama (nomic-embed-text) | Free, local, no API keys |
| Metadata | SQLite (modernc.org/sqlite) | Pure Go, no CGO, easy querying |
| CLI | cobra | Industry standard for Go CLIs |
| Config | viper | Layered config, env var support |
| Logging | slog (stdlib) | Zero-dependency structured logging |
| MCP | mark3labs/mcp-go | Most popular Go MCP SDK |
| HTTP | chi + net/http | Lightweight, idiomatic |
| Metrics | prometheus/client_golang | Industry standard |
| CI | GitHub Actions | Free for public repos |
