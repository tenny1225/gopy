package gopy

import (
	"fmt"
	"testing"
)

func TestPy(t *testing.T) {
	res, _ := RunPyDef(`
def main(a,b):
	return {"name":b["name"]}
		`, "main", []any{1, map[string]any{"name": "hello word!"}})
	fmt.Println(string(res))

	res, _ = RunPy(`print('hello python!')`)
	fmt.Println(string(res))
}

