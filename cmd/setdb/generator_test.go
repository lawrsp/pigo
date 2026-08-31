package setdb

import (
	"fmt"
	"os"
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

}
