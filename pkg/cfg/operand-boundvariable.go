package cfg

import "fmt"

type BOUND_VAR_SCOPE int

const (
	BOUND_VAR_SCOPE_GLOBAL = iota
	BOUND_VAR_SCOPE_LOCAL
	BOUND_VAR_SCOPE_OBJECT
	BOUND_VAR_SCOPE_FUNCTION
)

// Variable immune to SSA

type OperandBoundVariable struct {
	Name  Operand
	Value Operand
	Scope BOUND_VAR_SCOPE
	ByRef bool
	Extra Operand
	OperandAttributes
}

func NewOperandBoundVariable(nameOper Operand, valOper Operand, scope BOUND_VAR_SCOPE, byref bool, extreOper Operand) *OperandBoundVariable {

	return &OperandBoundVariable{
		Name:              nameOper,
		Value:             valOper,
		Scope:             scope,
		ByRef:             byref,
		Extra:             extreOper,
		OperandAttributes: NewOperandAttributes(false),
	}
}

func (obv *OperandBoundVariable) String() string {
	return fmt.Sprintf("OPERANDBOUNDVARNAME(%s)", obv.Name.String())
}
