package cfg

import (
	"errors"

	"github.com/VKCOM/php-parser/pkg/position"
)

type FuncModifFlag int

const (
	FUNC_MODIF_FLAG_PUBLIC      FuncModifFlag = 1
	FUNC_MODIF_FLAG_PROTECTED   FuncModifFlag = 1 << 1
	FUNC_MODIF_FLAG_PRIVATE     FuncModifFlag = 1 << 2
	FUNC_MODIF_FLAG_STATIC      FuncModifFlag = 1 << 3
	FUNC_MODIF_FLAG_ABSTRACT    FuncModifFlag = 1 << 4
	FUNC_MODIF_FLAG_FINAL       FuncModifFlag = 1 << 5
	FUNC_MODIF_FLAG_RETURNS_REF FuncModifFlag = 1 << 6
	FUNC_MODIF_FLAG_CLOSURE     FuncModifFlag = 1 << 7
)

// Extend Op
type Func struct {
	Name          string
	Flags         FuncModifFlag
	ReturnType    OpType
	FunctionClass *OperandString // Literal
	Params        []*OpExprParam //

	CFGBlock   *Block
	CallableOp Op
	OpGeneral

	FuncHasTaint bool
	Sources      []Op
	Calls        []Op
}

func NewFunc(name string, flags FuncModifFlag, returnType OpType, entryBlock *Block, position *position.Position) (*Func, error) {
	if entryBlock == nil {
		return nil, errors.New("entry Block cannot be nil")
	}
	return &Func{
		Name:          name,
		Flags:         flags,
		ReturnType:    returnType,
		FunctionClass: nil,
		Params:        make([]*OpExprParam, 0),
		CFGBlock:      entryBlock,
		OpGeneral:     NewOpGeneral(position),
		FuncHasTaint:  false,
	}, nil
}
func NewClassFunc(name string, flags FuncModifFlag, returnType OpType, entryBlock *Block, fclass OperandString, position *position.Position) (*Func, error) {
	if entryBlock == nil {
		return nil, errors.New("entry Block cannot be nil")
	}
	return &Func{
		Name:          name,
		Flags:         flags,
		ReturnType:    returnType,
		FunctionClass: &fclass,
		Params:        make([]*OpExprParam, 0),
		CFGBlock:      entryBlock,
		OpGeneral:     NewOpGeneral(position),
		FuncHasTaint:  false,
	}, nil
}
func (op *Func) GetScopedName() string {
	if op.FunctionClass != nil {
		className := op.FunctionClass.Val
		return className + "::" + op.Name
	}
	return op.Name
}

func (op *Func) AddModifier(flag FuncModifFlag) {
	op.Flags |= flag
}

func (op *Func) GetVisibility() FuncModifFlag {
	return FuncModifFlag(op.Flags & 7)
}

func (op *Func) IsPublic() bool {
	return op.Flags&FUNC_MODIF_FLAG_PUBLIC != 0
}

func (op *Func) IsPrivate() bool {
	return op.Flags&FUNC_MODIF_FLAG_PRIVATE != 0
}

func (op *Func) IsProtected() bool {
	return op.Flags&FUNC_MODIF_FLAG_PROTECTED != 0
}

func (op *Func) IsStatic() bool {
	return op.Flags&FUNC_MODIF_FLAG_STATIC != 0
}

func (op *Func) IsAbstract() bool {
	return op.Flags&FUNC_MODIF_FLAG_ABSTRACT != 0
}

func (op *Func) IsFinal() bool {
	return op.Flags&FUNC_MODIF_FLAG_FINAL != 0
}

func (op *Func) IsReturnRef() bool {
	return op.Flags&FUNC_MODIF_FLAG_RETURNS_REF != 0
}

func (op *Func) IsClosure() bool {
	return op.Flags&FUNC_MODIF_FLAG_CLOSURE != 0
}

func (op *Func) GetType() string {
	return "Func"
}

func (op *Func) Clone() Op {
	params := make([]*OpExprParam, len(op.Params))
	copy(params, op.Params)
	return &Func{
		OpGeneral:     op.OpGeneral,
		Name:          op.Name,
		Flags:         op.Flags,
		ReturnType:    op.ReturnType,
		FunctionClass: op.FunctionClass,
		Params:        params,
		CFGBlock:      op.CFGBlock,
		CallableOp:    op.CallableOp,
		FuncHasTaint:  op.FuncHasTaint,
	}
}
