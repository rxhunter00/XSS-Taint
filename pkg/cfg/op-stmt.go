package cfg

import "github.com/VKCOM/php-parser/pkg/position"

type ClassModifFlag int

const (
	CLASS_MODIF_PUBLIC    ClassModifFlag = 1
	CLASS_MODIF_PROTECTED ClassModifFlag = 1 << 1
	CLASS_MODIF_PRIVATE   ClassModifFlag = 1 << 2
	CLASS_MODIF_STATIC    ClassModifFlag = 1 << 3
	CLASS_MODIF_ABSTRACT  ClassModifFlag = 1 << 4
	CLASS_MODIF_FINAL     ClassModifFlag = 1 << 5
	CLASS_MODIF_READONLY  ClassModifFlag = 1 << 6
)

type OpStmtClass struct {
	OpGeneral
	Name       Operand
	Stmts      *Block
	Flags      ClassModifFlag // bitmask storing modifiers
	Extends    Operand
	Implements []Operand
	AttrGroups []*OpAttributeGroup
}

func NewOpStmtClass(name Operand, stmts *Block, flags ClassModifFlag, extends Operand, implements []Operand, attrGroups []*OpAttributeGroup, pos *position.Position) *OpStmtClass {
	Op := &OpStmtClass{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Name:       name,
		Stmts:      stmts,
		Flags:      flags,
		Extends:    extends,
		Implements: implements,
		AttrGroups: attrGroups,
	}

	AddReadRef(Op, name)

	return Op
}

func (op *OpStmtClass) IsPublic() bool {
	return op.Flags&CLASS_MODIF_PUBLIC == 1
}

func (op *OpStmtClass) IsProtected() bool {
	return op.Flags&CLASS_MODIF_PROTECTED == 1
}

func (op *OpStmtClass) IsPrivate() bool {
	return op.Flags&CLASS_MODIF_PRIVATE == 1
}

func (op *OpStmtClass) IsStatic() bool {
	return op.Flags&CLASS_MODIF_STATIC == 1
}

func (op *OpStmtClass) IsAbstract() bool {
	return op.Flags&CLASS_MODIF_ABSTRACT == 1
}

func (op *OpStmtClass) IsFinal() bool {
	return op.Flags&CLASS_MODIF_FINAL == 1
}

func (op *OpStmtClass) IsReadonly() bool {
	return op.Flags&CLASS_MODIF_READONLY == 1
}

func (op *OpStmtClass) GetType() string {
	return "StmtClass"
}

func (op *OpStmtClass) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Name":    op.Name,
		"Extends": op.Extends,
	}
}

func (op *OpStmtClass) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Name":
		op.Name = vr
	case "Extends":
		op.Extends = vr
	}
}

func (op *OpStmtClass) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Implements": op.Implements,
	}
}

func (op *OpStmtClass) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Implements":
		op.Implements = vr
	}
}

func (op *OpStmtClass) Clone() Op {
	implements := make([]Operand, len(op.Implements))
	copy(implements, op.Implements)
	return &OpStmtClass{
		OpGeneral:  op.OpGeneral,
		Name:       op.Name,
		Stmts:      op.Stmts,
		Flags:      op.Flags,
		Extends:    op.Extends,
		Implements: implements,
		AttrGroups: op.AttrGroups,
	}
}

type OpStmtClassMethod struct {
	OpGeneral
	Func       *Func
	AttrGroups []*OpAttributeGroup
	Visibility FuncModifFlag
	Static     bool
	Final      bool
	Abstract   bool
}

func NewOpStmtClassMethod(function *Func, attrGroups []*OpAttributeGroup, visibility FuncModifFlag, static bool, final bool, abstract bool, pos *position.Position) *OpStmtClassMethod {
	Op := &OpStmtClassMethod{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Func:       function,
		AttrGroups: attrGroups,
		Visibility: visibility,
		Static:     static,
		Final:      final,
		Abstract:   abstract,
	}

	return Op
}

func (op *OpStmtClassMethod) GetType() string {
	return "StmtClassMethod"
}

func (op *OpStmtClassMethod) GetFunc() *Func {
	return op.Func
}

func (op *OpStmtClassMethod) Clone() Op {
	attrGroups := make([]*OpAttributeGroup, len(op.AttrGroups))
	copy(attrGroups, op.AttrGroups)
	return &OpStmtClassMethod{
		OpGeneral:  op.OpGeneral,
		Func:       op.Func,
		AttrGroups: attrGroups,
		Visibility: op.Visibility,
		Static:     op.Static,
		Final:      op.Final,
		Abstract:   op.Abstract,
	}
}

type OpStmtFunc struct {
	OpGeneral
	Func       *Func
	AttrGroups []*OpAttributeGroup
}

func NewOpStmtFunc(function *Func, attrGroups []*OpAttributeGroup, pos *position.Position) *OpStmtFunc {
	Op := &OpStmtFunc{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Func:       function,
		AttrGroups: attrGroups,
	}

	return Op
}

func (op *OpStmtFunc) GetType() string {
	return "StmtFunc"
}

func (op *OpStmtFunc) GetFunc() *Func {
	return op.Func
}

func (op *OpStmtFunc) Clone() Op {
	attrGroups := make([]*OpAttributeGroup, len(op.AttrGroups))
	copy(attrGroups, op.AttrGroups)
	return &OpStmtFunc{
		OpGeneral:  op.OpGeneral,
		Func:       op.Func,
		AttrGroups: attrGroups,
	}
}

type OpStmtInterface struct {
	OpGeneral
	Name    Operand
	Stmts   *Block
	Extends []Operand
}

func NewOpStmtInterface(name Operand, stmts *Block, extends []Operand, pos *position.Position) *OpStmtInterface {
	Op := &OpStmtInterface{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Name:    name,
		Stmts:   stmts,
		Extends: extends,
	}

	AddReadRef(Op, name)

	return Op
}

func (op *OpStmtInterface) GetType() string {
	return "StmtInterface"
}

func (op *OpStmtInterface) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Name": op.Name,
	}
}

func (op *OpStmtInterface) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Name":
		op.Name = vr
	}
}

func (op *OpStmtInterface) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Extends": op.Extends,
	}
}

func (op *OpStmtInterface) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Extends":
		op.Extends = vr
	}
}

func (op *OpStmtInterface) Clone() Op {
	extends := make([]Operand, len(op.Extends))
	copy(extends, op.Extends)
	return &OpStmtInterface{
		OpGeneral: op.OpGeneral,
		Name:      op.Name,
		Stmts:     op.Stmts,
		Extends:   extends,
	}
}

type OpStmtJump struct {
	OpGeneral
	Target *Block
}

func NewOpStmtJump(target *Block, pos *position.Position) *OpStmtJump {
	Op := &OpStmtJump{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Target: target,
	}

	return Op
}

func (op *OpStmtJump) GetType() string {
	return "StmtJump"
}

func (op *OpStmtJump) Clone() Op {
	return &OpStmtJump{
		OpGeneral: op.OpGeneral,
		Target:    op.Target,
	}
}

type OpStmtJumpIf struct {
	OpGeneral
	Cond Operand
	If   *Block
	Else *Block
}

func NewOpStmtJumpIf(cond Operand, ifBlock *Block, elseBlock *Block, pos *position.Position) *OpStmtJumpIf {
	Op := &OpStmtJumpIf{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Cond: cond,
		If:   ifBlock,
		Else: elseBlock,
	}

	AddReadRef(Op, cond)

	return Op
}

func (op *OpStmtJumpIf) GetType() string {
	return "StmtJumpIf"
}

func (op *OpStmtJumpIf) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Cond": op.Cond,
	}
}

func (op *OpStmtJumpIf) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Cond":
		op.Cond = vr
	}
}

func (op *OpStmtJumpIf) Clone() Op {
	return &OpStmtJumpIf{
		OpGeneral: op.OpGeneral,
		Cond:      op.Cond,
		If:        op.If,
		Else:      op.Else,
	}
}

type OpStmtProperty struct {
	OpGeneral
	Name         Operand
	Visibility   ClassModifFlag
	Static       bool
	ReadOnly     bool
	AttrGroups   []*OpAttributeGroup
	DefaultVar   Operand
	DefaultBlock *Block
	DeclaredType OpType
}

func NewOpStmtProperty(name Operand, visibility ClassModifFlag, static bool, readOnly bool, attrGroups []*OpAttributeGroup,
	defaultVar Operand, defaultBlock *Block, declaredType OpType, pos *position.Position) *OpStmtProperty {
	Op := &OpStmtProperty{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Name:         name,
		Visibility:   visibility,
		Static:       static,
		ReadOnly:     readOnly,
		AttrGroups:   attrGroups,
		DefaultVar:   defaultVar,
		DefaultBlock: defaultBlock,
		DeclaredType: declaredType,
	}

	return Op
}

func (op *OpStmtProperty) IsPublic() bool {
	return op.Visibility == CLASS_MODIF_PUBLIC
}

func (op *OpStmtProperty) IsProtected() bool {
	return op.Visibility == CLASS_MODIF_PROTECTED
}

func (op *OpStmtProperty) IsPrivate() bool {
	return op.Visibility == CLASS_MODIF_PRIVATE
}

func (op *OpStmtProperty) GetType() string {
	return "StmtProperty"
}

func (op *OpStmtProperty) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Name":       op.Name,
		"DefaultVar": op.DefaultVar,
	}
}

func (op *OpStmtProperty) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Name":
		op.Name = vr
	case "DefaultVar":
		op.DefaultVar = vr
	}
}

func (op *OpStmtProperty) Clone() Op {
	attrGroups := make([]*OpAttributeGroup, len(op.AttrGroups))
	copy(attrGroups, op.AttrGroups)
	return &OpStmtProperty{
		OpGeneral:    op.OpGeneral,
		Name:         op.Name,
		Visibility:   op.Visibility,
		Static:       op.Static,
		ReadOnly:     op.ReadOnly,
		AttrGroups:   attrGroups,
		DefaultVar:   op.DefaultVar,
		DefaultBlock: op.DefaultBlock,
		DeclaredType: op.DeclaredType,
	}
}

type OpStmtSwitch struct {
	OpGeneral
	Cond          Operand
	Cases         []Operand
	Targets       []*Block
	DefaultTarget *Block
}

func NewOpStmtSwitch(cond Operand, cases []Operand, targets []*Block, defaultTarget *Block, pos *position.Position) *OpStmtSwitch {
	Op := &OpStmtSwitch{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Cond:          cond,
		Cases:         cases,
		Targets:       targets,
		DefaultTarget: defaultTarget,
	}

	AddReadRef(Op, cond)
	AddReadRefs(Op, cases...)

	return Op
}

func (op *OpStmtSwitch) GetType() string {
	return "StmtSwitch"
}

func (op *OpStmtSwitch) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Cond": op.Cond,
	}
}

func (op *OpStmtSwitch) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Cond":
		op.Cond = vr
	}
}

func (op *OpStmtSwitch) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Cases": op.Cases,
	}
}

func (op *OpStmtSwitch) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Cases":
		op.Cases = vr
	}
}

func (op *OpStmtSwitch) Clone() Op {
	return &OpStmtSwitch{
		OpGeneral:     op.OpGeneral,
		Cond:          op.Cond,
		Cases:         op.Cases,
		Targets:       op.Targets,
		DefaultTarget: op.DefaultTarget,
	}
}

type OpStmtTrait struct {
	OpGeneral
	Name  Operand
	Stmts *Block
}

func NewOpStmtTrait(name Operand, stmts *Block, pos *position.Position) *OpStmtTrait {
	Op := &OpStmtTrait{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Name:  name,
		Stmts: stmts,
	}

	AddReadRef(Op, name)

	return Op
}

func (op *OpStmtTrait) GetType() string {
	return "StmtTrait"
}

func (op *OpStmtTrait) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Name": op.Name,
	}
}

func (op *OpStmtTrait) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Name":
		op.Name = vr
	}
}

func (op *OpStmtTrait) Clone() Op {
	return &OpStmtTrait{
		OpGeneral: op.OpGeneral,
		Name:      op.Name,
		Stmts:     op.Stmts,
	}
}

type OpStmtTraitUse struct {
	OpGeneral
	Traits      []Operand
	Adaptations []Op
}

func NewOpStmtTraitUse(traits []Operand, adaptations []Op, pos *position.Position) *OpStmtTraitUse {
	Op := &OpStmtTraitUse{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Traits:      traits,
		Adaptations: adaptations,
	}

	return Op
}

func (op *OpStmtTraitUse) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Traits": op.Traits,
	}
}

func (op *OpStmtTraitUse) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Traits":
		op.Traits = vr
	}
}

func (op *OpStmtTraitUse) GetType() string {
	return "StmtTraitUse"
}

func (op *OpStmtTraitUse) Clone() Op {
	return &OpStmtTraitUse{
		OpGeneral:   op.OpGeneral,
		Traits:      op.Traits,
		Adaptations: op.Adaptations,
	}
}

type OpAlias struct {
	OpGeneral
	Trait       Operand
	Method      Operand
	NewName     Operand
	NewModifier ClassModifFlag
}

func NewOpAlias(trait, method, newName Operand, newModifier ClassModifFlag, pos *position.Position) *OpAlias {
	Op := &OpAlias{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Trait:       trait,
		Method:      method,
		NewName:     newName,
		NewModifier: newModifier,
	}

	return Op
}

func (op *OpAlias) GetType() string {
	return "Alias"
}

func (op *OpAlias) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Trait":   op.Trait,
		"Method":  op.Method,
		"NewName": op.NewName,
	}
}

func (op *OpAlias) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Trait":
		op.Trait = vr
	case "Method":
		op.Method = vr
	case "NewName":
		op.NewName = vr
	}
}

func (op *OpAlias) Clone() Op {
	return &OpAlias{
		OpGeneral:   op.OpGeneral,
		Trait:       op.Trait,
		Method:      op.Method,
		NewName:     op.NewName,
		NewModifier: op.NewModifier,
	}
}

type OpPrecedence struct {
	OpGeneral
	Trait     Operand
	Method    Operand
	InsteadOf []Operand
}

func NewOpPrecedence(trait, method Operand, insteadOf []Operand, pos *position.Position) *OpPrecedence {
	Op := &OpPrecedence{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Trait:     trait,
		Method:    method,
		InsteadOf: insteadOf,
	}

	return Op
}

func (op *OpPrecedence) GetType() string {
	return "Precedence"
}

func (op *OpPrecedence) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Trait":  op.Trait,
		"Method": op.Method,
	}
}

func (op *OpPrecedence) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Trait":
		op.Trait = vr
	case "Method":
		op.Method = vr
	}
}

func (op *OpPrecedence) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Insteadof": op.InsteadOf,
	}
}

func (op *OpPrecedence) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Insteadof":
		op.InsteadOf = vr
	}
}

func (op *OpPrecedence) Clone() Op {
	return &OpPrecedence{
		OpGeneral: op.OpGeneral,
		Trait:     op.Trait,
		Method:    op.Method,
		InsteadOf: op.InsteadOf,
	}
}

type OpConst struct {
	OpGeneral
	Name       Operand
	Value      Operand
	ValueBlock *Block
}

func NewOpConst(name, value Operand, block *Block, pos *position.Position) *OpConst {
	Op := &OpConst{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Name:       name,
		Value:      value,
		ValueBlock: block,
	}

	AddReadRefs(Op, name, value)

	return Op
}

func (op *OpConst) GetType() string {
	return "Const"
}

func (op *OpConst) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Name":  op.Name,
		"Value": op.Value,
	}
}

func (op *OpConst) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Name":
		op.Name = vr
	case "Value":
		op.Value = vr
	}
}

func (op *OpConst) Clone() Op {
	return &OpConst{
		OpGeneral:  op.OpGeneral,
		Name:       op.Name,
		Value:      op.Value,
		ValueBlock: op.ValueBlock,
	}
}

type OpEcho struct {
	OpGeneral
	Expr Operand
}

func NewOpEcho(expr Operand, pos *position.Position) *OpEcho {
	Op := &OpEcho{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr: expr,
	}

	AddReadRef(Op, expr)

	return Op
}

func (op *OpEcho) GetType() string {
	return "Echo"
}

func (op *OpEcho) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr": op.Expr,
	}
}

func (op *OpEcho) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	}
}

func (op *OpEcho) Clone() Op {
	return &OpEcho{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
	}
}

type OpExit struct {
	OpGeneral
	Expr Operand
}

func NewOpExit(expr Operand, pos *position.Position) *OpExit {
	Op := &OpExit{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr: expr,
	}

	AddReadRef(Op, expr)

	return Op
}

func (op *OpExit) GetType() string {
	return "Exit"
}

func (op *OpExit) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr": op.Expr,
	}
}

func (op *OpExit) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	}
}

func (op *OpExit) Clone() Op {
	return &OpExit{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
	}
}

type OpGlobalVar struct {
	OpGeneral
	Var Operand
}

func NewOpGlobalVar(vr Operand, pos *position.Position) *OpGlobalVar {
	Op := &OpGlobalVar{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Var: vr,
	}

	AddReadRef(Op, vr)

	return Op
}

func (op *OpGlobalVar) GetType() string {
	return "GlobalVar"
}

func (op *OpGlobalVar) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var": op.Var,
	}
}

func (op *OpGlobalVar) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	}
}

func (op *OpGlobalVar) Clone() Op {
	return &OpGlobalVar{
		OpGeneral: op.OpGeneral,
		Var:       op.Var,
	}
}

type OpReturn struct {
	OpGeneral
	Expr Operand
}

func NewOpReturn(expr Operand, pos *position.Position) *OpReturn {
	Op := &OpReturn{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr: expr,
	}

	AddReadRef(Op, expr)

	return Op
}

func (op *OpReturn) GetType() string {
	return "Return"
}

func (op *OpReturn) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr": op.Expr,
	}
}

func (op *OpReturn) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	}
}

func (op *OpReturn) Clone() Op {
	return &OpReturn{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
	}
}

type OpStaticVar struct {
	OpGeneral
	Var          Operand
	DefaultVar   Operand
	DefaultBlock *Block
}

func NewOpStaticVar(vr, defaultVr Operand, defaultBlock *Block, pos *position.Position) *OpStaticVar {
	Op := &OpStaticVar{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Var:          vr,
		DefaultVar:   defaultVr,
		DefaultBlock: defaultBlock,
	}

	AddReadRef(Op, defaultVr)
	AddWriteRef(Op, vr)

	return Op
}

func (op *OpStaticVar) GetType() string {
	return "StaticVar"
}

func (op *OpStaticVar) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var":        op.Var,
		"DefaultVar": op.DefaultVar,
	}
}

func (op *OpStaticVar) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	case "DefaultVar":
		op.DefaultVar = vr
	}
}

func (op *OpStaticVar) Clone() Op {
	return &OpStaticVar{
		OpGeneral:    op.OpGeneral,
		Var:          op.Var,
		DefaultVar:   op.DefaultVar,
		DefaultBlock: op.DefaultBlock,
	}
}

type OpThrow struct {
	OpGeneral
	Expr Operand
}

func NewOpThrow(expr Operand, pos *position.Position) *OpThrow {
	Op := &OpThrow{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr: expr,
	}

	AddReadRef(Op, expr)

	return Op
}

func (op *OpThrow) GetType() string {
	return "Throq"
}

func (op *OpThrow) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr": op.Expr,
	}
}

func (op *OpThrow) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	}
}

func (op *OpThrow) Clone() Op {
	return &OpThrow{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
	}
}

type OpUnset struct {
	OpGeneral
	Exprs []Operand

	ExprsPos []*position.Position
}

func NewOpUnset(exprs []Operand, exprsPos []*position.Position, pos *position.Position) *OpUnset {
	Op := &OpUnset{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Exprs:    exprs,
		ExprsPos: exprsPos,
	}

	AddReadRefs(Op, exprs...)

	return Op
}

func (op *OpUnset) GetType() string {
	return "Unset"
}

func (op *OpUnset) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Exprs": op.Exprs,
	}
}

func (op *OpUnset) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Exprs":
		op.Exprs = vr
	}
}

func (op *OpUnset) Clone() Op {
	return &OpUnset{
		OpGeneral: op.OpGeneral,
		Exprs:     op.Exprs,
	}
}
