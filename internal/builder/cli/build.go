package cli

import (
    "fmt"

    "github.com/Cl0udRs4/dinot/internal/builder/config"
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

        // Print the configuration (for now, will be replaced with actual build logic)
        fmt.Printf("Configuration:\n")
        fmt.Printf("  Protocols: %v\n", cfg.Protocols)
        fmt.Printf("  Domain: %s\n", cfg.Domain)
        fmt.Printf("  Servers: %v\n", cfg.Servers)
        fmt.Printf("  Modules: %v\n", cfg.Modules)
        fmt.Printf("  Encryption: %s\n", cfg.Encryption)
        fmt.Printf("  Debug: %v\n", cfg.Debug)
        fmt.Printf("  Version: %s\n", cfg.Version)
        fmt.Printf("  Signature: %v\n", cfg.Signature)

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
