package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/priya-sharma/documind/internal/models"
)

// CodeParser extracts semantic information from Go source code.
type CodeParser struct {
	fset *token.FileSet
}

// NewCodeParser creates a new CodeParser.
func NewCodeParser() *CodeParser {
	return &CodeParser{
		fset: token.NewFileSet(),
	}
}

// Parse extracts package docs, structs, and functions from Go code.
func (p *CodeParser) Parse(repoID, filePath string, source []byte) ([]models.Chunk, error) {
	node, err := parser.ParseFile(p.fset, filePath, source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing Go file: %w", err)
	}

	var chunks []models.Chunk
	packageName := node.Name.Name

	// 1. Package Documentation
	if node.Doc != nil {
		chunks = append(chunks, models.Chunk{
			RepoID:      repoID,
			FilePath:    filePath,
			Content:     node.Doc.Text(),
			HeadingPath: fmt.Sprintf("Package: %s", packageName),
			ChunkType:   models.ChunkTypeCode,
			Metadata: map[string]string{
				"type": "package_doc",
			},
		})
	}

	// 2. Traverse AST for structs and functions
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			// Structs/Interfaces
			if _, ok := x.Type.(*ast.StructType); ok {
				title := fmt.Sprintf("%s > struct %s", packageName, x.Name.Name)
				doc := ""
				if x.Doc != nil {
					doc = x.Doc.Text()
				}
				chunks = append(chunks, models.Chunk{
					RepoID:      repoID,
					FilePath:    filePath,
					Content:     fmt.Sprintf("%s\n%s", doc, x.Name.Name),
					HeadingPath: title,
					ChunkType:   models.ChunkTypeCode,
					Metadata: map[string]string{
						"type": "struct",
						"name": x.Name.Name,
					},
				})
			}

		case *ast.FuncDecl:
			// Functions/Methods
			funcName := x.Name.Name
			if x.Recv != nil {
				// It's a method — handle both value and pointer receivers
				for _, field := range x.Recv.List {
					switch t := field.Type.(type) {
					case *ast.Ident:
						funcName = fmt.Sprintf("(%s) %s", t.Name, funcName)
					case *ast.StarExpr:
						if ident, ok := t.X.(*ast.Ident); ok {
							funcName = fmt.Sprintf("(*%s) %s", ident.Name, funcName)
						}
					}
				}
			}

			doc := ""
			if x.Doc != nil {
				doc = x.Doc.Text()
			}

			title := fmt.Sprintf("%s > func %s", packageName, funcName)
			chunks = append(chunks, models.Chunk{
				RepoID:      repoID,
				FilePath:    filePath,
				Content:     fmt.Sprintf("%s\n%s", doc, funcName),
				HeadingPath: title,
				ChunkType:   models.ChunkTypeCode,
				Metadata: map[string]string{
					"type": "function",
					"name": funcName,
				},
			})
		}
		return true
	})

	return chunks, nil
}
