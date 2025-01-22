package cfg

import "github.com/VKCOM/php-parser/pkg/position"

type Op interface {
	GetType() string                 // Get String of Op Type
	GetPosition() *position.Position // Get Op Position
	SetFilePath(filePath string)     // Set Filepath
	GetFilePath() string             // Get Filepath

	GetOpVars() map[string]Operand       // Return Map with string as key and Operand
	GetOpListVars() map[string][]Operand // Return Map with string as key and slice of Operand

	ChangeOpVar(varName string, vr Operand)       // Change Operand value with given String Key in Map
	ChangeOpListVar(varName string, vr []Operand) // Replace array of Operand value with given String Key in Map

	SetBlock(*Block)
	GetBlock() *Block

	Clone() Op
}

type OpCallable interface {
	GetFunc() *Func
}
