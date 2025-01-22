package cfg

// Used to determine function parameter type
import "github.com/VKCOM/php-parser/pkg/position"

type OP_TYPE int

const (
	TYPE_LITERAL OP_TYPE = iota
	TYPE_VOID
	TYPE_REFERENCE
	TYPE_UNION
	TYPE_MIXED
)

type OpType interface {
	GetKind() OP_TYPE
	Nullable() bool // In case
}

type OpTypeLiteral struct {
	Name       string
	IsNullable bool
	OpGeneral
}

func NewOpTypeLiteral(name string, isnullable bool, pos *position.Position) *OpTypeLiteral {
	return &OpTypeLiteral{
		Name:       name,
		IsNullable: isnullable,
		OpGeneral:  NewOpGeneral(pos),
	}

}
func (otl *OpTypeLiteral) GetKind() OP_TYPE {
	return TYPE_LITERAL
}

func (otl *OpTypeLiteral) Nullable() bool {
	return otl.IsNullable
}

func (otl *OpTypeLiteral) GetType() string {
	return "TypeLiteral"
}

// void cannot return any type
type OpTypeVoid struct {
	IsNullable bool
	OpGeneral
}

func NewOpTypeVoid(pos *position.Position) *OpTypeVoid {
	return &OpTypeVoid{
		OpGeneral: NewOpGeneral(pos),
	}

}
func (otv *OpTypeVoid) GetKind() OP_TYPE {
	return TYPE_VOID
}

func (otv *OpTypeVoid) Nullable() bool {
	return false
}

func (otv *OpTypeVoid) GetType() string {
	return "TypeVoid"
}

// int|float Union
type OpTypeUnion struct {
	UnionSubtypes []OpType
	OpGeneral
}

func NewOpTypeUnion(types []OpType, pos *position.Position) *OpTypeUnion {
	return &OpTypeUnion{
		UnionSubtypes: types,
		OpGeneral:     NewOpGeneral(pos),
	}
}
func (otu *OpTypeUnion) GetKind() OP_TYPE {
	return TYPE_UNION
}
func (otu *OpTypeUnion) Nullable() bool {
	return false
}
func (otu *OpTypeUnion) GetType() string {
	return "TypeUnion"
}

// reference
type OpTypeReference struct {
	Declaration Operand
	IsNullable  bool
	OpGeneral
}

func (otr *OpTypeReference) GetKind() OP_TYPE {
	return TYPE_REFERENCE
}
func (otr *OpTypeReference) Nullable() bool {
	return otr.IsNullable
}
func (otr *OpTypeReference) GetType() string {
	return "TypeReference"
}

// check wahtever this is used
// mixed type
type OpTypeMixed struct {
	OpGeneral
}

func NewOpTypeMixed(pos *position.Position) *OpTypeMixed {
	return &OpTypeMixed{
		OpGeneral: NewOpGeneral(pos),
	}

}

func (otm *OpTypeMixed) GetKind() OP_TYPE {
	return TYPE_MIXED
}

func (otm *OpTypeMixed) Nullable() bool {
	return false
}

func (otm *OpTypeMixed) GetType() string {
	return "TypeMixed"
}
