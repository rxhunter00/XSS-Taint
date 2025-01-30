package cfg

import (
	"fmt"
	"log"
	"strings"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/position"
	"github.com/rxhunter00/XSS-Taint/pkg/asttraverser/astutils"
)

func (builder *CFGBuilder) parseStmtNodes(nodes []ast.Vertex, block *Block) (*Block, error) {
	if block == nil {
		return nil, fmt.Errorf("cannot parse nodes in nil block")
	}

	// set current block acoridng to statement block
	tmpBlock := builder.currentBlock
	builder.currentBlock = block
	for _, node := range nodes {
		builder.parseStmtNode(node)
	}
	// save current block
	endBlock := builder.currentBlock
	// reset builde block to block before we enter so they can continue the statement
	builder.currentBlock = tmpBlock

	if endBlock == nil {
		log.Fatalf("Error:parseStmtNodes: got nil as end block")
	}

	return endBlock, nil
}

func (builder *CFGBuilder) parseStmtNode(vertexNode ast.Vertex) {

	switch nodeType := vertexNode.(type) {
	case *ast.StmtExpression:
		builder.parseExprNode(nodeType.Expr)
	case *ast.StmtIf:
		builder.parseStmtIf(nodeType)
	case *ast.StmtBreak:
		// Modified in Loop Resolver
	case *ast.StmtCase:
		// Do Nothing Switch Statement
	case *ast.StmtCatch:
		// TODO
	case *ast.StmtEnum:
		// TODO
	case *ast.StmtClass:
		builder.parseStmtClass(nodeType)
	case *ast.StmtClassConstList:
		builder.parseStmtClassConstList(nodeType)
	case *ast.StmtClassMethod:
		builder.parseStmtClassMethod(nodeType)
	case *ast.StmtConstList:
		builder.parseStmtConstList(nodeType)
	case *ast.StmtConstant:
		builder.parseStmtConst(nodeType)
	case *ast.StmtContinue:
		// Modified in Loop Resolver
	case *ast.StmtDeclare:
		//TODO
	case *ast.StmtSwitch:
		builder.parseStmtSwitch(nodeType)
	case *ast.StmtDefault:
		// Do Nothing Switch Statement
	case *ast.StmtDo:
		builder.parseStmtDo(nodeType)
	case *ast.StmtEcho:
		builder.parseStmtEcho(nodeType)
	case *ast.StmtElse:
		//Do Nothing If Statement
	case *ast.StmtElseIf:
		//Do Nothing If Statement
	case *ast.StmtFinally:
		// TODO
	case *ast.StmtFor:
		builder.parseStmtFor(nodeType)
	case *ast.StmtForeach:
		builder.parseStmtForeach(nodeType)
	case *ast.StmtFunction:
		builder.parseStmtFunction(nodeType)
	case *ast.StmtGlobal:
		builder.parseStmtGlobal(nodeType)
	case *ast.StmtGoto:
		//Do Nothing
		builder.parseStmtGoto(nodeType)
	case *ast.StmtHaltCompiler:
		//do nothing
	case *ast.StmtInlineHtml:
		// TODO Parse for Context
	case *ast.StmtInterface:
		builder.parseStmtInterface(nodeType)
	case *ast.StmtLabel:
		builder.parseStmtLabel(nodeType)
	case *ast.StmtNamespace:
		builder.parseStmtNamespace(nodeType)
	case *ast.StmtNop:
		//Do Nothing
	case *ast.StmtProperty:
		//Handled
	case *ast.StmtPropertyList:
		builder.parseStmtPropertyList(nodeType)
	case *ast.StmtReturn:
		builder.parseStmtReturn(nodeType)
	case *ast.StmtStatic:
		builder.parseStmtStatic(nodeType)
	case *ast.StmtStaticVar:
		builder.parseStmtStaticVar(nodeType)
	case *ast.StmtStmtList:
		//Do nothing
	case *ast.StmtThrow:
		builder.parseStmtThrow(nodeType)
	case *ast.StmtTrait:
		builder.parseStmtTrait(nodeType)
	case *ast.StmtTraitUse:
		builder.parseStmtTraitUse(nodeType)
	case *ast.StmtTraitUseAlias:
		//do nothing
	case *ast.StmtTraitUsePrecedence:
		//do nothing
	case *ast.StmtTry:
		builder.parseStmtTry(nodeType)
	case *ast.StmtUnset:
		builder.parseStmtUnset(nodeType)
	case *ast.StmtUseList:
		//Do nothing
	case *ast.StmtGroupUseList:
		//do nothinh
	case *ast.StmtUse:
		//do nothing
	case *ast.StmtWhile:
		builder.parseStmtWhile(nodeType)

	}

}

func (cb *CFGBuilder) parseStmtUnset(stmt *ast.StmtUnset) {
	exprs, exprsPos := cb.parseExprList(stmt.Vars, PARSER_MODE_WRITE)
	op := NewOpUnset(exprs, exprsPos, stmt.Position)
	cb.currentBlock.AddInstructions(op)
}
func (cb *CFGBuilder) parseStmtTry(stmt *ast.StmtTry) {
	cb.parseStmtNodes(stmt.Stmts, cb.currentBlock)
}

func (builder *CFGBuilder) parseStmtWhile(stmt *ast.StmtWhile) {
	var err error
	// initialize 3 block in while loop
	initBlock := NewBlock(builder.GetBlockIdCount())
	bodyBlock := NewBlock(builder.GetBlockIdCount())
	endBlock := NewBlock(builder.GetBlockIdCount())

	// go to init block
	builder.currentBlock.AddInstructions(NewOpStmtJump(initBlock, stmt.Position))
	initBlock.AddPredecessor(builder.currentBlock)
	builder.currentBlock = initBlock

	// create branch to body and end block
	cond, err := builder.readVariable(builder.parseExprNode(stmt.Cond))
	if err != nil {
		log.Fatalf("Error in parseStmtWhile: %v", err)
	}
	builder.currentBlock.AddInstructions(NewOpStmtJumpIf(cond, bodyBlock, endBlock, stmt.Cond.GetPosition()))
	builder.currentBlock.IsConditionalBlock = true
	bodyBlock.AddPredecessor(builder.currentBlock)
	endBlock.AddPredecessor(builder.currentBlock)

	builder.FuncContex.PushCond(cond)
	bodyBlock.SetCondition(builder.FuncContex.CurrConds)

	stmts, err := astutils.GetStmtList(stmt.Stmt)
	if err != nil {
		log.Fatalf("Error in parseStmtWhile: %v", err)
	}
	builder.currentBlock, err = builder.parseStmtNodes(stmts, bodyBlock)
	// return condition
	builder.FuncContex.PopCond()
	if err != nil {
		log.Fatalf("Error in parseStmtWhile: %v", err)
	}

	// go back to init block
	builder.currentBlock.AddInstructions(NewOpStmtJump(initBlock, stmt.Position))
	initBlock.AddPredecessor(builder.currentBlock)

	// add condition to end block
	negatedCond := NewOpExprBooleanNot(cond, nil).Result
	builder.FuncContex.PushCond(negatedCond)
	endBlock.SetCondition(builder.FuncContex.CurrConds)
	builder.currentBlock = endBlock
}

func (builder *CFGBuilder) parseStmtTrait(stmt *ast.StmtTrait) {
	name := builder.parseExprNode(stmt.Name)
	prevClass := builder.currClassOper
	builder.currClassOper = name.(*OperandString)
	stmts, err := builder.parseStmtNodes(stmt.Stmts, NewBlock(builder.GetBlockIdCount()))
	if err != nil {
		log.Fatalf("Error in parseStmtTrait: %v", err)
	}
	builder.currentBlock.AddInstructions(NewOpStmtTrait(name, stmts, stmt.Position))
	builder.currClassOper = prevClass
}

func (builder *CFGBuilder) parseStmtTraitUse(stmt *ast.StmtTraitUse) {
	traits := make([]Operand, 0, len(stmt.Traits))
	adaptations := make([]Op, 0, len(stmt.Adaptations))

	for _, trait := range stmt.Traits {
		traitName, err := astutils.GetNameString(trait)
		if err != nil {
			log.Fatal("Error trait name in StmtTraitUse")
		}
		traits = append(traits, NewOperandString(traitName))
	}

	for _, adaptation := range stmt.Adaptations {
		switch a := adaptation.(type) {
		case *ast.StmtTraitUseAlias:
			trait := Operand(nil)
			methodStr, err := astutils.GetNameString(a.Method)
			if err != nil {
				log.Fatal("Error method name in StmtTraitUse")
			}
			method := NewOperandString(methodStr)
			newName := Operand(nil)
			newModifier := builder.parseClassModifier([]ast.Vertex{a.Modifier})
			if a.Trait != nil {
				traitStr, err := astutils.GetNameString(a.Trait)
				if err != nil {
					log.Fatal("Error trait name in StmtTraitUse")
				}
				trait = NewOperandString(traitStr)
			}
			if a.Alias != nil {
				aliasStr, err := astutils.GetNameString(a.Alias)
				if err != nil {
					log.Fatal("Error alias name in StmtTraitUse")
				}
				newName = NewOperandString(aliasStr)
			}

			aliasOp := NewOpAlias(trait, method, newName, newModifier, a.Position)
			adaptations = append(adaptations, aliasOp)
		case *ast.StmtTraitUsePrecedence:
			insteadOfs := make([]Operand, 0, len(a.Insteadof))
			trait := Operand(nil)
			methodStr, err := astutils.GetNameString(a.Method)
			if err != nil {
				log.Fatal("Error method name in StmtTraitUsePrecedence")
			}
			method := NewOperandString(methodStr)
			if a.Trait != nil {
				traitStr, err := astutils.GetNameString(a.Trait)
				if err != nil {
					log.Fatal("Error trait name in StmtTraitUsePrecedence")
				}
				trait = NewOperandString(traitStr)
			}
			for _, insteadOf := range a.Insteadof {
				insteadOfStr, err := astutils.GetNameString(insteadOf)
				if err != nil {
					log.Fatalf("Error insteadof in StmtTraitUsePrecedence")
				}
				insteadOfName := NewOperandString(insteadOfStr)
				insteadOfs = append(insteadOfs, insteadOfName)
			}

			precedenceOp := NewOpPrecedence(trait, method, insteadOfs, a.Position)
			adaptations = append(adaptations, precedenceOp)
		}
	}
	traitUseOp := NewOpStmtTraitUse(traits, adaptations, stmt.Position)
	builder.currentBlock.AddInstructions(traitUseOp)
}

func (builder *CFGBuilder) parseStmtThrow(stmt *ast.StmtThrow) {
	expr, err := builder.readVariable(builder.parseExprNode(stmt.Expr))
	if err != nil {
		log.Fatalf("Error in parseStmtThrow: %v", err)
	}
	op := NewOpThrow(expr, stmt.Position)
	builder.currentBlock.AddInstructions(op)
	// script after throw will be a dead code
	builder.currentBlock = NewBlock(builder.GetBlockIdCount())
	builder.currentBlock.Dead = true
}

func (builder *CFGBuilder) parseStmtStatic(stmt *ast.StmtStatic) {
	for _, vr := range stmt.Vars {
		builder.parseStmtNode(vr)
	}
}

func (builder *CFGBuilder) parseStmtStaticVar(stmt *ast.StmtStaticVar) {
	defaultVar := Operand(nil)
	defaultBlock := (*Block)(nil)
	if stmt.Expr != nil {
		tmp := builder.currentBlock
		defaultBlock = NewBlock(builder.GetBlockIdCount())
		builder.currentBlock = defaultBlock
		defaultVar = builder.parseExprNode(stmt.Expr)
		builder.currentBlock = tmp
	}

	vr := builder.writeVariable(NewOperandBoundVariable(builder.parseExprNode(stmt.Var), NewOperandNull(), BOUND_VAR_SCOPE_FUNCTION, true, nil))
	builder.currentBlock.AddInstructions(NewOpStaticVar(vr, defaultVar, defaultBlock, stmt.Position))
}

func (builder *CFGBuilder) parseStmtReturn(stmt *ast.StmtReturn) {
	expr := Operand(nil)
	if stmt.Expr != nil {
		var err error
		expr, err = builder.readVariable(builder.parseExprNode(stmt.Expr))
		if err != nil {
			log.Fatalf("Error in parseStmtReturn: %v", err)
		}
	}

	returnOp := NewOpReturn(expr, stmt.Position)
	builder.currentBlock.AddInstructions(returnOp)

	// script after return will be a dead code
	builder.currentBlock = NewBlock(builder.GetBlockIdCount())
	builder.currentBlock.Dead = true
}

func (builder *CFGBuilder) parseStmtPropertyList(stmt *ast.StmtPropertyList) {
	attrGroups := builder.parseAttributeGroups(stmt.AttrGroups)
	declaredType := builder.parseTypeNode(stmt.Type)
	visibility := ClassModifFlag(CLASS_MODIF_PUBLIC)
	static := false
	readonly := false
	// parse modifiers
	for _, modifier := range stmt.Modifiers {
		modifierStr, err := astutils.GetNameString(modifier)
		if err != nil {
			log.Fatal("Error modifier name in StmtPropertyList")
		}
		switch strings.ToLower(modifierStr) {
		case "public":
			visibility = CLASS_MODIF_PUBLIC
		case "protected":
			visibility = CLASS_MODIF_PROTECTED
		case "private":
			visibility = CLASS_MODIF_PRIVATE
		case "static":
			static = true
		case "readonly":
			readonly = true
		}
	}

	// parse each property
	for _, prop := range stmt.Props {
		defaultVar := Operand(nil)
		defaultBlock := (*Block)(nil)
		if prop.(*ast.StmtProperty).Expr != nil {
			defaultBlock = NewBlock(builder.GetBlockIdCount())
			tmp := builder.currentBlock
			builder.currentBlock = defaultBlock
			defaultVar = builder.parseExprNode(prop.(*ast.StmtProperty).Expr)
			builder.currentBlock = tmp
			if defaultVar.IsTainted() {
				builder.currentFunc.FuncHasTaint = true
				builder.currentBlock.HasTainted = true
			}
		}

		name := builder.parseExprNode(prop.(*ast.StmtProperty).Var)
		op := NewOpStmtProperty(name, visibility, static, readonly, attrGroups, defaultVar, defaultBlock, declaredType, prop.GetPosition())
		builder.currentBlock.AddInstructions(op)
	}
}
func (cb *CFGBuilder) parseStmtInterface(stmt *ast.StmtInterface) {
	name, err := cb.readVariable(cb.parseExprNode(stmt.Name))
	if err != nil {
		log.Fatalf("Error in parseStmtInterface: %v", err)
	}
	tmpClass := cb.currClassOper
	cb.currClassOper = name.(*OperandString)

	extends, _ := cb.parseExprList(stmt.Extends, PARSER_MODE_NONE)
	block, err := cb.parseStmtNodes(stmt.Stmts, cb.currentBlock)
	if err != nil {
		log.Fatalf("Error in parseStmtInterface: %v", err)
	}
	op := NewOpStmtInterface(name, block, extends, stmt.Position)
	cb.currentBlock.AddInstructions(op)

	cb.currClassOper = tmpClass
}

func (cb *CFGBuilder) parseStmtLabel(stmt *ast.StmtLabel) {
	labelName, err := astutils.GetNameString(stmt.Name)
	if err != nil {
		log.Fatal("Error label name in StmtLabel")
	}
	if _, ok := cb.FuncContex.GetLabel(labelName); ok {
		fmt.Println("Error: label '", labelName, "' have been defined")
		return
	}

	labelBlock := NewBlock(cb.GetBlockIdCount())
	jmp := NewOpStmtJump(labelBlock, stmt.Position)
	cb.currentBlock.AddInstructions(jmp)
	labelBlock.AddPredecessor(cb.currentBlock)

	// add condition to label block
	labelBlock.SetCondition(cb.FuncContex.CurrConds)

	// add jump to label block for every unresolved goto
	if unresolvedGotos, ok := cb.FuncContex.GetUnresolvedGotos(labelName); ok {
		for _, unresolvedGoto := range unresolvedGotos {
			jmp = NewOpStmtJump(labelBlock, nil)
			unresolvedGoto.AddInstructions(jmp)
			labelBlock.AddPredecessor(unresolvedGoto)
		}
		cb.FuncContex.RemoveGoto(labelName)
	}

	cb.FuncContex.Labels[labelName] = labelBlock
	cb.currentBlock = labelBlock
}

func (builder *CFGBuilder) parseStmtGoto(stmt *ast.StmtGoto) {
	labelName := ""
	var err error
	switch stmt.Label.(type) {
	case *ast.StmtLabel:
		labelName, err = astutils.GetNameString(stmt.Label.(*ast.StmtLabel).Name)
		if err != nil {

			fmt.Printf("Error in StmtGoto: %v\n", err)
			return
		}
	case *ast.Identifier:
		labelName, err = astutils.GetNameString(stmt.Label)
		if err != nil {

			fmt.Printf("Error in StmtGoto: %v\n", err)
			return
		}
	}

	if labelBlock, ok := builder.FuncContex.GetLabel(labelName); ok {
		builder.currentBlock.AddInstructions(NewOpStmtJump(labelBlock, stmt.Position))
		labelBlock.AddPredecessor(builder.currentBlock)
	} else {
		builder.FuncContex.AddUnresolvedGoto(labelName, builder.currentBlock)
	}

	// script after return will be a dead code
	builder.currentBlock = NewBlock(builder.GetBlockIdCount())
	builder.currentBlock.Dead = true
}

func (builder *CFGBuilder) parseStmtGlobal(stmt *ast.StmtGlobal) {
	for _, vr := range stmt.Vars {
		vrOper := builder.writeVariable(builder.parseExprNode(vr))
		op := NewOpGlobalVar(vrOper, vr.GetPosition())
		builder.currentBlock.AddInstructions(op)
	}
}

func (builder *CFGBuilder) parseStmtFunction(stmt *ast.StmtFunction) {
	// create OpFunc instance and append to script object
	name, err := astutils.GetNameString(stmt.Name)
	if err != nil {
		log.Fatal("Error func name in StmtFunction")
	}
	flags := FuncModifFlag(0)
	returnType := builder.parseTypeNode(stmt.ReturnType)
	if stmt.AmpersandTkn != nil {
		flags |= FUNC_MODIF_FLAG_RETURNS_REF
	}
	entryBlock := NewBlock(builder.GetBlockIdCount())
	fn, err := NewFunc(name, flags, returnType, entryBlock, stmt.Position)
	if err != nil {
		log.Fatalf("Error in parseStmtFunction: %v", err)
	}
	builder.Script.AddFunc(fn)

	// parse function
	builder.parseFunc(fn, stmt.Params, stmt.Stmts)
	attrGroups := builder.parseAttributeGroups(stmt.AttrGroups)
	opStmtFunc := NewOpStmtFunc(fn, attrGroups, stmt.Position)
	builder.currentBlock.AddInstructions(opStmtFunc)
	fn.CallableOp = opStmtFunc
}

func (builder *CFGBuilder) parseStmtNamespace(stmt *ast.StmtNamespace) {
	if stmt.Name == nil {
		return
	}
	nameSpace, err := astutils.GetNameString(stmt.Name)
	if err != nil {
		log.Fatal("Error namespace in StmtNameSpace")
	}
	builder.CurrNamespace = nameSpace
	builder.currentBlock, err = builder.parseStmtNodes(stmt.Stmts, builder.currentBlock)
	if err != nil {
		log.Fatalf("Error in StmtNameSpace: %v", err)
	}
}

func (builder *CFGBuilder) parseStmtEcho(stmt *ast.StmtEcho) {
	for _, expr := range stmt.Exprs {
		exprOper, err := builder.readVariable(builder.parseExprNode(expr))
		if err != nil {
			log.Fatalf("Error in parseStmtEcho: %v", err)
		}
		echoOp := NewOpEcho(exprOper, stmt.GetPosition())
		builder.currentBlock.AddInstructions(echoOp)
	}
}

func (builder *CFGBuilder) parseStmtDo(stmt *ast.StmtDo) {
	var err error

	bodyBlock := NewBlock(builder.GetBlockIdCount())
	bodyBlock.AddPredecessor(builder.currentBlock)
	endBlock := NewBlock(builder.GetBlockIdCount())
	builder.currentBlock.AddInstructions(NewOpStmtJump(bodyBlock, stmt.Position))

	// parse statements in the loop body
	// no need to add condition cause do block will always be executed
	builder.currentBlock = bodyBlock
	builder.currentBlock, err = builder.parseStmtNodes(stmt.Stmt.(*ast.StmtStmtList).Stmts, bodyBlock)
	if err != nil {
		log.Fatalf("Error in parseStmtNodes: %v", err)
	}
	cond, err := builder.readVariable(builder.parseExprNode(stmt.Cond))
	if err != nil {
		log.Fatalf("Error in parseStmtDo: %v", err)
	}
	builder.currentBlock.AddInstructions(NewOpStmtJumpIf(cond, bodyBlock, endBlock, stmt.Cond.GetPosition()))
	builder.currentBlock.IsConditionalBlock = true
	builder.processAssertion(cond, bodyBlock, endBlock)
	bodyBlock.AddPredecessor(builder.currentBlock)
	endBlock.AddPredecessor(builder.currentBlock)

	// add condition to end block
	negatedCond := NewOpExprBooleanNot(cond, nil).Result
	builder.FuncContex.PushCond(negatedCond)
	endBlock.SetCondition(builder.FuncContex.CurrConds)
	builder.currentBlock = endBlock
}

func (builder *CFGBuilder) parseStmtFor(stmt *ast.StmtFor) {
	var err error

	builder.parseExprList(stmt.Init, PARSER_MODE_READ)
	initBlock := NewBlock(builder.GetBlockIdCount())
	bodyBlock := NewBlock(builder.GetBlockIdCount())
	endBlock := NewBlock(builder.GetBlockIdCount())

	// go to init block
	builder.currentBlock.AddInstructions(NewOpStmtJump(initBlock, stmt.Position))
	initBlock.AddPredecessor(builder.currentBlock)
	builder.currentBlock = initBlock

	// check the condition
	cond := Operand(nil)
	if len(stmt.Cond) != 0 {
		vr, _ := builder.parseExprList(stmt.Cond, PARSER_MODE_NONE)
		cond, err = builder.readVariable(vr[len(vr)-1])
		if err != nil {
			log.Fatalf("Error in parseStmtFor: %v", err)
		}
	} else {
		cond = NewOperandBool(true)
	}
	builder.currentBlock.AddInstructions(NewOpStmtJumpIf(cond, bodyBlock, endBlock, stmt.Position))
	builder.currentBlock.IsConditionalBlock = true
	builder.processAssertion(cond, bodyBlock, endBlock)
	bodyBlock.AddPredecessor(builder.currentBlock)
	endBlock.AddPredecessor(builder.currentBlock)

	// add condition to block
	builder.FuncContex.PushCond(cond)
	bodyBlock.SetCondition(builder.FuncContex.CurrConds)

	stmts, err := astutils.GetStmtList(stmt.Stmt)
	if err != nil {
		log.Fatalf("Error in parseStmtFor: %v", err)
	}
	builder.currentBlock, err = builder.parseStmtNodes(stmts, bodyBlock)
	builder.FuncContex.PopCond()
	if err != nil {
		log.Fatalf("Error in parseStmtFor: %v", err)
	}
	builder.parseExprList(stmt.Loop, PARSER_MODE_READ)
	// go back to init block
	builder.currentBlock.AddInstructions(NewOpStmtJump(initBlock, stmt.Position))
	initBlock.AddPredecessor(builder.currentBlock)
	// add condition to endblock
	negatedCond := NewOpExprBooleanNot(cond, nil).Result
	builder.FuncContex.PushCond(negatedCond)
	endBlock.SetCondition(builder.FuncContex.CurrConds)
	builder.currentBlock = endBlock
}

func (builder *CFGBuilder) parseStmtForeach(stmt *ast.StmtForeach) {
	var err error
	iterable, err := builder.readVariable(builder.parseExprNode(stmt.Expr))
	if err != nil {
		log.Fatalf("Error in parseStmtForEach: %v", err)
	}
	builder.currentBlock.AddInstructions(NewOpReset(iterable, stmt.Expr.GetPosition()))

	initBlock := NewBlock(builder.GetBlockIdCount())
	bodyBlock := NewBlock(builder.GetBlockIdCount())
	endBlock := NewBlock(builder.GetBlockIdCount())

	// go to init block
	builder.currentBlock.AddInstructions(NewOpStmtJump(initBlock, stmt.Position))
	initBlock.AddPredecessor(builder.currentBlock)

	// create valid iterator
	validOp := NewOpExprValid(iterable, nil)
	initBlock.AddInstructions(validOp)

	// go to body block
	initBlock.AddInstructions(NewOpStmtJumpIf(validOp.Result, bodyBlock, endBlock, stmt.Position))
	initBlock.IsConditionalBlock = true
	builder.processAssertion(validOp.Result, bodyBlock, endBlock)
	bodyBlock.AddPredecessor(builder.currentBlock)
	endBlock.AddPredecessor(builder.currentBlock)

	// parse body
	builder.currentBlock = bodyBlock
	if stmt.Key != nil {
		keyOp := NewOpExprKey(iterable, stmt.Key.GetPosition())
		keyVar, err := builder.readVariable(builder.parseExprNode(stmt.Key))
		if err != nil {
			log.Fatalf("Error in parseStmtForEach (key): %v", err)
		}
		builder.currentBlock.AddInstructions(keyOp)
		assignOp := NewOpExprAssign(keyVar, keyOp.Result, stmt.Key.GetPosition(), stmt.Key.GetPosition(), stmt.Key.GetPosition())
		builder.currentBlock.AddInstructions(assignOp)
	}
	isRef := stmt.AmpersandTkn != nil
	valueOp := NewOpExprValue(iterable, isRef, stmt.Var.GetPosition())

	// assign each item to variable
	switch v := stmt.Var.(type) {
	case *ast.ExprList:
		builder.parseAssignList(v.Items, valueOp.Result, nil)
	case *ast.ExprArray:
		builder.parseAssignList(v.Items, valueOp.Result, nil)
	default:
		vr, err := builder.readVariable(builder.parseExprNode(stmt.Var))
		if err != nil {
			log.Fatalf("Error in parseStmtForEach (default): %v", err)
		}
		if isRef {
			builder.currentBlock.AddInstructions(NewOpExprAssignRef(vr, valueOp.Result, stmt.Var.GetPosition()))
		} else {
			builder.currentBlock.AddInstructions(NewOpExprAssign(vr, valueOp.Result, stmt.Var.GetPosition(), stmt.Var.GetPosition(), stmt.Var.GetPosition()))
		}
	}

	// parse statements inside loop body
	stmts, err := astutils.GetStmtList(stmt.Stmt)
	if err != nil {
		log.Fatalf("Error in parseStmtForEach: %v", err)
	}
	builder.currentBlock, err = builder.parseStmtNodes(stmts, builder.currentBlock)
	if err != nil {
		log.Fatalf("Error in parseStmtForEach: %v", err)
	}
	builder.currentBlock.AddInstructions(NewOpStmtJump(initBlock, stmt.Position))
	initBlock.AddPredecessor(builder.currentBlock)

	builder.currentBlock = endBlock
}

func (builder *CFGBuilder) parseStmtSwitch(stmt *ast.StmtSwitch) {
	var err error
	if isJumpTableSwitch(stmt) {
		// build jump table switch
		cond, err := builder.readVariable(builder.parseExprNode(stmt.Cond))
		if err != nil {
			log.Fatalf("Error in parseStmtSwitch: %v", err)
		}
		cases := make([]Operand, 0)
		targets := make([]*Block, 0)
		endBlock := NewBlock(builder.GetBlockIdCount())
		defaultBlock := endBlock
		prevBlock := (*Block)(nil)

		for _, caseNode := range stmt.Cases {
			caseBlock := NewBlock(builder.GetBlockIdCount())
			caseBlock.AddPredecessor(builder.currentBlock)

			// case will be fallthrough if no break (prevBlock dead)
			if prevBlock != nil && !prevBlock.Dead {
				jmp := NewOpStmtJump(caseBlock, caseNode.GetPosition())
				prevBlock.AddInstructions(jmp)
				caseBlock.AddPredecessor(prevBlock)
			}

			switch cn := caseNode.(type) {
			case *ast.StmtCase:
				caseValue := builder.parseExprNode(cn.Cond)
				caseCond := NewOpExprBinaryEqual(cond, caseValue, cn.Position).Result

				builder.FuncContex.PushCond(caseCond)
				caseBlock.SetCondition(builder.FuncContex.CurrConds)
				targets = append(targets, caseBlock)
				cases = append(cases, caseValue)
				prevBlock, err = builder.parseStmtNodes(cn.Stmts, caseBlock)

				builder.FuncContex.PopCond()
				if err != nil {
					log.Fatalf("Error in parseOpFunc: %v", err)
				}
			case *ast.StmtDefault:
				defaultBlock = caseBlock
				prevBlock, err = builder.parseStmtNodes(cn.Stmts, caseBlock)
				if err != nil {
					log.Fatalf("Error in parseOpFunc: %v", err)
				}
			default:
				log.Fatal("Error: Invalid case node type")
			}
		}

		switchOp := NewOpStmtSwitch(cond, cases, targets, defaultBlock, stmt.Position)
		builder.currentBlock.AddInstructions(switchOp)

		if prevBlock != nil && !prevBlock.Dead {
			jmp := NewOpStmtJump(endBlock, stmt.Position)
			prevBlock.AddInstructions(jmp)
			endBlock.AddPredecessor(prevBlock)
		}

		builder.currentBlock = endBlock
	} else {
		// build sequence of compare-and-jump
		cond := builder.parseExprNode(stmt.Cond)
		endBlock := NewBlock(builder.GetBlockIdCount())
		defaultBlock := endBlock
		prevBlock := (*Block)(nil)

		for _, caseNode := range stmt.Cases {
			ifBlock := NewBlock(builder.GetBlockIdCount())
			if prevBlock != nil && !prevBlock.Dead {
				jmp := NewOpStmtJump(ifBlock, caseNode.GetPosition())
				prevBlock.AddInstructions(jmp)
				ifBlock.AddPredecessor(prevBlock)
			}

			switch cn := caseNode.(type) {
			case *ast.StmtCase:
				caseExpr := builder.parseExprNode(cn.Cond)
				left, err := builder.readVariable(cond)
				if err != nil {
					log.Fatalf("Error in StmtCase: %v", err)
				}
				right, err := builder.readVariable(caseExpr)
				if err != nil {
					log.Fatalf("Error in StmtCase: %v", err)
				}
				opEqual := NewOpExprBinaryEqual(left, right, cn.Position)
				builder.currentBlock.AddInstructions(opEqual)

				elseBlock := NewBlock(builder.GetBlockIdCount())
				opJmpIf := NewOpStmtJumpIf(opEqual.Result, ifBlock, elseBlock, cn.Position)
				builder.currentBlock.AddInstructions(opJmpIf)
				builder.currentBlock.IsConditionalBlock = true
				ifBlock.AddPredecessor(builder.currentBlock)
				elseBlock.AddPredecessor(builder.currentBlock)
				builder.currentBlock = elseBlock
				// add condition to if Block
				builder.FuncContex.PushCond(opEqual.Result)
				ifBlock.SetCondition(builder.FuncContex.CurrConds)
				prevBlock, err = builder.parseStmtNodes(cn.Stmts, ifBlock)
				// return condition
				builder.FuncContex.PopCond()
				if err != nil {
					log.Fatalf("Error in parseStmtSwitch: %v", err)
				}
			case *ast.StmtDefault:
				defaultBlock = ifBlock
				prevBlock, err = builder.parseStmtNodes(cn.Stmts, ifBlock)
				if err != nil {
					log.Fatalf("Error in parseStmtSwitch: %v", err)
				}
			}
		}

		if prevBlock != nil && !prevBlock.Dead {
			jmp := NewOpStmtJump(endBlock, stmt.Position)
			prevBlock.AddInstructions(jmp)
			endBlock.AddPredecessor(prevBlock)
		}

		builder.currentBlock.AddInstructions(NewOpStmtJump(defaultBlock, stmt.Position))
		defaultBlock.AddPredecessor(builder.currentBlock)
		builder.currentBlock = endBlock
	}
}

func isJumpTableSwitch(stmt *ast.StmtSwitch) bool {
	for _, cs := range stmt.Cases {
		// all case must be a scalar
		switch csT := cs.(type) {
		case *ast.StmtCase:
			if !astutils.IsScalarNode(csT.Cond) {
				return false
			}
		}
	}
	return true
}

func (builder *CFGBuilder) parseStmtConst(stmt *ast.StmtConstant) {
	// create a new block for defining const
	tmp := builder.currentBlock
	valBlock := NewBlock(builder.GetBlockIdCount())
	builder.currentBlock = valBlock
	val := builder.parseExprNode(stmt.Expr)
	builder.currentBlock = tmp

	name, err := builder.readVariable(builder.parseExprNode(stmt.Name))
	if err != nil {
		log.Fatalf("Error in parseStmtConst: %v", err)
	}
	opConst := NewOpConst(name, val, valBlock, stmt.Position)
	builder.currentBlock.AddInstructions(opConst)

	// define the constant in this block
	nameStr, err := GetOperandName(name)
	if err != nil {
		log.Fatalf("Error in parseStmtConst: %v", err)
	}
	if builder.currentFunc == builder.Script.Main {
		builder.ConstsDef[nameStr] = val
	}
}

func (buiilder *CFGBuilder) parseStmtConstList(stmt *ast.StmtConstList) {
	for _, c := range stmt.Consts {
		buiilder.parseStmtNode(c)
	}
}

func (builder *CFGBuilder) parseStmtClassConstList(stmt *ast.StmtClassConstList) {
	if builder.currClassOper == nil {
		log.Fatal("Error: Unknown current class for a constants list")
	}

	for _, c := range stmt.Consts {
		builder.parseStmtNode(c)
	}
}

func (builder *CFGBuilder) parseStmtClassMethod(stmt *ast.StmtClassMethod) {
	if builder.currClassOper == nil {
		log.Fatal("Error: Unknown current class for a method")
	}

	name, err := astutils.GetNameString(stmt.Name)
	if err != nil {
		log.Fatal("Error method name in StmtClassMethod")
	}
	flags := builder.parseFuncModifier(stmt.Modifiers, stmt.AmpersandTkn != nil)
	returnType := builder.parseTypeNode(stmt.ReturnType)
	entryBlock := NewBlock(builder.GetBlockIdCount())
	fn, err := NewClassFunc(name, flags, returnType, entryBlock, *builder.currClassOper, stmt.Position)
	if err != nil {
		log.Fatalf("Error in parseStmtClassMethod: %v", err)
	}
	builder.Script.AddFunc(fn)

	// parse function
	stmts, err := astutils.GetStmtList(stmt.Stmt)
	if err != nil {
		log.Fatalf("Error in parseStmtClassMethod: %v", err)
	}
	builder.parseFunc(fn, stmt.Params, stmts)

	// create method op
	visibility := fn.GetVisibility()
	static := fn.IsStatic()
	final := fn.IsFinal()
	abstract := fn.IsAbstract()
	attrs := builder.parseAttributeGroups(stmt.AttrGroups)
	op := NewOpStmtClassMethod(fn, attrs, visibility, static, final, abstract, stmt.Position)
	builder.currentBlock.AddInstructions(op)
	fn.CallableOp = op
}

func (builder *CFGBuilder) parseStmtClass(stmt *ast.StmtClass) {
	name := builder.parseExprNode(stmt.Name)
	prevClass := builder.currClassOper
	builder.currClassOper = name.(*OperandString)
	attrGroups := builder.parseAttributeGroups(stmt.AttrGroups)
	stmts, err := builder.parseStmtNodes(stmt.Stmts, NewBlock(builder.GetBlockIdCount()))
	if err != nil {
		log.Fatalf("Error in parseStmtClass: %v", err)
	}
	modifFlags := builder.parseClassModifier(stmt.Modifiers)
	extends := builder.parseExprNode(stmt.Extends)
	implements, _ := builder.parseExprList(stmt.Implements, PARSER_MODE_NONE)

	op := NewOpStmtClass(name, stmts, modifFlags, extends, implements, attrGroups, stmt.Position)
	builder.currentBlock.AddInstructions(op)

	builder.currClassOper = prevClass
}

func (builder *CFGBuilder) parseStmtIf(vert *ast.StmtIf) {

	endBlock := NewBlock(builder.GetBlockIdCount())
	builder.parseIf(vert, endBlock)
	builder.currentBlock = endBlock

}

func (builder *CFGBuilder) parseIf(stmtVert ast.Vertex, endBlock *Block) {
	var stmts []ast.Vertex
	var err error
	condPosition := &position.Position{}
	cond := Operand(nil)
	switch nType := stmtVert.(type) {
	case *ast.StmtIf:
		condPosition = nType.Cond.GetPosition()
		condNode := builder.parseExprNode(nType.Cond)
		cond, err = builder.readVariable(condNode)
		if err != nil {

			log.Fatalf("parseIf: Error %v", err)
		}
		//parse stmt
		switch stmtT := nType.Stmt.(type) {
		case *ast.StmtStmtList:
			stmts = stmtT.Stmts
		case *ast.StmtExpression:
			stmtExpr := &ast.StmtExpression{
				Position: stmtT.Expr.GetPosition(),
				Expr:     stmtT.Expr,
			}
			stmts = []ast.Vertex{stmtExpr}
		}

	case *ast.StmtElseIf:
		condPosition = nType.Cond.GetPosition()
		condNode := builder.parseExprNode(nType.Cond)
		cond, err = builder.readVariable(condNode)
		if err != nil {

			log.Fatalf("parseIf: Error %v", err)
		}
		//parse stmt
		switch stmtT := nType.Stmt.(type) {
		case *ast.StmtStmtList:
			stmts = stmtT.Stmts
		case *ast.StmtExpression:
			stmtExpr := &ast.StmtExpression{
				Position: stmtT.Expr.GetPosition(),
				Expr:     stmtT.Expr,
			}
			stmts = []ast.Vertex{stmtExpr}
		}
	default:
		log.Fatalf("parseIf: invalid node	")

	}
	ifBlock := NewBlock(builder.GetBlockIdCount())
	ifBlock.AddPredecessor(builder.currentBlock)
	elseBlock := NewBlock(builder.GetBlockIdCount())
	elseBlock.AddPredecessor(builder.currentBlock)

	jmpIf := NewOpStmtJumpIf(cond, ifBlock, elseBlock, condPosition)
	builder.currentBlock.AddInstructions(jmpIf)
	builder.currentBlock.IsConditionalBlock = true
	builder.processAssertion(cond, ifBlock, elseBlock)

	builder.FuncContex.PushCond(cond)
	ifBlock.SetCondition(builder.FuncContex.CurrConds)
	builder.currentBlock, err = builder.parseStmtNodes(stmts, ifBlock)
	if err != nil {
		log.Fatalf("Error in parseIf: %v", err)
	}
	builder.FuncContex.PopCond()

	jmp := NewOpStmtJump(endBlock, stmtVert.GetPosition())
	builder.currentBlock.AddInstructions(jmp)
	endBlock.AddPredecessor(builder.currentBlock)
	builder.currentBlock = elseBlock

	if ifNode, ok := stmtVert.(*ast.StmtIf); ok {
		for _, elseIfNode := range ifNode.ElseIf {
			builder.parseIf(elseIfNode, endBlock)
		}
		if ifNode.Else != nil {
			// else if
			if elseIfNode, ok := ifNode.Else.(*ast.StmtElse).Stmt.(*ast.StmtIf); ok {
				builder.parseIf(elseIfNode, endBlock)
				return
			}

			stmts, err := astutils.GetStmtList(ifNode.Else.(*ast.StmtElse).Stmt)
			if err != nil {
				log.Fatalf("Error in parseIf: %v", err)
			}

			// add condition
			negatedCond := NewOpExprBooleanNot(cond, condPosition).Result
			builder.FuncContex.PushCond(negatedCond)
			elseBlock.SetCondition(builder.FuncContex.CurrConds)

			builder.currentBlock, err = builder.parseStmtNodes(stmts, builder.currentBlock)
			if err != nil {
				log.Fatalf("Error in parseIf: %v", err)
			}

			builder.FuncContex.PopCond()
		}
		jmp := NewOpStmtJump(endBlock, ifNode.Position)
		builder.currentBlock.AddInstructions(jmp)
		endBlock.AddPredecessor(builder.currentBlock)
	}
}
