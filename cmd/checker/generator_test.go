package checker

import (
	"fmt"
	"strings"

	"github.com/lawrsp/pigo/generator/parser"
)

type customPriter struct{}

func (*customPriter) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func ExampleNoEmptyProc() {
	ck := &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: parser.NewBasicType("int"),
	}
	checker := &Checker{
		c: ck,
		p: &customPriter{},
	}
	// noemtpy:emptyValue:CustomType:msgA:msgB:msgC...
	str := "noemtpy:0:CustomError:hello:is:empty"
	proc := NewNoEmptyProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc}
	checker.Next()

	ck = &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: &parser.ArrayType{},
	}
	checker = &Checker{
		c: ck,
		p: &customPriter{},
	}
	// noemtpy:emptyValue::msgA:msgB:msgC...
	str = "noemtpy:0::hello:"
	proc = NewNoEmptyProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc}
	checker.Next()

	ck = &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: parser.NewBasicType("string"),
	}
	checker = &Checker{
		c: ck,
		p: &customPriter{},
	}
	// noemtpy:emptyValue::msgA:msgB:msgC...
	str = "noemtpy:'null'"
	proc = NewNoEmptyProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc}
	checker.Next()

	// output:
	// mychecker.Assert(t.Hello != 0, "hello", CustomError, "hello", "is", "empty")
	// mychecker.Assert(len(t.Hello) != 0, "hello", "IsEmpty", "hello")
	// mychecker.Assert(t.Hello != "null", "hello", "IsEmpty")
}

func ExampleIsValidProc() {
	ck := &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: parser.NewBasicType("int"),
	}
	checker := &Checker{
		c: ck,
		p: &customPriter{},
	}

	// isvalid:valid-function:CustomType:MsgA:MsgB....
	str := "isvalid:CheckValid:CustomError:hello:is:not:valid"
	proc := NewIsValidProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc}
	checker.Next()

	ck = &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: &parser.ArrayType{},
	}
	checker = &Checker{
		c: ck,
		p: &customPriter{},
	}

	// isvalid:valid-function
	str = "isvalid:CheckValid"
	proc = NewIsValidProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc}
	checker.Next()

	ck = &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: parser.NewBasicType("string"),
	}
	checker = &Checker{
		c: ck,
		p: &customPriter{},
	}
	// isvalid:valid-function::MsgA:MsgB....
	str = "isvalid:CheckValid::not valid"
	proc = NewIsValidProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc}
	checker.Next()

	// output:
	// mychecker.Assert(CheckValid(t.Hello), "hello", CustomError, "hello", "is", "not", "valid")
	// mychecker.Assert(CheckValid(t.Hello), "hello", "Invalid")
	// mychecker.Assert(CheckValid(t.Hello), "hello", "Invalid", "not valid")
}

func ExampleCallProc() {
	ck := &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: parser.NewBasicType("int"),
	}
	checker := &Checker{
		c: ck,
		p: &customPriter{},
	}

	// call:some-function:error/bool:CustomType:MsgA:MsgB:....
	str := "call:CheckIt:error:CustomError:hello:is:not:valid"
	proc := NewCallProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc}
	checker.Next()

	ck = &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: &parser.ArrayType{},
	}
	checker = &Checker{
		c: ck,
		p: &customPriter{},
	}

	// call:some-function
	str = "call:CheckIt:false"
	proc = NewCallProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc}
	checker.Next()

	ck = &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: parser.NewBasicType("string"),
	}
	checker = &Checker{
		c: ck,
		p: &customPriter{},
	}

	// call:some-function:bool::MsgA:MsgB:....
	str = "call:CheckIt:bool::not valid"
	proc = NewCallProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc}
	checker.Next()

	// output:
	// mychecker.AssertError(t.Hello.CheckIt(), "hello", CustomError, "hello", "is", "not", "valid")
	// mychecker.Assert(!t.Hello.CheckIt(), "hello", "Invalid")
	// mychecker.Assert(t.Hello.CheckIt(), "hello", "Invalid", "not valid")
}

func ExampleCompareProc() {
	ck := &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: parser.NewBasicType("int"),
	}
	checker := &Checker{
		c: ck,
		p: &customPriter{},
	}

	// compare:!=><:value:CustomType:MsgA:MsgB...
	str := "compare:!=:'null':CustomError:hello:is:null"
	proc := NewCompareProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc}
	checker.Next()

	ck = &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: &parser.ArrayType{},
	}
	checker = &Checker{
		c: ck,
		p: &customPriter{},
	}

	// compare:!=><:value
	str = "compare:>:0"
	proc = NewCompareProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc}
	checker.Next()

	ck = &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: parser.NewBasicType("string"),
	}
	checker = &Checker{
		c: ck,
		p: &customPriter{},
	}

	// compare:!=><:value::MsgA:MsgB...
	str = "compare:<:100::too big"
	proc = NewCompareProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc}
	checker.Next()

	// output:
	// mychecker.Assert(t.Hello != "null", "hello", CustomError, "hello", "is", "null")
	// mychecker.Assert(t.Hello > 0, "hello", "Invalid")
	// mychecker.Assert(t.Hello < 100, "hello", "Invalid", "too big")
}

func ExampleMergeProc() {
	ck := &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: parser.NewBasicType("int"),
	}
	checker := &Checker{
		c: ck,
		p: &customPriter{},
	}

	// merge:other-valid-function
	str := "merge:OtherCheck"
	proc := NewMergeProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc}
	checker.Next()

	// output:
	// if err := t.Hello.OtherCheck(); err != nil {
	// mychecker.Merge(err)
	// }
}

func ExampleConvertProc() {
	ck := &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: parser.NewBasicType("int"),
	}
	checker := &Checker{
		c: ck,
		p: &customPriter{},
	}

	// convert:convert-to-type:custom-new-name
	str := "convert:int64:big"
	proc := NewConvertProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc}
	checker.Next()

	// output:
	// big := int64(t.Hello)
}

func ExampleStarProc() {
	ck := &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: parser.NewBasicType("int"),
	}
	checker := &Checker{
		c: ck,
		p: &customPriter{},
	}

	// stars:123
	str := "stars:1"
	proc := NewStarProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc}
	checker.Next()

	ck = &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: parser.NewBasicType("int"),
	}
	checker = &Checker{
		c: ck,
		p: &customPriter{},
	}

	str = "stars:2"
	proc = NewStarProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc, NewDefaultValueProc(strings.Split("default:20", ":"))}
	checker.Next()

	// output:
	// if t.Hello != nil {
	// }
	// if t.Hello != nil && *t.Hello != nil {
	// if **t.Hello == 0 {
	// **t.Hello = 20
	// }
	// }
}

// default:somevalue
// default:'stringvalue'

func ExampleDefaultValueProc() {
	ck := &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: parser.NewBasicType("int"),
	}
	checker := &Checker{
		c: ck,
		p: &customPriter{},
	}

	// default:value
	str := "default:100"
	proc := NewDefaultValueProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc}
	checker.Next()

	ck = &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: parser.NewBasicType("string"),
	}
	checker = &Checker{
		c: ck,
		p: &customPriter{},
	}

	str = "default:'world'"
	proc = NewDefaultValueProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc}
	checker.Next()

	// output:
	// if t.Hello == 0 {
	// t.Hello = 100
	// }
	// if t.Hello == "" {
	// t.Hello = "world"
	// }
}

func ExampleArrayProc() {
	typ := parser.ParseTypeString("[]int")
	ck := &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: typ,
	}
	checker := &Checker{
		c: ck,
		p: &customPriter{},
	}

	// arrays:123
	str := "arrays:1"
	proc := NewArrayProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc, NewDefaultValueProc(strings.Split("default:100", ":"))}
	checker.Next()

	typ = parser.ParseTypeString("[][]string")
	ck = &CheckerInfo{
		chk:        "mychecker",
		name:       "\"hello\"",
		expr:       "t.Hello",
		targetType: typ,
	}
	checker = &Checker{
		c: ck,
		p: &customPriter{},
	}

	str = "arrays:2"
	proc = NewArrayProc(strings.Split(str, ":"))
	checker.procs = []CheckerProc{proc, NewDefaultValueProc(strings.Split("default:'null'", ":"))}

	checker.Next()

	// output:
	// for i, it := range t.Hello {
	// if it == 0 {
	// t.Hello[i] = 100
	// }
	// }
	// for i, it := range t.Hello {
	// for ii, itt := range it {
	// if itt == "" {
	// it[ii] = "null"
	// }
	// }
	// }
}
