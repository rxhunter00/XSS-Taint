package cfgtraverser

import "github.com/rxhunter00/XSS-Taint/pkg/cfg"

const (
	REMOVE_BLOCK = iota
	REMOVE_OP
)

type BlockTraverser interface {
	EnterScript(*cfg.Script)
	LeaveScript(*cfg.Script)
	EnterFunc(*cfg.Func)
	LeaveFunc(*cfg.Func)
	EnterBlock(*cfg.Block, *cfg.Block)
	LeaveBlock(*cfg.Block, *cfg.Block)
	SkipBlock(*cfg.Block, *cfg.Block)
	EnterOp(cfg.Op, *cfg.Block)
	LeaveOp(cfg.Op, *cfg.Block)
}

type CFGTraverser struct {
	BlockTravs []BlockTraverser

	seenBlock map[*cfg.Block]struct{}
}

func NewTraverser() *CFGTraverser {
	return &CFGTraverser{
		BlockTravs: make([]BlockTraverser, 0),
	}
}

func (t *CFGTraverser) AddBlockTraverser(traverser BlockTraverser) {
	t.BlockTravs = append(t.BlockTravs, traverser)
}

func (t *CFGTraverser) Traverse(script *cfg.Script) {
	t.EnterScript(script)
	t.TraverseFunc(script.Main)
	for _, fn := range script.FuncsMap {
		t.TraverseFunc(fn)
	}
	t.LeaveScript(script)
}

func (t *CFGTraverser) TraverseFunc(fn *cfg.Func) {
	t.seenBlock = make(map[*cfg.Block]struct{})
	t.EnterFunc(fn)
	block := fn.CFGBlock
	if block != nil {
		t.TraverseBlock(block, nil)
	}
	t.LeaveFunc(fn)
	t.seenBlock = nil
}

func (t *CFGTraverser) TraverseBlock(block *cfg.Block, prior *cfg.Block) {
	if t.InSeenBlock(block) {
		t.SkipBlock(block, prior)
		return
	}
	t.AddSeenBlock(block)
	t.EnterBlock(block, prior)
	ops := block.Instructions

	for _, op := range ops {
		t.EnterOp(op, block)
		switch opT := op.(type) {
		case *cfg.OpStmtJumpIf:
			t.TraverseBlock(opT.If, block)
			t.TraverseBlock(opT.Else, block)
		default:
			for _, subblock := range cfg.GetSubBlocks(op) {
				t.TraverseBlock(subblock, block)
			}
		}
		t.LeaveOp(op, block)
	}

	t.LeaveBlock(block, prior)
}

func (t *CFGTraverser) EnterScript(script *cfg.Script) {
	for _, trav := range t.BlockTravs {
		trav.EnterScript(script)
	}
}

func (t *CFGTraverser) LeaveScript(script *cfg.Script) {
	for _, trav := range t.BlockTravs {
		trav.LeaveScript(script)
	}
}

func (t *CFGTraverser) EnterBlock(block *cfg.Block, prior *cfg.Block) {
	for _, trav := range t.BlockTravs {
		trav.EnterBlock(block, prior)
	}
}

func (t *CFGTraverser) LeaveBlock(block *cfg.Block, prior *cfg.Block) {
	for _, trav := range t.BlockTravs {
		trav.LeaveBlock(block, prior)
	}
}

func (t *CFGTraverser) SkipBlock(block *cfg.Block, prior *cfg.Block) {
	for _, trav := range t.BlockTravs {
		trav.SkipBlock(block, prior)
	}
}

func (t *CFGTraverser) EnterFunc(fn *cfg.Func) {
	for _, trav := range t.BlockTravs {
		trav.EnterFunc(fn)
	}
}

func (t *CFGTraverser) LeaveFunc(fn *cfg.Func) {
	for _, trav := range t.BlockTravs {
		trav.LeaveFunc(fn)
	}
}

func (t *CFGTraverser) EnterOp(op cfg.Op, block *cfg.Block) {
	for _, trav := range t.BlockTravs {
		trav.EnterOp(op, block)
	}
}

func (t *CFGTraverser) LeaveOp(op cfg.Op, block *cfg.Block) {
	for _, trav := range t.BlockTravs {
		trav.LeaveOp(op, block)
	}
}

func (t *CFGTraverser) AddSeenBlock(block *cfg.Block) {
	t.seenBlock[block] = struct{}{}
}

func (t *CFGTraverser) InSeenBlock(block *cfg.Block) bool {
	if _, ok := t.seenBlock[block]; ok {
		return true
	}
	return false
}
