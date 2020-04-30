package genrpc

import (
	"fmt"
	"os"
	"regexp"
	"testing"
)

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	fmt.Println("Test started")

	os.Exit(m.Run())
}

func TestRcallScan(t *testing.T) {
	// var receiver, call, params string

	var rcallRegexp = regexp.MustCompile("^(\\(([a-zA-Z0-9\\.\\*\\[\\]]*)\\))?([a-zA-Z0-9\\.]*)(\\(([a-zA-Z0-9,\\[\\]\\.\\*]*)\\))*$")

	src := "(adb.Z)M.Test(o.1i,b,d,f)"
	result := rcallRegexp.ReplaceAllString(src, "$2,$3,$5")
	fmt.Println(result)
	fmt.Println(NewRCallDsec(src))

	src = "M.Test(o.1i,b,d,f)"
	result = rcallRegexp.ReplaceAllString(src, "$2,$3,$5")
	fmt.Println(result)
	fmt.Println(NewRCallDsec(src))

	src = "M.Test"
	result = rcallRegexp.ReplaceAllString(src, "$2,$3,$5")
	fmt.Println(result)
	fmt.Println(NewRCallDsec(src))

	src = "(*user.ListParam)FilterByPath([]string)"
	result = rcallRegexp.ReplaceAllString(src, "$2,$3,$5")
	fmt.Println(result)
	fmt.Println(NewRCallDsec(src))
}
