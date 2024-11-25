# gopy

> Use golang to execute a python script and return the result



``` go get github.com/tenny1225/gopy```


```go
package main
import "github.com/tenny1225/gopy"
import "fmt"
func main() {
	res, _ := gopy.RunPyDef(`
def main(a,b):
	return {"name":b["name"]}
		`, "main", []any{1, map[string]any{"name": "hello word!"}})
	fmt.Println(string(res))

	res, _ = gopy.RunPy(`print('hello python!')`)
	fmt.Println(string(res))
}

```