package cfg

import (
	"fmt"
	"reflect"
)

type VarAssert struct {
	Var    Operand
	Assert Assertion
}

type Operand interface {
	AddUser(op Op)    // Add Op to the list of Op that use this Operand
	AddWriter(op Op)  // Add Op to the list of Op that define this Operand
	RemoveUser(op Op) // Remove Op to the list of Op that use this Operand

	AddAssertion(op Operand, assert Assertion, mode AssertionMode)
	GetAssertions() []VarAssert
	GetUsers() []Op
	GetWriter() Op      // Get last op
	GetWriterOps() []Op //Get list op

	AddCondUsage(block *Block)
	GetCondUsages() []*Block

	IsTainted() bool
	String() string
	IsWritten() bool
}

// Operand Null
type OperandNull struct {
	OperandAttributes
}

func NewOperandNull() *OperandNull {
	return &OperandNull{
		OperandAttributes: NewOperandAttributes(false),
	}
}

func (on *OperandNull) String() string {
	return "OPERANDNULL"
}

func GetOperNamed(oper Operand) *OperandString {
	if opT, ok := oper.(*TemporaryOperand); ok {
		if orig, ok := opT.Original.(*OperandVariable); ok {
			if name, ok := orig.VariableName.(*OperandString); ok {
				return name
			}
		}
	}
	return nil
}

func IsScalarOper(oper Operand) bool {
	switch oper.(type) {
	case *OperandBool, *OperandNumber, *OperandString:
		return true
	}
	return false
}

func GetStringOper(oper Operand) (string, bool) {
	if os, ok := oper.(*OperandString); ok {
		return os.Val, true
	}
	return "", false
}

func GetOperName(oper Operand) (string, error) {
	switch o := oper.(type) {
	case *OperandBoundVariable:
		return GetOperName(o.Name)
	case *OperandVariable:
		return GetOperName(o.VariableName)
	case *OperandString:
		return o.Val, nil
	case *TemporaryOperand:
		return GetOperName(o.Original)
	case *OperandBool, *OperandNumber, *OperandNull, *OperandSymbolic:
		return "", fmt.Errorf("operand doesn't have name '%v'", reflect.TypeOf(o))
	}
	return "", fmt.Errorf("operand doesn't have name '%v'", reflect.TypeOf(oper))
}

// Get the deepest value of operand
func GetOperVal(oper Operand) Operand {
	switch o := oper.(type) {
	case *OperandBoundVariable:
		return GetOperVal(o.Value)
	case *OperandVariable:
		return GetOperVal(o.VariableValue)
	case *TemporaryOperand:
		return GetOperVal(o.Original)
	case *OperandString, *OperandBool, *OperandNumber, *OperandNull, *OperandObject, *OperandSymbolic:
		return oper
	}
	return NewOperandNull()
}

// Set operand value
func SetOperVal(oper Operand, val Operand) {
	switch o := oper.(type) {
	case *TemporaryOperand:
		SetOperVal(o.Original, val)
	case *OperandBoundVariable:
		o.Value = val
	case *OperandVariable:
		o.VariableValue = val
	}

}
