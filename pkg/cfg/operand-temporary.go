package cfg

import "fmt"

type TemporaryOperand struct {
	Original Operand
	OperandAttributes
}

func NewTemporaryOperand(ori Operand) *TemporaryOperand {
	return &TemporaryOperand{
		Original:          ori,
		OperandAttributes: NewOperandAttributes(false),
	}
}

func (to *TemporaryOperand) String() string {
	// Write operand original value
	if to.Original != nil {

		return fmt.Sprintf("TEMPOF(%s)", to.Original.String())
	}
	return fmt.Sprintf("TEMPOF(%s)", "NIL")

}
