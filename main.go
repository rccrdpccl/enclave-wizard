package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/api"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/auth"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/config"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/tasks"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/validation"
)

type Options struct {
	Port         int    `help:"Port to listen on" short:"p" default:"8080"`
	EnclaveDir   string `help:"Path to the Enclave repository root" default:"../enclave"`
	PasswordFile string `help:"Path to the password file" default:"/etc/enclave-wizard/password"`
}

func SetupAPI(mux *http.ServeMux, enclaveDir string, authStore *auth.Store) (huma.API, *tasks.AnsibleRunner) {
	apiConfig := huma.DefaultConfig("Enclave Configuration Wizard", "0.1.0")
	apiConfig.Info.Description = "API for managing Enclave deployment configuration files on the Landing Zone."

	apiConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "opaque",
		},
	}

	humaAPI := humago.New(mux, apiConfig)

	reader := config.NewReader(enclaveDir)
	writer := config.NewWriter(enclaveDir)
	validator := validation.NewValidator(enclaveDir)

	api.NewAuthHandler(authStore).Register(humaAPI)
	api.NewConfigHandler(reader, writer, validator).Register(humaAPI)
	api.NewDefaultsHandler(enclaveDir).Register(humaAPI)
	api.NewPluginsHandler().Register(humaAPI)

	runner, err := tasks.NewAnsibleRunner(enclaveDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: task runner unavailable: %v (tasks API disabled)\n", err)
	} else {
		if err := runner.Recover(); err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: task recovery failed: %v\n", err)
		}
		api.NewTasksHandler(runner).Register(humaAPI)
	}

	return humaAPI, runner
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "hash-password" {
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: enclave-wizard hash-password <password>")
			os.Exit(1)
		}
		hashed, err := auth.HashPassword(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(hashed)
		return
	}

	cli := humacli.New(func(hooks humacli.Hooks, opts *Options) {
		dir := filepath.Dir(opts.PasswordFile)
		os.MkdirAll(dir, 0700)

		authStore := auth.NewStore(opts.PasswordFile)
		generatedPass, err := authStore.Init()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize auth: %v\n", err)
			os.Exit(1)
		}

		if generatedPass != "" {
			fmt.Println("")
			fmt.Println("  ┌────────────────────────────────────────────┐")
			fmt.Printf("  │  Initial admin password: %-18s │\n", generatedPass)
			fmt.Println("  │  (You must change it on first login)       │")
			fmt.Println("  └────────────────────────────────────────────┘")
			fmt.Println("")
			os.WriteFile("/tmp/enclave-wizard-init-pass", []byte(generatedPass+"\n"), 0644)
		}

		mux := http.NewServeMux()
		_, runner := SetupAPI(mux, opts.EnclaveDir, authStore)

		handler := api.BearerAuthMiddleware(authStore)(mux)

		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", opts.Port),
			Handler: handler,
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
