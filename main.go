package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/api"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/config"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/tasks"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/validation"
)

type Options struct {
	Port       int    `help:"Port to listen on" short:"p" default:"8080"`
	EnclaveDir string `help:"Path to the Enclave repository root" default:"../enclave"`
}

func SetupAPI(mux *http.ServeMux, enclaveDir string) (huma.API, *tasks.AnsibleRunner) {
	apiConfig := huma.DefaultConfig("Enclave Configuration Wizard", "0.1.0")
	apiConfig.Info.Description = "API for managing Enclave deployment configuration files on the Landing Zone."
	humaAPI := humago.New(mux, apiConfig)

	reader := config.NewReader(enclaveDir)
	writer := config.NewWriter(enclaveDir)
	validator := validation.NewValidator(enclaveDir)

	api.NewConfigHandler(reader, writer, validator).Register(humaAPI)
	api.NewDefaultsHandler(enclaveDir).Register(humaAPI)
	api.NewPluginsHandler().Register(humaAPI)

	runner, err := tasks.NewAnsibleRunner(enclaveDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: task runner unavailable: %v\n", err)
		os.Exit(1)
	}

	if err := runner.Recover(); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: task recovery failed: %v\n", err)
	}

	api.NewTasksHandler(runner).Register(humaAPI)
	return humaAPI, runner
}

func main() {
	cli := humacli.New(func(hooks humacli.Hooks, opts *Options) {
		mux := http.NewServeMux()
		_, runner := SetupAPI(mux, opts.EnclaveDir)

		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", opts.Port),
			Handler: mux,
		}

		hooks.OnStart(func() {
			fmt.Printf("Enclave Wizard API listening on :%d (enclave-dir: %s)\n", opts.Port, opts.EnclaveDir)
			fmt.Printf("API docs: http://localhost:%d/docs\n", opts.Port)

			go func() {
				sigCh := make(chan os.Signal, 1)
				signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
				<-sigCh

				fmt.Println("\nShutting down...")

				if runner != nil {
					shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					if err := runner.Shutdown(shutdownCtx); err != nil {
						fmt.Fprintf(os.Stderr, "runner shutdown error: %v\n", err)
					}
				}

				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				server.Shutdown(shutdownCtx)
			}()

			server.ListenAndServe()
		})
	})
	cli.Run()
}
