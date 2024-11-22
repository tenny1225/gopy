package gopy

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/google/uuid"
)

var pyCode = `
import sys
import json

def process(code,call):
    exec(code)
    return eval(call)


def parseParams(fn):
	with open(fn, 'r', encoding='utf-8') as file:
          return json.load(file)

# Using the special variable 
# __name__
if __name__=="__main__":
    args = sys.argv[1:]
    if(len(args)==1):
        eval(args[0])
    else:
        print(json.dumps(process(args[0],args[1])))
`
var pyCommand = "python3"
var tmpPyName = "__tmp__.py"

func init() {
	os.WriteFile(tmpPyName, []byte(pyCode), 0666)
}
func RegisterCommand(command string) {
	pyCommand = command
}

func RunPyDef(code, fun string, args []any) ([]byte, error) {
	bs, _ := json.Marshal(args)
	id := uuid.NewString() + ".json"
	os.WriteFile(id, bs, 0666)
	defer os.Remove(id)
	param := fmt.Sprintf(`%s(*parseParams('%s'))`, fun, id)
	cmd := exec.Command(pyCommand, tmpPyName, code, param)
	res, e := cmd.Output()
	if e != nil {
		switch t := e.(type) {
		case *exec.ExitError:
			return nil, errors.New(string(t.Stderr))
		}
		return nil, e
	}
	return res, nil
}
func RunPy(code string) ([]byte, error) {
	cmd := exec.Command(pyCommand, tmpPyName, code)
	res, e := cmd.Output()
	if e != nil {
		switch t := e.(type) {
		case *exec.ExitError:
			return nil, errors.New(string(t.Stderr))
		}
		return nil, e
	}
	return res, nil
}

