package sourcefinder

import (
	"log"

	"github.com/rxhunter00/XSS-Taint/pkg/cfg"
	"github.com/rxhunter00/XSS-Taint/pkg/cfgtraverser"
)

type SourceFinder struct {
	cfgtraverser.NullTraverser

	CurrScript *cfg.Script
	CurrFunc   *cfg.Func
}

func NewSourceFinder() *SourceFinder {
	return &SourceFinder{}
}

func (t *SourceFinder) EnterScript(script *cfg.Script) {
	t.CurrScript = script
}

func (t *SourceFinder) EnterFunc(fn *cfg.Func) {
	t.CurrFunc = fn
}

func (t *SourceFinder) LeaveFunc(fn *cfg.Func) {
	t.CurrFunc = nil
}

func (t *SourceFinder) EnterOp(op cfg.Op, block *cfg.Block) {
	// if source, add to sources
	if t.isSource(op) {
		t.CurrFunc.Sources = append(t.CurrFunc.Sources, op)
	}
}

func (t *SourceFinder) isSource(op cfg.Op) bool {

	switch opT := op.(type) {
	case *cfg.OpExprAssign:
		if right, ok := opT.Expr.(*cfg.OperandSymbolic); ok {
			switch right.Val {
			case "globalposts":
				return true
			case "globalgets":
				return true
			case "globalrequest":
				return true
			case "globalfiles":
				return true
			case "globalcookie":
				return true
			case "globalservers":
				return true
			}
		}
	case *cfg.OpExprFunctionCall:
		funcNameStr, _ := cfg.GetOperandName(opT.Name)
		switch funcNameStr {
		case "filter_input_array":
			if len(opT.Args) == 1 {
				return true
			} else {
				filter := opT.Args[1].GetWriter()
				switch filterOp := filter.(type) {
				case *cfg.OpExprConstFetch:
					constName, err := cfg.GetOperandName(filterOp.Name)
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
					constName, err := cfg.GetOperandName(filterOp.Name)
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
			case "globalposts":
				fallthrough
			case "globalgets":
				fallthrough
			case "globalrequest":
				fallthrough
			case "globalfiles":
				fallthrough
			case "globalcookie":
				fallthrough
			case "globalservers":
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
				case "globalposts":
					return true
				case "globalgets":
					return true
				case "globalrequest":
					return true
				case "globalfiles":
					return true
				case "globalcookie":
					return true
				case "globalservers":
					return true
				}
			}
		}
	}
	return false
}
