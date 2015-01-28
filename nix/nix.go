package nix

import (
	"bytes"
	"encoding/json"
	"github.com/offlinehacker/nixapi/utils"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"time"
)

// Interface for nix expressions
type ExpressionInterface interface {
	GetDerivations() []Derivation
}

// Interface for derivations
type DerivationInterface interface {
}

// Nix expression
type Expression struct {
	Path string
}

type ExpressionError struct {
	error
	stderr *bytes.Buffer
}

var RunCommand = utils.RunCommand

// Gets all derivations defined in expression
func (n Expression) GetDerivations(stop <-chan time.Time) (<-chan []Derivation, <-chan error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	_, filename, _, _ := runtime.Caller(1)
	errCh := make(chan error, 1)
	resultCh := make(chan []Derivation, 1)

	command := []string{"nix-instantiate", path.Join(filename, "packages.nix"), "--eval", "--strict"}
	cmd := RunCommand(command, stop, stdout, stderr)

	go func() {
		if err := <-cmd; err != nil {
			errCh <- ExpressionError{err, stderr}
			return
		}
		derivations := make([]Derivation, 0)
		if err := json.Unmarshal(stdout.Bytes(), &derivations); err != nil {
			errCh <- ExpressionError{err, stderr}
			return
		}

		for _, el := range derivations {
			el.Expression = n
		}

		resultCh <- derivations
	}()

	return resultCh, errCh
}

// Constucts new expression object
func NewExpression(path string) *Expression {
	p := new(Expression)
	p.Path = path

	return p
}

// Nix derivation
type Derivation struct {
	Expression Expression `-`
	Name       string     `json:'name'`
	AttrPath   string     `json:'attrPath'`
	Out        string     `json:'out'`
}

// Builds a derivation and returns
func (d Derivation) build(
	stop <-chan time.Time, stdout io.Writer, stderr io.Writer) (<-chan string, <-chan error) {
	resultCh := make(chan string, 1)
	errCh := make(chan error, 1)

	tmpPath, err := ioutil.TempDir("", "nix-result")
	if err != nil {
		errCh <- err
	}

	command := []string{"nix-build", d.Expression.Path}
	if d.AttrPath != "" && d.AttrPath != "-" {
		command = append(command, "-A", d.AttrPath)
	}
	command = append(command, "-o", path.Join(tmpPath, "result"))

	cmd := RunCommand(command, stop, stdout, stderr)
	go func() {
		if err := <-cmd; err != nil {
			errCh <- err
			return
		}

		link, err := os.Readlink(path.Join(tmpPath, "result"))
		if err != nil {
			errCh <- err
			return
		}

		if err := os.RemoveAll(tmpPath); err != nil {
			errCh <- err
			return
		}

		resultCh <- link
	}()

	return resultCh, errCh
}
