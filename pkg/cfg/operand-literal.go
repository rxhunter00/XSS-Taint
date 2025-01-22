package cfg

import (
	"fmt"
	"strconv"
)

// Scalar Operand

// String Type
type OperandString struct {
	Val string
	OperandAttributes
}

func NewOperandString(val string) *OperandString {
	return &OperandString{
		OperandAttributes: NewOperandAttributes(false),
		Val:               val,
	}
}

func (os *OperandString) String() string {
	return fmt.Sprintf("OPERANDLITERAL(%s)", os.Val)
}

// Bool Type
type OperandBool struct {
	Val bool
	OperandAttributes
}

func NewOperandBool(val bool) *OperandBool {
	return &OperandBool{
		OperandAttributes: NewOperandAttributes(false),
		Val:               val,
	}
}

func (ob *OperandBool) String() string {
	if ob.Val {
		return "OPERANDLITERAL(TRUE)"
	}
	return "OPERANDLITERAL(FALSE)"
}

// Number
type OperandNumber struct {
	Val float64
	OperandAttributes
}

func NewOperandNumber(val float64) *OperandNumber {
	return &OperandNumber{
		Val:               val,
		OperandAttributes: NewOperandAttributes(false),
	}
}
func (on *OperandNumber) String() string {
	formatVal := strconv.FormatFloat(on.Val, 'f', -1, 64)
	return fmt.Sprintf("OPERANDLITERAL(%s)", formatVal)
}

type OperandSymbolic struct {
	Val string
	OperandAttributes
}

func NewOperandSymbolic(val string, tainted bool) *OperandSymbolic {
	return &OperandSymbolic{
		Val:               val,
		OperandAttributes: NewOperandAttributes(tainted),
	}
}

func (oper *OperandSymbolic) String() string {
	return fmt.Sprintf("OPERANDSYMBOLIC(%s)", oper.Val)
}

type OperandObject struct {
	ClassName string
	OperandAttributes
}

func NewOperandObject(classname string) *OperandObject {

	return &OperandObject{
		ClassName:         classname,
		OperandAttributes: NewOperandAttributes(false),
	}
}

func (oo *OperandObject) String() string {
	return fmt.Sprintf("OPERANDOBJECT(%s)", oo.ClassName)
}
