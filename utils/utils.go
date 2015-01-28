package utils

import (
	log "github.com/Sirupsen/logrus"
	"io"
	"os/exec"
	"syscall"
	"time"
)

func RunCommand(command []string, stop <-chan time.Time, output io.Writer, stderr io.Writer) <-chan error {
	out := make(chan error, 1)

	cmd := exec.Command(command[0], command[1:len(command)]...)
	cmd.Stdout, cmd.Stderr = output, stderr

	cmd.Start()
	done := make(chan error, 1)

	go func() {
		done <- cmd.Wait()
	}()

	go func() {
		select {
		case <-stop:
			if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
				log.WithFields(log.Fields{
					"cmd": command,
					"pid": cmd.Process.Pid,
				}).Warn("cannot stop process")
			}

			select {
			case <-time.After(time.Duration(10) * time.Second):
				if err := cmd.Process.Kill(); err != nil {
					log.WithFields(log.Fields{
						"cmd": command,
						"pid": cmd.Process.Pid,
					}).Error("cannot kill process")
				}

				out <- <-done
			case err := <-done:
				out <- err
			}
		case err := <-done:
			out <- err
		}
	}()

	return out
}
