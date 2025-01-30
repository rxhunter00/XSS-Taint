package cfg

type Block struct {
	BlockId      int
	Instructions []Op
	Predecesors  []*Block
	Dead         bool
	BlockPhi     map[*OpPhi]struct{}

	HasTainted         bool
	IsConditionalBlock bool
	Conditions         []Operand
}

func NewBlock(id int) *Block {
	return &Block{
		BlockId:      id,
		Instructions: make([]Op, 0),
		Predecesors:  make([]*Block, 0),
		Dead:         false,
		BlockPhi:     make(map[*OpPhi]struct{}),
		HasTainted:   false,
	}
}

func (b *Block) AddInstructions(op Op) {
	b.Instructions = append(b.Instructions, op)
}

// Add predecessor block to current block
func (b *Block) AddPredecessor(block *Block) {
	for _, pred := range b.Predecesors {
		if block == pred {
			return
		}
	}
	b.Predecesors = append(b.Predecesors, block)
}

// Remove predecessor block from current block
func (b *Block) RemovePredecessor(block *Block) {
	for i, pred := range b.Predecesors {
		if block == pred {
			// Remove the block
			b.Predecesors = append(b.Predecesors[:i], b.Predecesors[i+1:]...)
			return
		}
	}
}

func (b *Block) AddPhi(phi *OpPhi) {
	b.BlockPhi[phi] = struct{}{}
}

func (b *Block) RemovePhi(phi *OpPhi) {
	delete(b.BlockPhi, phi)
}

// Get phi slice from map of current block phi
func (b *Block) GetPhi() []*OpPhi {
	res := make([]*OpPhi, 0, len(b.BlockPhi))
	for phi := range b.BlockPhi {
		res = append(res, phi)
	}
	return res
}
// Set condition of c
func (b *Block) SetCondition(conds []Operand) {
	for _, cond := range conds {
		cond.AddCondUsage(b)
	}
	b.Conditions = make([]Operand, len(conds))
	copy(b.Conditions, conds)
}
