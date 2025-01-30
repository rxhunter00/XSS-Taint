package pathgenerator

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/rxhunter00/XSS-Taint/pkg/cfg"
)

type PathGenerator struct {
	detectedPaths [][]cfg.Op
	currPath      []cfg.Op
	visited       map[cfg.Op]map[cfg.Operand]struct{}
}

func NewPathGenerator() *PathGenerator {
	return &PathGenerator{
		detectedPaths: make([][]cfg.Op, 0),
	}
}

func GeneratePath(scripts map[string]*cfg.Script) [][]cfg.Op {
	pg := NewPathGenerator()

	for _, script := range scripts {

		pg.visited = make(map[cfg.Op]map[cfg.Operand]struct{})
		pg.traverseScript(script)

	}
	return pg.detectedPaths
}

func (pg *PathGenerator) traverseScript(s *cfg.Script) {
	// Traverse Main Function
	pg.traverseFunc(*s.Main)

	for _, fn := range s.FuncsMap {
		// Traverse Other Function
		pg.traverseFunc(*fn)
	}
}

func (pg *PathGenerator) traverseFunc(fn cfg.Func) {

	for _, sourceOp := range fn.Sources {

		pg.currPath = []cfg.Op{sourceOp}
		// Get the result of tainted op
		sourceVar, err := pg.getPropagatedVar(sourceOp)
		if err != nil {
			continue
		}
		for _, sourceUser := range sourceVar.GetUsers() {
			// For each p that use this source
			err := pg.traceTaintFlow(sourceUser, sourceVar)
			if err != nil {
				log.Fatalf("traverseFunc:File '%s':  %v", fn.Filepath, err)
			}
		}
	}
}

func (pg *PathGenerator) traceTaintFlow(taintedUser cfg.Op, taintedVar cfg.Operand) error {

	if pg.isSink(taintedUser, taintedVar) {

		newPath := make([]cfg.Op, len(pg.currPath))
		copy(newPath, pg.currPath)
		newPath = append(newPath, taintedUser)
		pg.detectedPaths = append(pg.detectedPaths, newPath)

		return nil
	} else if pg.isSanitized(taintedUser, taintedVar) {
		return nil
	} else if pg.isAlreadyVisited(taintedUser, taintedVar) {
		return nil
	}
	pg.markVisited(taintedUser, taintedVar)

	// Get Next Operand that hold taint Value
	newTaint, err := pg.getPropagatedVar(taintedUser)
	if err != nil {
		return nil
	}

	// Get Op that use next tainted Operand
	for _, newTaintUser := range newTaint.GetUsers() {

		temp := pg.currPath
		newPath := make([]cfg.Op, len(pg.currPath))
		copy(newPath, pg.currPath)
		pg.currPath = append(newPath, newTaintUser)

		// Trace tainted for the result
		err := pg.traceTaintFlow(newTaintUser, newTaint)
		if err != nil {
			return err
		}
		pg.currPath = temp

	}

	return nil
}

func (pg *PathGenerator) isAlreadyVisited(op cfg.Op, taintedVar cfg.Operand) bool {
	_, ok := pg.visited[op]
	if ok {
		_, v := pg.visited[op][taintedVar]
		if v {
			return true
		}
	}
	return false

}
func (pg *PathGenerator) markVisited(op cfg.Op, taintedVar cfg.Operand) {

	if _, ok := pg.visited[op]; !ok {
		pg.visited[op] = make(map[cfg.Operand]struct{})
	}
	pg.visited[op][taintedVar] = struct{}{}

}

// Check if its sanitizer
func (pg *PathGenerator) isSanitized(op cfg.Op, taintedVar cfg.Operand) bool {

	switch opT := op.(type) {
	case *cfg.OpExprFunctionCall:
		funcNameStr, _ := cfg.GetOperandName(opT.Name)
		switch funcNameStr {
		// URL Context
		case "rawurlencode":
			return true
		case "urlencode":
			return true
		//Java Script
		case "json_encode":
			return true
		//Convertible()
		case "intval":
			return true
		case "floatval":
			return true
		case "doubleval":
			return true
		case "boolval":
			return true
		// filter_var with cosntants
		case "filter_var":
			constArg := opT.Args[1].GetWriter()
			switch filterOp := constArg.(type) {
			case *cfg.OpExprConstFetch:
				constName, err := cfg.GetOperandName(filterOp.Name)
				if err != nil {
					log.Fatalf("error in isSource: %v", err)
				}
				switch constName {
				case "FILTER_SANITIZE_NUMBER_INT":
					return true
				case "FILTER_SANITIZE_NUMBER_FLOAT":
					return true
				}
			}
		case "htmlentities":
			if len(opT.Args) > 1 {
				constArg := opT.Args[1].GetWriter()
				switch filterOp := constArg.(type) {
				case *cfg.OpExprConstFetch:

					constName, err := cfg.GetOperandName(filterOp.Name)
					if err != nil {
						log.Fatalf("error in IsSource: %v", err)
					}
					switch constName {
					case "ENT_COMPAT":
						return true
					case "ENT_QUOTES":
						return true
					case "ENT_NOQUOTES":
						return true
					}
				}
			} else {
				return true
			}

		case "htmlspecialchars":
			if len(opT.Args) > 1 {
				constArg := opT.Args[1].GetWriter()
				switch filterOp := constArg.(type) {
				case *cfg.OpExprConstFetch:
					constName, err := cfg.GetOperandName(filterOp.Name)
					if err != nil {
						log.Fatalf("error in IsSource: %v", err)
					}
					switch constName {
					case "ENT_COMPAT":
						return true
					case "ENT_QUOTES":
						return true
					case "ENT_NOQUOTES":
						return true
					}
				}
			} else {
				return true
			}

		}

	case *cfg.OpExprCastBool, *cfg.OpExprCastDouble, *cfg.OpExprCastInt:
		return true
	case *cfg.OpExprAssertion:
		switch assert := opT.Assertion.(type) {
		case *cfg.TypeAssertion:
			if typeVal, ok := assert.AssertionOperand.(*cfg.OperandString); ok {
				switch typeVal.Val {
				case "int", "float", "bool", "null":
					return true
				}
			}
		}

	case *cfg.OpExprArrayDimFetch:
		if opT.Dim == taintedVar {
			return true
		}

	}
	return false
}

// Check if its sink
func (pg *PathGenerator) isSink(op cfg.Op, taintedVar cfg.Operand) bool {

	switch opT := op.(type) {
	case *cfg.OpEcho:
		return true
	case *cfg.OpExprPrint:
		return true
	case *cfg.OpExprFunctionCall:
		funcNameStr, _ := cfg.GetOperandName(opT.Name)
		switch funcNameStr {
		case "printf":
			t := opT.Args[0]
			switch s := t.(type) {
			// Lazy Checking if there is any, should check based on args order and % position
			case *cfg.OperandString:
				if strings.Contains(s.Val, "%s") {
					return true
				}
			}
			return false
		case "header":
			if opT.Args[0] == taintedVar {
				return true
			}
			t := opT.Args[0].GetWriter()
			switch l := t.(type) {
			case *cfg.OpExprBinaryConcat:
				s, ok := l.Left.(*cfg.OperandString)
				if ok {
					if strings.Contains(s.Val, "Location") {
						return true
					}
				} else if l.Right == taintedVar {
					return true
				}

			}

		}

	}
	return false
}

func (pg *PathGenerator) getPropagatedVar(op cfg.Op) (cfg.Operand, error) {
	if assignmentOp, ok := op.(*cfg.OpExprAssign); ok {
		if assignmentOp.Var != nil {
			return assignmentOp.Var, nil
		}
		return assignmentOp.Result, nil
	} else if result, ok := op.GetOpVars()["Result"]; ok {
		if result != nil {
			return result, nil
		}
	}
	return nil, fmt.Errorf("getPropagatedVar:Error Unsupported type Operation '%v'", reflect.TypeOf(op))
}
