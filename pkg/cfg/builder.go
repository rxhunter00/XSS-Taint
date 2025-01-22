// https://github.com/ircmaxell/php-cfg/blob/master/lib/PHPCfg/Parser.php

package cfg

import (
	"log"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/conf"
	"github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/rxhunter00/XSS-Taint/pkg/asttraverser"
	"github.com/rxhunter00/XSS-Taint/pkg/asttraverser/astutils"
	"github.com/rxhunter00/XSS-Taint/pkg/asttraverser/nodetraverser/loopresolver"
	"github.com/rxhunter00/XSS-Taint/pkg/asttraverser/nodetraverser/magicconstansresolver"
	"github.com/rxhunter00/XSS-Taint/pkg/asttraverser/nodetraverser/namespaceresolver"
)

type ParserMode int

const (
	PARSER_MODE_NONE ParserMode = iota
	PARSER_MODE_READ
	PARSER_MODE_WRITE
)

type CFGBuilder struct {
	Script     *Script
	FuncContex FunctionContex

	VariableNames map[string]struct{}

	ConstsDef      map[string]Operand
	BlockIdCounter int
	AnnonIdCounter int

	CurrClass     *OperandString
	CurrNamespace string
	currentBlock  *Block
	currentFunc   *Func
}

func (builder *CFGBuilder) GetBlockIdCount() int {
	id := builder.BlockIdCounter
	builder.BlockIdCounter += 1
	return id
}
func (builder *CFGBuilder) GetAnonIdCount() int {
	id := builder.AnnonIdCounter
	builder.AnnonIdCounter += 1
	return id
}

func BuildCFG(src []byte, filePath string) *Script {
	builder := &CFGBuilder{
		VariableNames:  make(map[string]struct{}),
		ConstsDef:      make(map[string]Operand),
		BlockIdCounter: 0,
		AnnonIdCounter: 0,
	}
	fileName := filepath.Base(filePath)

	rootNode := builder.parseAST(src, fileName)

	// Start parsing Main function
	entryBlock := NewBlock(builder.GetBlockIdCount())
	mainFunction, err := NewFunc("Main", FUNC_MODIF_FLAG_PUBLIC,
		NewOpTypeVoid(nil), entryBlock, nil)

	if err != nil {
		log.Fatalf("BuildCFG:Error parsing root as main function: %v", err)
	}
	builder.Script = NewScript(mainFunction, filePath)
	builder.parseFunc(mainFunction, nil, rootNode.Stmts)
	return builder.Script
}

func (builder *CFGBuilder) parseAST(src []byte, filename string) *ast.Root {

	var parserErrors []*errors.Error
	errorHandler := func(e *errors.Error) {
		parserErrors = append(parserErrors, e)
	}
	config := conf.Config{
		Version:          &version.Version{Major: 8},
		ErrorHandlerFunc: errorHandler,
	}

	rootNode, err := parser.Parse(src, config)
	if err != nil {
		log.Fatal("parseAST:Error Parsing AST:" + err.Error())
	}
	root, ok := rootNode.(*ast.Root)

	if !ok {
		log.Fatalf("parseAST:Not Root\n")
	}

	astrav := asttraverser.NewTraverser()

	nsRes := namespaceresolver.NewNamespaceResolver()
	loopRes := loopresolver.NewLoopResolver()
	mCRes := magicconstansresolver.NewMagicConstantResolver(filename)

	astrav.AddNodeTraverser(nsRes)
	astrav.AddNodeTraverser(loopRes)
	astrav.AddNodeTraverser(mCRes)

	astrav.Traverse(root)

	return root

}

func (builder *CFGBuilder) parseFunc(functionF *Func, functionParams []ast.Vertex, functionStmt []ast.Vertex) {
	// Each function should be defined in own block

	// Switch builder function to the current function and save to prev value
	prevFunc := builder.currentFunc
	builder.currentFunc = functionF

	// Switch builder context to the current function a context
	// This is the context of current state in the function
	prevFuncContex := builder.FuncContex
	builder.FuncContex = NewFunctionContex()

	// We enter a function, so we will use the function block
	entryBlock := functionF.CFGBlock
	prevBlock := builder.currentBlock
	builder.currentBlock = entryBlock

	// Handle Function Parameter
	for _, paramVertex := range functionParams {

		param, ok := paramVertex.(*ast.Parameter)
		if !ok {
			log.Fatalf("parseFunc: Not a type of parameter")
		}
		var defaultVarOperand Operand = nil
		var defaultBlock *Block = nil
		if param.DefaultValue != nil {
			tempBlock := builder.currentBlock
			defaultBlock := NewBlock(builder.GetBlockIdCount())
			builder.currentBlock = defaultBlock
			defaultVarOperand = builder.parseExprNode(param.DefaultValue)
			builder.currentBlock = tempBlock

		}
		paramType := builder.parseTypeNode(param.Type)
		byRef := param.AmpersandTkn != nil
		paramName := builder.parseExprNode(param.Var.(*ast.ExprVariable).Name).(*OperandString)
		isVariadic := param.VariadicTkn != nil
		paramAttr := builder.parseAttributeGroups(param.AttrGroups)

		opParam := NewOpExprParam(
			paramName,
			byRef,
			isVariadic,
			paramAttr,
			defaultVarOperand,
			defaultBlock,
			paramType,
			param.Position,
		)

		opParam.Result.(*TemporaryOperand).Original = NewOperandVariable(paramName, nil)

		functionF.Params = append(functionF.Params, opParam)

		// Write param name in current scope
		builder.writeVariableName(paramName.Val, opParam.Result, entryBlock)
		entryBlock.AddInstructions(opParam)

	}

	endBlock, err := builder.parseStmtNodes(functionStmt, entryBlock) // it can create some blocks
	if err != nil {
		log.Fatalf("parseFunc: Error %v", err)
	}

	builder.currentBlock = prevBlock
	if endBlock.Dead {
		endBlock.AddInstructions(NewOpReturn(nil, nil))
	}

	builder.FuncContex.IsComplete = true
	// resolve all incomplete phis
	for block := range builder.FuncContex.IncompletePhis {
		for name, phi := range builder.FuncContex.IncompletePhis[block] {
			for _, pred := range block.Predecesors {

				if !pred.Dead {
					vr := builder.readVariableName(name, pred)
					phi.AddOperand(vr)
				}
			}
			// append complete phi to the list
			block.AddPhi(phi)
		}
	}
	builder.currentFunc = prevFunc
	builder.FuncContex = prevFuncContex

}

func (builder *CFGBuilder) parseTypeNode(parType ast.Vertex) OpType {
	switch parT := parType.(type) {
	case nil:
		// undefine type such as $test
		return NewOpTypeMixed(nil)
	case *ast.Name, *ast.NameFullyQualified:
		typename, _ := astutils.GetNameString(parT)
		if typename == "mixed" {
			return NewOpTypeMixed(parT.GetPosition())
		} else if typename == "void" {
			return NewOpTypeVoid(parT.GetPosition())
		} else if IsBuiltInType(typename) {
			return NewOpTypeLiteral(typename, false, parT.GetPosition())
		} else {
			//reference here
			return nil
		}
	case *ast.Nullable:
		// In case nullable parameter such as ?float
		// We go deep into the tree and get the type such as in case 1
		nullabletype := builder.parseTypeNode(parT.Expr)
		switch ntType := nullabletype.(type) {
		case *OpTypeReference:
			ntType.IsNullable = true
			return ntType
		case *OpTypeLiteral:
			ntType.IsNullable = true
			return ntType
		default:
			log.Fatalf("parseTypeNode:Error Parsing ast.Nulllable")

		}
	case *ast.Union:
		// In case union parameter such as float|int
		// We get all possible type such as float and int
		resTypes := make([]OpType, 0)
		for _, uType := range parT.Types {
			parsedType := builder.parseTypeNode(uType)
			resTypes = append(resTypes, parsedType)
		}
		return NewOpTypeUnion(resTypes, parT.Position)
	case *ast.Identifier:
		return NewOpTypeLiteral(string(parT.Value), false, parT.Position)
	default:
		log.Fatalf("parseTypeNode: Invalid Unhandle Type %v", reflect.TypeOf(parT))

	}
	return nil
}

/*
Example of AttributeGroup
https://github.com/VKCOM/php-parser/blob/master/internal/php8/parser_php8_test.go

public function createUser(

	#[ValidateLength(5, 20), ValidateEmail, ValidateRequired] string $username,
	#[ValidateEmail, ValidateRequired] string $email

	) {
	    echo "User created with username: $username";
	}

Attribute Group = #[ValidateLength(5, 20), ValidateEmail, ValidateRequired]
Atrribute = ValidateLength(5, 20)
*/
func (builder *CFGBuilder) parseAttributeGroups(attrGroups []ast.Vertex) []*OpAttributeGroup {
	resAttrGroup := make([]*OpAttributeGroup, 0)
	// For each attribute group
	// Haven't seen case where attrGroups len is >1
	for _, attrGroup := range attrGroups {
		attributeGroupNode, _ := attrGroup.(*ast.AttributeGroup)
		tempAttrGroup := builder.parseAttributeGroup(attributeGroupNode)
		resAttrGroup = append(resAttrGroup, tempAttrGroup)
	}
	return resAttrGroup

}

// #[ValidateLength(5, 20), ValidateEmail, ValidateRequired]
func (builder *CFGBuilder) parseAttributeGroup(attrGroup *ast.AttributeGroup) *OpAttributeGroup {
	//  Array of Attribute
	attrArray := make([]*OpAttribute, 0)
	for _, attributeVertex := range attrGroup.Attrs {
		attributeArgs := make([]Operand, 0)
		attributeNode, _ := attributeVertex.(*ast.Attribute)
		// Parse Argument

		for _, argVertex := range attributeNode.Args {
			// For each arguments in attribute
			argNode := builder.parseExprNode(argVertex.(*ast.Argument).Expr)
			arg, err := builder.readVariable(argNode)
			if err != nil {
				log.Fatalf("parseAttributeGroup:Error reading Argument")
			}
			attributeArgs = append(attributeArgs, arg)
		}
		//Parse Attribute Name
		attrNameNode := builder.parseExprNode(attributeVertex.(*ast.Attribute).Name)
		attrName, err := builder.readVariable(attrNameNode)
		if err != nil {
			log.Fatalf("parseAttributeGroup:Error reading Attribute Name")
		}
		attr := NewOpAttribute(attrName, attributeArgs, attributeVertex.GetPosition())
		attrArray = append(attrArray, attr)

	}
	return NewOpAttributeGroup(attrArray, attrGroup.Position)

}

func (cb *CFGBuilder) parseClassModifier(modifiers []ast.Vertex) ClassModifFlag {
	flags := ClassModifFlag(0)
	for _, modifier := range modifiers {
		switch strings.ToLower(string(modifier.(*ast.Identifier).Value)) {
		case "public":
			flags |= CLASS_MODIF_PUBLIC
		case "protected":
			flags |= CLASS_MODIF_PROTECTED
		case "private":
			flags |= CLASS_MODIF_PRIVATE
		case "static":
			flags |= CLASS_MODIF_STATIC
		case "abstract":
			flags |= CLASS_MODIF_ABSTRACT
		case "final":
			flags |= CLASS_MODIF_FINAL
		case "readonly":
			flags |= CLASS_MODIF_READONLY
		default:
			log.Fatal("Error: Unknown Identifier '", string(modifier.(*ast.Identifier).Value), "'")
		}
	}

	return flags
}

func (cb *CFGBuilder) parseFuncModifier(modifiers []ast.Vertex, isRef bool) FuncModifFlag {
	flags := FuncModifFlag(0)

	if isRef {
		flags |= FUNC_MODIF_FLAG_RETURNS_REF
	}

	for _, modifier := range modifiers {
		switch strings.ToLower(string(modifier.(*ast.Identifier).Value)) {
		case "public":
			flags |= FUNC_MODIF_FLAG_PUBLIC
		case "protected":
			flags |= FUNC_MODIF_FLAG_PROTECTED
		case "private":
			flags |= FUNC_MODIF_FLAG_PRIVATE
		case "static":
			flags |= FUNC_MODIF_FLAG_STATIC
		case "abstract":
			flags |= FUNC_MODIF_FLAG_ABSTRACT
		case "final":
			flags |= FUNC_MODIF_FLAG_FINAL
		default:
			log.Fatal("Error: Unknown Identifier '", string(modifier.(*ast.Identifier).Value), "'")
		}
	}

	return flags
}
