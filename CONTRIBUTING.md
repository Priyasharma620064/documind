# Contributing to DocuMind

Thank you for your interest in contributing to DocuMind! This project aims to build a robust documentation infrastructure for the cloud-native ecosystem.

## 🛠️ Development Environment

1. **Go 1.22+**: The core project is written in Go.
2. **Ollama**: Required for local embeddings.
3. **Git**: Used for repository ingestion.

### Building from Source

```bash
make build
```

### Running Tests

```bash
go test ./...
```

## 📜 Pull Request Guidelines

1. **Granular Commits**: Follow the pattern of small, descriptive commits.
2. **Documentation**: Update `docs/architecture.md` if you make architectural changes.
3. **Tests**: Add unit tests for new logic in `internal/`.
4. **Style**: Follow standard Go formatting (`go fmt`).

## 🧪 Testing your changes

If you add a new parser or ingestion logic, please test it against a real repository:

```bash
./bin/documind ingest --repo <your-repo-url>
./bin/documind search "query related to your change"
```

## ⚖️ License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
