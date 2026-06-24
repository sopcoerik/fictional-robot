package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sopcoerik/fictional-robot/internal/parser"
	"github.com/sopcoerik/fictional-robot/internal/sorter"
	"github.com/sopcoerik/fictional-robot/internal/starter"
)

type RunningService struct {
	Name        string
	Service     *parser.Service
	Ctx         context.Context
	Cancel      context.CancelFunc
	GlobalCtx   context.Context // parent context to create fresh child contexts from
	ServiceChan chan error
	LogChan     chan string
	Logs        []string
	LogMutex    sync.Mutex
}

type AppState struct {
	RunningServices map[string]*RunningService
	OrderedNames    []string
	GlobalCtx       context.Context
	GlobalCancel    context.CancelFunc
	ParentCtx       context.Context
	Config          *parser.Config
	Mu              sync.Mutex // guards RunningServices, GlobalCtx, GlobalCancel (UI reads them concurrently)
}

// GetService returns the running service for a name, or nil if it isn't
// present (e.g. mid-restart). Safe to call from the UI goroutine.
func (app *AppState) GetService(name string) *RunningService {
	app.Mu.Lock()
	defer app.Mu.Unlock()
	return app.RunningServices[name]
}

func CheckHealth(ctx context.Context, url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("health check timeout")
		case <-ticker.C:
			conn, err := net.Dial("tcp", url)
			if err == nil {
				conn.Close()
				return nil
			}
		}
	}
}

func NewRunningService(name string, service *parser.Service, globalCtx context.Context) *RunningService {
	ctx, cancel := context.WithCancel(globalCtx)

	rs := &RunningService{
		Name:        name,
		Service:     service,
		Ctx:         ctx,
		Cancel:      cancel,
		GlobalCtx:   globalCtx,
		ServiceChan: make(chan error, 1),
		LogChan:     make(chan string, 100),
		Logs:        []string{},
	}

	// one log collector for this service's whole lifetime: it drains LogChan and
	// exits when the global context is cancelled (shutdown or restart). This is
	// why Start() no longer spawns a collector each time it runs.
	go func() {
		for {
			select {
			case <-globalCtx.Done():
				return
			case log := <-rs.LogChan:
				rs.AddLog(log)
			}
		}
	}()

	return rs
}

func (rs *RunningService) AddLog(log string) {
	rs.LogMutex.Lock()
	defer rs.LogMutex.Unlock()
	rs.Logs = append(rs.Logs, log)
	// keep only last 100 logs in memory
	if len(rs.Logs) > 100 {
		rs.Logs = rs.Logs[1:]
	}
}

func (rs *RunningService) GetLogs() []string {
	rs.LogMutex.Lock()
	defer rs.LogMutex.Unlock()
	// Return a copy
	logsCopy := make([]string, len(rs.Logs))
	copy(logsCopy, rs.Logs)
	return logsCopy
}

func (rs *RunningService) Start() error {
	// create a fresh context as child of global context
	// this allows restarting after a previous cancellation
	rs.Ctx, rs.Cancel = context.WithCancel(rs.GlobalCtx)

	go starter.StartService(rs.Service, rs.Ctx, rs.ServiceChan, rs.LogChan)

	// wait for initial error report
	err := <-rs.ServiceChan
	if err != nil {
		return err
	}

	// health check
	serviceUrl := fmt.Sprintf("localhost:%d", rs.Service.Port)
	err = CheckHealth(rs.Ctx, serviceUrl, 5*time.Second)
	if err != nil {
		return fmt.Errorf("health check failed for %s: %w", rs.Name, err)
	}

	return nil
}

func (rs *RunningService) Stop() {
	rs.Cancel()
}

func StartAllServices(app *AppState) error {
	for _, sName := range app.OrderedNames {
		rs := app.GetService(sName)
		if rs == nil {
			continue
		}

		err := rs.Start()
		if err != nil {
			fmt.Printf("Error starting %s: %v\n", sName, err)
			return err
		}
	}
	return nil
}

func StopAllServices(app *AppState) {
	app.Mu.Lock()
	cancel := app.GlobalCancel
	app.Mu.Unlock()
	cancel()
}

func RestartAllServices(app *AppState) error {
	StopAllServices(app)

	// give services time to shut down
	time.Sleep(500 * time.Millisecond)

	// build the new context and services locally first
	globalCtx, globalCancel := context.WithCancel(app.ParentCtx)

	newServices := make(map[string]*RunningService)
	for _, sName := range app.OrderedNames {
		service := app.Config.Services[sName]
		newServices[sName] = NewRunningService(sName, &service, globalCtx)
	}

	// swap them in atomically under the lock (the UI reads these concurrently)
	app.Mu.Lock()
	app.GlobalCtx = globalCtx
	app.GlobalCancel = globalCancel
	app.RunningServices = newServices
	app.Mu.Unlock()

	return StartAllServices(app)
}

func StopService(app *AppState, serviceName string) {
	if rs := app.GetService(serviceName); rs != nil {
		rs.Stop()
	}
}

func main() {
	defer func() {
		fmt.Println("Shutting down")
	}()

	config, err := parser.ParseConfig("devenv.yaml")
	if err != nil {
		fmt.Println("failed to read config:", err)
		return
	}

	orderedNames, err := sorter.SortServices(config)
	if err != nil {
		fmt.Println("config error:", err)
		return
	}

	// parent context (survives restarts, killed only on program exit)
	parentCtx, parentCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parentCancel()

	// global context for current service lifecycle
	globalCtx, globalCancel := context.WithCancel(parentCtx)

	app := &AppState{
		RunningServices: make(map[string]*RunningService),
		OrderedNames:    orderedNames,
		GlobalCtx:       globalCtx,
		GlobalCancel:    globalCancel,
		ParentCtx:       parentCtx,
		Config:          config,
	}

	// initialize running services
	for _, sName := range orderedNames {
		service := config.Services[sName]
		app.RunningServices[sName] = NewRunningService(sName, &service, app.GlobalCtx)
	}

	err = StartAllServices(app)
	if err != nil {
		fmt.Printf("Failed to start services: %v\n", err)
		return
	}

	// start UI (blocks until user quits)
	err = RunUI(app)

	if err != nil {
		fmt.Println("UI error:", err)
	}

	// clean up when UI exits
	app.GlobalCancel()
}
