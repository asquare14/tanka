package jsonnet

import (
	"io/ioutil"

	jsonnet "github.com/google/go-jsonnet"
	"github.com/pkg/errors"

	"github.com/grafana/tanka/pkg/jsonnet/jpath"
	"github.com/grafana/tanka/pkg/jsonnet/native"
)

// Modifier allows to set optional parameters on the Jsonnet VM.
// See jsonnet.With* for this.
type Modifier func(vm *jsonnet.VM) error

// EvaluateFile opens the file, reads it into memory and evaluates it afterwards (`Evaluate()`)
func EvaluateFile(jsonnetFile string, mods ...Modifier) (string, error) {
	entrypoint, err := jpath.GetEntrypoint(jsonnetFile)
	if err != nil {
		return "", errors.Wrap(err, "resolving entrypoint")
	}

	bytes, err := ioutil.ReadFile(entrypoint)
	if err != nil {
		return "", err
	}

	resolved_jpath, _, _, err := jpath.Resolve(entrypoint)
	if err != nil {
		return "", errors.Wrap(err, "resolving jpath")
	}
	return Evaluate(entrypoint, string(bytes), resolved_jpath, mods...)
}

// Evaluate renders the given jsonnet into a string
func Evaluate(filename, sonnet string, jpath []string, mods ...Modifier) (string, error) {
	vm := jsonnet.MakeVM()
	vm.Importer(NewExtendedImporter(jpath))

	for _, mod := range mods {
		if err := mod(vm); err != nil {
			return "", err
		}
	}

	for _, nf := range native.Funcs() {
		vm.NativeFunction(nf)
	}

	return vm.EvaluateSnippet(filename, sonnet)
}

// WithExtCode allows to make the supplied snippet available to Jsonnet as an
// ext var
func WithExtCode(key, code string) Modifier {
	return func(vm *jsonnet.VM) error {
		vm.ExtCode(key, code)
		return nil
	}
}

// WithTLA allows to set the given code as a top level argument
func WithTLA(key, code string) Modifier {
	return func(vm *jsonnet.VM) error {
		vm.TLACode(key, code)
		return nil
	}
}
