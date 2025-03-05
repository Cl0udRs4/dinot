package template

import (
	"bytes"
	"text/template"
)

// ClientTemplate is the template for generating client code
const ClientTemplate = `package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Cl0udRs4/dinot/internal/client"
	"github.com/Cl0udRs4/dinot/internal/client/encryption"
	"github.com/Cl0udRs4/dinot/internal/client/module"
	{{range .Modules}}
	"github.com/Cl0udRs4/dinot/internal/client/module/{{.}}"
	{{end}}
)

// Version information
var (
	Version   = "{{.Version}}"
	BuildTime = "{{.BuildTime}}"
	Signature = "{{.Signature}}"
)

func main() {
	// Parse command-line flags
	debug := flag.Bool("debug", {{.Debug}}, "Enable debug mode")
	flag.Parse()

	// Create client configuration
	cfg := &client.Config{
		ClientID:          "{{.ClientID}}",
		Protocols:         []string{ {{range .Protocols}}"{{.}}", {{end}} },
		Domain:            "{{.Domain}}",
		Servers:           map[string]string{ {{range $key, $value := .Servers}}"{{$key}}": "{{$value}}", {{end}} },
		HeartbeatInterval: {{.HeartbeatInterval}},
		Encryption:        "{{.Encryption}}",
		Debug:             {{.Debug}},
	}

	// Create feedback configuration
	feedbackCfg := &client.FeedbackConfig{
		MaxRetries:         3,
		RetryInterval:      time.Second,
		MaxRetryInterval:   time.Minute,
		RetryBackoffFactor: 2.0,
	}

	// Create a new client
	c, err := client.NewClient(cfg, feedbackCfg)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}

	// Load modules
	{{range .Modules}}
	if err := c.LoadModule({{.}}.NewModule()); err != nil {
		fmt.Printf("Failed to load {{.}} module: %v\n", err)
	}
	{{end}}

	// Enable random heartbeat
	c.EnableRandomHeartbeat(time.Second*30, time.Hour*24)

	// Start the client
	if err := c.Start(context.Background()); err != nil {
		fmt.Printf("Failed to start client: %v\n", err)
		os.Exit(1)
	}

	// Wait for termination signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Stop the client
	c.Stop()
}
`

// GenerateClientCode generates client code based on the provided parameters
func GenerateClientCode(params map[string]interface{}) ([]byte, error) {
	tmpl, err := template.New("client").Parse(ClientTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
