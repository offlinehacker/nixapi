package utils

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestRunCommand(t *testing.T) {
	var stdout bytes.Buffer
	err := <-RunCommand([]string{"echo", "test"}, nil, &stdout, nil)

	assert.Nil(t, err)
	assert.Equal(t, "test\n", string(stdout.Bytes()[:]))
}

func TestRunCommandTimeout(t *testing.T) {
	timer := time.After(time.Duration(100) * time.Millisecond)
	err := <-RunCommand([]string{"sleep", "1"}, timer, nil, nil)
	assert.NotNil(t, err)
	assert.IsType(t, &exec.ExitError{}, err)
	assert.Equal(t, -1, err.(*exec.ExitError).Sys().(syscall.WaitStatus).ExitStatus())
}

func TestRunCommandStderr(t *testing.T) {
	var stderr bytes.Buffer
	_ = <-RunCommand([]string{"logger", "-s", "test"}, nil, nil, &stderr)

	assert.Equal(t, "test\n", string(stderr.Bytes()[:]))
}

func TestRunCommandError(t *testing.T) {
	err := <-RunCommand([]string{"ls", "pšđ"}, nil, nil, nil)
	assert.NotNil(t, err)
	assert.IsType(t, &exec.ExitError{}, err)
	assert.Equal(t, 2, err.(*exec.ExitError).Sys().(syscall.WaitStatus).ExitStatus())
}
