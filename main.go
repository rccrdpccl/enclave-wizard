package main

import (
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/api"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/config"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/validation"
)

type Options struct {
	Port       int    `help:"Port to listen on" short:"p" default:"8080"`
	EnclaveDir string `help:"Path to the Enclave repository root" default:"../enclave"`
}

func SetupAPI(mux *http.ServeMux, enclaveDir string) huma.API {
	apiConfig := huma.DefaultConfig("Enclave Configuration Wizard", "0.1.0")
	apiConfig.Info.Description = "API for managing Enclave deployment configuration files on the Landing Zone."
	humaAPI := humago.New(mux, apiConfig)

	reader := config.NewReader(enclaveDir)
	writer := config.NewWriter(enclaveDir)
	validator := validation.NewValidator(enclaveDir)

	api.NewConfigHandler(reader, writer, validator).Register(humaAPI)
	api.NewDefaultsHandler(enclaveDir).Register(humaAPI)
	api.NewPluginsHandler().Register(humaAPI)

	return humaAPI
}

func main() {
	cli := humacli.New(func(hooks humacli.Hooks, opts *Options) {
		mux := http.NewServeMux()
		SetupAPI(mux, opts.EnclaveDir)

		hooks.OnStart(func() {
			fmt.Printf("Enclave Wizard API listening on :%d (enclave-dir: %s)\n", opts.Port, opts.EnclaveDir)
			fmt.Printf("API docs: http://localhost:%d/docs\n", opts.Port)
			http.ListenAndServe(fmt.Sprintf(":%d", opts.Port), mux)
		})
	})
	cli.Run()
}
