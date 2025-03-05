package cli

import (
    "fmt"
    "os"

    "github.com/Cl0udRs4/dinot/internal/builder/config"
    "github.com/Cl0udRs4/dinot/internal/builder/generator"
    "github.com/Cl0udRs4/dinot/internal/builder/signature"
    "github.com/Cl0udRs4/dinot/internal/builder/validation"
    "github.com/spf13/cobra"
)

var (
    // Required flags
    protocolFlag string
    domainFlag   string
    serversFlag  string
    modulesFlag  string

    // Optional flags
    encryptionFlag string
    debugFlag      bool
    versionFlag    string
    signatureFlag  bool
)

var buildCmd = &cobra.Command{
    Use:   "build",
    Short: "Build a client executable",
    Long: `Build a client executable based on the specified parameters.
This command generates a client executable with the specified protocols, servers, and modules.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        // Create a new BuilderConfig
        cfg := config.NewBuilderConfig()

        // Parse required parameters
        cfg.Protocols = config.ParseProtocols(protocolFlag)
        cfg.Domain = domainFlag
        cfg.Servers = config.ParseServers(serversFlag)
        cfg.Modules = config.ParseModules(modulesFlag)

        // Parse optional parameters
        cfg.Encryption = encryptionFlag
        cfg.Debug = debugFlag
        cfg.Version = versionFlag
        cfg.Signature = signatureFlag

        // Validate the configuration
        if err := validation.ValidateConfig(cfg); err != nil {
            return fmt.Errorf("validation error: %w", err)
        }

        // Create output directory if it doesn't exist
        outputDir := "build"
        if err := os.MkdirAll(outputDir, 0755); err != nil {
            return fmt.Errorf("failed to create output directory: %w", err)
        }

        // Create a new generator
        gen := generator.NewGenerator(cfg, outputDir)

        // Set up signature manager if signature verification is enabled
        if cfg.Signature {
            sm := signature.NewSignatureManager(
                "keys/private.pem",
                "keys/public.pem",
            )

            // Create keys directory if it doesn't exist
            if err := os.MkdirAll("keys", 0755); err != nil {
                return fmt.Errorf("failed to create keys directory: %w", err)
            }

            // Load or generate keys
            err := sm.LoadKeys()
            if err == signature.ErrKeyNotFound {
                fmt.Println("Generating new key pair...")
                if err := sm.GenerateKeyPair(2048); err != nil {
                    return fmt.Errorf("failed to generate key pair: %w", err)
                }
            } else if err != nil {
                return fmt.Errorf("failed to load keys: %w", err)
            }

            gen.SetSignatureManager(sm)
        }

        // Generate the client
        outputPath, err := gen.Generate()
        if err != nil {
            return fmt.Errorf("failed to generate client: %w", err)
        }

        fmt.Printf("Client generated successfully: %s\n", outputPath)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(buildCmd)

    // Required flags
    buildCmd.Flags().StringVarP(&protocolFlag, "protocol", "p", "", "Protocols to support (comma-separated: tcp,udp,ws,icmp,dns)")
    buildCmd.Flags().StringVarP(&domainFlag, "domain", "d", "", "Domain for DNS protocol")
    buildCmd.Flags().StringVarP(&serversFlag, "servers", "s", "", "Server addresses (format: protocol1:address1,protocol2:address2)")
    buildCmd.Flags().StringVarP(&modulesFlag, "modules", "m", "", "Modules to include (comma-separated)")

    // Optional flags
    buildCmd.Flags().StringVarP(&encryptionFlag, "encryption", "e", "aes", "Encryption method (none, aes, chacha20)")
    buildCmd.Flags().BoolVar(&debugFlag, "debug", false, "Enable debug mode")
    buildCmd.Flags().StringVar(&versionFlag, "version", "1.0.0", "Client version")
    buildCmd.Flags().BoolVar(&signatureFlag, "signature", false, "Enable signature verification")

    // Mark required flags
    buildCmd.MarkFlagRequired("protocol")
    buildCmd.MarkFlagRequired("servers")
    buildCmd.MarkFlagRequired("modules")
}
