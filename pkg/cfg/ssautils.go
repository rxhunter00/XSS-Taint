package cfg

import (
	"fmt"
	"log"
	"reflect"
)

// Part Utils of Local Value Numbering Read Variable Name in current block
func (builder *CFGBuilder) readVariable(vr Operand) (Operand, error) {
	if vr == nil {
		return nil, fmt.Errorf("read nil operand")
	}
	switch v := vr.(type) {
	case *OperandBoundVariable:
		return v, nil
	case *OperandVariable:
		switch varName := v.VariableName.(type) {
		case *OperandString:
			
			return builder.readVariableName(varName.Val, builder.currentBlock), nil
		case *OperandVariable:
			_, err := builder.readVariable(varName)
			if err != nil {
				return nil, err
			}
			return vr, nil
		case *TemporaryOperand:
			_, err := builder.readVariable(varName)
			if err != nil {
				return nil, err
			}
			return vr, nil
		default:
			log.Fatalf("readVariable:Error '%v'", reflect.TypeOf(varName))
		}
	case *TemporaryOperand:
		if v.Original != nil {
			res, err := builder.readVariable(v.Original)
			if err != nil {
				return nil, err
			}
			return res, nil
		}
	}

	return vr, nil
}

// Read Variable Name in current block
// If a block currently contains no definition for a variable, we recursively look
// for a definition in its predecessors
func (builder *CFGBuilder) readVariableName(name string, block *Block) Operand {
	// Read from current block
	val, ok := builder.FuncContex.GetLocalVar(block, name)
	if ok {
		return val
	}

	switch name {
	case "$_GET":
		builder.currentFunc.FuncHasTaint = true
		builder.currentBlock.HasTainted = true
		return NewOperandSymbolic("getsymbolic", true)
	case "$_POST":
		builder.currentFunc.FuncHasTaint = true
		builder.currentBlock.HasTainted = true
		return NewOperandSymbolic("postsymbolic", true)
	case "$_REQUEST":
		builder.currentFunc.FuncHasTaint = true
		builder.currentBlock.HasTainted = true
		return NewOperandSymbolic("requestsymbolic", true)
	case "$_FILES":
		builder.currentFunc.FuncHasTaint = true
		builder.currentBlock.HasTainted = true
		return NewOperandSymbolic("filessymbolic", true)
	case "$_COOKIE":
		builder.currentFunc.FuncHasTaint = true
		builder.currentBlock.HasTainted = true
		return NewOperandSymbolic("cookiessymbolic", true)
	case "$_SERVERS":
		builder.currentFunc.FuncHasTaint = true
		builder.currentBlock.HasTainted = true
		return NewOperandSymbolic("serverssymbolic", true)
	}
	// Else search recursively from predecessors
	return builder.readVariableRecursive(name, block)
}

// Search Definition of Variable on all block
func (builder *CFGBuilder) readVariableRecursive(name string, block *Block) Operand {

	// Braun et al Global Var
	// If a block currently contains no definition for a variable, we recursively look
	// for a definition in its predecessors. If the block has a single predecessor, just
	// query it recursively for a definition,Otherwise, we collect the definitions from
	// all predecessors and construct a φ function
	//
	// Due to loops in the program, those might lead to endless recursion.
	//
	// Therefore, before recursing, we first create the phi function without operands and
	// record it as the current definition for the variable in the block

	vr := Operand(nil)
	if !builder.FuncContex.IsComplete {
		// Incomplete CFG, create an incomplete phi
		vr = NewTemporaryOperand(NewOperandVariable(NewOperandString(name), nil))
		phi := NewOpPhi(vr, block, nil)
		builder.FuncContex.AddIncompletePhi(block, name, phi)
		builder.writeVariableName(name, vr, block)
	} else if len(block.Predecesors) == 1 && !block.Predecesors[0].Dead {
		// If the block has a single predecessor, just query it recursively for a definition
		vr = builder.readVariableName(name, block.Predecesors[0])
		builder.writeVariableName(name, vr, block)
	} else {
		// Read Recursively from predecesors
		// create the phi function without operands
		vr = NewTemporaryOperand(NewOperandVariable(NewOperandString(name), nil))
		phi := NewOpPhi(vr, block, nil)
		block.AddPhi(phi)
		builder.writeVariableName(name, vr, block)

		// we collect the definitions from all predecessors
		for _, pred := range block.Predecesors {
			if !pred.Dead {
				oper := builder.readVariableName(name, pred)
				phi.AddOperand(oper)
			}
		}
	}

	return vr
}

// Add a new variable definition to current block scope
func (builder *CFGBuilder) writeVariable(vr Operand) Operand {
	// Get original Variable
	for {
		vrTemp, ok := vr.(*TemporaryOperand)
		if !ok || vrTemp.Original == nil {
			break
		}
		vr = vrTemp.Original
	}
	// Write variable name
	if vrVar, ok := vr.(*OperandVariable); ok {
		switch name := vrVar.VariableName.(type) {
		case *OperandVariable:
			builder.readVariable(name)
		case *OperandString:
			nameString := name.Val
			vr = NewTemporaryOperand(vr)
			builder.writeVariableName(nameString, vr, builder.currentBlock)
		}
	}
	return vr
}

// Write Variable name in current block scope
func (cb *CFGBuilder) writeVariableName(name string, val Operand, block *Block) {
	cb.VariableNames[name] = struct{}{}
	cb.FuncContex.SetLocalVar(block, name, val)
}

func (cb *CFGBuilder) processAssertion(oper Operand, ifBlock *Block, elseBlock *Block) {
	if ifBlock == nil {
		log.Fatalf("Error in processAssertion: ifBlock cannot be nil")
	} else if elseBlock == nil {
		log.Fatalf("Error in processAssertion: elseBlock cannot be nil")
	}
	block := cb.currentBlock
	for _, assert := range oper.GetAssertions() {
		// add assertion into if block
		cb.currentBlock = ifBlock
		read, err := cb.readVariable(assert.Var)
		if err != nil {
			log.Fatalf("Error in processAssertion (if): %v", err)
		}
		write := cb.writeVariable(assert.Var)
		a := cb.readAssertion(assert.Assert)
		opAssert := NewOpExprAssertion(read, write, a, nil)
		cb.currentBlock.AddInstructions(opAssert)

		// add negation of the assertion into else block
		cb.currentBlock = elseBlock
		read, err = cb.readVariable(assert.Var)
		if err != nil {
			log.Fatalf("Error in processAssertion (else): %v", err)
		}
		write = cb.writeVariable(assert.Var)
		a = cb.readAssertion(assert.Assert).GetNegation()
		opAssert = NewOpExprAssertion(read, write, a, nil)
		cb.currentBlock.AddInstructions(opAssert)
	}
	cb.currentBlock = block
}

func (cb *CFGBuilder) readAssertion(assert Assertion) Assertion {
	switch a := assert.(type) {
	case *TypeAssertion:
		vr, err := cb.readVariable(a.AssertionOperand)
		if err != nil {
			log.Fatalf("Error in readAssertion (if): %v", err)
		}
		return NewTypeAssertion(vr, a.IsNegated)
	case *CompositeAssertion:
		vrs := make([]Assertion, 0)
		for _, assertChild := range a.AssertionList {
			vrs = append(vrs, cb.readAssertion(assertChild))
		}
		return NewCompositeAssertion(vrs, a.Mode, a.IsNegated)
	}
	log.Fatal("Error: Wrong assertion type")
	return nil
}
