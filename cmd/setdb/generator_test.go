package setdb

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

type stdoutPrinter struct{}

func (stdoutPrinter) Printf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format, args...)
}

func TestBuildOnePtrClause(t *testing.T) {

	c := Condition{
		Config: &Config{},
		name:   "HelloLike",
	}

	c.InitFromTag("")

	fmt.Println(c)
	g := NewGenerator()

	g.buildOnePtrClause(stdoutPrinter{}, &c)

	x := " 1 2 3     "
	for _, xi := range strings.Split(x, " ") {
		fmt.Printf("x%sx", xi)
	}

}
