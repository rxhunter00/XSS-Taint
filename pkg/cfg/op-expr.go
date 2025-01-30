package cfg

import (
	"log"

	"github.com/VKCOM/php-parser/pkg/position"
)

type OpExprParam struct {
	Name         Operand
	ByRef        bool
	IsVariadic   bool
	AttrGroups   []*OpAttributeGroup
	DefaultVar   Operand
	DefaultBlock *Block
	DecalreType  OpType
	Result       Operand
	OpGeneral
}

func NewOpExprParam(name Operand,
	byRef bool,
	variadic bool,
	attrGroups []*OpAttributeGroup,
	defaultVar Operand,
	defaultBlock *Block,
	declaredType OpType,
	pos *position.Position) *OpExprParam {

	op := &OpExprParam{
		Name:         name,
		ByRef:        byRef,
		IsVariadic:   variadic,
		AttrGroups:   attrGroups,
		DefaultVar:   defaultVar,
		DefaultBlock: defaultBlock,
		DecalreType:  declaredType,
		Result:       NewTemporaryOperand(nil),
		OpGeneral:    NewOpGeneral(pos),
	}

	AddUseRefs(op, name, defaultVar)
	AddWriteRef(op, name)

	return op
}

func (op *OpExprParam) GetType() string {
	return "ExprParam"
}

func (op *OpExprParam) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Name":       op.Name,
		"DefaultVar": op.DefaultVar,
		"Result":     op.Result,
	}
}

func (op *OpExprParam) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Name":
		op.Name = vr
	case "DefaultVar":
		op.DefaultVar = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprParam) Clone() Op {
	attrGroups := make([]*OpAttributeGroup, len(op.AttrGroups))
	copy(attrGroups, op.AttrGroups)
	return &OpExprParam{
		OpGeneral:    op.OpGeneral,
		Name:         op.Name,
		ByRef:        op.ByRef,
		IsVariadic:   op.IsVariadic,
		AttrGroups:   attrGroups,
		DefaultVar:   op.DefaultVar,
		DefaultBlock: op.DefaultBlock,
		DecalreType:  op.DecalreType,
		Result:       op.Result,
	}
}

type OpExprConcatList struct {
	OpGeneral
	List   []Operand
	Result Operand

	ListPos []*position.Position
}

func NewOpExprConcatList(list []Operand, listPos []*position.Position, pos *position.Position) *OpExprConcatList {
	op := &OpExprConcatList{
		List:      list,
		Result:    NewTemporaryOperand(nil),
		ListPos:   listPos,
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, list...)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprConcatList) GetType() string {
	return "ExprConcatList"
}

func (op *OpExprConcatList) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Result": op.Result,
	}
}

func (op *OpExprConcatList) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprConcatList) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"List": op.List,
	}
}

func (op *OpExprConcatList) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "List":
		op.List = vr
	}
}

func (op *OpExprConcatList) GetOpVarListPos(vrName string, index int) *position.Position {
	switch vrName {
	case "List":
		return op.ListPos[index]
	}
	return nil
}

func (op *OpExprConcatList) Clone() Op {
	list := make([]Operand, len(op.List))
	copy(list, op.List)
	return &OpExprConcatList{
		OpGeneral: op.OpGeneral,
		List:      list,
		Result:    op.Result,
	}
}

type OpExprAssign struct {
	Var    Operand
	Expr   Operand
	Result Operand

	VarPos  *position.Position
	ExprPos *position.Position
	OpGeneral
}

func NewOpExprAssign(vr, expr Operand, varPos, exprPos, pos *position.Position) *OpExprAssign {
	op := &OpExprAssign{
		Var:       vr,
		Expr:      expr,
		Result:    NewTemporaryOperand(nil),
		VarPos:    varPos,  //  Check
		ExprPos:   exprPos, // Check
		OpGeneral: NewOpGeneral(pos),
	}
	// Write read ref to
	AddUseRef(op, expr)
	AddWriteRefs(op, op.Result, vr)

	return op
}

func (op *OpExprAssign) GetType() string {
	return "ExprAssign"
}

func (op *OpExprAssign) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var":    op.Var,
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

// func (op *OpExprAssign) ChangeOpVar(vrName string, vr Operand) {
// 	switch vrName {
// 	case "Var":
// 		op.Var = vr
// 	case "Expr":
// 		op.Expr = vr
// 	case "Result":
// 		op.Result = vr
// 	}
// }

// func (op *OpExprAssign) GetOpVarPos(vrName string) *position.Position {
// 	switch vrName {
// 	case "Var":
// 		return op.VarPos
// 	case "Expr":
// 		return op.ExprPos
// 	}
// 	return nil
// }

func (op *OpExprAssign) Clone() Op {
	return &OpExprAssign{
		Expr:      op.Expr,
		Var:       op.Var,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprAssignRef struct {
	Var    Operand
	Expr   Operand
	Result Operand
	OpGeneral
}

func NewOpExprAssignRef(vr, expr Operand, pos *position.Position) *OpExprAssignRef {
	op := &OpExprAssignRef{
		Var:       vr,
		Expr:      expr,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRef(op, expr)
	AddWriteRefs(op, op.Result, vr)

	return op
}

func (op *OpExprAssignRef) GetType() string {
	return "ExprAssignRef"
}

func (op *OpExprAssignRef) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var":    op.Var,
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprAssignRef) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprAssignRef) Clone() Op {
	return &OpExprAssignRef{
		Expr:      op.Expr,
		Var:       op.Var,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprArrayDimFetch struct {
	Var    Operand
	Dim    Operand
	Result Operand
	OpGeneral
}

func NewOpExprArrayDimFetch(vr, dim Operand, pos *position.Position) *OpExprArrayDimFetch {
	op := &OpExprArrayDimFetch{
		Var:       vr,
		Dim:       dim,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, vr, dim)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprArrayDimFetch) GetType() string {
	return "ExprArrayDimFetch"
}

func (op *OpExprArrayDimFetch) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var":    op.Var,
		"Dim":    op.Dim,
		"Result": op.Result,
	}
}

func (op *OpExprArrayDimFetch) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	case "Dim":
		op.Dim = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprArrayDimFetch) Clone() Op {
	return &OpExprArrayDimFetch{
		OpGeneral: op.OpGeneral,
		Var:       op.Var,
		Dim:       op.Dim,
		Result:    op.Result,
	}
}

func (op *OpExprArrayDimFetch) String() string {
	varName, _ := GetOperandName(op.Var)
	dimStr, err := GetOperandName(op.Dim)
	if err != nil {
		return ""
	}

	if varName != "" {
		return varName + "[" + dimStr + "]"
	} else {
		varDef := op.Var.GetWriter()
		if varDef == nil {
			return ""
		}
		if v, ok := varDef.(*OpExprArrayDimFetch); ok {
			varName = v.String()
			return varName + "[" + dimStr + "]"
		}
	}
	return ""
}

// Binary Op

type OpExprBinaryConcat struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand

	LeftPos  *position.Position
	RightPos *position.Position
}

func NewOpExprBinaryConcat(left, right Operand, leftPos, rightPos, pos *position.Position) *OpExprBinaryConcat {
	op := &OpExprBinaryConcat{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		LeftPos:   leftPos,
		RightPos:  rightPos,
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryConcat) GetType() string {
	return "ExprBinaryConcat"
}

func (op *OpExprBinaryConcat) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

// func (op *OpExprBinaryConcat) ChangeOpVar(vrName string, vr Operand) {
// 	switch vrName {
// 	case "Left":
// 		op.Left = vr
// 	case "Right":
// 		op.Right = vr
// 	case "Result":
// 		op.Result = vr
// 	}
// }

// func (op *OpExprBinaryConcat) GetOpVarPos(vrName string) *position.Position {
// 	switch vrName {
// 	case "Left":
// 		return op.LeftPos
// 	case "Right":
// 		return op.RightPos
// 	}

// 	return nil
// }

func (op *OpExprBinaryConcat) Clone() Op {
	return &OpExprBinaryConcat{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinaryBitwiseAnd struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryBitwiseAnd(left, right Operand, pos *position.Position) *OpExprBinaryBitwiseAnd {
	op := &OpExprBinaryBitwiseAnd{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryBitwiseAnd) GetType() string {
	return "ExprBinaryBitwiseAnd"
}

func (op *OpExprBinaryBitwiseAnd) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryBitwiseAnd) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryBitwiseAnd) Clone() Op {
	return &OpExprBinaryBitwiseAnd{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryBitwiseOr struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryBitwiseOr(left, right Operand, pos *position.Position) *OpExprBinaryBitwiseOr {
	op := &OpExprBinaryBitwiseOr{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryBitwiseOr) GetType() string {
	return "ExprBinaryBitwiseOr"
}

func (op *OpExprBinaryBitwiseOr) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryBitwiseOr) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryBitwiseOr) Clone() Op {
	return &OpExprBinaryBitwiseOr{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryBitwiseXor struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryBitwiseXor(left, right Operand, pos *position.Position) *OpExprBinaryBitwiseXor {
	op := &OpExprBinaryBitwiseXor{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryBitwiseXor) GetType() string {
	return "ExprBinaryBitwiseXor"
}

func (op *OpExprBinaryBitwiseXor) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryBitwiseXor) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryBitwiseXor) Clone() Op {
	return &OpExprBinaryBitwiseXor{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryCoalesce struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryCoalesce(left, right Operand, pos *position.Position) *OpExprBinaryCoalesce {
	op := &OpExprBinaryCoalesce{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryCoalesce) GetType() string {
	return "ExprBinaryCoalesce"
}

func (op *OpExprBinaryCoalesce) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryCoalesce) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryCoalesce) Clone() Op {
	return &OpExprBinaryCoalesce{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinaryDiv struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryDiv(left, right Operand, pos *position.Position) *OpExprBinaryDiv {
	op := &OpExprBinaryDiv{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryDiv) GetType() string {
	return "ExprBinaryDiv"
}

func (op *OpExprBinaryDiv) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryDiv) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryDiv) Clone() Op {
	return &OpExprBinaryDiv{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinaryMinus struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryMinus(left, right Operand, pos *position.Position) *OpExprBinaryMinus {
	op := &OpExprBinaryMinus{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryMinus) GetType() string {
	return "ExprBinaryMinus"
}

func (op *OpExprBinaryMinus) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryMinus) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryMinus) Clone() Op {
	return &OpExprBinaryMinus{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinaryMod struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryMod(left, right Operand, pos *position.Position) *OpExprBinaryMod {
	op := &OpExprBinaryMod{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryMod) GetType() string {
	return "ExprBinaryMod"
}

func (op *OpExprBinaryMod) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryMod) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryMod) Clone() Op {
	return &OpExprBinaryMod{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinaryMul struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryMul(left, right Operand, pos *position.Position) *OpExprBinaryMul {
	op := &OpExprBinaryMul{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryMul) GetType() string {
	return "ExprBinaryMul"
}

func (op *OpExprBinaryMul) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryMul) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryMul) Clone() Op {
	return &OpExprBinaryMul{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinaryPlus struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryPlus(left, right Operand, pos *position.Position) *OpExprBinaryPlus {
	op := &OpExprBinaryPlus{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryPlus) GetType() string {
	return "ExprBinaryPlus"
}

func (op *OpExprBinaryPlus) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryPlus) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryPlus) Clone() Op {
	return &OpExprBinaryPlus{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinaryPow struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryPow(left, right Operand, pos *position.Position) *OpExprBinaryPow {
	op := &OpExprBinaryPow{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryPow) GetType() string {
	return "ExprBinaryPow"
}

func (op *OpExprBinaryPow) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryPow) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryPow) Clone() Op {
	return &OpExprBinaryPow{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinaryShiftLeft struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryShiftLeft(left, right Operand, pos *position.Position) *OpExprBinaryShiftLeft {
	op := &OpExprBinaryShiftLeft{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryShiftLeft) GetType() string {
	return "ExprBinaryShiftLeft"
}

func (op *OpExprBinaryShiftLeft) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryShiftLeft) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryShiftLeft) Clone() Op {
	return &OpExprBinaryShiftLeft{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinaryShiftRight struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryShiftRight(left, right Operand, pos *position.Position) *OpExprBinaryShiftRight {
	op := &OpExprBinaryShiftRight{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryShiftRight) GetType() string {
	return "ExprBinaryShiftRight"
}

func (op *OpExprBinaryShiftRight) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryShiftRight) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryShiftRight) Clone() Op {
	return &OpExprBinaryShiftRight{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

// Binary Logical

type OpExprBinaryLogicalAnd struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryLogicalAnd(left, right Operand, pos *position.Position) *OpExprBinaryLogicalAnd {
	op := &OpExprBinaryLogicalAnd{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryLogicalAnd) GetType() string {
	return "ExprBinaryLogicalAnd"
}

func (op *OpExprBinaryLogicalAnd) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryLogicalAnd) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryLogicalAnd) Clone() Op {
	return &OpExprBinaryLogicalAnd{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinaryLogicalOr struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryLogicalOr(left, right Operand, pos *position.Position) *OpExprBinaryLogicalOr {
	op := &OpExprBinaryLogicalOr{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryLogicalOr) GetType() string {
	return "ExprBinaryLogicalOr"
}

func (op *OpExprBinaryLogicalOr) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryLogicalOr) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryLogicalOr) Clone() Op {
	return &OpExprBinaryLogicalAnd{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinaryLogicalXor struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryLogicalXor(left, right Operand, pos *position.Position) *OpExprBinaryLogicalXor {
	op := &OpExprBinaryLogicalXor{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryLogicalXor) GetType() string {
	return "ExprBinaryLogicalXor"
}

func (op *OpExprBinaryLogicalXor) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryLogicalXor) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryLogicalXor) Clone() Op {
	return &OpExprBinaryLogicalAnd{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinarySpaceship struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinarySpaceship(left, right Operand, pos *position.Position) *OpExprBinarySpaceship {
	op := &OpExprBinarySpaceship{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinarySpaceship) GetType() string {
	return "ExprBinarySpaceship"
}

func (op *OpExprBinarySpaceship) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinarySpaceship) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinarySpaceship) Clone() Op {
	return &OpExprBinarySpaceship{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinarySmallerOrEqual struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinarySmallerOrEqual(left, right Operand, pos *position.Position) *OpExprBinarySmallerOrEqual {
	op := &OpExprBinarySmallerOrEqual{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinarySmallerOrEqual) GetType() string {
	return "ExprBinarySmallerOrEqual"
}

func (op *OpExprBinarySmallerOrEqual) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinarySmallerOrEqual) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinarySmallerOrEqual) Clone() Op {
	return &OpExprBinarySmallerOrEqual{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinarySmaller struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinarySmaller(left, right Operand, pos *position.Position) *OpExprBinarySmaller {
	op := &OpExprBinarySmaller{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinarySmaller) GetType() string {
	return "ExprBinarySmaller"
}

func (op *OpExprBinarySmaller) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinarySmaller) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinarySmaller) Clone() Op {
	return &OpExprBinarySmaller{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinaryBiggerOrEqual struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryBiggerOrEqual(left, right Operand, pos *position.Position) *OpExprBinaryBiggerOrEqual {
	op := &OpExprBinaryBiggerOrEqual{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryBiggerOrEqual) GetType() string {
	return "ExprBinaryBiggerOrEqual"
}

func (op *OpExprBinaryBiggerOrEqual) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryBiggerOrEqual) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryBiggerOrEqual) Clone() Op {
	return &OpExprBinaryBiggerOrEqual{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinaryBigger struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryBigger(left, right Operand, pos *position.Position) *OpExprBinaryBigger {
	op := &OpExprBinaryBigger{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryBigger) GetType() string {
	return "ExprBinaryBigger"
}

func (op *OpExprBinaryBigger) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryBigger) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryBigger) Clone() Op {
	return &OpExprBinaryBigger{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinaryNotEqual struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryNotEqual(left, right Operand, pos *position.Position) *OpExprBinaryNotEqual {
	op := &OpExprBinaryNotEqual{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryNotEqual) GetType() string {
	return "ExprBinaryNotEqual"
}

func (op *OpExprBinaryNotEqual) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryNotEqual) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryNotEqual) Clone() Op {
	return &OpExprBinaryNotEqual{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinaryNotIdentical struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryNotIdentical(left, right Operand, pos *position.Position) *OpExprBinaryNotIdentical {
	op := &OpExprBinaryNotIdentical{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryNotIdentical) GetType() string {
	return "ExprBinaryNotIdentical"
}

func (op *OpExprBinaryNotIdentical) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryNotIdentical) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryNotIdentical) Clone() Op {
	return &OpExprBinaryNotIdentical{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinaryIdentical struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryIdentical(left, right Operand, pos *position.Position) *OpExprBinaryIdentical {
	op := &OpExprBinaryIdentical{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryIdentical) GetType() string {
	return "ExprBinaryIdentical"
}

func (op *OpExprBinaryIdentical) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryIdentical) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryIdentical) Clone() Op {
	return &OpExprBinaryIdentical{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBinaryEqual struct {
	Left   Operand
	Right  Operand
	Result Operand
	OpGeneral
}

func NewOpExprBinaryEqual(left, right Operand, pos *position.Position) *OpExprBinaryEqual {
	op := &OpExprBinaryEqual{
		Left:      left,
		Right:     right,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRefs(op, left, right)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBinaryEqual) GetType() string {
	return "ExprBinaryEqual"
}

func (op *OpExprBinaryEqual) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryEqual) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryEqual) Clone() Op {
	return &OpExprBinaryEqual{
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

// Cast
type OpExprCastBool struct {
	Expr   Operand
	Result Operand
	OpGeneral
}

func NewOpExprCastBool(expr Operand, pos *position.Position) *OpExprCastBool {
	Op := &OpExprCastBool{
		Expr:      expr,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprCastBool) GetType() string {
	return "ExprCastBool"
}

func (op *OpExprCastBool) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprCastBool) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprCastBool) Clone() Op {
	return &OpExprCastBool{
		Expr:      op.Expr,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprCastDouble struct {
	Expr   Operand
	Result Operand
	OpGeneral
}

func NewOpExprCastDouble(expr Operand, pos *position.Position) *OpExprCastDouble {
	Op := &OpExprCastDouble{
		Expr:      expr,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprCastDouble) GetType() string {
	return "ExprCastDouble"
}

func (op *OpExprCastDouble) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprCastDouble) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprCastDouble) Clone() Op {
	return &OpExprCastDouble{
		Expr:      op.Expr,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprCastInt struct {
	Expr   Operand
	Result Operand
	OpGeneral
}

func NewOpExprCastInt(expr Operand, pos *position.Position) *OpExprCastInt {
	Op := &OpExprCastInt{
		Expr:      expr,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprCastInt) GetType() string {
	return "ExprCastInt"
}

func (op *OpExprCastInt) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprCastInt) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprCastInt) Clone() Op {
	return &OpExprCastInt{
		Expr:      op.Expr,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprCastString struct {
	Expr   Operand
	Result Operand
	OpGeneral
}

func NewOpExprCastString(expr Operand, pos *position.Position) *OpExprCastString {
	Op := &OpExprCastString{
		Expr:      expr,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprCastString) GetType() string {
	return "ExprCastString"
}

func (op *OpExprCastString) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprCastString) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprCastString) Clone() Op {
	return &OpExprCastString{
		Expr:      op.Expr,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprCastObject struct {
	Expr   Operand
	Result Operand
	OpGeneral
}

func NewOpExprCastObject(expr Operand, pos *position.Position) *OpExprCastObject {
	Op := &OpExprCastObject{
		Expr:      expr,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprCastObject) GetType() string {
	return "ExprCastObject"
}

func (op *OpExprCastObject) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprCastObject) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprCastObject) Clone() Op {
	return &OpExprCastObject{
		Expr:      op.Expr,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprCastUnset struct {
	Expr   Operand
	Result Operand
	OpGeneral
}

func NewOpExprCastUnset(expr Operand, pos *position.Position) *OpExprCastUnset {
	Op := &OpExprCastUnset{
		Expr:      expr,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprCastUnset) GetType() string {
	return "ExprCastUnset"
}

func (op *OpExprCastUnset) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprCastUnset) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprCastUnset) Clone() Op {
	return &OpExprCastUnset{
		Expr:      op.Expr,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprCastArray struct {
	Expr   Operand
	Result Operand
	OpGeneral
}

func NewOpExprCastArray(expr Operand, pos *position.Position) *OpExprCastArray {
	Op := &OpExprCastArray{
		Expr:      expr,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprCastArray) GetType() string {
	return "ExprCastArray"
}

func (op *OpExprCastArray) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprCastArray) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprCastArray) Clone() Op {
	return &OpExprCastArray{
		Expr:      op.Expr,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

// Unary
type OpExprUnaryPlus struct {
	Expr   Operand
	Result Operand
	OpGeneral
}

func NewOpExprUnaryPlus(expr Operand, pos *position.Position) *OpExprUnaryPlus {
	Op := &OpExprUnaryPlus{
		Expr:      expr,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprUnaryPlus) GetType() string {
	return "ExprUnaryPlus"
}

func (op *OpExprUnaryPlus) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprUnaryPlus) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprUnaryPlus) Clone() Op {
	return &OpExprUnaryPlus{
		Expr:      op.Expr,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprUnaryMinus struct {
	Expr   Operand
	Result Operand
	OpGeneral
}

func NewOpExprUnaryMinus(expr Operand, pos *position.Position) *OpExprUnaryMinus {
	Op := &OpExprUnaryMinus{
		Expr:      expr,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprUnaryMinus) GetType() string {
	return "ExprUnaryMinus"
}

func (op *OpExprUnaryMinus) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprUnaryMinus) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprUnaryMinus) Clone() Op {
	return &OpExprUnaryMinus{
		Expr:      op.Expr,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

// Array
type OpExprArray struct {
	OpGeneral
	Keys   []Operand
	Vals   []Operand
	ByRef  []bool
	Result Operand
}

func NewOpExprArray(keys, vals []Operand, byRef []bool, pos *position.Position) *OpExprArray {
	op := &OpExprArray{
		OpGeneral: NewOpGeneral(pos),
		Keys:      keys,
		Vals:      vals,
		ByRef:     byRef,
		Result:    NewTemporaryOperand(nil),
	}

	AddUseRefs(op, keys...)
	AddUseRefs(op, vals...)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprArray) GetType() string {
	return "ExprArray"
}

func (op *OpExprArray) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Result": op.Result,
	}
}

func (op *OpExprArray) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprArray) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Keys":   op.Keys,
		"Values": op.Vals,
	}
}

func (op *OpExprArray) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Keys":
		op.Keys = vr
	case "Values":
		op.Vals = vr
	}
}

func (op *OpExprArray) Clone() Op {
	keys := make([]Operand, len(op.Keys))
	vals := make([]Operand, len(op.Vals))
	copy(keys, op.Keys)
	copy(vals, op.Vals)
	return &OpExprArray{
		OpGeneral: op.OpGeneral,
		Keys:      keys,
		Vals:      vals,
		ByRef:     op.ByRef,
		Result:    op.Result,
	}
}

type OpExprClosure struct {
	OpGeneral
	Func    *Func
	UseVars []Operand
	Result  Operand
}

func NewOpExprClosure(Func *Func, useVars []Operand, pos *position.Position) *OpExprClosure {
	Op := &OpExprClosure{
		OpGeneral: NewOpGeneral(pos),
		Func:      Func,
		UseVars:   useVars,
		Result:    NewTemporaryOperand(nil),
	}

	AddUseRefs(Op, useVars...)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprClosure) GetType() string {
	return "ExprClosure"
}

func (op *OpExprClosure) GetFunc() *Func {
	return op.Func
}

func (op *OpExprClosure) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Result": op.Result,
	}
}

func (op *OpExprClosure) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprClosure) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"UseVars": op.UseVars,
	}
}

func (op *OpExprClosure) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "UseVars":
		op.UseVars = vr
	}
}

func (op *OpExprClosure) Clone() Op {
	useVars := make([]Operand, len(op.UseVars))
	copy(useVars, op.UseVars)
	return &OpExprClosure{
		OpGeneral: op.OpGeneral,
		Func:      op.Func,
		UseVars:   useVars,
		Result:    op.Result,
	}
}

type OpExprBitwiseNot struct {
	Expr   Operand
	Result Operand
	OpGeneral
}

func NewOpExprBitwiseNot(expr Operand, pos *position.Position) *OpExprBitwiseNot {
	op := &OpExprBitwiseNot{
		Expr:      expr,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRef(op, expr)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBitwiseNot) GetType() string {
	return "ExprBitwiseNot"
}

func (op *OpExprBitwiseNot) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprBitwiseNot) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBitwiseNot) Clone() Op {
	return &OpExprBitwiseNot{
		Expr:      op.Expr,
		Result:    op.Result,
		OpGeneral: op.OpGeneral,
	}
}

type OpExprBooleanNot struct {
	Expr   Operand
	Result Operand
	OpGeneral
}

func NewOpExprBooleanNot(expr Operand, pos *position.Position) *OpExprBooleanNot {
	op := &OpExprBooleanNot{
		Expr:      expr,
		Result:    NewTemporaryOperand(nil),
		OpGeneral: NewOpGeneral(pos),
	}

	AddUseRef(op, expr)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprBooleanNot) GetType() string {
	return "ExprBooleanNot"
}

func (op *OpExprBooleanNot) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprBooleanNot) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBooleanNot) Clone() Op {
	return &OpExprBooleanNot{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprClassConstFetch struct {
	OpGeneral
	Class  Operand
	Name   Operand
	Result Operand
}

func NewOpExprClassConstFetch(class, name Operand, pos *position.Position) *OpExprClassConstFetch {
	Op := &OpExprClassConstFetch{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Class:  class,
		Name:   name,
		Result: NewTemporaryOperand(nil),
	}

	AddUseRefs(Op, class, name)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprClassConstFetch) GetType() string {
	return "ExprClassConstFetch"
}

func (op *OpExprClassConstFetch) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Class":  op.Class,
		"Name":   op.Name,
		"Result": op.Result,
	}
}

func (op *OpExprClassConstFetch) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Class":
		op.Class = vr
	case "Name":
		op.Name = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprClassConstFetch) Clone() Op {
	return &OpExprClassConstFetch{
		OpGeneral: op.OpGeneral,
		Class:     op.Class,
		Name:      op.Name,
		Result:    op.Result,
	}
}

type OpExprClone struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprClone(expr Operand, pos *position.Position) *OpExprClone {
	Op := &OpExprClone{
		OpGeneral: NewOpGeneral(pos),
		Expr:      expr,
		Result:    NewTemporaryOperand(nil),
	}

	AddUseRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprClone) GetType() string {
	return "ExprClone"
}

func (op *OpExprClone) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprClone) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprClone) Clone() Op {
	return &OpExprClone{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprConstFetch struct {
	OpGeneral
	Name   Operand
	Result Operand
}

func NewOpExprConstFetch(name Operand, pos *position.Position) *OpExprConstFetch {
	Op := &OpExprConstFetch{
		OpGeneral: NewOpGeneral(pos),
		Name:      name,
		Result:    NewTemporaryOperand(nil),
	}

	AddUseRef(Op, name)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprConstFetch) GetType() string {
	return "ExprConstFetch"
}

func (op *OpExprConstFetch) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Name":   op.Name,
		"Result": op.Result,
	}
}

func (op *OpExprConstFetch) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Name":
		op.Name = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprConstFetch) Clone() Op {
	return &OpExprConstFetch{
		OpGeneral: op.OpGeneral,
		Name:      op.Name,
		Result:    op.Result,
	}
}

type OpExprEmpty struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprEmpty(expr Operand, pos *position.Position) *OpExprEmpty {
	Op := &OpExprEmpty{
		OpGeneral: NewOpGeneral(pos),
		Expr:      expr,
		Result:    NewTemporaryOperand(nil),
	}

	AddUseRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprEmpty) GetType() string {
	return "ExprEmpty"
}

func (op *OpExprEmpty) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprEmpty) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprEmpty) Clone() Op {
	return &OpExprEmpty{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprEval struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprEval(expr Operand, pos *position.Position) *OpExprEval {
	Op := &OpExprEval{
		OpGeneral: NewOpGeneral(pos),
		Expr:      expr,
		Result:    NewTemporaryOperand(nil),
	}

	AddUseRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprEval) GetType() string {
	return "ExprEval"
}

func (op *OpExprEval) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprEval) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprEval) Clone() Op {
	return &OpExprEval{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprFunctionCall struct {
	OpGeneral
	Name   Operand
	Args   []Operand
	Result Operand

	CalledFunc *Func
	NamePos    *position.Position
	ArgsPos    []*position.Position
}

func NewOpExprFunctionCall(name Operand, args []Operand, namePos *position.Position, argsPos []*position.Position, pos *position.Position) *OpExprFunctionCall {
	op := &OpExprFunctionCall{
		OpGeneral: NewOpGeneral(pos),
		Name:      name,
		Args:      args,
		Result:    NewTemporaryOperand(nil),
		NamePos:   namePos,
		ArgsPos:   argsPos,
	}

	AddUseRef(op, name)
	AddUseRefs(op, args...)
	AddWriteRef(op, op.Result)

	return op
}

func (op *OpExprFunctionCall) GetType() string {
	return "ExprFunctionCall"
}

func (op *OpExprFunctionCall) GetName() string {
	funcName, _ := GetOperandName(op.Name)

	return funcName
}

func (op *OpExprFunctionCall) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Name":   op.Name,
		"Result": op.Result,
	}
}

func (op *OpExprFunctionCall) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Name":
		op.Name = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprFunctionCall) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Args": op.Args,
	}
}

func (op *OpExprFunctionCall) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Args":
		op.Args = vr
	}
}

// func (op *OpExprFunctionCall) GetOpVarPos(vrName string) *position.Position {
// 	switch vrName {
// 	case "Name":
// 		return op.NamePos
// 	}
// 	return nil
// }

// func (op *OpExprFunctionCall) GetOpVarListPos(vrName string, index int) *position.Position {
// 	switch vrName {
// 	case "Args":
// 		return op.ArgsPos[index]
// 	}
// 	return nil
// }

func (op *OpExprFunctionCall) Clone() Op {
	args := make([]Operand, len(op.Args))
	copy(args, op.Args)
	return &OpExprFunctionCall{
		OpGeneral:  op.OpGeneral,
		Name:       op.Name,
		Args:       args,
		CalledFunc: op.CalledFunc,
		Result:     op.Result,
	}
}

type INCLUDE_TYPE int

const (
	TYPE_INCLUDE INCLUDE_TYPE = iota
	TYPE_INCLUDE_ONCE
	TYPE_REQUIRE
	TYPE_REQUIRE_ONCE
)

type OpExprInclude struct {
	OpGeneral
	Type   INCLUDE_TYPE
	Expr   Operand
	Result Operand
}

func NewOpExprInclude(expr Operand, tp INCLUDE_TYPE, pos *position.Position) *OpExprInclude {
	Op := &OpExprInclude{
		OpGeneral: NewOpGeneral(pos),
		Type:      tp,
		Expr:      expr,
		Result:    NewTemporaryOperand(nil),
	}

	AddUseRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprInclude) IncludeTypeStr() string {
	switch op.Type {
	case TYPE_INCLUDE:
		return "Include"
	case TYPE_INCLUDE_ONCE:
		return "IncludeOnce"
	case TYPE_REQUIRE:
		return "Require"
	case TYPE_REQUIRE_ONCE:
		return "RequireOnce"
	}
	return ""
}

func (op *OpExprInclude) GetType() string {
	return "ExprInclude"
}

func (op *OpExprInclude) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprInclude) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprInclude) Clone() Op {
	return &OpExprInclude{
		OpGeneral: op.OpGeneral,
		Type:      op.Type,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprInstanceOf struct {
	OpGeneral
	Expr   Operand
	Class  Operand
	Result Operand
}

func NewOpExprInstanceOf(expr Operand, class Operand, pos *position.Position) *OpExprInstanceOf {
	Op := &OpExprInstanceOf{
		OpGeneral: NewOpGeneral(pos),
		Expr:      expr,
		Class:     class,
		Result:    NewTemporaryOperand(nil),
	}

	AddUseRefs(Op, expr, class)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprInstanceOf) GetType() string {
	return "ExprInstanceOf"
}

func (op *OpExprInstanceOf) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Class":  op.Class,
		"Result": op.Result,
	}
}

func (op *OpExprInstanceOf) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Class":
		op.Class = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprInstanceOf) Clone() Op {
	return &OpExprInstanceOf{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Class:     op.Class,
		Result:    op.Result,
	}
}

type OpExprIsset struct {
	OpGeneral
	Vars   []Operand
	Result Operand
}

func NewOpExprIsset(vars []Operand, pos *position.Position) *OpExprIsset {
	Op := &OpExprIsset{
		OpGeneral: NewOpGeneral(pos),
		Vars:      vars,
		Result:    NewTemporaryOperand(nil),
	}

	AddUseRefs(Op, vars...)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprIsset) GetType() string {
	return "ExprIsset"
}

func (op *OpExprIsset) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Result": op.Result,
	}
}

func (op *OpExprIsset) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprIsset) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Vars": op.Vars,
	}
}

func (op *OpExprIsset) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Vars":
		op.Vars = vr
	}
}

func (op *OpExprIsset) Clone() Op {
	vars := make([]Operand, len(op.Vars))
	copy(vars, op.Vars)
	return &OpExprIsset{
		OpGeneral: op.OpGeneral,
		Vars:      vars,
		Result:    op.Result,
	}
}

type OpExprMethodCall struct {
	OpGeneral
	Var    Operand
	Name   Operand
	Args   []Operand
	Result Operand

	IsNullSafe bool
	CalledFunc OpCallable
	VarPos     *position.Position
	NamePos    *position.Position
	ArgsPos    []*position.Position
}

func NewOpExprMethodCall(vr, name Operand, args []Operand, varPos, namePos *position.Position, argsPos []*position.Position, pos *position.Position) *OpExprMethodCall {
	Op := &OpExprMethodCall{
		OpGeneral:  NewOpGeneral(pos),
		Var:        vr,
		Name:       name,
		Args:       args,
		Result:     NewTemporaryOperand(nil),
		IsNullSafe: false,
		VarPos:     varPos,
		NamePos:    namePos,
		ArgsPos:    argsPos,
	}

	AddUseRefs(Op, vr, name)
	AddUseRefs(Op, args...)
	AddWriteRef(Op, Op.Result)

	return Op
}

func NewOpExprNullSafeMethodCall(vr, name Operand, args []Operand, varPos, namePos *position.Position, argsPos []*position.Position, pos *position.Position) *OpExprMethodCall {
	Op := &OpExprMethodCall{
		OpGeneral:  NewOpGeneral(pos),
		Var:        vr,
		Name:       name,
		Args:       args,
		Result:     NewTemporaryOperand(nil),
		IsNullSafe: true,
		VarPos:     varPos,
		NamePos:    namePos,
		ArgsPos:    argsPos,
	}

	AddUseRefs(Op, vr, name)
	AddUseRefs(Op, args...)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprMethodCall) GetType() string {
	return "ExprMethodCall"
}

func (op *OpExprMethodCall) GetName() string {
	className := ""
	switch c := op.Var.(type) {
	case *OperandObject:
		className = c.ClassName
	case *OperandVariable:
		if cv, ok := c.VariableValue.(*OperandObject); ok {
			className = cv.ClassName
		}
	case *TemporaryOperand:
		if co, ok := c.Original.(*OperandVariable); ok {
			if cv, ok := co.VariableValue.(*OperandObject); ok {
				className = cv.ClassName
			}
		}
	}
	funcName, _ := GetOperandName(op.Name)
	return className + "::" + funcName
}
func (op *OpExprMethodCall) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var":    op.Var,
		"Name":   op.Name,
		"Result": op.Result,
	}
}

func (op *OpExprMethodCall) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	case "Name":
		op.Name = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprMethodCall) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Args": op.Args,
	}
}

func (op *OpExprMethodCall) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Args":
		op.Args = vr
	}
}

// func (op *OpExprMethodCall) GetOpVarPos(vrName string) *position.Position {
// 	switch vrName {
// 	case "Var":
// 		return op.VarPos
// 	case "Name":
// 		return op.NamePos
// 	}
// 	return nil
// }

// func (op *OpExprMethodCall) GetOpVarListPos(vrName string, index int) *position.Position {
// 	switch vrName {
// 	case "Args":
// 		return op.ArgsPos[index]
// 	}
// 	return nil
// }

func (op *OpExprMethodCall) Clone() Op {
	args := make([]Operand, len(op.Args))
	copy(args, op.Args)
	return &OpExprMethodCall{
		OpGeneral:  op.OpGeneral,
		Var:        op.Var,
		Name:       op.Name,
		Args:       args,
		IsNullSafe: op.IsNullSafe,
		CalledFunc: op.CalledFunc,
		Result:     op.Result,
	}
}

type OpExprNew struct {
	OpGeneral
	Class  Operand
	Args   []Operand
	Result Operand
}

func NewOpExprNew(class Operand, args []Operand, pos *position.Position) *OpExprNew {
	Op := &OpExprNew{
		OpGeneral: NewOpGeneral(pos),
		Class:     class,
		Args:      args,
		Result:    NewTemporaryOperand(nil),
	}

	AddUseRef(Op, class)
	AddUseRefs(Op, args...)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprNew) GetType() string {
	return "ExprNew"
}

func (op *OpExprNew) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Class":  op.Class,
		"Result": op.Result,
	}
}

func (op *OpExprNew) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Class":
		op.Class = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprNew) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Args": op.Args,
	}
}

func (op *OpExprNew) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Args":
		op.Args = vr
	}
}

func (op *OpExprNew) Clone() Op {
	args := make([]Operand, len(op.Args))
	copy(args, op.Args)
	return &OpExprNew{
		OpGeneral: op.OpGeneral,
		Args:      args,
		Class:     op.Class,
		Result:    op.Result,
	}
}

type OpExprYield struct {
	OpGeneral
	Value  Operand
	Key    Operand
	Result Operand
}

func NewOpExprYield(value, key Operand, pos *position.Position) *OpExprYield {
	Op := &OpExprYield{
		OpGeneral: NewOpGeneral(pos),
		Value:     value,
		Key:       key,
		Result:    NewTemporaryOperand(nil),
	}

	AddUseRefs(Op, value, key)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprYield) GetType() string {
	return "ExprYield"
}

func (op *OpExprYield) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Value":  op.Value,
		"Key":    op.Key,
		"Result": op.Result,
	}
}

func (op *OpExprYield) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Value":
		op.Value = vr
	case "Key":
		op.Key = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprYield) Clone() Op {
	return &OpExprYield{
		OpGeneral: op.OpGeneral,
		Value:     op.Value,
		Key:       op.Key,
		Result:    op.Result,
	}
}

type OpExprAssertion struct {
	OpGeneral
	Expr      Operand
	Assertion Assertion
	Result    Operand
}

func NewOpExprAssertion(read, write Operand, assertion Assertion, pos *position.Position) *OpExprAssertion {
	Op := &OpExprAssertion{
		OpGeneral: NewOpGeneral(pos),
		Expr:      read,
		Assertion: assertion,
		Result:    write,
	}

	AddUseRef(Op, read)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprAssertion) GetType() string {
	return "ExprAssertion"
}

func (op *OpExprAssertion) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprAssertion) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprAssertion) Clone() Op {
	return &OpExprAssertion{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Assertion: op.Assertion,
		Result:    op.Result,
	}
}

type OpExprPrint struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprPrint(expr Operand, pos *position.Position) *OpExprPrint {
	Op := &OpExprPrint{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr:   expr,
		Result: NewTemporaryOperand(nil),
	}

	AddUseRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprPrint) GetType() string {
	return "ExprPrint"
}

func (op *OpExprPrint) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprPrint) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprPrint) Clone() Op {
	return &OpExprPrint{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprStaticCall struct {
	OpGeneral
	Class  Operand
	Name   Operand
	Args   []Operand
	Result Operand

	CalledFunc *Func
	ClassPos   *position.Position
	NamePos    *position.Position
	ArgsPos    []*position.Position
}

func NewOpExprStaticCall(class, name Operand, args []Operand, classPos, namePos *position.Position, argsPos []*position.Position, pos *position.Position) *OpExprStaticCall {
	Op := &OpExprStaticCall{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Class:    class,
		Name:     name,
		Args:     args,
		Result:   NewTemporaryOperand(nil),
		ClassPos: classPos,
		NamePos:  namePos,
		ArgsPos:  argsPos,
	}

	AddUseRefs(Op, class, name)
	AddUseRefs(Op, args...)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprStaticCall) GetType() string {
	return "ExprStaticCall"
}

func (op *OpExprStaticCall) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Class":  op.Class,
		"Name":   op.Name,
		"Result": op.Result,
	}
}

func (op *OpExprStaticCall) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Class":
		op.Class = vr
	case "Name":
		op.Name = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprStaticCall) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Args": op.Args,
	}
}

func (op *OpExprStaticCall) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Args":
		op.Args = vr
	}
}

// func (op *OpExprStaticCall) GetOpVarPos(vrName string) *position.Position {
// 	switch vrName {
// 	case "Class":
// 		return op.ClassPos
// 	case "Name":
// 		return op.NamePos
// 	}
// 	return nil
// }

// func (op *OpExprStaticCall) GetOpVarListPos(vrName string, index int) *position.Position {
// 	switch vrName {
// 	case "Args":
// 		return op.ArgsPos[index]
// 	}
// 	return nil
// }

func (op *OpExprStaticCall) GetName() string {
	// get class name
	className := ""
	switch c := op.Class.(type) {
	case *OperandObject:
		className = c.ClassName
	case *OperandVariable:
		if cv, ok := c.VariableValue.(*OperandObject); ok {
			className = cv.ClassName
		}
	case *TemporaryOperand:
		if co, ok := c.Original.(*OperandVariable); ok {
			if cv, ok := co.VariableValue.(*OperandObject); ok {
				className = cv.ClassName
			}
		}
	}
	funcName, err := GetOperandName(op.Name)
	if err != nil {
		log.Fatalf("Error in GetStaticCallName: %v", err)
	}
	return className + "::" + funcName
}

func (op *OpExprStaticCall) Clone() Op {
	args := make([]Operand, len(op.Args))
	copy(args, op.Args)
	return &OpExprStaticCall{
		OpGeneral:  op.OpGeneral,
		Class:      op.Class,
		Name:       op.Name,
		Args:       args,
		CalledFunc: op.CalledFunc,
		Result:     op.Result,
	}
}

type OpExprStaticPropertyFetch struct {
	OpGeneral
	Class  Operand
	Name   Operand
	Result Operand
}

func NewOpExprStaticPropertyFetch(class, name Operand, pos *position.Position) *OpExprStaticPropertyFetch {
	Op := &OpExprStaticPropertyFetch{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Class:  class,
		Name:   name,
		Result: NewTemporaryOperand(nil),
	}

	AddUseRefs(Op, class, name)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprStaticPropertyFetch) GetType() string {
	return "ExprStaticPropertyFetch"
}

func (op *OpExprStaticPropertyFetch) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Class":  op.Class,
		"Name":   op.Name,
		"Result": op.Result,
	}
}

func (op *OpExprStaticPropertyFetch) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Class":
		op.Class = vr
	case "Name":
		op.Name = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprStaticPropertyFetch) Clone() Op {
	return &OpExprStaticPropertyFetch{
		OpGeneral: op.OpGeneral,
		Class:     op.Class,
		Name:      op.Name,
		Result:    op.Result,
	}
}

type OpExprPropertyFetch struct {
	OpGeneral
	Var    Operand
	Name   Operand
	Result Operand
}

func NewOpExprPropertyFetch(vr, name Operand, pos *position.Position) *OpExprPropertyFetch {
	Op := &OpExprPropertyFetch{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Var:    vr,
		Name:   name,
		Result: NewTemporaryOperand(nil),
	}

	AddUseRefs(Op, vr, name)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprPropertyFetch) GetType() string {
	return "ExprPropertyFetch"
}

func (op *OpExprPropertyFetch) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var":    op.Var,
		"Name":   op.Name,
		"Result": op.Result,
	}
}

func (op *OpExprPropertyFetch) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	case "Name":
		op.Name = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprPropertyFetch) Clone() Op {
	return &OpExprPropertyFetch{
		OpGeneral: op.OpGeneral,
		Var:       op.Var,
		Name:      op.Name,
		Result:    op.Result,
	}
}

type OpReset struct {
	OpGeneral
	Var Operand
}

func NewOpReset(vr Operand, pos *position.Position) *OpReset {
	Op := &OpReset{
		OpGeneral: NewOpGeneral(pos),
		Var:       vr,
	}

	AddUseRef(Op, vr)

	return Op
}

func (op *OpReset) GetVar() Operand {
	return op.Var
}

func (op *OpReset) GetType() string {
	return "Reset"
}

func (op *OpReset) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var": op.Var,
	}
}

func (op *OpReset) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	}
}

func (op *OpReset) Clone() Op {
	return &OpReset{
		OpGeneral: op.OpGeneral,
		Var:       op.Var,
	}
}

// Iterator
type OpExprValid struct {
	OpGeneral
	Var    Operand
	Result Operand
}

func NewOpExprValid(vr Operand, pos *position.Position) *OpExprValid {
	Op := &OpExprValid{
		OpGeneral: NewOpGeneral(pos),
		Var:       vr,
		Result:    NewTemporaryOperand(nil),
	}

	AddUseRef(Op, vr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprValid) GetVar() Operand {
	return op.Var
}

func (op *OpExprValid) GetType() string {
	return "ExprValid"
}

func (op *OpExprValid) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var":    op.Var,
		"Result": op.Result,
	}
}

func (op *OpExprValid) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprValid) Clone() Op {
	return &OpExprValid{
		OpGeneral: op.OpGeneral,
		Var:       op.Var,
		Result:    op.Result,
	}
}

type OpExprKey struct {
	OpGeneral
	Var    Operand
	Result Operand
}

func NewOpExprKey(vr Operand, pos *position.Position) *OpExprKey {
	Op := &OpExprKey{
		OpGeneral: NewOpGeneral(pos),
		Var:       vr,
		Result:    NewTemporaryOperand(nil),
	}

	AddUseRef(Op, vr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprKey) GetVar() Operand {
	return op.Var
}

func (op *OpExprKey) GetType() string {
	return "ExprKey"
}

func (op *OpExprKey) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var":    op.Var,
		"Result": op.Result,
	}
}

func (op *OpExprKey) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprKey) Clone() Op {
	return &OpExprKey{
		OpGeneral: op.OpGeneral,
		Var:       op.Var,
		Result:    op.Result,
	}
}

type OpExprValue struct {
	OpGeneral
	Var    Operand
	ByRef  bool
	Result Operand
}

func NewOpExprValue(vr Operand, byRef bool, pos *position.Position) *OpExprValue {
	Op := &OpExprValue{
		OpGeneral: NewOpGeneral(pos),
		Var:       vr,
		ByRef:     byRef,
		Result:    NewTemporaryOperand(nil),
	}

	AddUseRef(Op, vr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprValue) GetVar() Operand {
	return op.Var
}

func (op *OpExprValue) GetType() string {
	return "ExprValue"
}

func (op *OpExprValue) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var":    op.Var,
		"Result": op.Result,
	}
}

func (op *OpExprValue) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprValue) Clone() Op {
	return &OpExprValue{
		OpGeneral: op.OpGeneral,
		Var:       op.Var,
		Result:    op.Result,
	}
}
