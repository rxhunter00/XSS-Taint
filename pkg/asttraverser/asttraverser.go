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

// Implements Visitor
type ASTTraverser struct {
	NodeTraversers []NodeTraverser
}

func NewTraverser() *ASTTraverser {
	return &ASTTraverser{NodeTraversers: make([]NodeTraverser, 0)}
}
func (t *ASTTraverser) AddNodeTraverser(nt ...NodeTraverser) {
	t.NodeTraversers = append(t.NodeTraversers, nt...)
}
func (t *ASTTraverser) Traverse(node ast.Vertex) ast.Vertex {
	if node == nil {
		return nil
	}

	for _, nt := range t.NodeTraversers {
		replacmentNode, returnTypeMode := nt.EnterNode(node)
		if replacmentNode != nil {
			if returnTypeMode == REPLACEMODE && isReplacementReasonable(node, replacmentNode) {
				return replacmentNode
			} else {
				log.Fatalf("Traverse:Enter error replacement of  '%v' - '%v'", reflect.TypeOf(node), reflect.TypeOf(replacmentNode))
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
				// log.Fatalf("Traverse:Leave error replacement of '%v' - '%v'", reflect.TypeOf(node), reflect.TypeOf(replacementNode))
			}
		}
	}

	return nil
}

func (t *ASTTraverser) TraverseNodes(nodes []ast.Vertex) []ast.Vertex {
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
						log.Fatalf("TraverseNodes: Invalid node replacement Enter'%v' - '%v'", reflect.TypeOf(n), reflect.TypeOf(returnedNode))
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
						log.Fatalf("TraverseNodes: Invalid node replacement on Leave '%v' - '%v'", reflect.TypeOf(n), reflect.TypeOf(returnedNode))
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

func (t *ASTTraverser) Root(n *ast.Root) {
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *ASTTraverser) Nullable(n *ast.Nullable) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) Parameter(n *ast.Parameter) {
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

func (t *ASTTraverser) Identifier(n *ast.Identifier) {

}

func (t *ASTTraverser) Argument(n *ast.Argument) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) MatchArm(n *ast.MatchArm) {
	n.Exprs = t.TraverseNodes(n.Exprs)

	if replacedNode := t.Traverse(n.ReturnExpr); replacedNode != nil {
		n.ReturnExpr = replacedNode
	}
}

func (t *ASTTraverser) Union(n *ast.Union) {
	n.Types = t.TraverseNodes(n.Types)
}

func (t *ASTTraverser) Intersection(n *ast.Intersection) {
	n.Types = t.TraverseNodes(n.Types)
}

func (t *ASTTraverser) Attribute(n *ast.Attribute) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
}

func (t *ASTTraverser) AttributeGroup(n *ast.AttributeGroup) {
	n.Attrs = t.TraverseNodes(n.Attrs)
}

func (t *ASTTraverser) StmtBreak(n *ast.StmtBreak) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) StmtCase(n *ast.StmtCase) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}

	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *ASTTraverser) StmtCatch(n *ast.StmtCatch) {
	n.Types = t.TraverseNodes(n.Types)

	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}

	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *ASTTraverser) StmtEnum(n *ast.StmtEnum) {
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

func (t *ASTTraverser) EnumCase(n *ast.EnumCase) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) StmtClass(n *ast.StmtClass) {
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

func (t *ASTTraverser) StmtClassConstList(n *ast.StmtClassConstList) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)
	n.Modifiers = t.TraverseNodes(n.Modifiers)
	n.Consts = t.TraverseNodes(n.Consts)
}

func (t *ASTTraverser) StmtClassMethod(n *ast.StmtClassMethod) {
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

func (t *ASTTraverser) StmtConstList(n *ast.StmtConstList) {
	n.Consts = t.TraverseNodes(n.Consts)
}

func (t *ASTTraverser) StmtConstant(n *ast.StmtConstant) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) StmtContinue(n *ast.StmtContinue) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) StmtDeclare(n *ast.StmtDeclare) {
	n.Consts = t.TraverseNodes(n.Consts)

	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *ASTTraverser) StmtDefault(n *ast.StmtDefault) {
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *ASTTraverser) StmtDo(n *ast.StmtDo) {
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}
}

func (t *ASTTraverser) StmtEcho(n *ast.StmtEcho) {
	n.Exprs = t.TraverseNodes(n.Exprs)
}

func (t *ASTTraverser) StmtElse(n *ast.StmtElse) {
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *ASTTraverser) StmtElseIf(n *ast.StmtElseIf) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *ASTTraverser) StmtExpression(n *ast.StmtExpression) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) StmtFinally(n *ast.StmtFinally) {
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *ASTTraverser) StmtFor(n *ast.StmtFor) {
	n.Init = t.TraverseNodes(n.Init)
	n.Cond = t.TraverseNodes(n.Cond)
	n.Loop = t.TraverseNodes(n.Loop)

	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *ASTTraverser) StmtForeach(n *ast.StmtForeach) {
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

func (t *ASTTraverser) StmtFunction(n *ast.StmtFunction) {
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

func (t *ASTTraverser) StmtGlobal(n *ast.StmtGlobal) {
	n.Vars = t.TraverseNodes(n.Vars)
}

func (t *ASTTraverser) StmtGoto(n *ast.StmtGoto) {
	if replacedNode := t.Traverse(n.Label); replacedNode != nil {
		n.Label = replacedNode
	}
}

func (t *ASTTraverser) StmtHaltCompiler(n *ast.StmtHaltCompiler) {

}

func (t *ASTTraverser) StmtIf(n *ast.StmtIf) {
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

func (t *ASTTraverser) StmtInlineHtml(n *ast.StmtInlineHtml) {

}

func (t *ASTTraverser) StmtInterface(n *ast.StmtInterface) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	n.Extends = t.TraverseNodes(n.Extends)
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *ASTTraverser) StmtLabel(n *ast.StmtLabel) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
}

func (t *ASTTraverser) StmtNamespace(n *ast.StmtNamespace) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *ASTTraverser) StmtNop(n *ast.StmtNop) {

}

func (t *ASTTraverser) StmtProperty(n *ast.StmtProperty) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) StmtPropertyList(n *ast.StmtPropertyList) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)
	n.Modifiers = t.TraverseNodes(n.Modifiers)

	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}

	n.Props = t.TraverseNodes(n.Props)
}

func (t *ASTTraverser) StmtReturn(n *ast.StmtReturn) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) StmtStatic(n *ast.StmtStatic) {
	n.Vars = t.TraverseNodes(n.Vars)
}

func (t *ASTTraverser) StmtStaticVar(n *ast.StmtStaticVar) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) StmtStmtList(n *ast.StmtStmtList) {
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *ASTTraverser) StmtSwitch(n *ast.StmtSwitch) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}

	n.Cases = t.TraverseNodes(n.Cases)
}

func (t *ASTTraverser) StmtThrow(n *ast.StmtThrow) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) StmtTrait(n *ast.StmtTrait) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *ASTTraverser) StmtTraitUse(n *ast.StmtTraitUse) {
	n.Traits = t.TraverseNodes(n.Traits)
	n.Adaptations = t.TraverseNodes(n.Adaptations)
}

func (t *ASTTraverser) StmtTraitUseAlias(n *ast.StmtTraitUseAlias) {
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

func (t *ASTTraverser) StmtTraitUsePrecedence(n *ast.StmtTraitUsePrecedence) {
	if replacedNode := t.Traverse(n.Trait); replacedNode != nil {
		n.Trait = replacedNode
	}
	if replacedNode := t.Traverse(n.Method); replacedNode != nil {
		n.Method = replacedNode
	}
	n.Insteadof = t.TraverseNodes(n.Insteadof)
}

func (t *ASTTraverser) StmtTry(n *ast.StmtTry) {
	n.Stmts = t.TraverseNodes(n.Stmts)
	n.Catches = t.TraverseNodes(n.Catches)

	if replacedNode := t.Traverse(n.Finally); replacedNode != nil {
		n.Finally = replacedNode
	}
}

func (t *ASTTraverser) StmtUnset(n *ast.StmtUnset) {
	n.Vars = t.TraverseNodes(n.Vars)
}

func (t *ASTTraverser) StmtUse(n *ast.StmtUseList) {
	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}
	n.Uses = t.TraverseNodes(n.Uses)
}

func (t *ASTTraverser) StmtGroupUse(n *ast.StmtGroupUseList) {
	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}
	if replacedNode := t.Traverse(n.Prefix); replacedNode != nil {
		n.Prefix = replacedNode
	}

	n.Uses = t.TraverseNodes(n.Uses)
}

func (t *ASTTraverser) StmtUseDeclaration(n *ast.StmtUse) {
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

func (t *ASTTraverser) StmtWhile(n *ast.StmtWhile) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *ASTTraverser) ExprArray(n *ast.ExprArray) {
	n.Items = t.TraverseNodes(n.Items)
}

func (t *ASTTraverser) ExprArrayDimFetch(n *ast.ExprArrayDimFetch) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Dim); replacedNode != nil {
		n.Dim = replacedNode
	}
}

func (t *ASTTraverser) ExprArrayItem(n *ast.ExprArrayItem) {
	if replacedNode := t.Traverse(n.Key); replacedNode != nil {
		n.Key = replacedNode
	}
	if replacedNode := t.Traverse(n.Val); replacedNode != nil {
		n.Val = replacedNode
	}
}

func (t *ASTTraverser) ExprArrowFunction(n *ast.ExprArrowFunction) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)
	n.Params = t.TraverseNodes(n.Params)

	if replacedNode := t.Traverse(n.ReturnType); replacedNode != nil {
		n.ReturnType = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprBitwiseNot(n *ast.ExprBitwiseNot) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprBooleanNot(n *ast.ExprBooleanNot) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprBrackets(n *ast.ExprBrackets) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprClassConstFetch(n *ast.ExprClassConstFetch) {
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}
	if replacedNode := t.Traverse(n.Const); replacedNode != nil {
		n.Const = replacedNode
	}
}

func (t *ASTTraverser) ExprClone(n *ast.ExprClone) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprClosure(n *ast.ExprClosure) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)
	n.Params = t.TraverseNodes(n.Params)
	n.Uses = t.TraverseNodes(n.Uses)

	if replacedNode := t.Traverse(n.ReturnType); replacedNode != nil {
		n.ReturnType = replacedNode
	}

	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *ASTTraverser) ExprClosureUse(n *ast.ExprClosureUse) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *ASTTraverser) ExprConstFetch(n *ast.ExprConstFetch) {
	if replacedNode := t.Traverse(n.Const); replacedNode != nil {
		n.Const = replacedNode
	}
}

func (t *ASTTraverser) ExprEmpty(n *ast.ExprEmpty) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprErrorSuppress(n *ast.ExprErrorSuppress) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprEval(n *ast.ExprEval) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprExit(n *ast.ExprExit) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprFunctionCall(n *ast.ExprFunctionCall) {
	if replacedNode := t.Traverse(n.Function); replacedNode != nil {
		n.Function = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
}

func (t *ASTTraverser) ExprInclude(n *ast.ExprInclude) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprIncludeOnce(n *ast.ExprIncludeOnce) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprInstanceOf(n *ast.ExprInstanceOf) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}
}

func (t *ASTTraverser) ExprIsset(n *ast.ExprIsset) {
	n.Vars = t.TraverseNodes(n.Vars)
}

func (t *ASTTraverser) ExprList(n *ast.ExprList) {
	n.Items = t.TraverseNodes(n.Items)
}

func (t *ASTTraverser) ExprMethodCall(n *ast.ExprMethodCall) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Method); replacedNode != nil {
		n.Method = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
}

func (t *ASTTraverser) ExprNullsafeMethodCall(n *ast.ExprNullsafeMethodCall) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Method); replacedNode != nil {
		n.Method = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
}

func (t *ASTTraverser) ExprNew(n *ast.ExprNew) {
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
}

func (t *ASTTraverser) ExprPostDec(n *ast.ExprPostDec) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *ASTTraverser) ExprPostInc(n *ast.ExprPostInc) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *ASTTraverser) ExprPreDec(n *ast.ExprPreDec) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *ASTTraverser) ExprPreInc(n *ast.ExprPreInc) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *ASTTraverser) ExprPrint(n *ast.ExprPrint) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprPropertyFetch(n *ast.ExprPropertyFetch) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Prop); replacedNode != nil {
		n.Prop = replacedNode
	}
}

func (t *ASTTraverser) ExprNullsafePropertyFetch(n *ast.ExprNullsafePropertyFetch) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Prop); replacedNode != nil {
		n.Prop = replacedNode
	}
}

func (t *ASTTraverser) ExprRequire(n *ast.ExprRequire) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprRequireOnce(n *ast.ExprRequireOnce) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprShellExec(n *ast.ExprShellExec) {
	n.Parts = t.TraverseNodes(n.Parts)
}

func (t *ASTTraverser) ExprStaticCall(n *ast.ExprStaticCall) {
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}
	if replacedNode := t.Traverse(n.Call); replacedNode != nil {
		n.Call = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
}

func (t *ASTTraverser) ExprStaticPropertyFetch(n *ast.ExprStaticPropertyFetch) {
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}
	if replacedNode := t.Traverse(n.Prop); replacedNode != nil {
		n.Prop = replacedNode
	}
}

func (t *ASTTraverser) ExprTernary(n *ast.ExprTernary) {
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

func (t *ASTTraverser) ExprUnaryMinus(n *ast.ExprUnaryMinus) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprUnaryPlus(n *ast.ExprUnaryPlus) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprVariable(n *ast.ExprVariable) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
}

func (t *ASTTraverser) ExprYield(n *ast.ExprYield) {
	if replacedNode := t.Traverse(n.Key); replacedNode != nil {
		n.Key = replacedNode
	}
	if replacedNode := t.Traverse(n.Val); replacedNode != nil {
		n.Val = replacedNode
	}
}

func (t *ASTTraverser) ExprYieldFrom(n *ast.ExprYieldFrom) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprAssign(n *ast.ExprAssign) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprAssignReference(n *ast.ExprAssignReference) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprAssignBitwiseAnd(n *ast.ExprAssignBitwiseAnd) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprAssignBitwiseOr(n *ast.ExprAssignBitwiseOr) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprAssignBitwiseXor(n *ast.ExprAssignBitwiseXor) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprAssignCoalesce(n *ast.ExprAssignCoalesce) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprAssignConcat(n *ast.ExprAssignConcat) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprAssignDiv(n *ast.ExprAssignDiv) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprAssignMinus(n *ast.ExprAssignMinus) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprAssignMod(n *ast.ExprAssignMod) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprAssignMul(n *ast.ExprAssignMul) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprAssignPlus(n *ast.ExprAssignPlus) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprAssignPow(n *ast.ExprAssignPow) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprAssignShiftLeft(n *ast.ExprAssignShiftLeft) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprAssignShiftRight(n *ast.ExprAssignShiftRight) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryBitwiseAnd(n *ast.ExprBinaryBitwiseAnd) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryBitwiseOr(n *ast.ExprBinaryBitwiseOr) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryBitwiseXor(n *ast.ExprBinaryBitwiseXor) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryBooleanAnd(n *ast.ExprBinaryBooleanAnd) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryBooleanOr(n *ast.ExprBinaryBooleanOr) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryCoalesce(n *ast.ExprBinaryCoalesce) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryConcat(n *ast.ExprBinaryConcat) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryDiv(n *ast.ExprBinaryDiv) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryEqual(n *ast.ExprBinaryEqual) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryGreater(n *ast.ExprBinaryGreater) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryGreaterOrEqual(n *ast.ExprBinaryGreaterOrEqual) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryIdentical(n *ast.ExprBinaryIdentical) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryLogicalAnd(n *ast.ExprBinaryLogicalAnd) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryLogicalOr(n *ast.ExprBinaryLogicalOr) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryLogicalXor(n *ast.ExprBinaryLogicalXor) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryMinus(n *ast.ExprBinaryMinus) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryMod(n *ast.ExprBinaryMod) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryMul(n *ast.ExprBinaryMul) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryNotEqual(n *ast.ExprBinaryNotEqual) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryNotIdentical(n *ast.ExprBinaryNotIdentical) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryPlus(n *ast.ExprBinaryPlus) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryPow(n *ast.ExprBinaryPow) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryShiftLeft(n *ast.ExprBinaryShiftLeft) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinaryShiftRight(n *ast.ExprBinaryShiftRight) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinarySmaller(n *ast.ExprBinarySmaller) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinarySmallerOrEqual(n *ast.ExprBinarySmallerOrEqual) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprBinarySpaceship(n *ast.ExprBinarySpaceship) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *ASTTraverser) ExprCastArray(n *ast.ExprCastArray) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprCastBool(n *ast.ExprCastBool) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprCastDouble(n *ast.ExprCastDouble) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprCastInt(n *ast.ExprCastInt) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprCastObject(n *ast.ExprCastObject) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprCastString(n *ast.ExprCastString) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprCastUnset(n *ast.ExprCastUnset) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ExprMatch(n *ast.ExprMatch) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
	n.Arms = t.TraverseNodes(n.Arms)
}

func (t *ASTTraverser) ExprThrow(n *ast.ExprThrow) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *ASTTraverser) ScalarDnumber(n *ast.ScalarDnumber) {

}

func (t *ASTTraverser) ScalarEncapsed(n *ast.ScalarEncapsed) {
	n.Parts = t.TraverseNodes(n.Parts)
}

func (t *ASTTraverser) ScalarEncapsedStringPart(n *ast.ScalarEncapsedStringPart) {

}

func (t *ASTTraverser) ScalarEncapsedStringVar(n *ast.ScalarEncapsedStringVar) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Dim); replacedNode != nil {
		n.Dim = replacedNode
	}
}

func (t *ASTTraverser) ScalarEncapsedStringBrackets(n *ast.ScalarEncapsedStringBrackets) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *ASTTraverser) ScalarHeredoc(n *ast.ScalarHeredoc) {
	n.Parts = t.TraverseNodes(n.Parts)
}

func (t *ASTTraverser) ScalarLnumber(n *ast.ScalarLnumber) {

}

func (t *ASTTraverser) ScalarMagicConstant(n *ast.ScalarMagicConstant) {

}

func (t *ASTTraverser) ScalarString(n *ast.ScalarString) {

}

func (t *ASTTraverser) NameName(n *ast.Name) {
	n.Parts = t.TraverseNodes(n.Parts)
}

func (t *ASTTraverser) NameFullyQualified(n *ast.NameFullyQualified) {
	n.Parts = t.TraverseNodes(n.Parts)
}

func (t *ASTTraverser) NameRelative(n *ast.NameRelative) {
	n.Parts = t.TraverseNodes(n.Parts)
}

func (t *ASTTraverser) NameNamePart(n *ast.NamePart) {

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
