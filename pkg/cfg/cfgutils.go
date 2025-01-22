package cfg

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

func IsBuiltInType(name string) bool {
	builtInTypes := map[string]struct{}{
		"self":     {},
		"parent":   {},
		"static":   {},
		"int":      {},
		"integer":  {},
		"long":     {},
		"float":    {},
		"double":   {},
		"real":     {},
		"array":    {},
		"object":   {},
		"bool":     {},
		"boolean":  {},
		"null":     {},
		"void":     {},
		"false":    {},
		"true":     {},
		"string":   {},
		"mixed":    {},
		"resource": {},
		"callable": {},
	}

	_, exists := builtInTypes[name]
	return exists
}

func GetTypeAssertFunc(funcName string) (string, bool) {
	funcName = strings.ToLower(funcName)
	switch funcName {
	case "is_array":
		return "array", true
	case "is_bool":
		return "bool", true
	case "is_callable":
		return "callable", true
	case "is_double":
		return "float", true
	case "is_float":
		return "float", true
	case "is_int":
		return "int", true
	case "is_integer":
		return "int", true
	case "is_long":
		return "int", true
	case "is_null":
		return "null", true
	case "is_numeric":
		return "numeric", true
	case "is_object":
		return "object", true
	case "is_real":
		return "float", true
	case "is_string":
		return "string", true
	case "is_resource":
		return "resource", true
	}
	return "", false
}

// Add op to list of each operand usage
func AddReadRefs(op Op, opers ...Operand) []Operand {
	result := make([]Operand, 0)
	for _, oper := range opers {
		if oper != nil {
			result = append(result, AddReadRef(op, oper))
		}
	}
	return result
}

// Add Op to list Op that that use/read this Operand
func AddReadRef(op Op, oper Operand) Operand {
	if oper == nil {
		return nil
	}

	oper.AddUser(op)
	return oper
}

// Add op to each list of operand that defined by this op
func AddWriteRefs(op Op, opers ...Operand) []Operand {
	result := make([]Operand, 0)
	for _, oper := range opers {
		if oper != nil {
			result = append(result, AddWriteRef(op, oper))
		}
	}
	return result
}

// Add Op to the Op that define/write to this Operand
func AddWriteRef(op Op, oper Operand) Operand {
	if oper == nil {
		return nil
	}

	oper.AddWriter(op)
	return oper
}

func GetSubBlocks(op Op) map[string]*Block {
	m := make(map[string]*Block)
	switch o := op.(type) {
	case *OpExprParam:
		if o.DefaultBlock != nil {
			m["DefaultBlock"] = o.DefaultBlock
		}
	case *OpStmtInterface:
		if o.Stmts != nil {
			m["Stmts"] = o.Stmts
		}
	case *OpStmtClass:
		if o.Stmts != nil {
			m["Stmts"] = o.Stmts
		} else {
			log.Fatal("nil stmts")
		}
	case *OpStmtTrait:
		if o.Stmts != nil {
			m["Stmts"] = o.Stmts
		}
	case *OpStmtJump:
		if o.Target != nil {
			m["Target"] = o.Target
		}
	case *OpStmtJumpIf:
		if o.If != nil {
			m["If"] = o.If
		}
		if o.Else != nil {
			m["Else"] = o.Else
		}
	case *OpStmtProperty:
		if o.DefaultBlock != nil {
			m["DefaultBlock"] = o.DefaultBlock
		}
	case *OpStmtSwitch:
		for i, subBlock := range o.Targets {
			s := fmt.Sprintf("Target[%d]", i)
			m[s] = subBlock
		}
	case *OpConst:
		if o.ValueBlock != nil {
			m["ValueBlock"] = o.ValueBlock
		}
	case *OpStaticVar:
		if o.DefaultBlock != nil {
			m["DefaultBlock"] = o.DefaultBlock
		}
	}

	return m
}

func ChangeSubBlock(op Op, subBlockName string, newBlock *Block) {
	switch o := op.(type) {
	case *OpExprParam:
		if subBlockName == "DefaultBlock" {
			o.DefaultBlock = newBlock
		} else {
			log.Fatalf("Error: Unknown OpExprParam subblock '%s'", subBlockName)
		}
	case *OpStmtInterface:
		if subBlockName == "Stmts" {
			o.Stmts = newBlock
		} else {
			log.Fatalf("Error: Unknown OpStmtInterface subblock '%s'", subBlockName)
		}
	case *OpStmtClass:
		if subBlockName == "Stmts" {
			o.Stmts = newBlock
		} else {
			log.Fatalf("Error: Unknown OpStmtClass subblock '%s'", subBlockName)
		}
	case *OpStmtTrait:
		if subBlockName == "Stmts" {
			o.Stmts = newBlock
		} else {
			log.Fatalf("Error: Unknown OpStmtTrait subblock '%s'", subBlockName)
		}
	case *OpStmtJump:
		if subBlockName == "Target" {
			o.Target = newBlock
		} else {
			log.Fatalf("Error: Unknown OpStmtJump subblock '%s'", subBlockName)
		}
	case *OpStmtJumpIf:
		if subBlockName == "If" {
			o.If = newBlock
		} else if subBlockName == "Else" {
			o.Else = newBlock
		} else {
			log.Fatalf("Error: Unknown OpStmtJumpIf subblock '%s'", subBlockName)
		}
	case *OpStmtProperty:
		if subBlockName == "DefaultBlock" {
			o.DefaultBlock = newBlock
		} else {
			log.Fatalf("Error: Unknown OpStmtProperty subblock '%s'", subBlockName)
		}
	case *OpStmtSwitch:
		startIdx := strings.Index(subBlockName, "[")
		endIdx := strings.Index(subBlockName, "]")
		if startIdx == -1 || endIdx == -1 {
			log.Fatalf("Error: Unknown OpStmtSwitch subblock '%s'", subBlockName)
		}
		idx, err := strconv.Atoi(subBlockName[startIdx+1 : endIdx])
		if err != nil || idx >= len(o.Targets) {
			log.Fatalf("Error: Unknown OpStmtSwitch subblock '%s'", subBlockName)
		}
		o.Targets[idx] = newBlock
	case *OpConst:
		if subBlockName == "ValueBlock" {
			o.ValueBlock = newBlock
		} else {
			log.Fatalf("Error: Unknown OpConst subblock '%s'", subBlockName)
		}
	case *OpStaticVar:
		if subBlockName == "DefaultBlock" {
			o.DefaultBlock = newBlock
		} else {
			log.Fatalf("Error: Unknown OpStaticVar subblock '%s'", subBlockName)
		}
	}
}

func IsWriteVar(op Op, varName string) bool {
	if varName == "Result" {
		return true
	} else if varName == "Var" {
		switch op.(type) {
		case *OpStaticVar:
			return true
		case *OpExprAssign:
			return true
		case *OpExprAssignRef:
			return true
		}
	}
	return false
}
