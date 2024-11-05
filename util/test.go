package util

import (
	"fmt"
	"os"
	"runtime"
	"slices"
	"testing"
)

func Fail(t *testing.T, depth int) {
	if _, file, line, ok := runtime.Caller(depth + 1); ok {
		fmt.Printf("[FAILED] %s:%d\n", file, line)
	}
	t.FailNow()
}

func TestAssert(t *testing.T, x bool) {
	if !x {
		Fail(t, 1)
	}
}

func TestEq[T comparable](t *testing.T, a, b T) {
	if a != b {
		fmt.Fprintf(os.Stdout, "`%v` != `%v`\n", a, b)
		Fail(t, 1)
	}
}

func TestEqSlice[S ~[]E, E comparable](t *testing.T, a, b S) {
	if a == nil || b == nil || !slices.Equal(a, b) {
		fmt.Fprintf(os.Stdout, "%v != %v\n", a, b)
		Fail(t, 1)
	}
}
