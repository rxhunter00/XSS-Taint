// ref https://github.com/ircmaxell/php-cfg/blob/master/lib/PHPCfg/AstVisitor/MagicStringResolver.php

package magicconstansresolver

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/position"
	"github.com/rxhunter00/XSS-Taint/pkg/asttraverser"
	"github.com/rxhunter00/XSS-Taint/pkg/asttraverser/astutils"
)

type MagicConstantResolver struct {
	classStack    []string
	parentStack   []string
	functionStack []string
	methodStack   []string
	currNamespace string
	filename      string
}

func NewMagicConstantResolver(filename string) *MagicConstantResolver {
	return &MagicConstantResolver{
		classStack:    make([]string, 0),
		parentStack:   make([]string, 0),
		functionStack: make([]string, 0),
		methodStack:   make([]string, 0),
		filename:      filename,
	}
}

func (mcr *MagicConstantResolver) EnterNode(n ast.Vertex) (ast.Vertex, asttraverser.ReturnModeFlag) {

	/*
		__CLASS__ return class name
		__DIR__ return directory of file
		__FILE__ file name
		__FUNCTION__ 	return function name
		__LINE__ return current line number
		__METHOD__ if inside method, return class and function name
		__NAMESPACE__ If inside a namespace, the name of the namespace is returned.
		__TRAIT__ If inside a trait, the trait name is returned.
		ClassName::class Returns the name of the specified class and the name of the namespace, if any.
	*/
	switch n := n.(type) {
	case *ast.StmtClass:
		// Append class name to class stack
		className := n.Name.(*ast.Identifier)
		classNameStr := string(className.Value)
		mcr.classStack = append(mcr.classStack, classNameStr)

		// Did Class extend parent?
		if n.Extends != nil {
			parentNameStr, err := astutils.GetNameString(n.Extends)
			if err != nil {
				log.Fatal("Error extends name in StmtClass")
			}
			mcr.parentStack = append(mcr.parentStack, parentNameStr)
		} else {
			mcr.parentStack = append(mcr.parentStack, "")
		}

	case *ast.StmtTrait:
		// Append trait name to class stack
		traitName := n.Name.(*ast.Identifier)
		traitNameStr := string(traitName.Value)
		mcr.classStack = append(mcr.classStack, traitNameStr)
		mcr.parentStack = append(mcr.parentStack, "")

	case *ast.StmtInterface:
		// Append trait name to class stack
		interfaceName := n.Name.(*ast.Identifier)
		interfaceNameStr := string(interfaceName.Value)
		mcr.classStack = append(mcr.classStack, interfaceNameStr)
		mcr.parentStack = append(mcr.parentStack, "")

	case *ast.StmtClassMethod:
		// Append method name to method and function stack
		functionName := n.Name.(*ast.Identifier)
		functionNameStr := string(functionName.Value)
		mcr.functionStack = append(mcr.functionStack, functionNameStr)

		currClassName := mcr.classStack[len(mcr.classStack)-1]
		methodNameStr := fmt.Sprintf("%s::%s", currClassName, functionNameStr)
		mcr.methodStack = append(mcr.methodStack, methodNameStr)

	case *ast.StmtFunction:
		// Append function name to function stack
		functionName := n.Name.(*ast.Identifier)
		functionNameStr := string(functionName.Value)
		mcr.functionStack = append(mcr.functionStack, functionNameStr)

	case *ast.StmtNamespace:
		// set current namespace context
		if n.Name != nil {
			nameSpaceStr := concatNameParts(n.Name.(*ast.Name).Parts)
			mcr.currNamespace = nameSpaceStr
		}

	case *ast.Name:
		// Get name string
		nodeName := concatNameParts(n.Parts)
		if nodeName == "self" {

			if len(mcr.classStack) == 0 {
				log.Printf("No Active Class")
			}
			// convert self to the current class name
			currClassName := mcr.classStack[len(mcr.classStack)-1]
			return &ast.NameFullyQualified{
				Position: n.Position,
				Parts:    createNameParts(currClassName, n.Position),
			}, asttraverser.REPLACEMODE
		} else if nodeName == "parent" {
			// Error 'parent' constant not recognized
			if len(mcr.parentStack) == 0 {
				log.Printf("No Active Class")
			}

			// convert 'parent' to the current parent name
			parentName := mcr.parentStack[len(mcr.parentStack)-1]

			return &ast.NameFullyQualified{
				Position: n.Position,
				Parts:    createNameParts(parentName, n.Position),
			}, asttraverser.REPLACEMODE
		}

	case *ast.ScalarMagicConstant:
		magicConstStr := string(n.Value)

		if magicConstStr == "__CLASS__" {
			var currClassName string

			// If not in class scope, convert to empty string
			if len(mcr.classStack) == 0 {
				currClassName = ""
			} else {
				currClassName = mcr.classStack[len(mcr.classStack)-1]
			}

			return &ast.ScalarString{
				Position: n.Position,
				Value:    []byte(currClassName),
			}, asttraverser.REPLACEMODE
		} else if magicConstStr == "__TRAIT__" {
			var currTraitName string

			// If not in trait scope, convert to empty string
			if len(mcr.classStack) == 0 {
				currTraitName = ""
			} else {
				currTraitName = mcr.classStack[len(mcr.classStack)-1]
			}

			return &ast.ScalarString{
				Position: n.Position,
				Value:    []byte(currTraitName),
			}, asttraverser.REPLACEMODE
		} else if magicConstStr == "__NAMESPACE__" {
			return &ast.ScalarString{
				Position: n.Position,
				Value:    []byte(mcr.currNamespace),
			}, asttraverser.REPLACEMODE
		} else if magicConstStr == "__FUNCTION__" {
			var functionName string

			// If not in function scope, convert to empty string
			if len(mcr.classStack) == 0 {
				functionName = ""
			} else {
				functionName = mcr.functionStack[len(mcr.functionStack)-1]
			}

			return &ast.ScalarString{
				Position: n.Position,
				Value:    []byte(functionName),
			}, asttraverser.REPLACEMODE
		} else if magicConstStr == "__METHOD__" {
			var methodName string

			// If not in method scope, convert to empty string
			if len(mcr.methodStack) == 0 {
				methodName = ""
			} else {
				methodName = mcr.methodStack[len(mcr.methodStack)-1]
			}

			return &ast.ScalarString{
				Position: n.Position,
				Value:    []byte(methodName),
			}, asttraverser.REPLACEMODE
		} else if magicConstStr == "__LINE__" {
			return &ast.ScalarLnumber{
				Position: n.Position,
				Value:    []byte(strconv.Itoa(n.Position.StartLine)),
			}, asttraverser.REPLACEMODE
		} else if magicConstStr == "__FILE__" {
			return &ast.ScalarString{
				Position: n.Position,
				Value:    []byte(mcr.filename),
			}, asttraverser.REPLACEMODE
		} else if magicConstStr == "__DIR__" {
			dir := filepath.Dir(mcr.filename)
			return &ast.ScalarString{
				Position: n.Position,
				Value:    []byte(dir),
			}, asttraverser.REPLACEMODE
		} else {
			log.Printf("Invalid Magic Constant: %s", magicConstStr)
		}
	}

	return nil, asttraverser.REPLACEMODE
}

func (mcr *MagicConstantResolver) LeaveNode(n ast.Vertex) (ast.Vertex, asttraverser.ReturnModeFlag) {
	switch n := n.(type) {
	case *ast.StmtClass:
		popStringStack(&mcr.classStack)
		popStringStack(&mcr.parentStack)
	case *ast.StmtTrait, *ast.StmtInterface:
		popStringStack(&mcr.classStack)
	case *ast.StmtFunction:
		popStringStack(&mcr.functionStack)
	case *ast.StmtClassMethod:
		popStringStack(&mcr.methodStack)
	case *ast.StmtNamespace:
		if len(n.Stmts) > 0 {
			mcr.currNamespace = ""
		}
	}

	return nil, asttraverser.REPLACEMODE
}

func popStringStack(st *[]string) string {
	if len(*st) == 0 {
		log.Printf("popStringStack:Pop on Empty")
	}
	top := (*st)[len(*st)-1]
	*st = (*st)[:len(*st)-1]

	return top
}

func createNameParts(name string, pos *position.Position) []ast.Vertex {
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

func concatNameParts(parts ...[]ast.Vertex) string {
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
