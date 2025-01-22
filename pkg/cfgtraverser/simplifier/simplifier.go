package simplifier

import (
	"github.com/rxhunter00/XSS-Taint/pkg/cfg"
	"github.com/rxhunter00/XSS-Taint/pkg/cfgtraverser"
)

type Simplifier struct {
	Removed             map[*cfg.Block]struct{}
	RecursionProtection map[cfg.Op]struct{}
	TrivPhiCandidate    map[*cfg.OpPhi]*cfg.Block
	FilePath            string

	cfgtraverser.NullTraverser
}

func NewSimplifier() *Simplifier {

	return &Simplifier{}
}
func (t *Simplifier) EnterScript(script *cfg.Script) {
	t.FilePath = script.Filepath
}
func (t *Simplifier) EnterFunc(fn *cfg.Func) {
	t.Removed = make(map[*cfg.Block]struct{})
	t.RecursionProtection = make(map[cfg.Op]struct{})
}

func (t *Simplifier) LeaveFunc(fn *cfg.Func) {
	// remove trivial phi
	if fn.CFGBlock != nil {
		t.TrivPhiCandidate = make(map[*cfg.OpPhi]*cfg.Block)
		t.removeTrivialPhi(fn.CFGBlock)
	}
}

func (t *Simplifier) EnterOp(op cfg.Op, block *cfg.Block) {
	// Save Op to RecursionProtection as base case
	if InOpSet(t.RecursionProtection, op) {
		return
	}
	AddToOpSet(t.RecursionProtection, op)
	op.SetFilePath(t.FilePath)

	for targetName, target := range cfg.GetSubBlocks(op) {
		// Target block has no instruction
		if len(target.Instructions) <= 0 {
			continue
		}
		// jmpOp
		jmpOp, ok := target.Instructions[0].(*cfg.OpStmtJump)
		if !ok {
			continue
		}
		// get jump target
		jmpTarget := jmpOp.Target

		if InBlockSet(t.Removed, target) {
			// Target have been removed already
			cfg.ChangeSubBlock(op, targetName, jmpTarget)
			jmpTarget.AddPredecessor(block)
		} else {
			// Optimize child block
			t.EnterOp(jmpOp, target)

			// Prevent kill infinite tight loop
			if jmpOp.Target == target {
				continue
			}

			foundPhis := make([]*cfg.OpPhi, 0)

			// Get block phi
			jmptargetPhis := target.GetPhi()

			for _, phi := range jmptargetPhis {
				// get phi from child block
				for targetPhi := range jmpTarget.BlockPhi {
					// childPhi use phi value
					if targetPhi.HasOperand(phi.Result) {
						foundPhis = append(foundPhis, targetPhi)
						break
					}
				}
			}
			// Not all phi used by subblock
			if len(foundPhis) != len(target.BlockPhi) {
				continue
			}
			// here, we can remove phi node and jmp
			for i := 0; i < len(target.BlockPhi); i++ {
				phi := jmptargetPhis[i]
				foundPhi := foundPhis[i]
				// we can actually remove the phi node and teh jump
				foundPhi.RemoveOperand(phi.Result)
				for oper := range phi.Vars {
					foundPhi.AddOperand(oper)
				}
			}
			// empty block phi
			target.BlockPhi = make(map[*cfg.OpPhi]struct{})
			AddToBlockSet(t.Removed, target)
			target.Dead = true

			// remove target from list of preds
			jmpTarget.RemovePredecessor(target)
			jmpTarget.AddPredecessor(block)

			cfg.ChangeSubBlock(op, targetName, jmpTarget)
		}
	}
	RemoveFromOpSet(t.RecursionProtection, op)
}

func (t *Simplifier) removeTrivialPhi(block *cfg.Block) {
	toReplace := make(map[*cfg.Block]struct{})
	replaced := make(map[*cfg.Block]struct{})
	AddToBlockSet(toReplace, block)
	for len(toReplace) > 0 {
		for currBlock := range toReplace {
			RemoveFromBlockSet(toReplace, currBlock)
			AddToBlockSet(replaced, currBlock)
			for phi := range currBlock.BlockPhi {
				if t.tryRemoveTrivialPhi(phi, currBlock) {
					currBlock.RemovePhi(phi)
				}
			}
			for _, op := range currBlock.Instructions {
				for _, subBlock := range cfg.GetSubBlocks(op) {
					if !InBlockSet(replaced, subBlock) {
						AddToBlockSet(toReplace, subBlock)
					}
				}
			}
		}
	}
	for len(t.TrivPhiCandidate) > 0 {
		for phi, currBlock := range t.TrivPhiCandidate {
			delete(t.TrivPhiCandidate, phi)
			if t.tryRemoveTrivialPhi(phi, currBlock) {
				currBlock.RemovePhi(phi)
			}
		}
	}
}

func (t *Simplifier) tryRemoveTrivialPhi(phi *cfg.OpPhi, block *cfg.Block) bool {
	// phi variables more than 1, not trivial
	if len(phi.Vars) > 1 {
		return false
	}
	var vr cfg.Operand
	if len(phi.Vars) == 0 {
		return true
	} else {
		vr = phi.GetVars()[0]
		
		t.replaceVariables(phi.Result, vr, block)
	}

	return true
}

// remove operand which become trivial from a phi
func (t *Simplifier) replaceVariables(from, to cfg.Operand, block *cfg.Block) {
	toReplace := make(map[*cfg.Block]struct{})
	replaced := make(map[*cfg.Block]struct{})
	AddToBlockSet(toReplace, block)
	for len(toReplace) > 0 {
		for block := range toReplace {
			RemoveFromBlockSet(toReplace, block)
			AddToBlockSet(replaced, block)
			for phi := range block.BlockPhi {
				if phi.HasOperand(from) {
					// removing operand from phi, hence phi maybe become trivial
					t.TrivPhiCandidate[phi] = block
					phi.RemoveOperand(from)
					phi.AddOperand(to)
				}
			}
			for _, op := range block.Instructions {
				t.replaceOpVariable(from, to, op)
				for _, subBlock := range cfg.GetSubBlocks(op) {
					if !InBlockSet(replaced, subBlock) {
						AddToBlockSet(toReplace, subBlock)
					}
				}
				// propagate new value
				switch o := op.(type) {
				case *cfg.OpExprAssign:
					result := cfg.Operand(nil)
					switch r := o.Expr.(type) {
					case *cfg.OperandBool, *cfg.OperandNumber, *cfg.OperandObject, *cfg.OperandString, *cfg.OperandSymbolic:
						result = o.Expr
					case *cfg.OperandVariable:
						result = r.VariableValue
					case *cfg.TemporaryOperand:
						if rv, ok := r.Original.(*cfg.OperandVariable); ok {
							result = rv.VariableValue
						}
					}

					if result != nil {
						o.Result = result
						// get left variable, then give the value
						switch l := o.Var.(type) {
						case *cfg.OperandVariable:
							l.VariableValue = o.Result
						case *cfg.TemporaryOperand:
							if lv, ok := l.Original.(*cfg.OperandVariable); ok {
								lv.VariableValue = o.Result
							}
						}
					}
				}
			}
		}
	}
}

func (t *Simplifier) replaceOpVariable(from, to cfg.Operand, op cfg.Op) {
	for vrName, vr := range op.GetOpVars() {
		if vr == from {
			// change previous operand which is trivial phi
			op.ChangeOpVar(vrName, to)
			from.RemoveUser(op)
			if cfg.IsWriteVar(op, vrName) {
				to.AddWriter(op)
			} else {
				to.AddUser(op)
			}
		}
	}
	for vrName, vrList := range op.GetOpListVars() {
		new := make([]cfg.Operand, len(vrList))
		for i, vr := range vrList {
			if vr == from {
				new[i] = to
				to.AddUser(op)
				from.RemoveUser(op)
			} else {
				new[i] = vr
			}
		}
		op.ChangeOpListVar(vrName, new)
	}
}

func AddToBlockSet(set map[*cfg.Block]struct{}, item *cfg.Block) {
	set[item] = struct{}{}
}

func RemoveFromBlockSet(set map[*cfg.Block]struct{}, item *cfg.Block) {
	delete(set, item)
}

func InBlockSet(set map[*cfg.Block]struct{}, item *cfg.Block) bool {
	if _, ok := set[item]; ok {
		return true
	}
	return false
}

func AddToOpSet(set map[cfg.Op]struct{}, item cfg.Op) {
	set[item] = struct{}{}
}

func RemoveFromOpSet(set map[cfg.Op]struct{}, item cfg.Op) {
	delete(set, item)
}

func InOpSet(set map[cfg.Op]struct{}, item cfg.Op) bool {
	if _, ok := set[item]; ok {
		return true
	}
	return false
}
