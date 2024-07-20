package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/adhocore/gronx/pkg/tasker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func run() error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	streamForMetricsServer := make(chan error)
	go runMetricsServer(streamForMetricsServer)
	streamForCronRunner := make(chan error)
	go runCronRunner(streamForCronRunner)

	for {
		select {
		case s := <-sig:
			log.Printf("signal received: %s", s.String())
			return nil
		case err := <-streamForMetricsServer:
			return fmt.Errorf("metricsServer failed: %w", err)
		case err := <-streamForCronRunner:
			return fmt.Errorf("cronRunner failed: %w", err)
		}
	}
}

func runMetricsServer(errCh chan<- error) {
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	errCh <- http.ListenAndServe(config.metricsBindAddr, nil)
}

func runCronRunner(_ chan<- error) {
	taskr := tasker.New(tasker.Option{
		Verbose: true,
		Tz:      config.timezone,
		Out:     config.output,
	})
	taskr.Task(config.cronExpression, taskRunner{taskr.Log, config}.cronTask).Run()
}

type taskRunner struct {
	logger *log.Logger
	config Config
}

func (tr taskRunner) cronTask(ctx context.Context) (resCode int, resErr error) {
	ctx = startMeasurement(ctx)

	logger := tr.logger
	config := tr.config
	succeeded := true
	buf := &bytes.Buffer{}
	var mw io.Writer
	if config.regexMatcher != nil {
		mw = io.MultiWriter(buf, logger.Writer())
	} else {
		mw = logger.Writer()
	}

	ctx, cancelFunc := context.WithTimeout(ctx, time.Duration(config.timeoutMillisecond)*time.Millisecond)
	defer cancelFunc()
	cmd := exec.CommandContext(ctx, config.commands[0], config.commands[1:]...)
	cmd.Stdout = mw
	cmd.Stderr = mw
	if err := cmd.Run(); err != nil {
		if !config.ignoreExitCode {
			succeeded = false
			switch err.(type) {
			case *exec.ExitError:
				resCode = cmd.ProcessState.ExitCode()
				resErr = err
			case *exec.Error:
				resCode = cmd.ProcessState.ExitCode()
			}
		}
	}
	if config.regexMatcher != nil {
		if !config.regexMatcher.Match(buf.Bytes()) {
			succeeded = false
			resErr = fmt.Errorf("regex pattern is not matched with outputs")
			resCode = 1
		}
	}
	finishMeasurement(ctx, succeeded)
	return
}
