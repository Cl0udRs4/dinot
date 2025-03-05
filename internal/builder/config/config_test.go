package config

import (
    "reflect"
    "testing"
)

func TestParseProtocols(t *testing.T) {
    tests := []struct {
        name        string
        protocolStr string
        want        []string
    }{
        {
            name:        "Empty string",
            protocolStr: "",
            want:        []string{},
        },
        {
            name:        "Single protocol",
            protocolStr: "tcp",
            want:        []string{"tcp"},
        },
        {
            name:        "Multiple protocols",
            protocolStr: "tcp,udp,ws",
            want:        []string{"tcp", "udp", "ws"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ParseProtocols(tt.protocolStr)
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseProtocols() = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestParseServers(t *testing.T) {
    tests := []struct {
        name      string
        serverStr string
        want      map[string]string
    }{
        {
            name:      "Empty string",
            serverStr: "",
            want:      map[string]string{},
        },
        {
            name:      "Single server",
            serverStr: "tcp:localhost:8080",
            want:      map[string]string{"tcp": "localhost:8080"},
        },
        {
            name:      "Multiple servers",
            serverStr: "tcp:localhost:8080,udp:localhost:8081",
            want:      map[string]string{"tcp": "localhost:8080", "udp": "localhost:8081"},
        },
        {
            name:      "Invalid format",
            serverStr: "invalid",
            want:      map[string]string{},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ParseServers(tt.serverStr)
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseServers() = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestParseModules(t *testing.T) {
    tests := []struct {
        name      string
        moduleStr string
        want      []string
    }{
        {
            name:      "Empty string",
            moduleStr: "",
            want:      []string{},
        },
        {
            name:      "Single module",
            moduleStr: "shell",
            want:      []string{"shell"},
        },
        {
            name:      "Multiple modules",
            moduleStr: "shell,process,file",
            want:      []string{"shell", "process", "file"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ParseModules(tt.moduleStr)
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseModules() = %v, want %v", got, tt.want)
            }
        })
    }
}
