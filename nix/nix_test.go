package nix

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
	"time"
)

var derivations = []byte(`[
	{
		"name": "test",
		"attrPath": "-",
		"out": "/nix/store"
	},
	{
		"name": "test2",
		"attrPath": "-",
		"out": "/nix/store"
	}
]`)

func TestDerivation(t *testing.T) {

	keys := make([]Derivation, 0)
	err := json.Unmarshal(derivations, &keys)

	assert.Nil(t, err)
	assert.Equal(t, keys[0].Name, "test")
	assert.Equal(t, keys[0].AttrPath, "-")
	assert.Equal(t, keys[0].Out, "/nix/store")
}

func TestGetDerivations(t *testing.T) {
	RunCommand = func(command []string, stop <-chan time.Time, output io.Writer, stderr io.Writer) <-chan error {
		output.Write(derivations)
		out := make(chan error, 1)
		out <- nil
		return out
	}

	expression := NewExpression("/home/offlinehacker/nix/workdirs/master")
	res, err := expression.GetDerivations(nil)

	select {
	case <-err:
		assert.Fail(t, "no error should be returned")
	case derivations := <-res:
		assert.Equal(t, "test", derivations[0].Name)
	}
}

func TestGetDerivationsError(t *testing.T) {
	RunCommand = func(command []string, stop <-chan time.Time, output io.Writer, stderr io.Writer) <-chan error {
		stderr.Write([]byte(`stderr`))
		out := make(chan error, 1)
		out <- errors.New("undefined error")
		return out
	}

	expression := NewExpression("/home/offlinehacker/nix/workdirs/master")
	res, err := expression.GetDerivations(nil)

	select {
	case <-res:
		assert.Fail(t, "should not return result")
	case err := <-err:
		assert.NotNil(t, err)
		assert.Equal(t, []byte(`stderr`), err.(ExpressionError).stderr.Bytes())
	}
}

func TestBuildDerivation(t *testing.T) {
	RunCommand = func(command []string, stop <-chan time.Time, output io.Writer, stderr io.Writer) <-chan error {
		output.Write([]byte(`stdout`))
		stderr.Write([]byte(`stderr`))
		out := make(chan error, 1)

		if err := os.Symlink("/nix/store/abcdef", command[3]); err != nil {
			out <- err
		} else {
			out <- nil
		}
		return out
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	derivation := Derivation{Expression: Expression{Path: ""}}
	pathCh, errCh := derivation.build(nil, &stdout, &stderr)

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case path := <-pathCh:
		assert.Equal(t, "/nix/store/abcdef", path)
		assert.Equal(t, stdout.Bytes(), []byte(`stdout`))
		assert.Equal(t, stderr.Bytes(), []byte(`stderr`))
	}
}

func TestBuildDerivationError(t *testing.T) {
	RunCommand = func(command []string, stop <-chan time.Time, output io.Writer, stderr io.Writer) <-chan error {
		out := make(chan error, 1)
		out <- errors.New("some error")
		return out
	}

	derivation := Derivation{Expression: Expression{Path: ""}}
	pathCh, errCh := derivation.build(nil, nil, nil)

	select {
	case err := <-errCh:
		assert.EqualError(t, err, "some error")
	case <-pathCh:
		assert.Fail(t, "should not return path")
	}

}
