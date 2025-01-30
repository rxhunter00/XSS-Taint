package cfg

import "github.com/VKCOM/php-parser/pkg/position"

type OpPhi struct {
	Vars      map[Operand]struct{}
	PhiResult Operand
	PhiBlock  *Block
	OpGeneral
}

func NewOpPhi(result Operand, block *Block, position *position.Position) *OpPhi {
	op := &OpPhi{
		Vars:      make(map[Operand]struct{}),
		PhiResult: result,
		PhiBlock:  block,
		OpGeneral: NewOpGeneral(position),
	}
	AddWriteRef(op, result)
	return op
}

func (op *OpPhi) GetType() string {
	return "Phi"
}

// func (op *OpPhi) GetPosition() *position.Position {
// 	panic("not implemented") // TODO: Implement
// }

// func (op *OpPhi) SetFilePath(filePath string) {
// 	panic("not implemented") // TODO: Implement
// }

// func (op *OpPhi) GetFilePath() string {
// 	panic("not implemented") // TODO: Implement
// }

func (op *OpPhi) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Result": op.PhiResult,
	}
}

// func (op *OpPhi) GetOpListVars() map[string][]Operand {
// 	panic("not implemented") // TODO: Implement
// }

// func (op *OpPhi) ChangeOpVar(varName string, vr Operand) {
// 	panic("not implemented") // TODO: Implement
// }

// func (op *OpPhi) ChangeOpListVar(varName string, vr []Operand) {
// 	panic("not implemented") // TODO: Implement
// }

// func (op *OpPhi) GetOpVarPos(varName string) *position.Position {
// 	panic("not implemented") // TODO: Implement
// }

// func (op *OpPhi) GetOpVarListPos(varName string, index int) *position.Position {
// 	panic("not implemented") // TODO: Implement
// }

// func (op *OpPhi) SetBlock(_ *Block) {
// 	panic("not implemented") // TODO: Implement
// }

// func (op *OpPhi) GetBlock() *Block {
// 	panic("not implemented") // TODO: Implement
// }

func (op *OpPhi) Clone() Op {
	return &OpPhi{
		Vars:      op.Vars,
		PhiResult: op.PhiResult,
		PhiBlock:  op.PhiBlock,
		OpGeneral: op.OpGeneral,
	}
}

func (op *OpPhi) HasOperand(operand Operand) bool {
	//operand exist
	_, ok := op.Vars[operand]

	return ok
}

// Collects all the operands stored in op.Vars return as slice of Operand
func (op *OpPhi) GetPhiOperands() []Operand {
	pVars := make([]Operand, 0, len(op.Vars))
	for variable := range op.Vars {
		pVars = append(pVars, variable)
	}
	return pVars
}

// Add Operand
func (op *OpPhi) AddOperandtoPhi(oper Operand) {
	var empty struct{}
	// add if operand have not been in vars and not phi itself
	if _, ok := op.Vars[oper]; !ok && op.PhiResult != oper {
		tmp := AddUseRef(op, oper)
		op.Vars[tmp] = empty
	}
}

// Remove an operand from phi vars
func (op *OpPhi) RemoveOperandfromPhi(oper Operand) {
	if _, ok := op.Vars[oper]; ok {
		oper.RemoveUser(op)
		delete(op.Vars, oper)
	}
}
