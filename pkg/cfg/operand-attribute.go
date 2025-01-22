package cfg

type OperandAttributes struct {
	AssertionsList []VarAssert
	WriterOps      []Op // which op define this operand
	UserOps        []Op // which op use this operand
	CondBlock      []*Block

	Tainted bool
}

func NewOperandAttributes(isTainted bool) OperandAttributes {
	return OperandAttributes{
		AssertionsList: make([]VarAssert, 0),
		WriterOps:      make([]Op, 0),
		UserOps:        make([]Op, 0),
		Tainted:        isTainted,
	}
}

// Add Op to the list of Op that use this Operand / the operand is right side of op
func (oa *OperandAttributes) AddUser(op Op) {
	for _, usage := range oa.UserOps {
		if usage == op {
			return
		}
	}
	oa.UserOps = append(oa.UserOps, op)

}

// Add Op to the list of Op that define this operand / the operand is left side of assignment
func (oa *OperandAttributes) AddWriter(op Op) {
	for _, usage := range oa.WriterOps {
		if usage == op {
			return
		}
	}
	oa.WriterOps = append(oa.WriterOps, op)
}

// Remove Op to the list of Op that use this Operand / the operand is right side of op
func (oa *OperandAttributes) RemoveUser(op Op) {

	for i, usage := range oa.UserOps {
		if usage == op {
			oa.UserOps = append(oa.UserOps[:i], oa.UserOps[i+1:]...)
		}
	}
}

func (oa *OperandAttributes) AddAssertion(oper Operand, assert Assertion, mode AssertionMode) {
	// VarAssert contain assertion and operand
	for i, currVarAssertion := range oa.AssertionsList {
		// Check if operand already in one of VarAssertion.var
		if oper == currVarAssertion.Var {
			// if so make new slice of assertion with those two assertion
			temp := []Assertion{currVarAssertion.Assert, assert}
			oa.AssertionsList[i].Assert = NewCompositeAssertion(temp, mode, false)
			return
		}
		operName := GetOperNamed(oper)
		varName := GetOperNamed(currVarAssertion.Var)
		if operName != nil && varName != nil && operName.Val == varName.Val {
			temp := []Assertion{currVarAssertion.Assert, assert}
			oa.AssertionsList[i].Assert = NewCompositeAssertion(temp, mode, false)
			return
		}
	}
}

func (oa *OperandAttributes) GetAssertions() []VarAssert {
	return oa.AssertionsList
}

func (oa *OperandAttributes) GetUsers() []Op {
	return oa.UserOps
}

func (oa *OperandAttributes) GetWriter() Op {
	definerLen := len(oa.WriterOps)
	if definerLen > 0 {
		return oa.WriterOps[definerLen-1]
	}
	return nil
}
func (oa *OperandAttributes) GetWriterOps() []Op {

	return oa.WriterOps
}

func (oa *OperandAttributes) AddCondUsage(block *Block) {
	for _, blockCond := range oa.CondBlock {
		if block == blockCond {
			return
		}
	}
	oa.CondBlock = append(oa.CondBlock, block)
}

func (oa *OperandAttributes) GetCondUsages() []*Block {
	return oa.CondBlock
}

func (oa *OperandAttributes) IsTainted() bool {
	return oa.Tainted
}

func (oa *OperandAttributes) String() string {
	return "DEFAULT"
}

// Is an op already write to this
func (oa *OperandAttributes) IsWritten() bool {
	return len(oa.WriterOps) > 0
}
