package sourcefinder

import (
	"log"
	"strings"

	"github.com/rxhunter00/XSS-Taint/pkg/cfg"
	"github.com/rxhunter00/XSS-Taint/pkg/cfgtraverser"
)

type SourceFinder struct {
	cfgtraverser.NullTraverser

	CurrScript     *cfg.Script
	CurrFunc       *cfg.Func
	ArrVars        map[*cfg.OpExprArrayDimFetch]string
	UnresolvedArrs map[*cfg.OpExprArrayDimFetch]string
	CurrClass      *cfg.OpStmtClass
}

func NewSourceFinder() *SourceFinder {
	return &SourceFinder{}
}

func (t *SourceFinder) EnterScript(script *cfg.Script) {
	t.CurrScript = script
}

func (t *SourceFinder) EnterFunc(fn *cfg.Func) {
	t.CurrFunc = fn
	t.ArrVars = make(map[*cfg.OpExprArrayDimFetch]string)
	t.UnresolvedArrs = make(map[*cfg.OpExprArrayDimFetch]string)
}

func (t *SourceFinder) LeaveFunc(fn *cfg.Func) {
	t.CurrFunc = nil
	t.ArrVars = nil
	t.UnresolvedArrs = nil
}

func (t *SourceFinder) EnterOp(op cfg.Op, block *cfg.Block) {
	// if source, add to sources
	if IsSource(op) {
		t.CurrFunc.Sources = append(t.CurrFunc.Sources, op)
	}

	// Resolve ArrayDimFetch
	switch opT := op.(type) {
	case *cfg.OpExprArrayDimFetch:
		arrDimStr := opT.ToString()
		for arr, arrStr := range t.ArrVars {
			if strings.HasPrefix(arrDimStr, arrStr) && opT != arr {
				arr.Result.AddUser(opT)
			}
		}
		t.UnresolvedArrs[opT] = arrDimStr
	case *cfg.OpExprAssign:
		for _, left := range opT.Var.GetWriterOps() {
			if left != nil {
				leftDef, ok := left.(*cfg.OpExprArrayDimFetch)
				leftDefStr := ""
				if ok {
					leftDefStr = leftDef.ToString()
				}
				if leftDefStr != "" {
					t.ArrVars[leftDef] = leftDefStr
					for arr, arrStr := range t.UnresolvedArrs {
						if strings.HasPrefix(arrStr, leftDefStr) {
							leftDef.Result.AddUser(arr)
						}
					}
				}
			}
		}
	}
}

func IsSource(op cfg.Op) bool {

	switch opT := op.(type) {
	case *cfg.OpExprAssign:
		if right, ok := opT.Expr.(*cfg.OperandSymbolic); ok {
			switch right.Val {
			case "postsymbolic":
				return true
			case "getsymbolic":
				return true
			case "requestsymbolic":
				return true
			case "filessymbolic":
				return true
			case "cookiesymbolic":
				return true
			case "serverssymbolic":
				return true
			}
		}
	case *cfg.OpExprFunctionCall:
		funcNameStr, _ := cfg.GetOperName(opT.Name)
		switch funcNameStr {
		case "filter_input_array":
			if len(opT.Args) == 1 {
				return true
			} else {
				filter := opT.Args[1].GetWriter()
				switch filterOp := filter.(type) {
				case *cfg.OpExprConstFetch:
					constName, err := cfg.GetOperName(filterOp.Name)
					if err != nil {
						log.Fatalf("IsSource: %v", err)
					}
					switch constName {
					case "FILTER_SANITIZE_NUMBER_INT":
						fallthrough
					case "FILTER_SANITIZE_NUMBER_FLOAT":
						return false
					default:
						return true
					}
				}
			}
		case "filter_input":
			if len(opT.Args) <= 2 {
				return true
			} else {
				filter := opT.Args[2].GetWriter()
				switch filterOp := filter.(type) {
				case *cfg.OpExprConstFetch:
					constName, err := cfg.GetOperName(filterOp.Name)
					if err != nil {
						log.Fatalf("IsSource: %v", err)
					}
					switch constName {
					case "FILTER_SANITIZE_NUMBER_INT":
						fallthrough
					case "FILTER_SANITIZE_NUMBER_FLOAT":
						return false
					default:
						return true
					}
				}
			}
		case "getallheaders":
			fallthrough
		case "apache_request_headers":
			return true
		}
	case *cfg.OpReset:
		return false
	case *cfg.OpExprArrayDimFetch:
		if right, ok := opT.Var.(*cfg.OperandSymbolic); ok {
			switch right.Val {
			case "postsymbolic":
				fallthrough
			case "getsymbolic":
				fallthrough
			case "requestsymbolic":
				fallthrough
			case "filessymbolic":
				fallthrough
			case "cookiesymbolic":
				fallthrough
			case "serverssymbolic":
				return true
			}
		} else if varName, ok := cfg.GetOperVal(opT.Var).(*cfg.OperandString); ok {
			if !ok {
				return false
			}
			switch varName.Val {
			case "$_POST":
				fallthrough
			case "$_GET":
				fallthrough
			case "$_REQUEST":
				fallthrough
			case "$_FILES":
				fallthrough
			case "$_COOKIE":
				fallthrough
			case "$_SERVERS":
				return true
			}
		}
	default:
		for _, vr := range op.GetOpVars() {
			if vr, ok := vr.(*cfg.OperandSymbolic); ok {
				switch vr.Val {
				case "postsymbolic":
					return true
				case "getsymbolic":
					return true
				case "requestsymbolic":
					return true
				case "filessymbolic":
					return true
				case "cookiesymbolic":
					return true
				case "serverssymbolic":
					return true
				}
			}
		}
	}
	return false
}
