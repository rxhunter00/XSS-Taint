package cfg

import "fmt"

type OperandVariable struct {
	VariableName  Operand
	VariableValue Operand
	OperandAttributes
}

func NewOperandVariable(varname Operand, varvalue Operand) *OperandVariable {
	if varvalue == nil {
		varvalue = NewOperandNull()
	}
	return &OperandVariable{
		VariableName:      varname,
		VariableValue:     varvalue, // Should be Literal or Variable
		OperandAttributes: NewOperandAttributes(false),
	}
}

func (ov *OperandVariable) String() string {

	return fmt.Sprintf("OPERANDVARIABLE(%s):(%s)", ov.VariableName.String(), ov.VariableValue.String())

}
