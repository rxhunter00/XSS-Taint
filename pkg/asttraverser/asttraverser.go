package asttraverser

// implement https://github.com/VKCOM/php-parser/blob/master/pkg/ast/ast.go

import (
	"log"
	"reflect"

	"github.com/VKCOM/php-parser/pkg/ast"
)

type ReturnModeFlag int

const (
	REPLACEMODE ReturnModeFlag = iota
	INSERTMODE
)

type InsertedNode struct {
	Idx  int
	Node ast.Vertex
}

type NodeTraverser interface {
	EnterNode(n ast.Vertex) (ast.Vertex, ReturnModeFlag)
	LeaveNode(n ast.Vertex) (ast.Vertex, ReturnModeFlag)
}

type Traverser struct {
	NodeTraversers []NodeTraverser
}

func NewTraverser() *Traverser {
	return &Traverser{NodeTraversers: make([]NodeTraverser, 0)}
}
func (t *Traverser) AddNodeTraverser(nt ...NodeTraverser) {
	t.NodeTraversers = append(t.NodeTraversers, nt...)
}
func (t *Traverser) Traverse(node ast.Vertex) ast.Vertex {
	if node == nil {
		return nil
	}

	for _, nt := range t.NodeTraversers {
		replacmentNode, returnTypeMode := nt.EnterNode(node)
		if replacmentNode != nil {
			if returnTypeMode == REPLACEMODE && isReplacementReasonable(node, replacmentNode) {
				return replacmentNode
			} else {
				log.Fatalf("Invalid node replacement '%v' - '%v'", reflect.TypeOf(node), reflect.TypeOf(replacmentNode))
			}
		}
	}
	node.Accept(t)
	for _, nt := range t.NodeTraversers {
		replacementNode, returnTypeMode := nt.LeaveNode(node)
		if replacementNode != nil {
			if returnTypeMode == REPLACEMODE && isReplacementReasonable(node, replacementNode) {
				return replacementNode
			} else {
				log.Fatalf("Invalid node replacement '%v' - '%v'", reflect.TypeOf(node), reflect.TypeOf(replacementNode))
			}
		}
	}

	return nil
}

func (t *Traverser) TraverseNodes(nodes []ast.Vertex) []ast.Vertex {
	var insertedNodes []InsertedNode = make([]InsertedNode, 0)

	for i, n := range nodes {
		// Enter Node
		for _, nt := range t.NodeTraversers {
			returnedNode, nType := nt.EnterNode(n)
			if returnedNode != nil {
				if nType == REPLACEMODE {
					if isReplacementReasonable(n, returnedNode) {
						nodes[i] = returnedNode
					} else {
						log.Fatalf("TraverseNodes: Invalid node replacement '%v' - '%v'", reflect.TypeOf(n), reflect.TypeOf(returnedNode))
					}
				} else {
					log.Fatalf("TraverseNodes: Invalid Replacement Mode")
				}
			}
		}

		n.Accept(t)

		// Leave Node
		for _, nt := range t.NodeTraversers {
			returnedNode, nType := nt.LeaveNode(n)
			if returnedNode != nil {
				if nType == REPLACEMODE {
					if isReplacementReasonable(n, returnedNode) {
						nodes[i] = returnedNode
					} else {
						log.Fatalf("TraverseNodes: Invalid node replacement '%v' - '%v'", reflect.TypeOf(n), reflect.TypeOf(returnedNode))
					}
				} else if nType == INSERTMODE {
					insertedNodes = append(insertedNodes, InsertedNode{Idx: i, Node: returnedNode})
				} else {
					log.Fatalf("TraverseNodes: Invalid Replacement Mode")
				}
			}
		}
	}

	// inserting nodes
	for i := len(insertedNodes) - 1; i >= 0; i-- {
		idx := insertedNodes[i].Idx
		nd := insertedNodes[i].Node

		// if there is other node in the right, append it
		if idx < len(nodes)-1 {
			left := nodes[:idx+1]
			right := append([]ast.Vertex{nd}, nodes[idx+1:]...)
			nodes = append(left, right...)
		} else {
			nodes = append(nodes, nd)
		}
	}

	return nodes
}

/**
Traverse Vertex and Vertex List
**/

func (t *Traverser) Root(n *ast.Root) {
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *Traverser) Nullable(n *ast.Nullable) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) Parameter(n *ast.Parameter) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)

	n.Modifiers = t.TraverseNodes(n.Modifiers)

	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.DefaultValue); replacedNode != nil {
		n.DefaultValue = replacedNode
	}
}

func (t *Traverser) Identifier(n *ast.Identifier) {

}

func (t *Traverser) Argument(n *ast.Argument) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) MatchArm(n *ast.MatchArm) {
	n.Exprs = t.TraverseNodes(n.Exprs)

	if replacedNode := t.Traverse(n.ReturnExpr); replacedNode != nil {
		n.ReturnExpr = replacedNode
	}
}

func (t *Traverser) Union(n *ast.Union) {
	n.Types = t.TraverseNodes(n.Types)
}

func (t *Traverser) Intersection(n *ast.Intersection) {
	n.Types = t.TraverseNodes(n.Types)
}

func (t *Traverser) Attribute(n *ast.Attribute) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
}

func (t *Traverser) AttributeGroup(n *ast.AttributeGroup) {
	n.Attrs = t.TraverseNodes(n.Attrs)
}

func (t *Traverser) StmtBreak(n *ast.StmtBreak) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) StmtCase(n *ast.StmtCase) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}

	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtCatch(n *ast.StmtCatch) {
	n.Types = t.TraverseNodes(n.Types)

	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}

	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtEnum(n *ast.StmtEnum) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}

	n.Implements = t.TraverseNodes(n.Implements)
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *Traverser) EnumCase(n *ast.EnumCase) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) StmtClass(n *ast.StmtClass) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)
	n.Modifiers = t.TraverseNodes(n.Modifiers)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
	n.Implements = t.TraverseNodes(n.Implements)

	if replacedNode := t.Traverse(n.Extends); replacedNode != nil {
		n.Extends = replacedNode
	}

	n.Implements = t.TraverseNodes(n.Implements)
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtClassConstList(n *ast.StmtClassConstList) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)
	n.Modifiers = t.TraverseNodes(n.Modifiers)
	n.Consts = t.TraverseNodes(n.Consts)
}

func (t *Traverser) StmtClassMethod(n *ast.StmtClassMethod) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)
	n.Modifiers = t.TraverseNodes(n.Modifiers)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	n.Params = t.TraverseNodes(n.Params)

	if replacedNode := t.Traverse(n.ReturnType); replacedNode != nil {
		n.ReturnType = replacedNode
	}
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *Traverser) StmtConstList(n *ast.StmtConstList) {
	n.Consts = t.TraverseNodes(n.Consts)
}

func (t *Traverser) StmtConstant(n *ast.StmtConstant) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) StmtContinue(n *ast.StmtContinue) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) StmtDeclare(n *ast.StmtDeclare) {
	n.Consts = t.TraverseNodes(n.Consts)

	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *Traverser) StmtDefault(n *ast.StmtDefault) {
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtDo(n *ast.StmtDo) {
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}
}

func (t *Traverser) StmtEcho(n *ast.StmtEcho) {
	n.Exprs = t.TraverseNodes(n.Exprs)
}

func (t *Traverser) StmtElse(n *ast.StmtElse) {
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *Traverser) StmtElseIf(n *ast.StmtElseIf) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *Traverser) StmtExpression(n *ast.StmtExpression) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) StmtFinally(n *ast.StmtFinally) {
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtFor(n *ast.StmtFor) {
	n.Init = t.TraverseNodes(n.Init)
	n.Cond = t.TraverseNodes(n.Cond)
	n.Loop = t.TraverseNodes(n.Loop)

	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *Traverser) StmtForeach(n *ast.StmtForeach) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
	if replacedNode := t.Traverse(n.Key); replacedNode != nil {
		n.Key = replacedNode
	}
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *Traverser) StmtFunction(n *ast.StmtFunction) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	n.Params = t.TraverseNodes(n.Params)

	if replacedNode := t.Traverse(n.ReturnType); replacedNode != nil {
		n.ReturnType = replacedNode
	}

	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtGlobal(n *ast.StmtGlobal) {
	n.Vars = t.TraverseNodes(n.Vars)
}

func (t *Traverser) StmtGoto(n *ast.StmtGoto) {
	if replacedNode := t.Traverse(n.Label); replacedNode != nil {
		n.Label = replacedNode
	}
}

func (t *Traverser) StmtHaltCompiler(n *ast.StmtHaltCompiler) {

}

func (t *Traverser) StmtIf(n *ast.StmtIf) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}

	n.ElseIf = t.TraverseNodes(n.ElseIf)

	if replacedNode := t.Traverse(n.Else); replacedNode != nil {
		n.Else = replacedNode
	}
}

func (t *Traverser) StmtInlineHtml(n *ast.StmtInlineHtml) {

}

func (t *Traverser) StmtInterface(n *ast.StmtInterface) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	n.Extends = t.TraverseNodes(n.Extends)
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtLabel(n *ast.StmtLabel) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
}

func (t *Traverser) StmtNamespace(n *ast.StmtNamespace) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtNop(n *ast.StmtNop) {

}

func (t *Traverser) StmtProperty(n *ast.StmtProperty) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) StmtPropertyList(n *ast.StmtPropertyList) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)
	n.Modifiers = t.TraverseNodes(n.Modifiers)

	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}

	n.Props = t.TraverseNodes(n.Props)
}

func (t *Traverser) StmtReturn(n *ast.StmtReturn) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) StmtStatic(n *ast.StmtStatic) {
	n.Vars = t.TraverseNodes(n.Vars)
}

func (t *Traverser) StmtStaticVar(n *ast.StmtStaticVar) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) StmtStmtList(n *ast.StmtStmtList) {
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtSwitch(n *ast.StmtSwitch) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}

	n.Cases = t.TraverseNodes(n.Cases)
}

func (t *Traverser) StmtThrow(n *ast.StmtThrow) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) StmtTrait(n *ast.StmtTrait) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtTraitUse(n *ast.StmtTraitUse) {
	n.Traits = t.TraverseNodes(n.Traits)
	n.Adaptations = t.TraverseNodes(n.Adaptations)
}

func (t *Traverser) StmtTraitUseAlias(n *ast.StmtTraitUseAlias) {
	if replacedNode := t.Traverse(n.Trait); replacedNode != nil {
		n.Trait = replacedNode
	}
	if replacedNode := t.Traverse(n.Method); replacedNode != nil {
		n.Method = replacedNode
	}
	if replacedNode := t.Traverse(n.Modifier); replacedNode != nil {
		n.Modifier = replacedNode
	}
	if replacedNode := t.Traverse(n.Alias); replacedNode != nil {
		n.Alias = replacedNode
	}
}

func (t *Traverser) StmtTraitUsePrecedence(n *ast.StmtTraitUsePrecedence) {
	if replacedNode := t.Traverse(n.Trait); replacedNode != nil {
		n.Trait = replacedNode
	}
	if replacedNode := t.Traverse(n.Method); replacedNode != nil {
		n.Method = replacedNode
	}
	n.Insteadof = t.TraverseNodes(n.Insteadof)
}

func (t *Traverser) StmtTry(n *ast.StmtTry) {
	n.Stmts = t.TraverseNodes(n.Stmts)
	n.Catches = t.TraverseNodes(n.Catches)

	if replacedNode := t.Traverse(n.Finally); replacedNode != nil {
		n.Finally = replacedNode
	}
}

func (t *Traverser) StmtUnset(n *ast.StmtUnset) {
	n.Vars = t.TraverseNodes(n.Vars)
}

func (t *Traverser) StmtUse(n *ast.StmtUseList) {
	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}
	n.Uses = t.TraverseNodes(n.Uses)
}

func (t *Traverser) StmtGroupUse(n *ast.StmtGroupUseList) {
	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}
	if replacedNode := t.Traverse(n.Prefix); replacedNode != nil {
		n.Prefix = replacedNode
	}

	n.Uses = t.TraverseNodes(n.Uses)
}

func (t *Traverser) StmtUseDeclaration(n *ast.StmtUse) {
	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}
	if replacedNode := t.Traverse(n.Use); replacedNode != nil {
		n.Use = replacedNode
	}
	if replacedNode := t.Traverse(n.Alias); replacedNode != nil {
		n.Alias = replacedNode
	}
}

func (t *Traverser) StmtWhile(n *ast.StmtWhile) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *Traverser) ExprArray(n *ast.ExprArray) {
	n.Items = t.TraverseNodes(n.Items)
}

func (t *Traverser) ExprArrayDimFetch(n *ast.ExprArrayDimFetch) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Dim); replacedNode != nil {
		n.Dim = replacedNode
	}
}

func (t *Traverser) ExprArrayItem(n *ast.ExprArrayItem) {
	if replacedNode := t.Traverse(n.Key); replacedNode != nil {
		n.Key = replacedNode
	}
	if replacedNode := t.Traverse(n.Val); replacedNode != nil {
		n.Val = replacedNode
	}
}

func (t *Traverser) ExprArrowFunction(n *ast.ExprArrowFunction) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)
	n.Params = t.TraverseNodes(n.Params)

	if replacedNode := t.Traverse(n.ReturnType); replacedNode != nil {
		n.ReturnType = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprBitwiseNot(n *ast.ExprBitwiseNot) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprBooleanNot(n *ast.ExprBooleanNot) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprBrackets(n *ast.ExprBrackets) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprClassConstFetch(n *ast.ExprClassConstFetch) {
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}
	if replacedNode := t.Traverse(n.Const); replacedNode != nil {
		n.Const = replacedNode
	}
}

func (t *Traverser) ExprClone(n *ast.ExprClone) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprClosure(n *ast.ExprClosure) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)
	n.Params = t.TraverseNodes(n.Params)
	n.Uses = t.TraverseNodes(n.Uses)

	if replacedNode := t.Traverse(n.ReturnType); replacedNode != nil {
		n.ReturnType = replacedNode
	}

	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *Traverser) ExprClosureUse(n *ast.ExprClosureUse) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *Traverser) ExprConstFetch(n *ast.ExprConstFetch) {
	if replacedNode := t.Traverse(n.Const); replacedNode != nil {
		n.Const = replacedNode
	}
}

func (t *Traverser) ExprEmpty(n *ast.ExprEmpty) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprErrorSuppress(n *ast.ExprErrorSuppress) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprEval(n *ast.ExprEval) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprExit(n *ast.ExprExit) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprFunctionCall(n *ast.ExprFunctionCall) {
	if replacedNode := t.Traverse(n.Function); replacedNode != nil {
		n.Function = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
}

func (t *Traverser) ExprInclude(n *ast.ExprInclude) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprIncludeOnce(n *ast.ExprIncludeOnce) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprInstanceOf(n *ast.ExprInstanceOf) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}
}

func (t *Traverser) ExprIsset(n *ast.ExprIsset) {
	n.Vars = t.TraverseNodes(n.Vars)
}

func (t *Traverser) ExprList(n *ast.ExprList) {
	n.Items = t.TraverseNodes(n.Items)
}

func (t *Traverser) ExprMethodCall(n *ast.ExprMethodCall) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Method); replacedNode != nil {
		n.Method = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
}

func (t *Traverser) ExprNullsafeMethodCall(n *ast.ExprNullsafeMethodCall) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Method); replacedNode != nil {
		n.Method = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
}

func (t *Traverser) ExprNew(n *ast.ExprNew) {
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
}

func (t *Traverser) ExprPostDec(n *ast.ExprPostDec) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *Traverser) ExprPostInc(n *ast.ExprPostInc) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *Traverser) ExprPreDec(n *ast.ExprPreDec) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *Traverser) ExprPreInc(n *ast.ExprPreInc) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *Traverser) ExprPrint(n *ast.ExprPrint) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprPropertyFetch(n *ast.ExprPropertyFetch) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Prop); replacedNode != nil {
		n.Prop = replacedNode
	}
}

func (t *Traverser) ExprNullsafePropertyFetch(n *ast.ExprNullsafePropertyFetch) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Prop); replacedNode != nil {
		n.Prop = replacedNode
	}
}

func (t *Traverser) ExprRequire(n *ast.ExprRequire) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprRequireOnce(n *ast.ExprRequireOnce) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprShellExec(n *ast.ExprShellExec) {
	n.Parts = t.TraverseNodes(n.Parts)
}

func (t *Traverser) ExprStaticCall(n *ast.ExprStaticCall) {
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}
	if replacedNode := t.Traverse(n.Call); replacedNode != nil {
		n.Call = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
}

func (t *Traverser) ExprStaticPropertyFetch(n *ast.ExprStaticPropertyFetch) {
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}
	if replacedNode := t.Traverse(n.Prop); replacedNode != nil {
		n.Prop = replacedNode
	}
}

func (t *Traverser) ExprTernary(n *ast.ExprTernary) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}
	if replacedNode := t.Traverse(n.IfTrue); replacedNode != nil {
		n.IfTrue = replacedNode
	}
	if replacedNode := t.Traverse(n.IfFalse); replacedNode != nil {
		n.IfFalse = replacedNode
	}
}

func (t *Traverser) ExprUnaryMinus(n *ast.ExprUnaryMinus) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprUnaryPlus(n *ast.ExprUnaryPlus) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprVariable(n *ast.ExprVariable) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
}

func (t *Traverser) ExprYield(n *ast.ExprYield) {
	if replacedNode := t.Traverse(n.Key); replacedNode != nil {
		n.Key = replacedNode
	}
	if replacedNode := t.Traverse(n.Val); replacedNode != nil {
		n.Val = replacedNode
	}
}

func (t *Traverser) ExprYieldFrom(n *ast.ExprYieldFrom) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssign(n *ast.ExprAssign) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignReference(n *ast.ExprAssignReference) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignBitwiseAnd(n *ast.ExprAssignBitwiseAnd) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignBitwiseOr(n *ast.ExprAssignBitwiseOr) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignBitwiseXor(n *ast.ExprAssignBitwiseXor) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignCoalesce(n *ast.ExprAssignCoalesce) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignConcat(n *ast.ExprAssignConcat) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignDiv(n *ast.ExprAssignDiv) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignMinus(n *ast.ExprAssignMinus) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignMod(n *ast.ExprAssignMod) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignMul(n *ast.ExprAssignMul) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignPlus(n *ast.ExprAssignPlus) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignPow(n *ast.ExprAssignPow) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignShiftLeft(n *ast.ExprAssignShiftLeft) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignShiftRight(n *ast.ExprAssignShiftRight) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprBinaryBitwiseAnd(n *ast.ExprBinaryBitwiseAnd) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryBitwiseOr(n *ast.ExprBinaryBitwiseOr) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryBitwiseXor(n *ast.ExprBinaryBitwiseXor) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryBooleanAnd(n *ast.ExprBinaryBooleanAnd) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryBooleanOr(n *ast.ExprBinaryBooleanOr) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryCoalesce(n *ast.ExprBinaryCoalesce) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryConcat(n *ast.ExprBinaryConcat) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryDiv(n *ast.ExprBinaryDiv) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryEqual(n *ast.ExprBinaryEqual) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryGreater(n *ast.ExprBinaryGreater) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryGreaterOrEqual(n *ast.ExprBinaryGreaterOrEqual) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryIdentical(n *ast.ExprBinaryIdentical) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryLogicalAnd(n *ast.ExprBinaryLogicalAnd) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryLogicalOr(n *ast.ExprBinaryLogicalOr) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryLogicalXor(n *ast.ExprBinaryLogicalXor) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryMinus(n *ast.ExprBinaryMinus) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryMod(n *ast.ExprBinaryMod) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryMul(n *ast.ExprBinaryMul) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryNotEqual(n *ast.ExprBinaryNotEqual) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryNotIdentical(n *ast.ExprBinaryNotIdentical) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryPlus(n *ast.ExprBinaryPlus) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryPow(n *ast.ExprBinaryPow) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryShiftLeft(n *ast.ExprBinaryShiftLeft) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryShiftRight(n *ast.ExprBinaryShiftRight) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinarySmaller(n *ast.ExprBinarySmaller) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinarySmallerOrEqual(n *ast.ExprBinarySmallerOrEqual) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinarySpaceship(n *ast.ExprBinarySpaceship) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprCastArray(n *ast.ExprCastArray) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprCastBool(n *ast.ExprCastBool) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprCastDouble(n *ast.ExprCastDouble) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprCastInt(n *ast.ExprCastInt) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprCastObject(n *ast.ExprCastObject) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprCastString(n *ast.ExprCastString) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprCastUnset(n *ast.ExprCastUnset) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprMatch(n *ast.ExprMatch) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
	n.Arms = t.TraverseNodes(n.Arms)
}

func (t *Traverser) ExprThrow(n *ast.ExprThrow) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ScalarDnumber(n *ast.ScalarDnumber) {

}

func (t *Traverser) ScalarEncapsed(n *ast.ScalarEncapsed) {
	n.Parts = t.TraverseNodes(n.Parts)
}

func (t *Traverser) ScalarEncapsedStringPart(n *ast.ScalarEncapsedStringPart) {

}

func (t *Traverser) ScalarEncapsedStringVar(n *ast.ScalarEncapsedStringVar) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Dim); replacedNode != nil {
		n.Dim = replacedNode
	}
}

func (t *Traverser) ScalarEncapsedStringBrackets(n *ast.ScalarEncapsedStringBrackets) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *Traverser) ScalarHeredoc(n *ast.ScalarHeredoc) {
	n.Parts = t.TraverseNodes(n.Parts)
}

func (t *Traverser) ScalarLnumber(n *ast.ScalarLnumber) {

}

func (t *Traverser) ScalarMagicConstant(n *ast.ScalarMagicConstant) {

}

func (t *Traverser) ScalarString(n *ast.ScalarString) {

}

func (t *Traverser) NameName(n *ast.Name) {
	n.Parts = t.TraverseNodes(n.Parts)
}

func (t *Traverser) NameFullyQualified(n *ast.NameFullyQualified) {
	n.Parts = t.TraverseNodes(n.Parts)
}

func (t *Traverser) NameRelative(n *ast.NameRelative) {
	n.Parts = t.TraverseNodes(n.Parts)
}

func (t *Traverser) NameNamePart(n *ast.NamePart) {

}

func isReplacementReasonable(v1 ast.Vertex, v2 ast.Vertex) bool {
	isV1Stmt := isStmt(v1)
	isV2Stmt := isStmt(v2)

	return isV1Stmt == isV2Stmt
}

func isStmt(v ast.Vertex) bool {
	switch v.(type) {
	case *ast.StmtBreak, *ast.StmtCase, *ast.StmtCatch, *ast.StmtEnum, *ast.EnumCase, *ast.StmtClass,
		*ast.StmtClassConstList, *ast.StmtClassMethod, *ast.StmtConstList, *ast.StmtConstant,
		*ast.StmtContinue, *ast.StmtDeclare, *ast.StmtDefault, *ast.StmtDo, *ast.StmtEcho, *ast.StmtElse,
		*ast.StmtElseIf, *ast.StmtExpression, *ast.StmtFinally, *ast.StmtFor, *ast.StmtForeach,
		*ast.StmtFunction, *ast.StmtGlobal, *ast.StmtGoto, *ast.StmtHaltCompiler, *ast.StmtIf,
		*ast.StmtInlineHtml, *ast.StmtInterface, *ast.StmtLabel, *ast.StmtNamespace, *ast.StmtNop,
		*ast.StmtProperty, *ast.StmtPropertyList, *ast.StmtReturn, *ast.StmtStatic, *ast.StmtStaticVar,
		*ast.StmtStmtList, *ast.StmtSwitch, *ast.StmtThrow, *ast.StmtTrait, *ast.StmtTraitUse,
		*ast.StmtTraitUseAlias, *ast.StmtTraitUsePrecedence, *ast.StmtTry, *ast.StmtUnset, *ast.StmtUse,
		*ast.StmtGroupUseList, *ast.StmtUseList, *ast.StmtWhile:
		return true
	}
	return false
}
