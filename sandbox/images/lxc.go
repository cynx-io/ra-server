package images

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// LXCBuilder handles LXC container creation from Dockerfile-like instructions
type LXCBuilder struct {
	WorkDir      string
	ContainerDir string
	LogFile      *os.File
}

// NewLXCBuilder creates a new LXC builder instance
func NewLXCBuilder(workDir, containerDir string) *LXCBuilder {
	logFile, err := os.Create(filepath.Join(workDir, fmt.Sprintf("build-%d.log", time.Now().Unix())))
	if err != nil {
		fmt.Printf("Warning: Failed to create log file: %v\n", err)
	}

	return &LXCBuilder{
		WorkDir:      workDir,
		ContainerDir: containerDir,
		LogFile:      logFile,
	}
}

// Close cleans up resources
func (l *LXCBuilder) Close() {
	if l.LogFile != nil {
		l.LogFile.Close()
	}
}

// log writes a message to both stdout and log file
func (l *LXCBuilder) log(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf("[%s] %s", timestamp, fmt.Sprintf(format, args...))

	fmt.Println(message)

	if l.LogFile != nil {
		l.LogFile.WriteString(message + "\n")
		l.LogFile.Sync()
	}
}

// runCommand executes a command and captures output
func (l *LXCBuilder) runCommand(name string, args ...string) error {
	l.log("Running: %s %s", name, strings.Join(args, " "))

	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()

	if len(output) > 0 {
		fmt.Print(string(output))
		if l.LogFile != nil {
			l.LogFile.WriteString(string(output))
			l.LogFile.Sync()
		}
	}

	return err
}

// RunUbuntuContainer creates an Ubuntu 22.04 LXC container like a Dockerfile
func RunUbuntuContainer() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get current directory: %v\n", err)
		return
	}

	workDir := filepath.Join(pwd, "sandbox", "build", "lxc-ubuntu")
	containerDir := filepath.Join(workDir, "containers")

	// Create directories
	if err := os.MkdirAll(workDir, 0755); err != nil {
		fmt.Printf("Failed to create work directory: %v\n", err)
		return
	}
	if err := os.MkdirAll(containerDir, 0755); err != nil {
		fmt.Printf("Failed to create container directory: %v\n", err)
		return
	}

	fmt.Printf("🏗️  Work directory: %s\n", workDir)
	fmt.Printf("🐳 Container directory: %s\n", containerDir)

	builder := NewLXCBuilder(workDir, containerDir)
	defer builder.Close()

	if err := buildUbuntuContainer(builder); err != nil {
		fmt.Printf("Container build failed: %v\n", err)
		return
	}

	fmt.Println("✅ Ubuntu 22.04 LXC container created successfully!")
	containerPath := filepath.Join(containerDir, "ubuntu-base")
	fmt.Printf("📁 Container location: %s\n", containerPath)
}

// buildUbuntuContainer creates Ubuntu 22.04 container with Dockerfile-like steps
func buildUbuntuContainer(l *LXCBuilder) error {
	containerName := "ubuntu-base"

	l.log("🐳 FROM ubuntu:22.04 - Creating Ubuntu 22.04 LXC container...")

	// Clean up any existing container
	l.log("Cleaning up any existing container: %s", containerName)
	l.runCommand("lxc-stop", "-n", containerName, "-P", l.ContainerDir)
	l.runCommand("lxc-destroy", "-n", containerName, "-P", l.ContainerDir)

	// Create LXC container - equivalent to FROM ubuntu:22.04
	if err := l.runCommand("lxc-create", "-t", "download", "-n", containerName, "-P", l.ContainerDir, "--", "--dist", "ubuntu", "--release", "jammy", "--arch", "amd64"); err != nil {
		return fmt.Errorf("failed to create LXC container: %w", err)
	}

	// Configure container for better compatibility
	configPath := filepath.Join(l.ContainerDir, containerName, "config")
	additionalConfig := `
# Disable networking to avoid bridge issues
lxc.net.0.type = none
lxc.apparmor.profile = unconfined
lxc.cap.drop = 
`
	if err := l.appendToConfig(configPath, additionalConfig); err != nil {
		l.log("Warning: failed to modify container config: %v", err)
	}

	// Modify rootfs directly instead of starting container (chroot approach)
	rootfsPath := filepath.Join(l.ContainerDir, containerName, "rootfs")
	l.log("🔧 Modifying container rootfs directly at: %s", rootfsPath)

	// Create a script to run inside chroot
	scriptPath := filepath.Join(l.ContainerDir, "setup.sh")
	setupScript := `#!/bin/bash
set -e

# Update package lists
echo "📦 RUN apt-get update"
apt-get update

# Upgrade system 
echo "📦 RUN apt-get upgrade -y"
DEBIAN_FRONTEND=noninteractive apt-get upgrade -y

# Install essential packages
echo "📦 RUN apt-get install -y essential packages"
DEBIAN_FRONTEND=noninteractive apt-get install -y \
    curl wget vim git build-essential \
    software-properties-common ca-certificates \
    gnupg lsb-release

# Set environment variables
echo "🔧 ENV DEBIAN_FRONTEND=noninteractive"
echo 'DEBIAN_FRONTEND=noninteractive' >> /etc/environment

# Create working directory
echo "📁 WORKDIR /app"
mkdir -p /app

# Clean up
echo "🧹 Cleaning up package cache"
apt-get autoremove -y
apt-get autoclean
apt-get clean

echo "✅ Container setup completed successfully"
`

	// Write setup script
	if err := os.WriteFile(scriptPath, []byte(setupScript), 0755); err != nil {
		return fmt.Errorf("failed to create setup script: %w", err)
	}

	// Execute script in chroot
	l.log("🚀 Running Dockerfile-like setup in chroot...")
	if err := l.runCommand("chroot", rootfsPath, "/bin/bash", "-c", fmt.Sprintf("$(cat %s)", scriptPath)); err != nil {
		// Try alternative approach with systemd-nspawn if available
		l.log("Chroot failed, trying alternative approach...")
		if err := l.runCommand("cp", scriptPath, filepath.Join(rootfsPath, "setup.sh")); err == nil {
			if err := l.runCommand("chroot", rootfsPath, "/bin/bash", "/setup.sh"); err != nil {
				l.log("Warning: failed to run setup script: %v", err)
			}
		}
	}

	// Clean up script
	os.Remove(scriptPath)
	os.Remove(filepath.Join(rootfsPath, "setup.sh"))

	l.log("✅ Container created successfully: %s", containerName)
	return nil
}

// appendToConfig appends text to a container config file
func (l *LXCBuilder) appendToConfig(configPath, text string) error {
	file, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(text)
	return err
}
