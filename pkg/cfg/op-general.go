package cfg

import "github.com/VKCOM/php-parser/pkg/position"

type OpGeneral struct {
	Position *position.Position
	Filepath string
	OpBlock  *Block // which block op is
}

func NewOpGeneral(pos *position.Position) OpGeneral {
	return OpGeneral{
		Filepath: "",
		OpBlock:  nil,
		Position: pos,
	}
}

// Should Delete if conflicting
func (og *OpGeneral) GetType() string {
	return "OpGeneral"
}

func (og *OpGeneral) GetPosition() *position.Position {
	return og.Position
}

func (og *OpGeneral) SetFilePath(filePath string) {
	og.Filepath = filePath
}

func (og *OpGeneral) GetFilePath() string {
	return og.Filepath
}

func (og *OpGeneral) SetBlock(block *Block) {
	og.OpBlock = block
}

func (og *OpGeneral) GetBlock() *Block {
	return og.OpBlock
}

func (og *OpGeneral) GetOpVars() map[string]Operand {
	return map[string]Operand{}
}

func (og *OpGeneral) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{}
}

func (og *OpGeneral) ChangeOpVar(varName string, vr Operand) {

}

func (og *OpGeneral) ChangeOpListVar(varName string, vr []Operand) {

}

// func (og *OpGeneral) GetOpVarPos(varName string) *position.Position {
// 	return nil
// }

// func (og *OpGeneral) GetOpVarListPos(varName string, index int) *position.Position {
// 	return nil
// }
