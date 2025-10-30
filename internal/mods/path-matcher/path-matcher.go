package path_matcher

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/indigo-sadland/jsxcite/internal/models"
	checkpath "github.com/indigo-sadland/jsxcite/internal/utils/check-path"
	"github.com/indigo-sadland/jsxcite/internal/utils/filter"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"

	"log"
	"os"
	"regexp"
	"strings"
)

func ExtractPaths(targetJs string, showSkipped bool) ([]models.Path, error) {
	var err error
	var pathStruct models.Path
	var pathStructs []models.Path
	var sourceCode []byte

	// Check source of the target JS files
	isLocal, _ := checkpath.IsLocalPath(targetJs)
	if !isLocal {
		isRemote := checkpath.IsWebPath(targetJs)
		if !isRemote {
			return []models.Path{}, fmt.Errorf("%s is not a valid path", targetJs)
		} else {
			response, err := http.Get(targetJs)
			if err != nil {
				return []models.Path{}, fmt.Errorf("failed to get remote file %s: %w", targetJs, err)
			}

			// read response body
			sourceCode, err = io.ReadAll(response.Body)
			if err != nil {
				return []models.Path{}, fmt.Errorf("failed to get remote file %s: %w", targetJs, err)
			}
		}
	} else {
		sourceCode, err = os.ReadFile(targetJs)
		if err != nil {
			return []models.Path{}, fmt.Errorf("failed to read file %s: %w", targetJs, err)
		}
	}

	if strings.HasPrefix(targetJs, "~") {
		targetJs, err = checkpath.ExpandPath(targetJs)
		if err != nil {
			return []models.Path{}, err
		}
	}

	// Create JavaScript parser
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())

	// Parse the code
	tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
	if err != nil {
		log.Fatal(err)
	}
	defer tree.Close()

	// Query to match all strings and template strings
	queryStr := `
  (string) @str
  (template_string) @str
  (comment) @comment
 `
	// Create query
	query, err := sitter.NewQuery([]byte(queryStr), javascript.GetLanguage())
	if err != nil {
		log.Fatal(err)
	}
	defer query.Close()

	// Execute query
	qc := sitter.NewQueryCursor()
	defer qc.Close()
	qc.Exec(query, tree.RootNode())

	// Iterate through matches
	//fmt.Printf("Found URL/URI-like strings (excluding imports/requires):\n")
	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			node := capture.Node
			text := string(sourceCode[node.StartByte():node.EndByte()])

			// Match in comments
			if node.Type() == "comment" {
				urls := ExtractFromComment(text)
				for _, url := range urls {
					pathStruct = models.Path{
						URL:    url,
						Source: targetJs,
						Type:   node.Type(),
					}
					pathStructs = append(pathStructs, pathStruct)
				}
				continue
			}

			// Extract the actual string content (remove quotes/backticks)
			content := extractStringContent(text)

			// Check if it looks like a URL/URI
			if !filter.IsURL(content) {
				continue
			}

			// Check the nodeContext - skip if it's in an import/require
			nodeContext := getNodeContext(node, sourceCode)
			if nodeContext.isImport || nodeContext.isRequire {
				if showSkipped {
					t := fmt.Sprintf("SKIPPED - %s", nodeContext.contextType)
					pathStruct = models.Path{
						URL:    content,
						Source: targetJs,
						Type:   t,
					}
					pathStructs = append(pathStructs, pathStruct)
				}
				continue
			}

			// Other valid matches
			pathStruct = models.Path{
				URL:    content,
				Source: targetJs,
				Type:   nodeContext.contextType,
			}
			pathStructs = append(pathStructs, pathStruct)
		}
	}

	//fmt.Println("\n--- Full tree structure for reference ---")
	//fmt.Println(tree.RootNode().String())

	return pathStructs, nil
}

// NodeContext contains information about where a string was found
type NodeContext struct {
	isImport    bool
	isRequire   bool
	contextType string
}

func ExtractFromComment(comment string) []string {
	var urls []string
	// Clean comment markers
	cleaned := comment
	cleaned = regexp.MustCompile(`^//\s*`).ReplaceAllString(cleaned, "")
	cleaned = regexp.MustCompile(`^/\*+\s*|\s*\*+/$`).ReplaceAllString(cleaned, "")
	cleaned = regexp.MustCompile(`(?m)^\s*\*\s*`).ReplaceAllString(cleaned, "")

	// Split and check each word
	words := regexp.MustCompile(`[\s,;(){}[\]<>"'`+"`"+`]+`).Split(cleaned, -1)
	for _, word := range words {
		word = strings.TrimRight(word, ".,;:!?")
		if filter.IsURL(word) {
			urls = append(urls, word)
		}
	}

	return urls
}

// getNodeContext walks up the tree to determine the context of a node
func getNodeContext(node *sitter.Node, sourceCode []byte) NodeContext {
	ctx := NodeContext{contextType: "other"}

	// Walk up the parent chain
	current := node.Parent()
	for current != nil {
		nodeType := current.Type()

		switch nodeType {
		case "import_statement":
			ctx.isImport = true
			ctx.contextType = "import"
			return ctx
		case "call_expression":
			// Check if it's a require() call
			if current.ChildCount() > 0 {
				fnNode := current.Child(0)
				if fnNode != nil && fnNode.Type() == "identifier" {
					// Safely extract function name
					start := fnNode.StartByte()
					end := fnNode.EndByte()
					if start < end && end <= uint32(len(sourceCode)) {
						fnText := string(sourceCode[start:end])
						if fnText == "require" {
							ctx.isRequire = true
							ctx.contextType = "require"
							return ctx
						}
					}
				}
			}
			// Check for other function calls
			ctx.contextType = "function_call"
		case "variable_declarator":
			ctx.contextType = "variable"
		case "assignment_expression":
			ctx.contextType = "assignment"
		case "member_expression":
			ctx.contextType = "member_access"
		}

		current = current.Parent()
	}

	return ctx
}

// extractStringContent removes quotes or backticks from string literals
func extractStringContent(s string) string {
	if len(s) < 2 {
		return s
	}

	// Remove quotes
	if (s[0] == '"' || s[0] == '\'' || s[0] == '`') && s[0] == s[len(s)-1] {
		return s[1 : len(s)-1]
	}

	return s
}
