package gopy

import (
	"fmt"
	"testing"
)

var pystr = `
def main(a,b):
    import time
    return a+b+add_bb(12,14)
`
var pystr1 = `
def add_bb(a):
	return a
`

func TestPy(t *testing.T) {
	p := NewPy()
	//p.Timeout=time.Second*20
	p.Func("add_bb", func(call PyFunctionCall) PyValue {
		v1, _ := call.Argument(0)
		v2, _ := call.Argument(1)
		f1 := v1.Float64()
		f2 := v2.Float64()
		return PyValue{f1 + f2}
	})
	bs, e := p.Run(pystr, "main", []any{1, 2})
	fmt.Println(string(bs), e)
}
