package starter

import (
	"os/exec"
	"os"
	"syscall"
	"context"

	"github.com/sopcoerik/fictional-robot/internal/parser"
)


func StartService(service *parser.Service, ctx context.Context, serviceChan chan error) {
	cmd := exec.Command("sh", "-c", service.Command)	

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	serviceChan <- err

	go func() {
		<-ctx.Done()
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}()

	cmd.Wait()
}
