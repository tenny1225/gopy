package gopy

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"
)

var pyCode = `
%s

%s



if __name__=="__main__":
    import sys
    import json
    import sys
    data = input()
    sys.stdout.write(json.dumps({'status': 0, 'result': %s(*json.loads(data))}) + "\n")

`

func NewPyContext(ctx context.Context) *py {
	return &py{
		funcs:   make(map[string]func(call PyFunctionCall) PyValue),
		Timeout: 10 * 60 * time.Second,
		ctx:     ctx,
	}
}
func NewPy() *py {
	return &py{
		funcs:   make(map[string]func(call PyFunctionCall) PyValue),
		Timeout: 10 * 60 * time.Second,
		ctx:     context.Background(),
	}
}

type PyFunctionCall struct {
	arguments []PyValue
}

func (c PyFunctionCall) Argument(i int) (PyValue, error) {
	if i > len(c.arguments)-1 {
		return PyValue{}, errors.New("index error")
	}
	return c.arguments[i], nil
}

func (c PyFunctionCall) Arguments() []PyValue {

	return c.arguments
}

type pyRPC struct {
	Status int    `json:"status"`
	Result any    `json:"result"`
	Call   string `json:"call"`
	Params []any  `json:"params"`
}
type PyValue struct {
	Value any `json:"value"`
}

func (p PyValue) String() string {
	switch t := p.Value.(type) {
	case string:
		return t
	}
	return ""
}
func (p PyValue) Float64() float64 {
	switch t := p.Value.(type) {
	case float64:
		return t
	}
	return 0
}

type py struct {
	funcs   map[string]func(call PyFunctionCall) PyValue
	Timeout time.Duration
	ctx     context.Context
}

func (p *py) Func(name string, f func(call PyFunctionCall) PyValue) {
	p.funcs[name] = f
}

func (p *py) Run(code, call string, args []any) ([]byte, error) {
	bs, _ := json.Marshal(args)
	funcs := `

`
	for k := range p.funcs {
		funcs += fmt.Sprintf(`
def %s(*args):
    import sys
    import json
    import sys
    data = []
    for i, arg in enumerate(args):
        data.append(arg)
    sys.stdout.write(json.dumps({"status":1,"call":"%s","params":data})+"\n")
    command= input()
    return json.loads(command)["result"]["value"]

`, k, k)
	}

	script := fmt.Sprintf(pyCode, code, funcs, call)
	cmd := exec.Command("python3", "-c", script)
	er, e := cmd.StderrPipe()
	if e != nil {
		return nil, e
	}
	r, e := cmd.StdoutPipe()
	if e != nil {
		return nil, e
	}
	reader := bufio.NewReader(r)

	w, e := cmd.StdinPipe()
	if e != nil {
		return nil, e
	}
	writer := bufio.NewWriter(w)

	if e := cmd.Start(); e != nil {
		return nil, e
	}
	writer.Write(append(bs, '\n'))
	writer.Flush()
	message := ""
	current := time.Now()
	ctx, cancel := context.WithCancel(p.ctx)
	defer cancel()
	var err error
	go func() {
		for {
			select {
			case <-time.After(time.Second):
				if current.Add(p.Timeout).Before(time.Now()) {
					cmd.Process.Kill()
					err = errors.New("script timeout")
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		line, _, e := reader.ReadLine()
		if e != nil {
			break
		}
		//fmt.Println("line", string(line))
		current = time.Now()
		message += string(line)
		var rpc pyRPC
		if e := json.Unmarshal([]byte(message), &rpc); e != nil {
			continue
		}
		message = ""
		if rpc.Status == 0 {
			if rpc.Result == nil {
				rpc.Result = "{}"
			}
			return json.Marshal(rpc.Result)
		} else if rpc.Status == 1 {
			arguments := make([]PyValue, 0)
			for _, p := range rpc.Params {
				arguments = append(arguments, PyValue{p})
			}

			v := p.funcs[rpc.Call](PyFunctionCall{
				arguments: arguments,
			})
			bs, e := json.Marshal(pyRPC{
				Result: v,
				Status: 2,
			})
			if e != nil {
				return nil, e
			}
			writer.Write([]byte(string(bs) + "\n"))
			writer.Flush()
		}

	}
	bs, e = io.ReadAll(er)
	return nil, errors.Join(errors.New(string(bs)), e, err)
}
