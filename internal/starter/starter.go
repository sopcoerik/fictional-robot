package starter

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"
	"time"

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

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	err := cmd.Start()

	if err != nil {
		serviceChan <- err
		return
	}

	// give the command time to fail (bad syntax etc.)
	time.Sleep(50 * time.Millisecond)

	stderrStr := strings.TrimSpace(stderr.String())
	if stderrStr != "" {
		serviceChan <- fmt.Errorf("ERROR: %s: %s", service.Command, stderrStr)
		return
	}

	serviceChan <- nil

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
