package pathgenerator

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/rxhunter00/XSS-Taint/pkg/cfg"
)

type PathGenerator struct {
	paths    [][]cfg.Op
	currPath []cfg.Op
	visited  map[cfg.Op]map[cfg.Operand]struct{}
}

func NewPathGenerator() *PathGenerator {
	return &PathGenerator{
		paths: make([][]cfg.Op, 0),
	}
}

func GeneratePath(scripts map[string]*cfg.Script) [][]cfg.Op {
	pg := NewPathGenerator()
	for _, script := range scripts {
		pg.visited = make(map[cfg.Op]map[cfg.Operand]struct{})
		pg.traverseFunc(*script.Main)
		for _, fn := range script.FuncsMap {
			pg.traverseFunc(*fn)
		}
	}
	return pg.paths
}

func (pg *PathGenerator) traverseFunc(fn cfg.Func) {

	for _, source := range fn.Sources {

		sourceVar, err := GetTaintedVar(source)
		if err != nil {
			continue
		}
		pg.currPath = []cfg.Op{source}
		for _, sourceUser := range sourceVar.GetUsers() {
			// for each op that use this source
			err := pg.traceTaintedVar(sourceUser, sourceVar)
			if err != nil {
				log.Fatalf("traverseFunc:Error in '%s': %v", fn.Filepath, err)
			}
		}
	}
}
func (pg *PathGenerator) traceTaintedVar(userNode cfg.Op, taintedVar cfg.Operand) error {

	if IsSink(userNode, taintedVar) {

		newPath := make([]cfg.Op, len(pg.currPath))
		copy(newPath, pg.currPath)
		newPath = append(newPath, userNode)

		pg.paths = append(pg.paths, newPath)

		return nil
	} else if !IsPropagated(userNode, taintedVar) {
		return nil
	} else if _, ok := pg.visited[userNode]; ok {
		if _, ok := pg.visited[userNode][taintedVar]; ok {
			return nil
		}
	}

	if _, ok := pg.visited[userNode]; !ok {
		pg.visited[userNode] = make(map[cfg.Operand]struct{})
	}
	// Mark visted
	pg.visited[userNode][taintedVar] = struct{}{}

	// Get left operand or the result of op
	nodeVar, err := GetTaintedVar(userNode)
	if err != nil {
		return nil
	}

	// Get op that use that operand
	for _, taintedUsage := range nodeVar.GetUsers() {

		newPath := make([]cfg.Op, len(pg.currPath))
		copy(newPath, pg.currPath)
		newPath = append(newPath, taintedUsage)

		tmp := pg.currPath
		pg.currPath = newPath
		// Trace tainted for the result
		err := pg.traceTaintedVar(taintedUsage, nodeVar)
		if err != nil {
			return err
		}
		pg.currPath = tmp
		newPath = nil

	}

	return nil
}

func GetTaintedVar(op cfg.Op) (cfg.Operand, error) {
	if assignOp, ok := op.(*cfg.OpExprAssign); ok {
		if assignOp.Var != nil {
			return assignOp.Var, nil
		} else {
			return assignOp.Result, nil
		}
	} else if result, ok := op.GetOpVars()["Result"]; ok {
		if result != nil {
			return result, nil
		}
	}
	return nil, fmt.Errorf("GetTaintedVar:Error var '%v'", reflect.TypeOf(op))
}

// On Sanitizer return false,otherwise return true
func IsPropagated(op cfg.Op, taintedVar cfg.Operand) bool {

	switch opT := op.(type) {
	case *cfg.OpExprFunctionCall:
		funcNameStr, _ := cfg.GetOperName(opT.Name)
		switch funcNameStr {
		case "intval":
			return false
		case "floatval":
			return false
		case "boolval":
			return false
		case "doubleval":
			return false
		// URL Context
		case "rawurlencode":
			return false
		case "urlencode":
			return false
		//Java Script
		case "json_encode":
			return false
		}

	case *cfg.OpExprCastBool, *cfg.OpExprCastDouble, *cfg.OpExprCastInt, *cfg.OpExprCastUnset, *cfg.OpUnset:
		return false
	case *cfg.OpExprAssertion:
		switch assert := opT.Assertion.(type) {
		case *cfg.TypeAssertion:
			if typeVal, ok := assert.AssertionOperand.(*cfg.OperandString); ok {
				switch typeVal.Val {
				case "int", "float", "bool", "null":
					return false
				}
			}
		}

	case *cfg.OpExprArrayDimFetch:
		if opT.Dim == taintedVar {
			return false
		}

	}

	// Function that need constant as parameter
	if fnCall, ok := op.(*cfg.OpExprFunctionCall); ok {
		fnName, err := cfg.GetOperName(fnCall.Name)
		if err != nil {
			log.Fatalf("error in IsPropagated: %v", err)
		}
		switch fnName {
		case "filter_var":
			constArg := fnCall.Args[1].GetWriter()
			// log.Printf("filter_var Args %v %#v", len(fnCall.Args), fnCall.Args)
			switch filterOp := constArg.(type) {
			case *cfg.OpExprConstFetch:
				constName, err := cfg.GetOperName(filterOp.Name)
				if err != nil {
					log.Fatalf("error in IsSource: %v", err)
				}
				switch constName {
				case "FILTER_SANITIZE_NUMBER_INT":
					fallthrough
				case "FILTER_SANITIZE_NUMBER_FLOAT":
					return false
				}
			}
		case "htmlentities":
			if len(fnCall.Args) > 1 {
				constArg := fnCall.Args[1].GetWriter()
				switch filterOp := constArg.(type) {
				case *cfg.OpExprConstFetch:

					constName, err := cfg.GetOperName(filterOp.Name)
					if err != nil {
						log.Fatalf("error in IsSource: %v", err)
					}
					switch constName {
					case "ENT_COMPAT":
						fallthrough
					case "ENT_QUOTES":
						return false
					case "ENT_NOQUOTES":
						return false
					}
				}
			} else {
				return false
			}

		case "htmlspecialchars":
			if len(fnCall.Args) > 1 {
				constArg := fnCall.Args[1].GetWriter()
				switch filterOp := constArg.(type) {
				case *cfg.OpExprConstFetch:
					constName, err := cfg.GetOperName(filterOp.Name)
					if err != nil {
						log.Fatalf("error in IsSource: %v", err)
					}
					switch constName {
					case "ENT_COMPAT":
						fallthrough
					case "ENT_QUOTES":
						return false
					case "ENT_NOQUOTES":
						return false
					}
				}
			} else {
				return false
			}

		}

	}

	return true
}

func IsSink(op cfg.Op, taintedVar cfg.Operand) bool {

	switch opT := op.(type) {
	case *cfg.OpEcho:
		return true
	case *cfg.OpExprPrint:
		return true
	case *cfg.OpExprFunctionCall:
		funcNameStr, _ := cfg.GetOperName(opT.Name)
		switch funcNameStr {
		case "printf":
			t := opT.Args[0]
			switch s := t.(type) {
			case *cfg.OperandString:
				if strings.Contains(s.Val, "%s") {
					return true
				}
			}
			return false
		case "header":
			t := opT.Args[0].GetWriter()
			switch l := t.(type) {
			case *cfg.OpExprBinaryConcat:
				s, ok := l.Left.(*cfg.OperandString)
				if ok {
					if strings.Contains(s.Val, "Location") {
						return true
					}
				}

			}

		}

	}
	return false
}
