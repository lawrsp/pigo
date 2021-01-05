package convert

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	fmt.Println("Test started")

	os.Exit(m.Run())
}

func TestIncreaseName(t *testing.T) {
	b := [10]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Println(b[5:10])

	if increaseName("user") != "user1" {
		t.Error("increase 0 Error")
	}

	if increaseName("user9") != "user10" {
		t.Error("increate 9 Error")
	}

	if increaseName("10") != "11" {
		t.Error("increate number 10 Error")
	}
}

func TestGenAssign(t *testing.T) {

}
