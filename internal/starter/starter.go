package starter

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"syscall"

	"github.com/sopcoerik/fictional-robot/internal/parser"
)

func StartService(service *parser.Service, ctx context.Context, serviceChan chan error, logChan chan string) {

	// run the process once; restarting is driven by the user (Start / Restart All),
	// which calls this again with a fresh context.
	if err := ctx.Err(); err != nil {
		return
	}

	cmd := exec.Command("sh", "-c", service.Command)

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		serviceChan <- err
		return
	}

	// the process started; whether it actually comes up is verified by the
	// health check in the caller
	serviceChan <- nil

	// stream BOTH stdout and stderr into the log channel, line by line
	pumpLogs := func(pipe io.Reader, source string) {
		scanner := bufio.NewScanner(pipe)

		for scanner.Scan() {
			logChan <- fmt.Sprintf("[%s]: %s\n", source, scanner.Text())
		}
	}

	go pumpLogs(stdoutPipe, "STDOUT")
	go pumpLogs(stderrPipe, "STDERR")

	go func() {
		<-ctx.Done()
		syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
	}()

	cmd.Wait()
}
