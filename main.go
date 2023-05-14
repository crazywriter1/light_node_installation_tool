package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}

	celestiaNodeDir := filepath.Join(homeDir, "celestia-node")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter the IP address: ")
	ipAddress, _ := reader.ReadString('\n')
	ipAddress = strings.TrimSpace(ipAddress)

	fmt.Print("Enter the port: ")
	port, _ := reader.ReadString('\n')
	port = strings.TrimSpace(port)

	fmt.Print("Enter the network: ")
	network, _ := reader.ReadString('\n')
	network = strings.TrimSpace(network)

	commands := []struct {
		cmd        string
		workingDir string
	}{
		{"sudo apt update", ""},
		{"sudo apt upgrade -y", ""},
		{"sudo apt install curl tar wget clang pkg-config libssl-dev jq build-essential git make ncdu -y", ""},
		{"rm -rf celestia-node", homeDir},
		{"git clone https://github.com/celestiaorg/celestia-node.git", homeDir},
		{"git checkout tags/v0.9.3", celestiaNodeDir},
		{"make build", celestiaNodeDir},
		{"make install", celestiaNodeDir},
		{"make cel-key", celestiaNodeDir},
		{"celestia version", ""},
		{"celestia light init --p2p.network " + network, ""},
	}

	for _, command := range commands {
		fmt.Printf("Running command: %s\n", command.cmd)
		err := runCommand(command.cmd, command.workingDir)
		if err != nil {
			log.Fatalf("Command failed: %s\nError: %v\n", command.cmd, err)
		}
	}

	celestiaServiceCmd := fmt.Sprintf(`sudo tee <<EOF >/dev/null /etc/systemd/system/celestia-lightd.service
[Unit]
Description=celestia-lightd Light Node
After=network-online.target

[Service]
User=$USER
ExecStart=/usr/local/bin/celestia light start --core.ip %s --core.rpc.port 26657 --core.grpc.port 9090 --keyring.accname my_celes_key --metrics.tls=false --metrics --metrics.endpoint otel.celestia.tools:4318 --gateway --gateway.addr localhost --gateway.port 26659 --p2p.network %s
Restart=on-failure
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
EOF`, ipAddress, network)

	err = runCommand(celestiaServiceCmd, "")
	if err != nil {
		log.Fatalf("Command failed: %s\nError: %v\n", celestiaServiceCmd, err)
	}

	fmt.Println("All commands executed successfully.")
}

func runCommand(command, workingDir string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if workingDir != "" {
		cmd.Dir = workingDir
	}

	err := cmd.Run()
	return err
}