package cfg

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/position"
	"github.com/rxhunter00/XSS-Taint/pkg/asttraverser/astutils"
)

func (builder *CFGBuilder) parseExprNode(exprVertex ast.Vertex) Operand {
	if exprVertex == nil {
		return nil
	}

	switch exprT := exprVertex.(type) {
	case *ast.ScalarDnumber:
		return builder.parseScalarDnumber(exprT)
	case *ast.ScalarLnumber:
		return builder.parseScalarLnumber(exprT)
	case *ast.ScalarString:
		return builder.parseScalarString(exprT)
	case *ast.Name, *ast.NameFullyQualified, *ast.NameRelative, *ast.Identifier:
		nameStr, _ := astutils.GetNameString(exprT)
		return NewOperandString(nameStr)
	case *ast.ScalarEncapsed:
		// Can result into new Op
		parts, partsPos := builder.parseExprList(exprT.Parts, PARSER_MODE_READ)
		op := NewOpExprConcatList(parts, partsPos, exprT.Position)
		builder.currentBlock.Instructions = append(builder.currentBlock.Instructions, op)
		return op.Result
	case *ast.ScalarEncapsedStringVar:
		return NewOperandString("")
	case *ast.ScalarEncapsedStringPart:
		return builder.parseScalarEncapsedStringPart(exprT)
	case *ast.ScalarEncapsedStringBrackets:
		return builder.parseExprNode(exprT.Var)
	case *ast.ScalarHeredoc:
		return builder.parseScalarhereDoc(exprT)
	case *ast.ExprVariable:
		return builder.parseExprVariable(exprT)
	case *ast.Argument:
		return builder.parseArgument(exprT)
	case *ast.ExprAssign:
		return builder.parseExprAssign(exprT)
	case *ast.ExprAssignReference:
		return builder.parseExprAssignReference(exprT)
	case *ast.ExprAssignBitwiseAnd, *ast.ExprAssignBitwiseOr, *ast.ExprAssignBitwiseXor, *ast.ExprAssignCoalesce, *ast.ExprAssignConcat, *ast.ExprAssignDiv, *ast.ExprAssignMinus, *ast.ExprAssignMod, *ast.ExprAssignMul, *ast.ExprAssignPlus, *ast.ExprAssignPow, *ast.ExprAssignShiftLeft, *ast.ExprAssignShiftRight:
		// Assignment with operator such as =%
		return builder.parseExprAssignOperation(exprT)
		// type casting such as (bool)$a
	case *ast.ExprCastArray, *ast.ExprCastBool, *ast.ExprCastDouble, *ast.ExprCastInt, *ast.ExprCastObject, *ast.ExprCastString, *ast.ExprCastUnset:
		return builder.parseExprCast(exprT)
	case *ast.ExprBinaryBitwiseAnd, *ast.ExprBinaryBitwiseOr, *ast.ExprBinaryBitwiseXor,
		*ast.ExprBinaryBooleanAnd, *ast.ExprBinaryBooleanOr,
		*ast.ExprBinaryCoalesce, *ast.ExprBinaryConcat,
		*ast.ExprBinaryEqual, *ast.ExprBinaryNotEqual, *ast.ExprBinaryIdentical, *ast.ExprBinaryNotIdentical,
		*ast.ExprBinaryGreater, *ast.ExprBinaryGreaterOrEqual, *ast.ExprBinarySmaller, *ast.ExprBinarySmallerOrEqual,
		*ast.ExprBinaryLogicalAnd, *ast.ExprBinaryLogicalOr, *ast.ExprBinaryLogicalXor,
		*ast.ExprBinaryMinus, *ast.ExprBinaryMod, *ast.ExprBinaryMul, *ast.ExprBinaryDiv, *ast.ExprBinaryPlus, *ast.ExprBinaryPow,
		*ast.ExprBinaryShiftLeft, *ast.ExprBinaryShiftRight, *ast.ExprBinarySpaceship:
		return builder.parseExprBinaryLogical(exprT)
	case *ast.ExprUnaryPlus:
		return builder.parseUnaryPlus(exprT)
	case *ast.ExprUnaryMinus:
		return builder.parseUnaryMinus(exprT)
	case *ast.ExprArray:
		return builder.parseExprArray(exprT)
	case *ast.ExprArrayDimFetch:
		return builder.parseExprArrayDimFetch(exprT)
	case *ast.ExprArrowFunction:
		return builder.parseExprArrowFunction(exprT)
	case *ast.ExprClosure:
		return builder.parseExprClosure(exprT)
	case *ast.ExprBrackets:
		return builder.parseExprNode(exprT.Expr)
	case *ast.ExprBitwiseNot:
		oper, err := builder.readVariable(builder.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprBitwiseNot: %v", err)
		}
		op := NewOpExprBitwiseNot(oper, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBooleanNot:
		cond, err := builder.readVariable(builder.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprBooleanNot: %v", err)
		}
		op := NewOpExprBooleanNot(cond, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprConstFetch:
		return builder.parseExprConstFetch(exprT)
	case *ast.ExprClassConstFetch:
		class, err := builder.readVariable(builder.parseExprNode(exprT.Class))
		if err != nil {
			log.Fatalf("Error in ExprClassConstFetch (class): %v", err)
		}
		name, err := builder.readVariable(builder.parseExprNode(exprT.Class))
		if err != nil {
			log.Fatalf("Error in ExprClassConstFetch (name): %v", err)
		}
		op := NewOpExprClassConstFetch(class, name, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprClone:
		clone, err := builder.readVariable(builder.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprClone: %v", err)
		}
		op := NewOpExprClone(clone, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprEmpty:
		empty, err := builder.readVariable(builder.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprEmpty: %v", err)
		}
		op := NewOpExprEmpty(empty, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprErrorSuppress:
		return builder.parseExprErrorSuppress(exprT)
	case *ast.ExprEval:
		eval, err := builder.readVariable(builder.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprEval: %v", err)
		}
		op := NewOpExprEval(eval, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprExit:
		return builder.parseExprExit(exprT)
	case *ast.ExprFunctionCall:
		return builder.parseExprFuncCall(exprT)
	case *ast.ExprInclude:
		include, err := builder.readVariable(builder.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprInclude: %v", err)
		}

		if includeStr, ok := include.(*OperandString); ok {
			builder.Script.IncludeFiles = append(builder.Script.IncludeFiles, includeStr.Val)
		}
		op := NewOpExprInclude(include, TYPE_INCLUDE, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprIncludeOnce:
		include, err := builder.readVariable(builder.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprInclude: %v", err)
		}

		if includeStr, ok := include.(*OperandString); ok {
			builder.Script.IncludeFiles = append(builder.Script.IncludeFiles, includeStr.Val)
		}
		op := NewOpExprInclude(include, TYPE_INCLUDE_ONCE, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprRequire:
		include, err := builder.readVariable(builder.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprInclude: %v", err)
		}
		// add to include file
		if includeStr, ok := include.(*OperandString); ok {
			builder.Script.IncludeFiles = append(builder.Script.IncludeFiles, includeStr.Val)
		}
		op := NewOpExprInclude(include, TYPE_REQUIRE, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprRequireOnce:
		include, err := builder.readVariable(builder.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprInclude: %v", err)
		}
		// add to include file
		if includeStr, ok := include.(*OperandString); ok {
			builder.Script.IncludeFiles = append(builder.Script.IncludeFiles, includeStr.Val)
		}
		op := NewOpExprInclude(include, TYPE_REQUIRE_ONCE, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprInstanceOf:
		vr, err := builder.readVariable(builder.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprInstanceOf (var): %v", err)
		}
		class, err := builder.readVariable(builder.parseExprNode(exprT.Class))
		if err != nil {
			log.Fatalf("Error in ExprInstanceOf (class): %v", err)
		}
		op := NewOpExprInstanceOf(vr, class, exprT.Position)
		op.Result.AddAssertion(vr, NewTypeAssertion(class, false), ASSERTION_MODE_INTERSECTION)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprIsset:
		isset, _ := builder.parseExprList(exprT.Vars, PARSER_MODE_READ)
		op := NewOpExprIsset(isset, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprMethodCall:
		vr, err := builder.readVariable(builder.parseExprNode(exprT.Var))
		if err != nil {
			log.Fatalf("Error in ExprMethodCall (var): %v", err)
		}
		name, err := builder.readVariable(builder.parseExprNode(exprT.Method))
		if err != nil {
			log.Fatalf("Error in ExprMethodCall (name): %v", err)
		}
		args, argsPos := builder.parseExprList(exprT.Args, PARSER_MODE_READ)
		op := NewOpExprMethodCall(vr, name, args, exprT.Var.GetPosition(), exprT.Method.GetPosition(), argsPos, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		builder.currentFunc.Calls = append(builder.currentFunc.Calls, op)
		return op.Result
	case *ast.ExprNullsafeMethodCall:
		vr, err := builder.readVariable(builder.parseExprNode(exprT.Var))
		if err != nil {
			log.Fatalf("Error in ExprMethodCall (var): %v", err)
		}
		name, err := builder.readVariable(builder.parseExprNode(exprT.Method))
		if err != nil {
			log.Fatalf("Error in ExprMethodCall (name): %v", err)
		}
		args, argsPos := builder.parseExprList(exprT.Args, PARSER_MODE_READ)
		op := NewOpExprNullSafeMethodCall(vr, name, args, exprT.Var.GetPosition(), exprT.Method.GetPosition(), argsPos, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		builder.currentFunc.Calls = append(builder.currentFunc.Calls, op)
		return op.Result
	case *ast.ExprPostDec:
		vr := builder.parseExprNode(exprT.Var)
		read, err := builder.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprPostDec: %v", err)
		}
		write := builder.writeVariable(vr)
		opMinus := NewOpExprBinaryMinus(read, NewOperandNumber(1), exprT.Position)
		opAssign := NewOpExprAssign(write, opMinus.Result, exprT.Var.GetPosition(), opMinus.Position, exprT.Position)
		builder.currentBlock.AddInstructions(opMinus)
		builder.currentBlock.AddInstructions(opAssign)
		return read
	case *ast.ExprPostInc:
		vr := builder.parseExprNode(exprT.Var)
		read, err := builder.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprPostInc: %v", err)
		}
		write := builder.writeVariable(vr)
		opPlus := NewOpExprBinaryPlus(read, NewOperandNumber(1), exprT.Position)
		opAssign := NewOpExprAssign(write, opPlus.Result, exprT.Var.GetPosition(), opPlus.Position, exprT.Position)
		builder.currentBlock.AddInstructions(opPlus)
		builder.currentBlock.AddInstructions(opAssign)
		return read
	case *ast.ExprPreDec:
		vr := builder.parseExprNode(exprT.Var)
		read, err := builder.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprPreDec: %v", err)
		}
		write := builder.writeVariable(vr)
		opMinus := NewOpExprBinaryMinus(read, NewOperandNumber(1), exprT.Position)
		opAssign := NewOpExprAssign(write, opMinus.Result, exprT.Var.GetPosition(), opMinus.Position, exprT.Position)
		builder.currentBlock.AddInstructions(opMinus)
		builder.currentBlock.AddInstructions(opAssign)
		return opMinus.Result
	case *ast.ExprPreInc:
		vr := builder.parseExprNode(exprT.Var)
		read, err := builder.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprPreInc: %v", err)
		}
		write := builder.writeVariable(vr)
		opPlus := NewOpExprBinaryPlus(read, NewOperandNumber(1), exprT.Position)
		opAssign := NewOpExprAssign(write, opPlus.Result, exprT.Var.GetPosition(), opPlus.Position, exprT.Position)
		builder.currentBlock.AddInstructions(opPlus)
		builder.currentBlock.AddInstructions(opAssign)
		return opPlus.Result
	case *ast.ExprNew:
		return builder.parseExprNew(exprT)
	case *ast.ExprTernary:
		return builder.parseExprTernary(exprT)
	case *ast.ExprYield:
		return builder.parseExprYield(exprT)
	case *ast.ExprShellExec:
		args, argsPos := builder.parseExprList(exprT.Parts, PARSER_MODE_READ)
		argOp := NewOpExprConcatList(args, argsPos, exprT.Position)
		builder.currentBlock.AddInstructions(argOp)
		funcCallOp := NewOpExprFunctionCall(NewOperandString("shell_exec"), []Operand{argOp.Result}, exprT.Position, argsPos, exprT.Position)
		builder.currentBlock.AddInstructions(funcCallOp)
		return argOp.Result
	case *ast.ExprPrint:
		print, err := builder.readVariable(builder.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprPrint: %v", err)
		}
		op := NewOpExprPrint(print, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprPropertyFetch:
		vr, err := builder.readVariable(builder.parseExprNode(exprT.Var))
		if err != nil {
			log.Fatalf("Error in ExprPropertyFetch (var): %v", err)
		}
		prop, err := builder.readVariable(builder.parseExprNode(exprT.Prop))
		if err != nil {
			log.Fatalf("Error in ExprPropertyFetch (name): %v", err)
		}
		op := NewOpExprPropertyFetch(vr, prop, exprT.Position)

		varName, _ := GetOperandName(vr)
		propStr, ok := GetOperVal(prop).(*OperandString)
		if varName != "" && ok {
			propFetchName := "<propfetch>" + varName[1:] + "->" + propStr.Val
			op.Result = NewOperandVariable(NewOperandString(propFetchName), nil)
		}

		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprNullsafePropertyFetch:
		vr, err := builder.readVariable(builder.parseExprNode(exprT.Var))
		if err != nil {
			log.Fatalf("Error in ExprPropertyFetch (var): %v", err)
		}
		prop, err := builder.readVariable(builder.parseExprNode(exprT.Prop))
		if err != nil {
			log.Fatalf("Error in ExprPropertyFetch (name): %v", err)
		}
		op := NewOpExprPropertyFetch(vr, prop, exprT.Position)
		builder.currentBlock.AddInstructions(op)

		varName, _ := GetOperandName(vr)
		propStr, ok := GetOperVal(prop).(*OperandString)
		if varName != "" && ok {
			propFetchName := "<propfetch>" + varName[1:] + "->" + propStr.Val
			op.Result = NewOperandVariable(NewOperandString(propFetchName), nil)
		}
	case *ast.ExprStaticPropertyFetch:
		classVar, err := builder.readVariable(builder.parseExprNode(exprT.Class))
		if err != nil {
			log.Fatalf("Error in ExprStaticCall (class): %v", err)
		}
		prop, err := builder.readVariable(builder.parseExprNode(exprT.Prop))
		if err != nil {
			log.Fatalf("Error in ExprStaticCall (name): %v", err)
		}
		op := NewOpExprStaticPropertyFetch(classVar, prop, exprT.Position)

		className, _ := GetOperandName(classVar)
		propStr, ok := GetOperVal(prop).(*OperandString)
		if className != "" && ok {
			propFetchName := "<staticpropfetch>" + className[1:] + "->" + propStr.Val
			op.Result = NewOperandVariable(NewOperandString(propFetchName), nil)
		}

		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprStaticCall:
		class, err := builder.readVariable(builder.parseExprNode(exprT.Class))
		if err != nil {
			log.Fatalf("Error in ExprStaticCall (class): %v", err)
		}
		name, err := builder.readVariable(builder.parseExprNode(exprT.Call))
		if err != nil {
			log.Fatalf("Error in ExprStaticCall (name): %v", err)
		}
		args, argsPos := builder.parseExprList(exprT.Args, PARSER_MODE_READ)
		op := NewOpExprStaticCall(class, name, args, exprT.Class.GetPosition(), exprT.Call.GetPosition(), argsPos, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprMatch, *ast.ExprYieldFrom, *ast.ExprThrow:
		log.Fatal("Error: Cannot parse expression node, wrong type '", reflect.TypeOf(exprVertex), "'")
	default:
		log.Printf("%+v", exprT)
		log.Fatalf("Error: Cannot parse expression node, wrong type %v'\n", reflect.TypeOf(exprVertex))
	}

	return nil
}

func (builder *CFGBuilder) parseExprTernary(expr *ast.ExprTernary) Operand {
	cond, err := builder.readVariable(builder.parseExprNode(expr.Cond))
	if err != nil {
		log.Fatalf("Error in parseExprTernary (cond): %v", err)
	}
	ifBlock := NewBlock(builder.GetBlockIdCount())
	elseBlock := NewBlock(builder.GetBlockIdCount())
	endBlock := NewBlock(builder.GetBlockIdCount())

	jmpIf := NewOpStmtJumpIf(cond, ifBlock, elseBlock, expr.Position)
	builder.currentBlock.AddInstructions(jmpIf)
	builder.currentBlock.IsConditionalBlock = true
	builder.processAssertion(cond, ifBlock, elseBlock)
	ifBlock.AddPredecessor(builder.currentBlock)
	elseBlock.AddPredecessor(builder.currentBlock)

	// add condition to if block
	builder.FuncContex.PushCond(cond)
	ifBlock.SetCondition(builder.FuncContex.CurrConds)
	// build ifTrue block
	builder.currentBlock = ifBlock
	ifVar := NewTemporaryOperand(nil)
	var ifAssignOp *OpExprAssign
	// if there is ifTrue value, assign ifVar with it
	// else, assign with 1
	if expr.IfTrue != nil {
		ifVal, err := builder.readVariable(builder.parseExprNode(expr.IfTrue))
		if err != nil {
			log.Fatalf("Error in parseExprTernary (if): %v", err)
		}
		ifAssignOp = NewOpExprAssign(ifVar, ifVal, nil, expr.IfTrue.GetPosition(), expr.Position)
	} else {
		ifAssignOp = NewOpExprAssign(ifVar, NewOperandNumber(1), nil, expr.Position, expr.Position)
	}
	builder.currentBlock.AddInstructions(ifAssignOp)
	// add jump op to end block
	jmp := NewOpStmtJump(endBlock, expr.Position)
	builder.currentBlock.AddInstructions(jmp)
	// return the condition
	builder.FuncContex.PopCond()

	// add condition to else block
	negatedCond := NewOpExprBooleanNot(cond, nil).Result
	builder.FuncContex.PushCond(negatedCond)
	elseBlock.SetCondition(builder.FuncContex.CurrConds)
	// build ifFalse block
	builder.currentBlock = elseBlock
	elseVar := NewTemporaryOperand(nil)
	elseVal, err := builder.readVariable(builder.parseExprNode(expr.IfFalse))
	if err != nil {
		log.Fatalf("Error in parseExprTernary (else): %v", err)
	}
	elseAssignOp := NewOpExprAssign(elseVar, elseVal, nil, expr.IfFalse.GetPosition(), expr.Position)
	builder.currentBlock.AddInstructions(elseAssignOp)
	// add jump to end block
	jmp = NewOpStmtJump(endBlock, expr.Position)
	builder.currentBlock.AddInstructions(jmp)
	endBlock.AddPredecessor(builder.currentBlock)
	// return else block
	builder.FuncContex.PopCond()

	// build end block
	builder.currentBlock = endBlock
	result := NewTemporaryOperand(nil)
	phi := NewOpPhi(result, builder.currentBlock, expr.Position)
	phi.AddOperandtoPhi(ifVar)
	phi.AddOperandtoPhi(elseVar)
	builder.currentBlock.AddPhi(phi)

	// return phi
	return result
}

func (builder *CFGBuilder) parseExprYield(expr *ast.ExprYield) Operand {
	var key Operand
	var val Operand
	var err error

	if expr.Key != nil {
		key, err = builder.readVariable(builder.parseExprNode(expr.Key))
		if err != nil {
			log.Fatalf("Error in parseExprYield (key): %v", err)
		}
	}
	if expr.Val != nil {
		val, err = builder.readVariable(builder.parseExprNode(expr.Val))
		if err != nil {
			log.Fatalf("Error in parseExprYield (val): %v", err)
		}
	}

	yieldOp := NewOpExprYield(val, key, expr.Position)
	builder.currentBlock.AddInstructions(yieldOp)

	return yieldOp.Result
}
func (cb *CFGBuilder) parseExprNew(expr *ast.ExprNew) Operand {
	var className Operand
	switch ec := expr.Class.(type) {
	case *ast.StmtClass:
		// anonymous class
		className = cb.parseExprNode(ec.Name)
	default:
		className = cb.parseExprNode(ec)
	}

	args, _ := cb.parseExprList(expr.Args, PARSER_MODE_READ)
	opNew := NewOpExprNew(className, args, expr.Position)
	cb.currentBlock.AddInstructions(opNew)

	// set result type to object operand
	if _, isString := className.(*OperandString); isString {
		opNew.Result = NewOperandObject(className.(*OperandString).Val)
	}

	return opNew.Result
}
func (cb *CFGBuilder) parseExprFuncCall(expr *ast.ExprFunctionCall) Operand {
	args, argsPos := cb.parseExprList(expr.Args, PARSER_MODE_READ)
	nameNode := cb.parseExprNode(expr.Function)
	functionName, err := cb.readVariable(nameNode)
	if err != nil {
		log.Fatalf("Error in parseExprFuncCall (name): %v", err)
	}

	// Adding read ref for function name, and argument
	opFuncCall := NewOpExprFunctionCall(functionName, args, expr.Function.GetPosition(), argsPos, expr.Position)

	// Only handle assertion type
	if nameStr, ok := functionName.(*OperandString); ok {
		if assertionType, ok := GetTypeAssertFunc(nameStr.Val); ok {
			assert := NewTypeAssertion(NewOperandString(assertionType), false)
			opFuncCall.Result.AddAssertion(args[0], assert, ASSERTION_MODE_INTERSECTION)
		} else if nameStr.Val == "settype" {
			read, err := cb.readVariable(opFuncCall.Args[0])
			if err != nil {
				log.Fatalf("Error in ExprFuncCall: %v", err)
			}
			write := cb.writeVariable(opFuncCall.Args[0])
			tp := opFuncCall.Args[1]
			if tpStr, ok := GetOperVal(tp).(*OperandString); ok {
				switch tpStr.Val {
				case "boolean", "bool":
					op := NewOpExprCastBool(read, nil)
					cb.currentBlock.AddInstructions(op)
					assign := NewOpExprAssign(write, op.Result, nil, op.Position, nil)
					cb.currentBlock.AddInstructions(assign)
				case "integer", "int":
					op := NewOpExprCastInt(read, nil)
					cb.currentBlock.AddInstructions(op)
					assign := NewOpExprAssign(write, op.Result, nil, op.Position, nil)
					cb.currentBlock.AddInstructions(assign)
				case "float", "double":
					op := NewOpExprCastDouble(read, nil)
					cb.currentBlock.AddInstructions(op)
					assign := NewOpExprAssign(write, op.Result, nil, op.Position, nil)
					cb.currentBlock.AddInstructions(assign)
				case "string":
					op := NewOpExprCastString(read, nil)
					cb.currentBlock.AddInstructions(op)
					assign := NewOpExprAssign(write, op.Result, nil, op.Position, nil)
					cb.currentBlock.AddInstructions(assign)
				case "array":
					op := NewOpExprCastArray(read, nil)
					cb.currentBlock.AddInstructions(op)
					assign := NewOpExprAssign(write, op.Result, nil, op.Position, nil)
					cb.currentBlock.AddInstructions(assign)
				case "object":
					op := NewOpExprCastObject(read, nil)
					cb.currentBlock.AddInstructions(op)
					assign := NewOpExprAssign(write, op.Result, nil, op.Position, nil)
					cb.currentBlock.AddInstructions(assign)
				case "null":
					op := NewOpExprCastUnset(read, nil)
					cb.currentBlock.AddInstructions(op)
					assign := NewOpExprAssign(write, op.Result, nil, op.Position, nil)
					cb.currentBlock.AddInstructions(assign)
				}
			}
		}
	}

	cb.currentBlock.AddInstructions(opFuncCall)
	cb.currentFunc.Calls = append(cb.currentFunc.Calls, opFuncCall)

	return opFuncCall.Result
}

func (builder *CFGBuilder) parseExprExit(expr *ast.ExprExit) Operand {
	var e Operand = nil
	var err error
	if expr.Expr != nil {
		e, err = builder.readVariable(builder.parseExprNode(expr.Expr))
		if err != nil {
			log.Fatalf("Error in parseExprExit (expr): %v", err)
		}
	}

	// create exit op
	exitOp := NewOpExit(e, expr.Position)
	builder.currentBlock.AddInstructions(exitOp)
	// ignore all code after exit
	builder.currentBlock = NewBlock(builder.GetBlockIdCount())
	builder.currentBlock.Dead = true

	return NewOperandNumber(1)
}

func (builder *CFGBuilder) parseExprErrorSuppress(expr *ast.ExprErrorSuppress) Operand {
	// create new error supress block
	errSupressBlock := NewBlock(builder.GetBlockIdCount())
	// add instruction to jump into error supress block
	jmp := NewOpStmtJump(errSupressBlock, expr.Position)
	builder.currentBlock.AddInstructions(jmp)
	errSupressBlock.AddPredecessor(builder.currentBlock)
	builder.currentBlock = errSupressBlock

	// parse expression
	result := builder.parseExprNode(expr.Expr)
	// create new block as end block
	endBlock := NewBlock(builder.GetBlockIdCount())
	jmp = NewOpStmtJump(endBlock, expr.Position)
	builder.currentBlock.AddInstructions(jmp)
	endBlock.AddPredecessor(builder.currentBlock)
	builder.currentBlock = endBlock

	return result
}
func (builder *CFGBuilder) parseExprConstFetch(expr *ast.ExprConstFetch) Operand {
	nameStr, err := astutils.GetNameString(expr.Const)
	if err != nil {
		log.Fatal("Error const name in ExprConstFetch")
	}
	lowerName := strings.ToLower(nameStr)
	switch lowerName {
	case "null":
		return NewOperandNull()
	case "true":
		return NewOperandBool(true)
	case "false":
		return NewOperandBool(false)
	}

	name := builder.parseExprNode(expr.Const)
	op := NewOpExprConstFetch(name, expr.Position)
	builder.currentBlock.AddInstructions(op)

	// find the constant definition
	if val, ok := builder.ConstsDef[nameStr]; ok {
		op.Result = val
	}

	return op.Result
}
func (builder *CFGBuilder) parseExprClosure(expr *ast.ExprClosure) Operand {
	// Example <? function($a, $b) use (&$c, $d) {}
	// Parse each Variable in Use
	uses := make([]Operand, len(expr.Uses))
	for i, exprUse := range expr.Uses {
		eu := exprUse.(*ast.ExprClosureUse)
		nodeVar := builder.parseExprNode(eu.Var)
		nameVar, err := builder.readVariable(nodeVar)
		if err != nil {
			log.Fatalf("Error in parseExprClosure: %v", err)
		}
		useByRef := eu.AmpersandTkn != nil
		// Change to Variable inside the closure does not affect the original
		uses[i] = NewOperandBoundVariable(nameVar, NewOperandNull(), BOUND_VAR_SCOPE_LOCAL, useByRef, nil)
	}

	// Create function
	byRef := expr.AmpersandTkn != nil
	isStatic := expr.StaticTkn != nil
	name := fmt.Sprintf("{anonymous}#%d", builder.GetAnonIdCount())
	types := builder.parseTypeNode(expr.ReturnType)
	entryBlock := NewBlock(builder.GetBlockIdCount())
	opFunc, err := NewFunc(name, FUNC_MODIF_FLAG_CLOSURE, types, entryBlock, expr.Position)
	if err != nil {
		log.Fatalf("Error in parseExprClosure: %v", err)
	}
	// Add bit flag
	if byRef {
		opFunc.AddModifier(FUNC_MODIF_FLAG_RETURNS_REF)
	}
	if isStatic {
		opFunc.AddModifier(FUNC_MODIF_FLAG_STATIC)
	}
	builder.currentBlock.AddInstructions(opFunc)

	// Build the CFG
	builder.parseFunc(opFunc, expr.Params, expr.Stmts)
	builder.Script.AddFunc(opFunc)

	// create op closure
	closure := NewOpExprClosure(opFunc, uses, expr.Position)
	opFunc.CallableOp = closure

	builder.currentBlock.AddInstructions(closure)
	return closure.Result
}

func (builder *CFGBuilder) parseExprArrowFunction(expr *ast.ExprArrowFunction) Operand {
	// Create opFunction
	byRef := expr.AmpersandTkn != nil
	isStatic := expr.StaticTkn != nil
	name := fmt.Sprintf("{anonymous}#%d", builder.GetAnonIdCount())
	types := builder.parseTypeNode(expr.ReturnType)
	entryBlock := NewBlock(builder.GetBlockIdCount())
	opFunc, err := NewFunc(name, FUNC_MODIF_FLAG_CLOSURE, types, entryBlock, expr.Position)
	if err != nil {
		log.Fatalf("Error in parseExprClosure: %v", err)
	}
	if byRef {
		opFunc.AddModifier(FUNC_MODIF_FLAG_RETURNS_REF)
	}
	if isStatic {
		opFunc.AddModifier(FUNC_MODIF_FLAG_STATIC)
	}
	builder.currentBlock.AddInstructions(opFunc)

	// build cfg for the closure
	stmtExpr := &ast.StmtExpression{
		Position: expr.Expr.GetPosition(),
		Expr:     expr.Expr,
	}
	stmts := []ast.Vertex{stmtExpr}
	builder.parseFunc(opFunc, expr.Params, stmts)
	builder.Script.AddFunc(opFunc)

	// create op closure
	closure := NewOpExprClosure(opFunc, nil, expr.Position)
	opFunc.CallableOp = closure

	builder.currentBlock.AddInstructions(closure)
	return closure.Result
}

func (builder *CFGBuilder) parseExprArrayDimFetch(expr *ast.ExprArrayDimFetch) Operand {
	varNode := builder.parseExprNode(expr.Var)
	vr, err := builder.readVariable(varNode)
	if err != nil {
		log.Fatalf("parseExprArrayDimFetch: parsing var: %v", err)
	}
	var dim Operand
	if expr.Dim != nil {
		dimNode := builder.parseExprNode(expr.Dim)
		dim, err = builder.readVariable(dimNode)
		if err != nil {
			log.Fatalf("parseExprArrayDimFetch: parsing (dim): %v", err)
		}
	} else {
		dim = NewOperandNull()
	}

	op := NewOpExprArrayDimFetch(vr, dim, expr.Position)
	builder.currentBlock.AddInstructions(op)

	return op.Result
}

func (builder *CFGBuilder) parseExprArray(expr *ast.ExprArray) Operand {
	keys := make([]Operand, 0)
	vals := make([]Operand, 0)
	byRefs := make([]bool, 0)

	if expr.Items != nil {
		for _, arrItem := range expr.Items {
			item, ok := arrItem.(*ast.ExprArrayItem)
			if !ok {
				log.Fatalf("parseExprArray:wrong type vertex %v", reflect.TypeOf(arrItem))
			}
			if item.Val == nil {
				continue
			}

			if item.Key != nil {
				keyNode := builder.parseExprNode(item.Key)
				key, err := builder.readVariable(keyNode)
				if err != nil {
					log.Fatalf("parseExprArray: error parsing key: %v", err)
				}
				keys = append(keys, key)
			} else {
				keys = append(keys, NewOperandNull())
			}
			valNode := builder.parseExprNode(item.Val)
			val, err := builder.readVariable(valNode)
			if err != nil {
				log.Fatalf("parseExprArray: error parsing val: %v", err)
			}
			vals = append(vals, val)

			if item.AmpersandTkn != nil {
				byRefs = append(byRefs, true)
			} else {
				byRefs = append(byRefs, false)
			}
		}
	}

	op := NewOpExprArray(keys, vals, byRefs, expr.Position)
	builder.currentBlock.AddInstructions(op)

	return op.Result
}

func (builder *CFGBuilder) parseUnaryPlus(node *ast.ExprUnaryPlus) Operand {
	exprNode := builder.parseExprNode(node.Expr)
	varOperand, err := builder.readVariable(exprNode)
	if err != nil {
		log.Fatalf("parseUnaryPlus:Error parsing argument %v", err)
	}
	op := NewOpExprUnaryPlus(varOperand, node.Position)
	builder.currentBlock.AddInstructions(op)
	return op.Result
}

func (builder *CFGBuilder) parseUnaryMinus(node *ast.ExprUnaryMinus) Operand {
	exprNode := builder.parseExprNode(node.Expr)
	varOperand, err := builder.readVariable(exprNode)
	if err != nil {
		log.Fatalf("parseUnaryMinus:Error parsing argument %v", err)
	}
	op := NewOpExprUnaryMinus(varOperand, node.Position)
	builder.currentBlock.AddInstructions(op)
	return op.Result
}
func (builder *CFGBuilder) parseExprList(exprs []ast.Vertex, mode ParserMode) ([]Operand, []*position.Position) {
	vars := make([]Operand, 0, len(exprs))
	positions := make([]*position.Position, 0, len(exprs))
	switch mode {
	case PARSER_MODE_READ:
		for _, expr := range exprs {
			exprNode := builder.parseExprNode(expr)
			vr, err := builder.readVariable(exprNode)
			if err != nil {
				log.Fatalf("Error in parseExprList (var): %v", err)
			}
			vars = append(vars, vr)
			positions = append(positions, expr.GetPosition())
		}
	case PARSER_MODE_WRITE:
		for _, expr := range exprs {
			exprNode := builder.parseExprNode(expr)
			vars = append(vars, builder.writeVariable(exprNode))
			positions = append(positions, expr.GetPosition())
		}
	case PARSER_MODE_NONE:
		for _, expr := range exprs {
			exprNode := builder.parseExprNode(expr)
			vars = append(vars, exprNode)
			positions = append(positions, expr.GetPosition())
		}
	}

	return vars, positions
}

// Variable
func (builder *CFGBuilder) parseExprVariable(expr *ast.ExprVariable) Operand {

	varNameString, err := astutils.GetNameString(expr.Name)
	operandName := builder.parseExprNode(expr.Name)
	if varNameString == "this" && err != nil {
		return NewOperandBoundVariable(operandName, nil, BOUND_VAR_SCOPE_OBJECT, false, builder.currClassOper)

	}

	// initialize variable value as nil
	return NewOperandVariable(operandName, nil)

}

// Decimal Number
func (builder *CFGBuilder) parseScalarDnumber(scalarNumNode *ast.ScalarDnumber) Operand {
	var dnum float64
	var err error
	// check if value is a hex
	if string(scalarNumNode.Value[:2]) == "0x" {
		intNum, err := strconv.ParseInt(string(scalarNumNode.Value), 0, 64)
		if err != nil {
			log.Fatalf("parseScalarDnumber:error parsing Int")
		}
		dnum = float64(intNum)
	} else {
		dnum, err = strconv.ParseFloat(string(scalarNumNode.Value), 64)
		if err != nil {
			log.Fatalf("parseScalarDnumber:error parsing float")
		}

	}
	return NewOperandNumber(dnum)
}

// Real number
func (builder *CFGBuilder) parseScalarLnumber(scalarNumNode *ast.ScalarLnumber) Operand {
	var lnum float64
	var err error
	// check if value is a hex
	if string(scalarNumNode.Value[:2]) == "0x" {
		intNum, err := strconv.ParseInt(string(scalarNumNode.Value), 0, 64)
		if err != nil {
			log.Fatalf("parseScalarLnumber:error parsing Int")
		}
		lnum = float64(intNum)
	} else {
		lnum, err = strconv.ParseFloat(string(scalarNumNode.Value), 64)
		if err != nil {
			log.Fatalf("parseScalarLnumber:error parsing float")
		}

	}
	return NewOperandNumber(lnum)
}

// String
func (builder *CFGBuilder) parseScalarString(strNode *ast.ScalarString) Operand {
	strVal := string(strNode.Value)
	return NewOperandString(strVal)
}

func (builder *CFGBuilder) parseScalarEncapsedStringPart(strNode *ast.ScalarEncapsedStringPart) Operand {
	strVal := string(strNode.Value)
	return NewOperandString(strVal)
}

// Scalar heredoc
func (builder *CFGBuilder) parseScalarhereDoc(shdode *ast.ScalarHeredoc) Operand {
	parts, partsPos := builder.parseExprList(shdode.Parts, PARSER_MODE_READ)
	op := NewOpExprConcatList(parts, partsPos, shdode.Position)
	builder.currentBlock.Instructions = append(builder.currentBlock.Instructions, op)
	return op.Result
}

func (builder *CFGBuilder) parseArgument(anode *ast.Argument) Operand {
	expr := builder.parseExprNode(anode.Expr)
	varOperand, err := builder.readVariable(expr)
	if err != nil {
		log.Fatalf("parseArgument:Error parsing argument %v", err)
	}
	return varOperand
}

func (builder *CFGBuilder) parseExprAssign(anode *ast.ExprAssign) Operand {
	/*
		Noteable AST Vertex

		Var      Vertex // Left
		Expr     Vertex // Right
	*/
	// Right
	exprNode := builder.parseExprNode(anode.Expr)
	// Search current definition
	rightOperand, err := builder.readVariable(exprNode)
	if err != nil {
		log.Fatalf("parseExprAssign: Error parsing %v", err)
	}

	//Left
	// Handle Array List
	/*
		list($a, $b, $c) = $arr;
	*/
	switch e := anode.Var.(type) {
	case *ast.ExprList:
		builder.parseAssignList(e.Items, rightOperand, e.Position)
		return rightOperand
	case *ast.ExprArray:
		builder.parseAssignList(e.Items, rightOperand, e.Position)
		return rightOperand
	}

	varNode := builder.parseExprNode(anode.Var)
	leftOperand := builder.writeVariable(varNode)

	// Register Op
	assignOp := NewOpExprAssign(leftOperand, rightOperand, anode.Var.GetPosition(), anode.Expr.GetPosition(), anode.Position)
	builder.currentBlock.AddInstructions(assignOp)

	switch rightOperValue := GetOperVal(rightOperand).(type) {
	// literal
	case *OperandNumber, *OperandString, *OperandBool, *OperandSymbolic, *OperandObject:
		// Result should be the right op
		assignOp.Result = rightOperValue
		// Set value of left operand to right operand value
		SetOperVal(leftOperand, rightOperValue)

	}
	// return result/shoudl be right op
	return assignOp.Result

}

// TODO CHECK Function
func (builder *CFGBuilder) parseAssignList(items []ast.Vertex, arrVar Operand, pos *position.Position) {
	var err error
	counter := 0
	for _, item := range items {
		if item == nil {
			continue
		}
		var key Operand = nil
		arrItem := item.(*ast.ExprArrayItem)
		if arrItem.Val == nil {
			continue
		}

		// if no key, set key to cnt (considered as array)
		if arrItem.Key == nil {
			//create key
			key = NewOperandNumber(float64(counter))
			counter += 1
		} else {
			keyNode := builder.parseExprNode(arrItem.Key)
			key, err = builder.readVariable(keyNode)
			if err != nil {
				log.Fatalf("Error in parseAssignList (key): %v", err)
			}
		}

		// set array's item value
		vr := arrItem.Val
		fetch := NewOpExprArrayDimFetch(arrVar, key, pos)
		builder.currentBlock.AddInstructions(fetch)

		// assign recursively
		switch e := vr.(type) {
		case *ast.ExprList:
			builder.parseAssignList(e.Items, fetch.Result, e.Position)
			continue
		case *ast.ExprArray:
			builder.parseAssignList(e.Items, fetch.Result, e.Position)
			continue
		}

		// assign item with corresponding value
		left := builder.writeVariable(builder.parseExprNode(vr))
		assign := NewOpExprAssign(left, fetch.Result, vr.GetPosition(), fetch.Position, pos)
		builder.currentBlock.AddInstructions(assign)
	}
}

func (builder *CFGBuilder) parseExprAssignReference(arnode *ast.ExprAssignReference) Operand {
	/*
		Noteable AST Vertex

		Var      Vertex // Left
		Expr     Vertex // Right
	*/
	// Right
	left := builder.writeVariable(builder.parseExprNode(arnode.Var))
	right, err := builder.readVariable(builder.parseExprNode(arnode.Expr))
	if err != nil {
		log.Fatalf("Error in parseExprAssignRef: %v", err)
	}

	assign := NewOpExprAssignRef(left, right, arnode.Position)
	return assign.Result

}

func (builder *CFGBuilder) parseExprAssignOperation(vertexAssign ast.Vertex) Operand {
	var vr, e Operand
	var read, write Operand
	var err error

	switch exprT := vertexAssign.(type) {
	case *ast.ExprAssignConcat:
		vr = builder.parseExprNode(exprT.Var)
		read, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("parseExprAssignOperation:Error parsing ExprAssignConcat: %v", err)
		}
		write = builder.writeVariable(vr)
		e = builder.parseExprNode(exprT.Expr)
		// Return OpExprBinaryConcat interface of Op
		op := NewOpExprBinaryConcat(read, e, exprT.Var.GetPosition(), exprT.Expr.GetPosition(), exprT.Position)
		builder.currentBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		builder.currentBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignBitwiseAnd:
		vr = builder.parseExprNode(exprT.Var)
		read, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignBitwiseAnd: %v", err)
		}
		write = builder.writeVariable(vr)
		e = builder.parseExprNode(exprT.Expr)
		// Return OpExprBinaryBitwiseAnd interface of Op
		op := NewOpExprBinaryBitwiseAnd(read, e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		builder.currentBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignBitwiseOr:
		vr = builder.parseExprNode(exprT.Var)
		read, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignBitwiseOr: %v", err)
		}
		write = builder.writeVariable(vr)
		e = builder.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryBitwiseOr(read, e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		builder.currentBlock.AddInstructions(assign)
		return op.Result

	case *ast.ExprAssignBitwiseXor:
		vr = builder.parseExprNode(exprT.Var)
		read, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignBitwiseXor: %v", err)
		}
		write = builder.writeVariable(vr)
		e = builder.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryBitwiseXor(read, e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		builder.currentBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignCoalesce:
		vr = builder.parseExprNode(exprT.Var)
		read, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignCoalesce: %v", err)
		}
		write = builder.writeVariable(vr)
		e = builder.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryCoalesce(read, e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		builder.currentBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignDiv:
		vr = builder.parseExprNode(exprT.Var)
		read, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignBitwiseXor: %v", err)
		}
		write = builder.writeVariable(vr)
		e = builder.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryBitwiseXor(read, e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		builder.currentBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignMinus:
		vr = builder.parseExprNode(exprT.Var)
		read, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignMinus: %v", err)
		}
		write = builder.writeVariable(vr)
		e = builder.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryMinus(read, e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		builder.currentBlock.AddInstructions(assign)
		return op.Result

	case *ast.ExprAssignMod:
		vr = builder.parseExprNode(exprT.Var)
		read, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignMod: %v", err)
		}
		write = builder.writeVariable(vr)
		e = builder.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryMod(read, e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		builder.currentBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignMul:

		vr = builder.parseExprNode(exprT.Var)

		read, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignMul: %v", err)
		}
		write = builder.writeVariable(vr)
		e = builder.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryMul(read, e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		builder.currentBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignPlus:
		vr = builder.parseExprNode(exprT.Var)
		read, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignPlus: %v", err)
		}
		write = builder.writeVariable(vr)
		e = builder.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryPlus(read, e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		builder.currentBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignPow:
		vr = builder.parseExprNode(exprT.Var)
		read, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignpow: %v", err)
		}
		write = builder.writeVariable(vr)
		e = builder.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryPow(read, e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		builder.currentBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignShiftLeft:
		vr = builder.parseExprNode(exprT.Var)
		read, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignShiftLeft: %v", err)
		}
		write = builder.writeVariable(vr)
		e = builder.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryShiftLeft(read, e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		builder.currentBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignShiftRight:
		vr = builder.parseExprNode(exprT.Var)
		read, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignShiftRight: %v", err)
		}
		write = builder.writeVariable(vr)
		e = builder.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryShiftRight(read, e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		builder.currentBlock.AddInstructions(assign)
		return op.Result
	default:
		log.Printf("parseExprAssignOperation: Unhandled Node assigment	")

	}
	return nil

}

// handle type casting
func (builder *CFGBuilder) parseExprCast(vertexCast ast.Vertex) Operand {
	var vr, e Operand
	var err error

	switch exprT := vertexCast.(type) {

	case *ast.ExprCastBool:
		vr = builder.parseExprNode(exprT.Expr)
		e, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("parseExprCast: Error parsing %v", err)
		}
		op := NewOpExprCastBool(e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprCastDouble:
		vr = builder.parseExprNode(exprT.Expr)
		e, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("parseExprCast: Error parsing %v", err)
		}
		op := NewOpExprCastDouble(e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprCastInt:
		vr = builder.parseExprNode(exprT.Expr)
		e, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("parseExprCast: Error parsing %v", err)
		}
		op := NewOpExprCastInt(e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprCastString:
		vr = builder.parseExprNode(exprT.Expr)
		e, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("parseExprCast: Error parsing %v", err)
		}
		op := NewOpExprCastString(e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprCastObject:
		vr = builder.parseExprNode(exprT.Expr)
		e, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("parseExprCast: Error parsing %v", err)
		}
		op := NewOpExprCastObject(e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprCastUnset:
		vr = builder.parseExprNode(exprT.Expr)
		e, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("parseExprCast: Error parsing %v", err)
		}
		op := NewOpExprCastUnset(e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprCastArray:
		vr = builder.parseExprNode(exprT.Expr)
		e, err = builder.readVariable(vr)
		if err != nil {
			log.Fatalf("parseExprCast: Error parsing %v", err)
		}
		op := NewOpExprCastArray(e, exprT.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	default:
		log.Printf("parseExprCast: Unhandled Type Casting")

	}
	return nil

}

func (builder *CFGBuilder) parseExprBinaryLogical(vertexBinary ast.Vertex) Operand {
	switch e := vertexBinary.(type) {
	case *ast.ExprBinaryBitwiseAnd:
		// Handle ExprBinaryBitwiseAnd
		leftNode := builder.parseExprNode(e.Left)
		// Read Definition
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: Parsng left of: %v", err)
		}
		right, err := builder.readVariable(builder.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: Parsng left of: %v", err)
		}
		op := NewOpExprBinaryBitwiseAnd(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryBitwiseOr:
		// Handle ExprBinaryBitwiseOr
		leftNode := builder.parseExprNode(e.Left)

		// Read Definition
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		right, err := builder.readVariable(builder.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryBitwiseOr(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryBitwiseXor:
		// Handle ExprBinaryBitwiseXor
		leftNode := builder.parseExprNode(e.Left)

		// Read Definition
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		right, err := builder.readVariable(builder.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryBitwiseXor(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryBooleanAnd:
		// Handle ExprBinaryBooleanAnd and
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryLogicalAnd(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryBooleanOr:
		// Handle ExprBinaryBooleanOr
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryLogicalOr(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryCoalesce:
		// Handle ExprBinaryCoalesce

		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryCoalesce(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryConcat:
		// Handle ExprBinaryConcat
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryConcat(left, right, e.Left.GetPosition(), e.Right.GetPosition(), e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryEqual:
		// Handle ExprBinaryEqual
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryEqual(left, right, e.Position)

		//Check if any op has been defined
		if left.IsWritten() {
			// TODO Handle Function Get
		}
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryNotEqual:
		// Handle ExprBinaryNotEqual
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryNotEqual(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryIdentical:
		// Handle ExprBinaryIdentical
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryIdentical(left, right, e.Position)

		//Check if any op has been defined
		if left.IsWritten() {
			// TODO Handle Function Get
		}
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryNotIdentical:
		// Handle ExprBinaryNotIdentical
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryNotIdentical(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryGreater:
		// Handle ExprBinaryGreater
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryBigger(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryGreaterOrEqual:
		// Handle ExprBinaryGreaterOrEqual
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryBigger(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinarySmaller:
		// Handle ExprBinarySmaller
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinarySmaller(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinarySmallerOrEqual:
		// Handle ExprBinarySmallerOrEqual
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinarySmallerOrEqual(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryLogicalAnd:
		// Handle ExprBinaryLogicalAnd
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryLogicalAnd(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryLogicalOr:
		// Handle ExprBinaryLogicalOr
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryLogicalOr(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryLogicalXor:
		// Handle ExprBinaryLogicalXor
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryLogicalXor(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryMinus:
		// Handle ExprBinaryMinus
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryMinus(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryMod:
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryMod(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryMul:
		// Handle ExprBinaryMul
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryMul(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryDiv:
		// Handle ExprBinaryDiv
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryDiv(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryPlus:
		// Handle ExprBinaryPlus
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryPlus(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryPow:
		// Handle ExprBinaryPow
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryPow(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryShiftLeft:
		// Handle ExprBinaryShiftLeft
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryShiftLeft(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinaryShiftRight:
		// Handle ExprBinaryShiftRight
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinaryShiftRight(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	case *ast.ExprBinarySpaceship:
		// Handle ExprBinarySpaceship
		leftNode := builder.parseExprNode(e.Left)
		left, err := builder.readVariable(leftNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing left of: %v", err)
		}
		rightNode := builder.parseExprNode(e.Right)
		right, err := builder.readVariable(rightNode)
		if err != nil {
			log.Fatalf("parseExprBinaryLogical: parsing right of: %v", err)
		}
		op := NewOpExprBinarySpaceship(left, right, e.Position)
		builder.currentBlock.AddInstructions(op)
		return op.Result

	default:
		fmt.Println("Unknown expression type")
	}
	return nil

}
