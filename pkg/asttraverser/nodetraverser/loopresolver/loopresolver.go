// https://github.com/ircmaxell/php-cfg/blob/master/lib/PHPCfg/AstVisitor/LoopResolver.php

package loopresolver

import (
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"strconv"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/rxhunter00/XSS-Taint/pkg/asttraverser"
	"github.com/rxhunter00/XSS-Taint/pkg/asttraverser/astutils"
)

type LoopResolver struct {
	labelCounter int
	contStack    []ast.StmtLabel
	breakStack   []ast.StmtLabel
}

func NewLoopResolver() *LoopResolver {
	return &LoopResolver{
		contStack:    make([]ast.StmtLabel, 0),
		breakStack:   make([]ast.StmtLabel, 0),
		labelCounter: 0,
	}
}

// Stub Implement lr *LoopResolver NodeTraverser
func (lr *LoopResolver) EnterNode(n ast.Vertex) (ast.Vertex, asttraverser.ReturnModeFlag) {
	switch n := n.(type) {
	case *ast.StmtBreak:
		lr.handlerBreakStatement(n)
	case *ast.StmtContinue:
		lr.handlerContinueStatement(n)
	case *ast.StmtForeach, *ast.StmtFor, *ast.StmtWhile, *ast.StmtDo:
		lr.breakStack = append(lr.breakStack, lr.createLabelStatement())
		lr.contStack = append(lr.contStack, lr.createLabelStatement())	
	case *ast.StmtSwitch:
		label := lr.createLabelStatement()
		lr.breakStack = append(lr.breakStack, label)
		lr.contStack = append(lr.contStack, label)

	}
	return nil, asttraverser.REPLACEMODE
}

func (lr *LoopResolver) LeaveNode(n ast.Vertex) (ast.Vertex, asttraverser.ReturnModeFlag) {
	switch n := n.(type) {
	case *ast.StmtForeach:
		if _, ok := n.Stmt.(*ast.StmtStmtList); ok {
			n.Stmt.(*ast.StmtStmtList).Stmts = append(n.Stmt.(*ast.StmtStmtList).Stmts, astutils.PopLabelStack(&lr.contStack))
			return astutils.PopLabelStack(&lr.breakStack), asttraverser.INSERTMODE
		}
	case *ast.StmtFor:
		if _, ok := n.Stmt.(*ast.StmtStmtList); ok {
			n.Stmt.(*ast.StmtStmtList).Stmts = append(n.Stmt.(*ast.StmtStmtList).Stmts, astutils.PopLabelStack(&lr.contStack))
			return astutils.PopLabelStack(&lr.breakStack), asttraverser.INSERTMODE
		}
	case *ast.StmtWhile:
		if _, ok := n.Stmt.(*ast.StmtStmtList); ok {
			n.Stmt.(*ast.StmtStmtList).Stmts = append(n.Stmt.(*ast.StmtStmtList).Stmts, astutils.PopLabelStack(&lr.contStack))
			return astutils.PopLabelStack(&lr.breakStack), asttraverser.INSERTMODE
		}
	case *ast.StmtDo:
		if _, ok := n.Stmt.(*ast.StmtStmtList); ok {
			n.Stmt.(*ast.StmtStmtList).Stmts = append(n.Stmt.(*ast.StmtStmtList).Stmts, astutils.PopLabelStack(&lr.contStack))
			return astutils.PopLabelStack(&lr.breakStack), asttraverser.INSERTMODE
		}

	case *ast.StmtSwitch:
		astutils.PopLabelStack(&lr.contStack)
		return astutils.PopLabelStack(&lr.breakStack), asttraverser.INSERTMODE
	}
	return nil, asttraverser.REPLACEMODE
}

func (lr *LoopResolver) createLabelStatement() ast.StmtLabel {
	labelName := fmt.Sprintf("compiled_label_%d_%d", rand.Int(), lr.labelCounter)
	lr.labelCounter += 1
	return ast.StmtLabel{
		Name: &ast.Identifier{
			Value: []byte(labelName),
		},
	}
}

func (lr *LoopResolver) handlerBreakStatement(n *ast.StmtBreak) *ast.StmtGoto {

	// break;
	if n.Expr == nil {
		label := astutils.TopLabelStack(&lr.breakStack)
		return &ast.StmtGoto{Label: label}
	}
	// break 2;
	if nExpr, ok := n.Expr.(*ast.ScalarLnumber); ok {
		paramNum, err := strconv.Atoi(string(nExpr.Value))
		if err != nil || paramNum <= 0 {
			log.Fatalf("'break' operator accepts only positive integers %v", err)
		}

		// too much break
		if paramNum > len(lr.breakStack) {
			log.Fatalf("Cannot 'break' %d level\n", paramNum)
		}

		// get appropriate break location
		labelIdx := len(lr.breakStack) - paramNum
		labelLoc := lr.breakStack[labelIdx]
		return &ast.StmtGoto{Label: &labelLoc}
	} else {
		log.Fatalf("'break' operator accepts only positive integers")
	}

	return nil
}
func (lr *LoopResolver) handlerContinueStatement(n *ast.StmtContinue) *ast.StmtGoto {

	if n.Expr == nil {
		label := astutils.TopLabelStack(&lr.contStack)
		return &ast.StmtGoto{Label: label}
	}
	if nExpr, ok := n.Expr.(*ast.ScalarLnumber); ok {
		paramNum, err := strconv.Atoi(string(nExpr.Value))
		if err != nil || paramNum <= 0 {
			log.Fatalf("'continue' operator accepts only positive integers: %v", err)
		}

		if paramNum > len(lr.contStack) {
			log.Fatalf("Cannot 'continue' %d level\n", paramNum)
		}
		// get continue location
		labelIdx := len(lr.contStack) - paramNum
		labelLoc := lr.contStack[labelIdx]
		return &ast.StmtGoto{Label: &labelLoc}
	} else if nExpr, ok := n.Expr.(*ast.ScalarDnumber); ok {
		paramNum, err := strconv.Atoi(string(nExpr.Value))
		if err != nil || paramNum <= 0 {
			log.Fatalf("'continue' operator accepts only positive integers: %v", err)
		}
		if paramNum > len(lr.contStack) {
			log.Fatalf("Cannot 'continue' %d level\n", paramNum)
		}
		labelIdx := len(lr.contStack) - paramNum
		labelLoc := lr.contStack[labelIdx]
		return &ast.StmtGoto{Label: &labelLoc}
	} else if nExpr, ok := n.Expr.(*ast.ExprBrackets); ok {
		paramNum, err := strconv.Atoi(string(nExpr.Expr.(*ast.ScalarLnumber).Value))
		if err != nil || paramNum <= 0 {
			log.Fatalf("'continue' operator accepts only positive integers: %v", err)
		}
		if paramNum > len(lr.contStack) {
			log.Fatalf("Cannot 'continue' %d level\n", paramNum)
		}
		// get appropriate continue location
		labelIdx := len(lr.contStack) - paramNum
		labelLoc := lr.contStack[labelIdx]
		return &ast.StmtGoto{Label: &labelLoc}
	} else {
		log.Fatalf("'continue' operator accepts only positive integers: '%v'", reflect.TypeOf(n.Expr))
	}

	return nil
}
