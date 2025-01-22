package cfg

type AssertionMode int

const (
	ASSERTION_MODE_NONE AssertionMode = iota
	ASSERTION_MODE_UNION
	ASSERTION_MODE_INTERSECTION
)

type Assertion interface {
	GetNegation() Assertion
	Negated() bool
}

type TypeAssertion struct {
	AssertionOperand Operand
	IsNegated        bool
}

func NewTypeAssertion(oper Operand, isNegated bool) *TypeAssertion {
	return &TypeAssertion{
		AssertionOperand: oper,
		IsNegated:        isNegated,
	}

}
func (ta *TypeAssertion) GetNegation() Assertion {
	return &TypeAssertion{
		AssertionOperand: ta.AssertionOperand,
		IsNegated:        !ta.IsNegated,
	}
}

func (ta *TypeAssertion) Negated() bool {
	return ta.IsNegated
}

type CompositeAssertion struct {
	AssertionList []Assertion
	Mode          AssertionMode

	IsNegated bool
}

func NewCompositeAssertion(assertions []Assertion, mode AssertionMode, isNegated bool) *CompositeAssertion {
	return &CompositeAssertion{
		AssertionList: assertions,
		Mode:          mode,
		IsNegated:     isNegated,
	}

}
func (ca *CompositeAssertion) GetNegation() Assertion {
	return &CompositeAssertion{
		AssertionList: ca.AssertionList,
		Mode:          ca.Mode,
		IsNegated:     !ca.IsNegated,
	}
}

func (ca *CompositeAssertion) Negated() bool {
	return ca.IsNegated
}
