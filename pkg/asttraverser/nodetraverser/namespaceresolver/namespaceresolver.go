// https://github.com/nikic/PHP-Parser/blob/master/lib/PhpParser/NodeVisitor/NameResolver.php


package namespaceresolver

import (
	"fmt"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/rxhunter00/XSS-Taint/pkg/asttraverser"
	"github.com/rxhunter00/XSS-Taint/pkg/asttraverser/astutils"
)

type NamespaceResolver struct {
	NamespaceCtx *Namespace

	goDeep           bool
	anonClassCounter int
}

// NewNamespaceResolver NamespaceResolver type constructor
func NewNamespaceResolver() *NamespaceResolver {
	return &NamespaceResolver{
		NamespaceCtx:     NewNamespace(""),
		goDeep:           true,
		anonClassCounter: 0,
	}
}

func (nsr *NamespaceResolver) EnterNode(n ast.Vertex) (ast.Vertex, asttraverser.ReturnModeFlag) {
	switch n := n.(type) {
	case *ast.StmtNamespace:
		return nsr.StmtNamespace(n), asttraverser.REPLACEMODE
	case *ast.StmtUseList:
		return nsr.StmtUse(n), asttraverser.REPLACEMODE
	case *ast.StmtGroupUseList:
		return nsr.StmtGroupUse(n), asttraverser.REPLACEMODE
	case *ast.StmtClass:
		return nsr.StmtClass(n), asttraverser.REPLACEMODE
	case *ast.StmtInterface:
		return nsr.StmtInterface(n), asttraverser.REPLACEMODE
	case *ast.StmtTrait:
		return nsr.StmtTrait(n), asttraverser.REPLACEMODE
	case *ast.StmtFunction:
		return nsr.StmtFunction(n), asttraverser.REPLACEMODE
	case *ast.StmtClassMethod:
		return nsr.StmtClassMethod(n), asttraverser.REPLACEMODE
	case *ast.ExprClosure:
		return nsr.ExprClosure(n), asttraverser.REPLACEMODE
	case *ast.StmtPropertyList:
		return nsr.StmtPropertyList(n), asttraverser.REPLACEMODE
	case *ast.StmtConstList:
		return nsr.StmtConstList(n), asttraverser.REPLACEMODE
	case *ast.ExprStaticCall:
		return nsr.ExprStaticCall(n), asttraverser.REPLACEMODE
	case *ast.ExprStaticPropertyFetch:
		return nsr.ExprStaticPropertyFetch(n), asttraverser.REPLACEMODE
	case *ast.ExprClassConstFetch:
		return nsr.ExprClassConstFetch(n), asttraverser.REPLACEMODE
	case *ast.ExprNew:
		return nsr.ExprNew(n), asttraverser.REPLACEMODE
	case *ast.ExprInstanceOf:
		return nsr.ExprInstanceOf(n), asttraverser.REPLACEMODE
	case *ast.StmtCatch:
		return nsr.StmtCatch(n), asttraverser.REPLACEMODE
	case *ast.ExprFunctionCall:
		return nsr.ExprFunctionCall(n), asttraverser.REPLACEMODE
	case *ast.ExprConstFetch:
		return nsr.ExprConstFetch(n), asttraverser.REPLACEMODE
	case *ast.StmtTraitUse:
		return nsr.StmtTraitUse(n), asttraverser.REPLACEMODE
	}

	return nil, asttraverser.REPLACEMODE
}

func (nsr *NamespaceResolver) LeaveNode(n ast.Vertex) (ast.Vertex, asttraverser.ReturnModeFlag) {
	// do nothing
	return nil, asttraverser.REPLACEMODE
}

func (nsr *NamespaceResolver) StmtNamespace(n *ast.StmtNamespace) ast.Vertex {
	if n.Name == nil {
		nsr.NamespaceCtx = NewNamespace("")
	} else {
		NSParts := n.Name.(*ast.Name).Parts
		nsr.NamespaceCtx = NewNamespace(astutils.ConcatNameParts(NSParts))
	}

	return nil
}

func (nsr *NamespaceResolver) StmtUse(n *ast.StmtUseList) ast.Vertex {
	useType := ""
	if n.Type != nil {
		useType = string(n.Type.(*ast.Identifier).Value)
	}

	for _, nn := range n.Uses {
		nsr.AddAlias(useType, nn, nil)
	}

	nsr.goDeep = false

	return nil
}

func (nsr *NamespaceResolver) StmtGroupUse(n *ast.StmtGroupUseList) ast.Vertex {
	useType := ""
	if n.Type != nil {
		useType = string(n.Type.(*ast.Identifier).Value)
	}

	for _, nn := range n.Uses {
		nsr.AddAlias(useType, nn, n.Prefix.(*ast.Name).Parts)
	}

	nsr.goDeep = false

	return nil
}

func (nsr *NamespaceResolver) StmtClass(n *ast.StmtClass) ast.Vertex {
	if n.Extends != nil {
		nsr.ResolveName(n.Extends, "")
	}

	if n.Implements != nil {
		for _, interfaceName := range n.Implements {
			nsr.ResolveName(interfaceName, "")
		}
	}

	if n.Name != nil {
		nsr.AddNamespacedName(n.Name.(*ast.Identifier), string(n.Name.(*ast.Identifier).Value))
	} else {
		// anonymous class
		anonName := fmt.Sprintf("{anonymousClass}#%d", nsr.anonClassCounter)
		n.Name = &ast.Identifier{Value: []byte(anonName)}
		nsr.AddNamespacedName(n.Name.(*ast.Identifier), anonName)
		nsr.anonClassCounter += 1
	}

	return nil
}

func (nsr *NamespaceResolver) StmtInterface(n *ast.StmtInterface) ast.Vertex {
	if n.Extends != nil {
		for _, interfaceName := range n.Extends {
			nsr.ResolveName(interfaceName, "")
		}
	}

	nsr.AddNamespacedName(n.Name.(*ast.Identifier), string(n.Name.(*ast.Identifier).Value))

	return nil
}

func (nsr *NamespaceResolver) StmtTrait(n *ast.StmtTrait) ast.Vertex {
	nsr.AddNamespacedName(n.Name.(*ast.Identifier), string(n.Name.(*ast.Identifier).Value))

	return nil
}

func (nsr *NamespaceResolver) StmtFunction(n *ast.StmtFunction) ast.Vertex {
	nsr.AddNamespacedName(n.Name.(*ast.Identifier), string(n.Name.(*ast.Identifier).Value))

	for _, parameter := range n.Params {
		nsr.ResolveType(parameter.(*ast.Parameter).Type)
	}

	if n.ReturnType != nil {
		nsr.ResolveType(n.ReturnType)
	}

	return nil
}

func (nsr *NamespaceResolver) StmtClassMethod(n *ast.StmtClassMethod) ast.Vertex {
	for _, parameter := range n.Params {
		nsr.ResolveType(parameter.(*ast.Parameter).Type)
	}

	if n.ReturnType != nil {
		nsr.ResolveType(n.ReturnType)
	}

	return nil
}

func (nsr *NamespaceResolver) ExprClosure(n *ast.ExprClosure) ast.Vertex {
	for _, parameter := range n.Params {
		nsr.ResolveType(parameter.(*ast.Parameter).Type)
	}

	if n.ReturnType != nil {
		nsr.ResolveType(n.ReturnType)
	}

	return nil
}

func (nsr *NamespaceResolver) StmtPropertyList(n *ast.StmtPropertyList) ast.Vertex {
	if n.Type != nil {
		nsr.ResolveType(n.Type)
	}

	return nil
}

func (nsr *NamespaceResolver) StmtConstList(n *ast.StmtConstList) ast.Vertex {
	for _, constant := range n.Consts {
		constant := constant.(*ast.StmtConstant)
		nsr.AddNamespacedName(constant.Name.(*ast.Identifier), string(constant.Name.(*ast.Identifier).Value))
	}

	return nil
}

func (nsr *NamespaceResolver) ExprStaticCall(n *ast.ExprStaticCall) ast.Vertex {
	nsr.ResolveName(n.Class, "")
	return nil
}

func (nsr *NamespaceResolver) ExprStaticPropertyFetch(n *ast.ExprStaticPropertyFetch) ast.Vertex {
	nsr.ResolveName(n.Class, "")

	return nil
}

func (nsr *NamespaceResolver) ExprClassConstFetch(n *ast.ExprClassConstFetch) ast.Vertex {
	nsr.ResolveName(n.Class, "")

	return nil
}

func (nsr *NamespaceResolver) ExprNew(n *ast.ExprNew) ast.Vertex {
	nsr.ResolveName(n.Class, "")

	return nil
}

func (nsr *NamespaceResolver) ExprInstanceOf(n *ast.ExprInstanceOf) ast.Vertex {
	nsr.ResolveName(n.Class, "")

	return nil
}

func (nsr *NamespaceResolver) StmtCatch(n *ast.StmtCatch) ast.Vertex {
	for _, t := range n.Types {
		nsr.ResolveName(t, "")
	}

	return nil
}

func (nsr *NamespaceResolver) ExprFunctionCall(n *ast.ExprFunctionCall) ast.Vertex {
	nsr.ResolveName(n.Function, "function")

	return nil
}

func (nsr *NamespaceResolver) ExprConstFetch(n *ast.ExprConstFetch) ast.Vertex {
	nsr.ResolveName(n.Const, "const")

	return nil
}

func (nsr *NamespaceResolver) StmtTraitUse(n *ast.StmtTraitUse) ast.Vertex {
	for _, t := range n.Traits {
		nsr.ResolveName(t, "")
	}

	for _, a := range n.Adaptations {
		switch aa := a.(type) {
		case *ast.StmtTraitUsePrecedence:
			refTrait := aa.Trait
			if refTrait != nil {
				nsr.ResolveName(refTrait, "")
			}
			for _, insteadOf := range aa.Insteadof {
				nsr.ResolveName(insteadOf, "")
			}

		case *ast.StmtTraitUseAlias:
			refTrait := aa.Trait
			if refTrait != nil {
				nsr.ResolveName(refTrait, "")
			}
		}
	}

	return nil
}

// AddAlias adds a new alias
func (nsr *NamespaceResolver) AddAlias(useType string, nodename ast.Vertex, prefix []ast.Vertex) {
	switch use := nodename.(type) {
	case *ast.StmtUse:
		if use.Type != nil {
			useType = string(use.Type.(*ast.Identifier).Value)
		}

		useNameParts := use.Use.(*ast.Name).Parts
		var alias string
		if use.Alias == nil {
			alias = string(useNameParts[len(useNameParts)-1].(*ast.NamePart).Value)
		} else {
			alias = string(use.Alias.(*ast.Identifier).Value)
		}

		nsr.NamespaceCtx.AddAlias(useType, astutils.ConcatNameParts(prefix, useNameParts), alias)
	}
}

// AddNamespacedName adds namespaced name by node
func (nsr *NamespaceResolver) AddNamespacedName(nn *ast.Identifier, nodeName string) {
	var resolvedName string
	if nsr.NamespaceCtx.Namespace == "" {
		resolvedName = nodeName
	} else {
		resolvedName = nsr.NamespaceCtx.Namespace + "\\" + nodeName
	}

	nn.Value = []byte(resolvedName)
}

// ResolveName adds a resolved fully qualified name by node
func (nsr *NamespaceResolver) ResolveName(nameNode ast.Vertex, aliasType string) {
	resolved, err := nsr.NamespaceCtx.ResolveName(nameNode, aliasType)
	if err == nil {
		switch nameNode := nameNode.(type) {
		case *ast.Name:
			nameNode.Parts = astutils.CreateNameParts(resolved, nameNode.Position)
		case *ast.NameFullyQualified:
			nameNode.Parts = astutils.CreateNameParts(resolved, nameNode.Position)
		case *ast.NameRelative:
			nameNode.Parts = astutils.CreateNameParts(resolved, nameNode.Position)
		}
	}
}

// ResolveType adds a resolved fully qualified type name
func (nsr *NamespaceResolver) ResolveType(n ast.Vertex) {
	switch nodetype := n.(type) {
	case *ast.Nullable:
		nsr.ResolveType(nodetype.Expr)
	case *ast.Name:
		nsr.ResolveName(n, "")
	case *ast.NameRelative:
		nsr.ResolveName(n, "")
	case *ast.NameFullyQualified:
		nsr.ResolveName(n, "")
	}
}
