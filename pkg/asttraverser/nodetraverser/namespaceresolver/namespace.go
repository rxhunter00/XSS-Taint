package namespaceresolver

import (
	"errors"
	"strings"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/rxhunter00/XSS-Taint/pkg/asttraverser/astutils"
)

// Namespace context
type Namespace struct {
	Namespace string
	Aliases   map[string]map[string]string
}

// NewNamespace constructor
func NewNamespace(NSName string) *Namespace {
	return &Namespace{
		Namespace: NSName,
		Aliases: map[string]map[string]string{
			"":         {},
			"const":    {},
			"function": {},
		},
	}
}

// AddAlias adds a new alias
func (ns *Namespace) AddAlias(aliasType string, aliasName string, alias string) {
	aliasType = strings.ToLower(aliasType)

	if aliasType == "const" {
		ns.Aliases[aliasType][alias] = aliasName
	} else {
		ns.Aliases[aliasType][strings.ToLower(alias)] = aliasName
	}
}

// ResolveName returns a resolved fully qualified name
func (ns *Namespace) ResolveName(nameNode ast.Vertex, aliasType string) (string, error) {
	switch n := nameNode.(type) {
	case *ast.NameFullyQualified:

		return astutils.ConcatNameParts(n.Parts), nil

	case *ast.NameRelative:
		if ns.Namespace == "" {
			return astutils.ConcatNameParts(n.Parts), nil
		}
		return ns.Namespace + "\\" + astutils.ConcatNameParts(n.Parts), nil

	case *ast.Name:
		if aliasType == "const" && len(n.Parts) == 1 {
			part := string(n.Parts[0].(*ast.NamePart).Value)
			lowerPart := strings.ToLower(part)
			if lowerPart == "true" || lowerPart == "false" || lowerPart == "null" {
				return part, nil
			}
			switch part {
			case "INPUT_GET":
				fallthrough
			case "INPUT_POST":
				fallthrough
			case "INPUT_COOKIE":
				fallthrough
			case "INPUT_SERVER":
				fallthrough
			case "INPUT_ENV":
				fallthrough
			case "INPUT_SESSION":
				fallthrough
			case "INPUT_REQUEST":
				return part, nil
			}
		}

		if aliasType == "function" && len(n.Parts) == 1 {
			part := strings.ToLower(string(n.Parts[0].(*ast.NamePart).Value))
			switch part {
			case "define":
				fallthrough
			case "defined":
				fallthrough
			case "settype":
				fallthrough
			case "gettype":
				fallthrough
			case "is_array":
				fallthrough
			case "is_null":
				fallthrough
			case "is_bool":
				fallthrough
			case "is_float":
				fallthrough
			case "is_int":
				fallthrough
			case "is_string":
				fallthrough
			case "is_object":
				fallthrough
			case "is_resource":
				fallthrough
			case "var_dump":
				fallthrough
			case "boolval":
				fallthrough
			case "intval":
				fallthrough
			case "floatval":
				fallthrough
			case "strval":
				fallthrough
			case "is_numeric":
				fallthrough
			case "filter_input":
				fallthrough
			case "filter_input_array":
				return part, nil
			}
		}

		if aliasType == "" && len(n.Parts) == 1 {
			part := strings.ToLower(string(n.Parts[0].(*ast.NamePart).Value))

			switch part {
			case "self":
				fallthrough
			case "static":
				fallthrough
			case "parent":
				fallthrough
			case "int":
				fallthrough
			case "float":
				fallthrough
			case "bool":
				fallthrough
			case "string":
				fallthrough
			case "void":
				fallthrough
			case "iterable":
				fallthrough
			case "mixed":
				fallthrough
			case "object":
				fallthrough
			case "define":
				return part, nil
			}
		}

		aliasName, err := ns.ResolveAlias(nameNode, aliasType)
		if err != nil {
			// resolve as relative name if alias not found
			if ns.Namespace == "" {
				return astutils.ConcatNameParts(n.Parts), nil
			}
			return ns.Namespace + "\\" + astutils.ConcatNameParts(n.Parts), nil
		}

		if len(n.Parts) > 1 {
			// if name qualified, replace first part by alias
			return aliasName + "\\" + astutils.ConcatNameParts(n.Parts[1:]), nil
		}

		return aliasName, nil
	}

	return "", errors.New("must be instance of name.Names")
}

// ResolveAlias returns alias or error if not found
func (ns *Namespace) ResolveAlias(nameNode ast.Vertex, aliasType string) (string, error) {
	aliasType = strings.ToLower(aliasType)
	nameParts := nameNode.(*ast.Name).Parts

	firstPartStr := string(nameParts[0].(*ast.NamePart).Value)

	if len(nameParts) > 1 { // resolve aliases for qualified names, always against class alias type
		firstPartStr = strings.ToLower(firstPartStr)
		aliasType = ""
	} else {
		if aliasType != "const" { // constants are case-sensitive
			firstPartStr = strings.ToLower(firstPartStr)
		}
	}

	aliasName, ok := ns.Aliases[aliasType][firstPartStr]
	if !ok {
		return "", errors.New("not found")
	}

	return aliasName, nil
}
