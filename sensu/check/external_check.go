package check

import (
	"bytes"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type ExternalCheck struct {
	Command string
}

func (c *ExternalCheck) Execute() CheckOutput {
	command := strings.Split(c.Command, " ")

	t0 := time.Now()
	cmd := exec.Command(command[0], command[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Start(); err != nil {
		return CheckOutput{
			Error,
			err.Error(),
			time.Now().Sub(t0).Seconds(),
			t0.Unix(),
		}
	}

	status := Success

	if err := cmd.Wait(); err != nil {
		status = Error
		if exiterr, ok := err.(*exec.ExitError); ok {
			if statusReturn, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				status = ExitStatus(statusReturn)
			}
		}
	}

	return CheckOutput{
		status,
		out.String(),
		time.Now().Sub(t0).Seconds(),
		t0.Unix(),
	}
}
