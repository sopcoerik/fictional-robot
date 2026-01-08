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
	backReport := false

	for {

		if err := ctx.Err(); err != nil {
			fmt.Printf("%s context: %v", service.Command, err)
			return
		}

		cmd := exec.Command("sh", "-c", service.Command)

		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		stdoutPipe, _ := cmd.StdoutPipe()
		stderrPipe, _ := cmd.StderrPipe()

		err := cmd.Start()

		if !backReport {
			serviceChan <- err
			backReport = true
		}

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
}
