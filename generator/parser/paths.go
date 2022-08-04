package parser

import (
	"fmt"
	"log"
)

type TPath struct {
	Source  Type
	Target  Type
	D       Direction
	Arg     interface{}
	parent  *TPath
	sibling *TPath
}

func NewTPath(src Type, dst Type) *TPath {
	return &TPath{Source: src, Target: dst}
}
func (tp *TPath) Copy() *TPath {
	n := &TPath{}
	*n = *tp
	return n
}
func (tp *TPath) WithFunction(funcType Type) *TPath {
	n := &TPath{}
	*n = *tp
	n.D = D_CallFunction
	n.Arg = funcType
	return n
}

func (tp *TPath) WithTypeConversion(t Type) *TPath {
	n := &TPath{}
	*n = *tp
	n.D = D_TypeConversion
	n.Arg = t
	return n
}

func (tp *TPath) WithDirection(d Direction) *TPath {
	n := &TPath{}
	*n = *tp
	n.D = d
	return n
}

func (tp *TPath) WithArg(arg interface{}) *TPath {
	n := tp.Copy()
	n.Arg = arg
	return n
}

func (tp *TPath) withParent(parent *TPath) *TPath {
	n := &TPath{}
	*n = *tp
	n.parent = parent
	return n
}

func (tp *TPath) withSibling(sibling *TPath) *TPath {
	n := &TPath{}
	*n = *tp
	n.sibling = sibling
	return n
}

func (tp *TPath) ConflictWith(ot *TPath) bool {
	if tp.Source.EqualTo(ot.Source) && tp.Target.EqualTo(ot.Target) {
		return true
	}
	return false
}

func (tp *TPath) String() string {
	return fmt.Sprintf("[%s,%s,%d,%v]", tp.Source, tp.Target, tp.D, tp.Arg)
}

func debugPrintPaths(paths []*TPath, format string, args ...interface{}) {
	// return
	log.Printf(format, args...)
	for _, tp := range paths {
		log.Printf(".%s", tp)
	}
	if len(paths) > 0 {
		log.Printf("\n")
	}
}

type Direction int

const (
	d_skip Direction = iota //named
	D_Self
	D_SkipPointer
	D_SkipBracket
	D_SpreadFields
	D_CallFunction //call function
	D_AddPointer
	D_AddBracket
	D_TypeConversion
)

func useExists(a Type, exists []*TPath) (paths []*TPath) {
	if exists == nil {
		return
	}
	for _, tp := range exists {
		if TypeEqual(a, tp.Source) {
			paths = append(paths, tp.Copy())
		}
	}
	return
}

func downDirections(t Type) (paths []*TPath) {
	trip := t
	// var file *File
	if x, ok := trip.(*filedType); ok {
		// file = x.file
		trip = x.Type
	}
	switch m := trip.(type) {
	case *namedType:
		// log.Printf("namedType: %s :%s (nil?%v)", t, m.Type, m.Type == nil)
		// namedType: type A b => b(x)
		if m.Type != nil {
			switch m.Type.(type) {
			case *namedType:
				paths = append(paths, NewTPath(trip, m.Type).WithTypeConversion(m.Type))
			case *BasicType:
				paths = append(paths, NewTPath(trip, m.Type).WithTypeConversion(m.Type))
			default:
				// log.Printf("====m.Type is %v", m.Type)
				paths = append(paths, NewTPath(trip, m.Type).WithDirection(d_skip))
			}

		}
	case *PointerType:
		dst := TypeSkipPointer(m, 1)
		paths = append(paths, NewTPath(trip, dst).WithDirection(D_SkipPointer))
	case *ArrayType:
		dst := TypeSkipBracket(m, 1)
		// log.Printf("%s ===>>> %s", trip, dst)
		paths = append(paths, NewTPath(trip, dst).WithDirection(D_SkipBracket))
	case *StructType:
		for _, field := range m.Fields {
			ResolveUnknownField(field)
			tp := NewTPath(trip, field.Type).WithDirection(D_SpreadFields).WithArg(field)
			paths = append(paths, tp)
		}
	default:
		// log.Printf("no type: %s", t)
	}

	// debugPrintPaths(paths, "%s down:\n", t)
	return
}

func upDirections(t Type) (paths []*TPath) {
	//up
	m := TypeWithSlice(t)
	paths = append(paths, NewTPath(t, m).WithDirection(D_AddBracket))
	m = TypeWithPointer(t)
	paths = append(paths, NewTPath(t, m).WithDirection(D_AddPointer))
	return
}

func typeDirections(t Type, lastD Direction, knowns []*TPath) []*TPath {
	//exists
	paths := useExists(t, knowns)

	//down
	downs := downDirections(t)
	if len(downs) > 0 {
		paths = append(paths, downs...)
	}
	//up
	ups := upDirections(t)
	if len(ups) > 0 {
		paths = append(paths, ups...)
	}
	return paths
}

func checkExists(nowPaths []*TPath, tp *TPath) bool {
	for _, old := range nowPaths {
		if tp.Source.EqualTo(old.Source) && tp.Target.EqualTo(old.Target) {
			return true
		}
	}
	return false
}

func allDirections(nowPaths []*TPath, starts []*TPath, knowns []*TPath) (results []*TPath) {
	for _, tp := range starts {
		nextPaths := typeDirections(tp.Target, tp.D, knowns)
		var sibling *TPath = nil
		for _, np := range nextPaths {
			if checkExists(nowPaths, np) {
				continue
			}

			nextP := np.withParent(tp)
			if sibling != nil {
				nextP = nextP.withSibling(sibling)
			}
			results = append(results, nextP)
			sibling = nextP
		}
	}

	// debugPrintPaths(paths, ">>>>>> \nfrom: ")
	// debugPrintPaths(newPaths, "add: ")
	// debugPrintPaths(results, "get: ")

	return
}

func checkReach(paths []*TPath, b Type) *TPath {
	// debugPrintPaths(paths, "check reached:")
	for _, tp := range paths {
		if TypeEqual(tp.Target, b) {
			return tp
		}

		// if tp.Target.Underlying().EqualTo(b.Underlying()) {
		//	log.Printf("=========underlying equal!!!!!")
		//	// panic("hello")
		//	return tp
		// }

	}

	return nil
}

func collectTPath(endTp *TPath) []*TPath {
	result := []*TPath{}
	for x := endTp; x != nil; x = x.parent {
		switch x.D {
		case d_skip:
		default:
			result = append(result, x)
		}
	}

	ln := len(result)
	sorted := make([]*TPath, ln)
	for i := 0; i < ln; i++ {
		sorted[i] = result[ln-i-1]
	}

	return sorted
}

//optimize
func optimizTPaths(input []*TPath) []*TPath {
	output := []*TPath{}
	index := -1
	for _, tp := range input {
		if tp.D == D_SpreadFields && index >= 0 && output[index].D == D_SkipPointer {
			output[index] = tp
		} else {
			output = append(output, tp)
			index += 1
		}
	}
	return output
}

//a to b
func TypeToType(a Type, b Type, knowns []*TPath) ([]*TPath, bool) {
	allPaths := []*TPath{}
	checkPaths := []*TPath{NewTPath(a, a).WithDirection(D_Self)}

	var endTp *TPath = nil

	var deep int = 0
	for ; endTp == nil && deep < 7; deep += 1 {
		if endTp = checkReach(checkPaths, b); endTp != nil {
			break
		}
		allPaths = append(allPaths, checkPaths...)
		checkPaths = allDirections(allPaths, checkPaths, knowns)
	}

	// log.Printf("endTp: %v, deep :%d", endTp, deep)
	if endTp == nil {
		return nil, false
	}

	results := collectTPath(endTp)
	// results = optimizTPaths(results)
	return results, true
}

func InspectUnderlyingStruct(t Type, inspect func(*Field) bool) bool {
	t = t.Underlying()

	for {
		if x, ok := t.(*PointerType); ok {
			t = x.Base.Underlying()
		} else {
			break
		}
	}

	st, ok := t.(*StructType)
	if !ok {
		return false
	}

	for _, fd := range st.Fields {
		//ignore
		ResolveUnknownField(fd)
		stepInto := inspect(fd)
		if stepInto {
			InspectUnderlyingStruct(fd.Type, inspect)
		}
	}

	return true
}
