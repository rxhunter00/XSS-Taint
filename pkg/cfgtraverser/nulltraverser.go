package cfgtraverser

import "github.com/rxhunter00/XSS-Taint/pkg/cfg"

type NullTraverser struct {
}

func (t *NullTraverser) EnterScript(*cfg.Script) {}

func (t *NullTraverser) LeaveScript(*cfg.Script) {}

func (t *NullTraverser) EnterFunc(*cfg.Func) {}

func (t *NullTraverser) LeaveFunc(*cfg.Func) {}

func (t *NullTraverser) EnterBlock(*cfg.Block, *cfg.Block) {}

func (t *NullTraverser) LeaveBlock(*cfg.Block, *cfg.Block) {}

func (t *NullTraverser) SkipBlock(*cfg.Block, *cfg.Block) {}

func (t *NullTraverser) EnterOp(cfg.Op, *cfg.Block) {}

func (t *NullTraverser) LeaveOp(cfg.Op, *cfg.Block) {}
