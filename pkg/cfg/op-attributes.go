package cfg

import "github.com/VKCOM/php-parser/pkg/position"

type OpAttribute struct {
	Name Operand
	Args []Operand
	OpGeneral
}

type OpAttributeGroup struct {
	Attrs []*OpAttribute
	OpGeneral
}

//check
func NewOpAttributeGroup(attrs []*OpAttribute, pos *position.Position) *OpAttributeGroup {
	Op := &OpAttributeGroup{
		OpGeneral: NewOpGeneral(pos),
		Attrs:     attrs,
	}

	return Op
}
func (op *OpAttributeGroup) GetType() string {
	return "AttributeGroup"
}

func (op *OpAttributeGroup) Clone() Op {
	attrs := make([]*OpAttribute, len(op.Attrs))
	copy(attrs, op.Attrs)
	return &OpAttributeGroup{
		OpGeneral: op.OpGeneral,
		Attrs:     attrs,
	}
}

func NewOpAttribute(name Operand, args []Operand, pos *position.Position) *OpAttribute {

	op := &OpAttribute{
		Name:      name,
		Args:      args,
		OpGeneral: NewOpGeneral(pos),
	}
	AddReadRef(op, name)
	return op
}

func (op *OpAttribute) GetType() string {
	return "Attribute"
}

func (op *OpAttribute) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Name": op.Name,
	}
}

func (op *OpAttribute) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Name":
		op.Name = vr
	}
}

func (op *OpAttribute) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Args": op.Args,
	}
}

func (op *OpAttribute) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Args":
		op.Args = vr
	}
}

func (op *OpAttribute) Clone() Op {
	args := make([]Operand, len(op.Args))
	copy(args, op.Args)
	return &OpAttribute{
		OpGeneral: op.OpGeneral,
		Name:      op.Name,
		Args:      args,
	}
}
