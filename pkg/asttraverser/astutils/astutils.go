package astutils

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/position"
)

func GetStmtList(node ast.Vertex) ([]ast.Vertex, error) {
	switch typeNode := node.(type) {
	case *ast.StmtStmtList:
		return typeNode.Stmts, nil
	case *ast.StmtNop:
		return make([]ast.Vertex, 0), nil
	default:
		return []ast.Vertex{typeNode}, nil
	}
}

func GetNameString(nameNode ast.Vertex) (string, error) {
	switch name := nameNode.(type) {
	case *ast.Name:
		return ConcatNameParts(name.Parts), nil
	case *ast.NameFullyQualified:
		return ConcatNameParts(name.Parts), nil
	case *ast.NameRelative:
		return ConcatNameParts(name.Parts), nil
	case *ast.Identifier:
		return string(name.Value), nil
	}
	return "", fmt.Errorf("incompatible name type '%s'", reflect.TypeOf(nameNode))
}

func IsScalarNode(n ast.Vertex) bool {
	if n == nil {
		return false
	}
	switch n.(type) {
	case *ast.ScalarDnumber, *ast.ScalarString, *ast.ScalarLnumber:
		return true
	default:
		return false
	}

}

func ConcatNameParts(parts ...[]ast.Vertex) string {
	str := ""

	for _, p := range parts {
		for _, n := range p {
			if str == "" {
				str = string(n.(*ast.NamePart).Value)
			} else {
				str = str + "\\" + string(n.(*ast.NamePart).Value)
			}
		}
	}

	return str
}

func PopLabelStack(stack *[]ast.StmtLabel) *ast.StmtLabel {
	if IsLabelStackEmpty(stack) {
		log.Fatal("popLabelStack on empty stack")
	}
	label := (*stack)[len(*stack)-1]
	*stack = (*stack)[:len(*stack)-1]
	return &label
}
func TopLabelStack(stack *[]ast.StmtLabel) *ast.StmtLabel {
	if IsLabelStackEmpty(stack) {
		log.Fatal("topLabelStack on empty stack")
	}
	label := (*stack)[len(*stack)-1]

	return &label
}
func IsLabelStackEmpty(stack *[]ast.StmtLabel) bool {
	return len(*stack) == 0
}

func CreateNameParts(name string, pos *position.Position) []ast.Vertex {
	nameParts := make([]ast.Vertex, 0, 5)
	parts := strings.Split(name, "\\")

	for _, p := range parts {
		namePart := &ast.NamePart{
			Position: pos,
			Value:    []byte(p),
		}
		nameParts = append(nameParts, namePart)
	}

	return nameParts
}
