package cfg

type FunctionContex struct {
	Labels          map[string]*Block
	UnresolvedGotos map[string][]*Block
	LocalVariables  map[*Block]map[string]Operand // Used to store local varaible definiton in each block
	IncompletePhis  map[*Block]map[string]*OpPhi
	CurrConds       []Operand
	IsComplete      bool
}

func NewFunctionContex() FunctionContex {
	return FunctionContex{
		Labels:          make(map[string]*Block),
		UnresolvedGotos: make(map[string][]*Block),
		LocalVariables:  make(map[*Block]map[string]Operand), // Store Local Var in each block scope
		IncompletePhis:  make(map[*Block]map[string]*OpPhi),
		CurrConds:       make([]Operand, 0),
		IsComplete:      false, // Flag for complete CFG
	}
}

func (funcctx *FunctionContex) SetLocalVar(scopeblock *Block, variablename string, variablevalue Operand) {
	if funcctx.LocalVariables[scopeblock] == nil {
		funcctx.LocalVariables[scopeblock] = make(map[string]Operand)
	}
	funcctx.LocalVariables[scopeblock][variablename] = variablevalue
}

func (funcctx *FunctionContex) GetLocalVar(scopeblock *Block, variablename string) (Operand, bool) {
	if funcctx.IsLocalVar(scopeblock, variablename) {
		return funcctx.LocalVariables[scopeblock][variablename], true
	}
	return nil, false

}
func (funcctx *FunctionContex) IsLocalVar(scopeblock *Block, variablename string) bool {
	// Check if map for the current block has been initialized
	variableMap, ok := funcctx.LocalVariables[scopeblock]

	if ok {
		_, isVarThere := variableMap[variablename]
		return isVarThere
	}
	return false

}

func (funcctx *FunctionContex) AddIncompletePhi(block *Block, varname string, phiop *OpPhi) {
	// Check if the current block has been initialized, if not initialized it
	if funcctx.IncompletePhis[block] == nil {
		funcctx.IncompletePhis[block] = make(map[string]*OpPhi)
	}
	// Save variable value in current block
	funcctx.IncompletePhis[block][varname] = phiop
}

func (funcctx *FunctionContex) GetLabel(name string) (*Block, bool) {
	a, ok := funcctx.Labels[name]

	return a, ok
}

// UnresolvedGotos map[string]*Block
func (funcctx *FunctionContex) AddUnresolvedGoto(name string, block *Block) {
	if funcctx.UnresolvedGotos[name] == nil {
		funcctx.UnresolvedGotos[name] = make([]*Block, 0)
	}

	funcctx.UnresolvedGotos[name] = append(funcctx.UnresolvedGotos[name], block)
}

func (funcctx *FunctionContex) GetUnresolvedGotos(name string) ([]*Block, bool) {
	a, ok := funcctx.UnresolvedGotos[name]

	return a, ok
}

func (funcctx *FunctionContex) RemoveGoto(name string) {
	delete(funcctx.UnresolvedGotos, name)
}

func (funcctx *FunctionContex) PushCond(cond Operand) {
	funcctx.CurrConds = append(funcctx.CurrConds, cond)
}

func (funcctx *FunctionContex) PopCond() {
	funcctx.CurrConds = funcctx.CurrConds[:len(funcctx.CurrConds)-1]
}
