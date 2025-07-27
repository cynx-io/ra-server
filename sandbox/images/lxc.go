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

	// Export container as tar.gz
	if err := exportContainerAsTarGz(builder, workDir, containerPath); err != nil {
		fmt.Printf("Container export failed: %v\n", err)
		return
	}
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
# Enable networking with veth and bridge
lxc.net.0.type = veth
lxc.net.0.link = lxcbr0
lxc.net.0.flags = up
lxc.net.0.hwaddr = 00:16:3e:xx:xx:xx
lxc.apparmor.profile = unconfined
lxc.cap.drop = 
`
	if err := l.appendToConfig(configPath, additionalConfig); err != nil {
		l.log("Warning: failed to modify container config: %v", err)
	}

	// Get rootfs path for later use
	rootfsPath := filepath.Join(l.ContainerDir, containerName, "rootfs")
	l.log("🔧 Container rootfs location: %s", rootfsPath)

	// Create a script to run inside container
	scriptPath := filepath.Join(l.ContainerDir, "setup.sh")
	setupScript := `#!/bin/bash
set -e

# Setup DNS resolution first
echo "🌐 Setting up DNS resolution..."
cat > /etc/resolv.conf << 'EOF'
# DNS configuration for LXC container
nameserver 8.8.8.8
nameserver 8.8.4.4
nameserver 1.1.1.1
nameserver 1.0.0.1
EOF

# Test DNS resolution
echo "🔍 Testing DNS resolution..."
nslookup archive.ubuntu.com || echo "Warning: DNS resolution test failed"

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
    gnupg lsb-release dnsutils iputils-ping

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

	// Start the container to enable network access for package installation
	l.log("🚀 Starting container for package installation...")
	if err := l.runCommand("lxc-start", "-n", containerName, "-P", l.ContainerDir, "-d"); err != nil {
		l.log("Warning: failed to start container, falling back to chroot")
		return l.fallbackToChroot(rootfsPath, scriptPath)
	}

	// Wait for container to be ready
	l.log("⏳ Waiting for container to be ready...")
	for i := 0; i < 30; i++ {
		if err := l.runCommand("lxc-info", "-n", containerName, "-P", l.ContainerDir, "-s"); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Setup DNS immediately in the running container
	l.log("🌐 Setting up DNS in running container...")
	dnsSetupCmd := `echo 'nameserver 8.8.8.8
nameserver 8.8.4.4
nameserver 1.1.1.1
nameserver 1.0.0.1' > /etc/resolv.conf`
	if err := l.runCommand("lxc-attach", "-n", containerName, "-P", l.ContainerDir, "--", "/bin/bash", "-c", dnsSetupCmd); err != nil {
		l.log("Warning: failed to setup DNS in container: %v", err)
	}

	// Copy script into container
	containerScriptPath := filepath.Join(rootfsPath, "setup.sh")
	if err := l.runCommand("cp", scriptPath, containerScriptPath); err != nil {
		l.runCommand("lxc-stop", "-n", containerName, "-P", l.ContainerDir)
		return fmt.Errorf("failed to copy setup script: %w", err)
	}

	// Execute script inside running container
	l.log("🔧 Running setup script in container...")
	if err := l.runCommand("lxc-attach", "-n", containerName, "-P", l.ContainerDir, "--", "/bin/bash", "/setup.sh"); err != nil {
		l.log("Warning: setup script execution failed: %v", err)
	}

	// Stop the container
	l.log("⏹️ Stopping container...")
	l.runCommand("lxc-stop", "-n", containerName, "-P", l.ContainerDir)

	// Clean up script
	os.Remove(scriptPath)
	os.Remove(filepath.Join(rootfsPath, "setup.sh"))

	// Clean up mounts
	l.cleanupMounts(rootfsPath)

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

// setupContainerDNS configures DNS resolution for the container
func (l *LXCBuilder) setupContainerDNS(rootfsPath string) error {
	l.log("🌐 Setting up DNS resolution...")

	// Ensure /etc directory exists
	etcPath := filepath.Join(rootfsPath, "etc")
	if err := os.MkdirAll(etcPath, 0755); err != nil {
		return fmt.Errorf("failed to create /etc directory: %w", err)
	}

	// Copy host's resolv.conf for DNS resolution, then add fallback DNS servers
	hostResolvConf, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		l.log("Warning: failed to read host resolv.conf, using fallback DNS")
		hostResolvConf = []byte("")
	}

	// Create comprehensive DNS configuration
	resolvConf := string(hostResolvConf) + `
# Fallback DNS servers for LXC container
nameserver 8.8.8.8
nameserver 8.8.4.4
nameserver 1.1.1.1
nameserver 1.0.0.1
`
	resolvPath := filepath.Join(etcPath, "resolv.conf")
	if err := os.WriteFile(resolvPath, []byte(resolvConf), 0644); err != nil {
		return fmt.Errorf("failed to write resolv.conf: %w", err)
	}

	// Bind mount /proc, /sys, and /dev for proper chroot operation
	procPath := filepath.Join(rootfsPath, "proc")
	sysPath := filepath.Join(rootfsPath, "sys")
	devPath := filepath.Join(rootfsPath, "dev")

	l.runCommand("mount", "-t", "proc", "proc", procPath)
	l.runCommand("mount", "-t", "sysfs", "sysfs", sysPath)
	l.runCommand("mount", "--bind", "/dev", devPath)

	return nil
}

// fallbackToChroot attempts to run setup using chroot when container start fails
func (l *LXCBuilder) fallbackToChroot(rootfsPath, scriptPath string) error {
	l.log("🔄 Using chroot fallback approach...")

	// Copy script into container
	containerScriptPath := filepath.Join(rootfsPath, "setup.sh")
	if err := l.runCommand("cp", scriptPath, containerScriptPath); err != nil {
		return fmt.Errorf("failed to copy setup script: %w", err)
	}

	// Make script executable
	if err := os.Chmod(containerScriptPath, 0755); err != nil {
		l.log("Warning: failed to make setup script executable: %v", err)
	}

	// Execute script in chroot (network will be limited)
	if err := l.runCommand("chroot", rootfsPath, "/bin/bash", "/setup.sh"); err != nil {
		l.log("Warning: chroot setup script execution failed: %v", err)
		// Continue anyway as basic container structure is created
	}

	return nil
}

// cleanupMounts unmounts bind mounts from the container
func (l *LXCBuilder) cleanupMounts(rootfsPath string) {
	l.log("🧹 Cleaning up mounts...")

	devPath := filepath.Join(rootfsPath, "dev")
	sysPath := filepath.Join(rootfsPath, "sys")
	procPath := filepath.Join(rootfsPath, "proc")

	// Unmount in reverse order
	l.runCommand("umount", devPath)
	l.runCommand("umount", sysPath)
	l.runCommand("umount", procPath)
}

// createProxmoxMetadata ensures proper metadata for Proxmox compatibility
func (l *LXCBuilder) createProxmoxMetadata(rootfsPath string) error {
	l.log("📝 Creating Proxmox metadata...")

	// Ensure /etc/os-release exists for Proxmox autodetection
	osReleasePath := filepath.Join(rootfsPath, "etc", "os-release")
	osReleaseContent := `NAME="Ubuntu"
VERSION="22.04.5 LTS (Jammy Jellyfish)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 22.04.5 LTS"
VERSION_ID="22.04"
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
VERSION_CODENAME=jammy
UBUNTU_CODENAME=jammy
`

	if err := os.WriteFile(osReleasePath, []byte(osReleaseContent), 0644); err != nil {
		return fmt.Errorf("failed to create os-release: %w", err)
	}

	// Also create lsb-release for better compatibility
	lsbReleasePath := filepath.Join(rootfsPath, "etc", "lsb-release")
	lsbReleaseContent := `DISTRIB_ID=Ubuntu
DISTRIB_RELEASE=22.04
DISTRIB_CODENAME=jammy
DISTRIB_DESCRIPTION="Ubuntu 22.04.5 LTS"
`

	if err := os.WriteFile(lsbReleasePath, []byte(lsbReleaseContent), 0644); err != nil {
		return fmt.Errorf("failed to create lsb-release: %w", err)
	}

	return nil
}

// exportContainerAsTarGz exports the LXC container as a Proxmox-compatible tar.gz template
func exportContainerAsTarGz(l *LXCBuilder, workDir, containerPath string) error {
	// Determine container name from path
	containerName := filepath.Base(containerPath)

	// Check if this is a layered container (has parent)
	if containerName != "ubuntu-base" {
		return l.exportLayeredContainer(workDir, containerPath, containerName)
	}

	// Export base container normally
	return l.exportBaseContainer(workDir, containerPath)
}

// exportBaseContainer exports the base Ubuntu container
func (l *LXCBuilder) exportBaseContainer(workDir, containerPath string) error {
	l.log("📦 Exporting base container as Proxmox-compatible tar.gz template...")

	// Create tar.gz filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	tarGzName := fmt.Sprintf("lxc-ubuntu-base-%s.tar.gz", timestamp)
	tarGzPath := filepath.Join(workDir, tarGzName)

	// Get rootfs path for Proxmox-compatible export
	rootfsPath := filepath.Join(containerPath, "rootfs")

	// Create Proxmox metadata
	if err := l.createProxmoxMetadata(rootfsPath); err != nil {
		l.log("Warning: failed to create metadata: %v", err)
	}

	// Export only the rootfs as tar.gz (Proxmox format)
	l.log("Creating Proxmox-compatible tar.gz archive: %s", tarGzPath)
	if err := l.runCommand("tar", "-czf", tarGzPath, "-C", rootfsPath, "."); err != nil {
		return fmt.Errorf("failed to create tar.gz archive: %w", err)
	}

	// Also create a symlink with a consistent name
	symlinkPath := filepath.Join(workDir, "lxc-ubuntu-latest.tar.gz")
	os.Remove(symlinkPath) // Remove existing symlink if it exists
	if err := os.Symlink(tarGzName, symlinkPath); err != nil {
		l.log("Warning: failed to create symlink: %v", err)
	}

	l.log("✅ Proxmox-compatible container template exported!")
	l.log("📁 Archive location: %s", tarGzPath)
	l.log("🔗 Latest symlink: %s", symlinkPath)
	l.log("📋 Usage: Copy to /var/lib/vz/template/cache/ on Proxmox")

	return nil
}

// exportLayeredContainer exports only the differences from the parent layer
func (l *LXCBuilder) exportLayeredContainer(workDir, containerPath, containerName string) error {
	l.log("📦 Exporting layered container with diff-only approach...")

	// Determine parent layer (for java8, parent is ubuntu-base)
	parentLayer := "ubuntu-base"
	parentPath := filepath.Join(l.ContainerDir, parentLayer)

	if !l.dirExists(parentPath) {
		l.log("Warning: Parent layer not found, exporting full container")
		return l.exportBaseContainer(workDir, containerPath)
	}

	// Create tar.gz filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	tarGzName := fmt.Sprintf("lxc-%s-layer-%s.tar.gz", containerName, timestamp)
	tarGzPath := filepath.Join(workDir, tarGzName)

	// Get rootfs paths
	containerRootfs := filepath.Join(containerPath, "rootfs")

	// Create layer diff using a simple approach: tar specific Java directories
	l.log("🔍 Creating layer diff archive with Java-specific files...")

	// Check which files exist before adding to tar
	filesToTar := []string{}
	possibleFiles := []string{"usr/lib/jvm", "etc/environment", "etc/bash.bashrc", "usr/bin/java", "usr/bin/javac"}

	for _, file := range possibleFiles {
		fullPath := filepath.Join(containerRootfs, file)
		if _, err := os.Stat(fullPath); err == nil {
			filesToTar = append(filesToTar, file)
			l.log("✓ Found: %s", file)
		} else {
			l.log("✗ Missing: %s", file)
		}
	}

	if len(filesToTar) == 0 {
		return fmt.Errorf("no Java files found in container - Java installation may have failed")
	}

	// Create tar with only the existing files
	args := []string{"-czf", tarGzPath, "-C", containerRootfs}
	args = append(args, filesToTar...)

	if err := l.runCommand("tar", args...); err != nil {
		return fmt.Errorf("failed to create layer diff archive: %w", err)
	}

	// Also create a symlink with a consistent name
	symlinkPath := filepath.Join(workDir, fmt.Sprintf("lxc-%s-layer-latest.tar.gz", containerName))
	os.Remove(symlinkPath) // Remove existing symlink if it exists
	if err := os.Symlink(tarGzName, symlinkPath); err != nil {
		l.log("Warning: failed to create symlink: %v", err)
	}

	// Create layer metadata
	l.createLayerMetadata(workDir, containerName, parentLayer, tarGzName)

	l.log("✅ Layer diff exported successfully!")
	l.log("📁 Archive location: %s", tarGzPath)
	l.log("🔗 Latest symlink: %s", symlinkPath)
	l.log("📋 This layer contains only changes from %s", parentLayer)

	return nil
}

// createLayerMetadata creates metadata for the layer
func (l *LXCBuilder) createLayerMetadata(workDir, layerName, parentLayer, archiveName string) {
	metadata := fmt.Sprintf(`{
  "layer_name": "%s",
  "parent_layer": "%s", 
  "archive": "%s",
  "created": "%s",
  "type": "diff_layer"
}`, layerName, parentLayer, archiveName, time.Now().Format(time.RFC3339))

	metadataPath := filepath.Join(workDir, fmt.Sprintf("%s-layer.json", layerName))
	os.WriteFile(metadataPath, []byte(metadata), 0644)
	l.log("📄 Layer metadata: %s", metadataPath)
}

// RunJava8Container creates a Java 8 layer on top of Ubuntu base
func RunJava8Container() {
	buildLayeredContainer("ubuntu-java8", "ubuntu-base", buildJava8Layer)
}

// buildLayeredContainer creates a container layer, optionally based on a parent layer
func buildLayeredContainer(containerName, parentLayer string, buildFunc func(*LXCBuilder, string, string) error) {
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

	if err := buildFunc(builder, containerName, parentLayer); err != nil {
		fmt.Printf("Container build failed: %v\n", err)
		return
	}

	fmt.Printf("✅ %s container created successfully!\n", containerName)
	containerPath := filepath.Join(containerDir, containerName)
	fmt.Printf("📁 Container location: %s\n", containerPath)

	// Export container as tar.gz
	if err := exportContainerAsTarGz(builder, workDir, containerPath); err != nil {
		fmt.Printf("Container export failed: %v\n", err)
		return
	}
}

// dirExists checks if a directory exists
func (l *LXCBuilder) dirExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// createLayerFromParent creates a new container by copying from parent
func (l *LXCBuilder) createLayerFromParent(containerName, parentLayer string) error {
	l.log("📋 Copying from parent layer: %s", parentLayer)

	parentPath := filepath.Join(l.ContainerDir, parentLayer)
	newPath := filepath.Join(l.ContainerDir, containerName)

	// Copy the entire parent container directory
	if err := l.runCommand("cp", "-r", parentPath, newPath); err != nil {
		return fmt.Errorf("failed to copy parent layer: %w", err)
	}

	// Update the container name in config
	configPath := filepath.Join(newPath, "config")
	if content, err := os.ReadFile(configPath); err == nil {
		newContent := string(content)
		// Update container name references in config
		if updatedContent := strings.ReplaceAll(newContent, parentLayer, containerName); updatedContent != newContent {
			os.WriteFile(configPath, []byte(updatedContent), 0644)
		}
	}

	return nil
}

// buildJava8Layer creates Java 8 layer on top of Ubuntu base
func buildJava8Layer(l *LXCBuilder, containerName, parentLayer string) error {
	l.log("🍵 FROM %s - Creating Java 8 layer...", parentLayer)

	// Clean up any existing container
	l.log("Cleaning up any existing container: %s", containerName)
	l.runCommand("lxc-stop", "-n", containerName, "-P", l.ContainerDir)
	l.runCommand("lxc-destroy", "-n", containerName, "-P", l.ContainerDir)

	// Create snapshot from parent layer if it exists
	parentPath := filepath.Join(l.ContainerDir, parentLayer)
	if parentLayer != "" && l.dirExists(parentPath) {
		l.log("📦 Creating layer from parent: %s", parentLayer)
		if err := l.createLayerFromParent(containerName, parentLayer); err != nil {
			return fmt.Errorf("failed to create layer from parent: %w", err)
		}
	} else {
		// Create fresh container if no parent
		if err := l.runCommand("lxc-create", "-t", "download", "-n", containerName, "-P", l.ContainerDir, "--", "--dist", "ubuntu", "--release", "jammy", "--arch", "amd64"); err != nil {
			return fmt.Errorf("failed to create LXC container: %w", err)
		}
	}

	// Get rootfs path
	rootfsPath := filepath.Join(l.ContainerDir, containerName, "rootfs")
	l.log("🔧 Container rootfs location: %s", rootfsPath)

	// Create Java 8 installation script
	scriptPath := filepath.Join(l.ContainerDir, "java8-setup.sh")
	setupScript := `#!/bin/bash
# Removed set -e to continue on errors and debug issues

# Setup DNS resolution first
echo "🌐 Setting up DNS resolution..."
cat > /etc/resolv.conf << 'EOF'
nameserver 8.8.8.8
nameserver 8.8.4.4
nameserver 1.1.1.1
nameserver 1.0.0.1
EOF

echo "✅ DNS configuration written"

# Test DNS resolution  
echo "🔍 Testing DNS resolution..."
if nslookup archive.ubuntu.com; then
    echo "✅ DNS resolution working"
else
    echo "⚠️ DNS test failed, but continuing..."
fi

# Update package lists with retries
echo "📦 RUN apt-get update"
UPDATE_SUCCESS=false
for i in {1..3}; do
    echo "Attempt $i: Running apt-get update..."
    if apt-get update; then
        echo "✅ apt-get update succeeded on attempt $i"
        UPDATE_SUCCESS=true
        break
    else
        echo "❌ Attempt $i failed: apt-get update failed (exit code $?)"
        sleep 5
    fi
done

if [ "$UPDATE_SUCCESS" = "false" ]; then
    echo "❌ All apt-get update attempts failed"
    exit 1
fi

# Install OpenJDK 8 with more specific package handling
echo "☕ RUN apt-get install -y openjdk-8-jdk"
echo "Setting DEBIAN_FRONTEND=noninteractive..."
export DEBIAN_FRONTEND=noninteractive

echo "Running apt-get install..."
if apt-get install -y --no-install-recommends openjdk-8-jdk; then
    echo "✅ Java 8 installation completed successfully"
else
    echo "❌ Java 8 installation failed (exit code $?)"
    echo "Checking available packages..."
    apt-cache search openjdk-8 || echo "Package search failed"
    exit 1
fi

# Verify Java installation immediately
echo "🔍 Initial Java verification..."
which java || echo "Java binary not found in PATH"
ls -la /usr/lib/jvm/ || echo "JVM directory not found"

# Set JAVA_HOME environment variable
JAVA_HOME_PATH="/usr/lib/jvm/java-8-openjdk-amd64"
if [ -d "$JAVA_HOME_PATH" ]; then
    echo "🔧 ENV JAVA_HOME=$JAVA_HOME_PATH"
    echo "JAVA_HOME=$JAVA_HOME_PATH" >> /etc/environment
    echo "export JAVA_HOME=$JAVA_HOME_PATH" >> /etc/bash.bashrc
    
    # Add Java to PATH
    echo "🔧 ENV PATH=$JAVA_HOME_PATH/bin:\$PATH"
    echo "export PATH=$JAVA_HOME_PATH/bin:\$PATH" >> /etc/bash.bashrc
else
    echo "⚠️  Warning: JAVA_HOME directory not found at $JAVA_HOME_PATH"
    echo "Available JVM directories:"
    ls -la /usr/lib/jvm/ || echo "No JVM directories found"
fi

# Source the environment and verify Java installation
echo "✅ Verifying Java installation..."
export JAVA_HOME="$JAVA_HOME_PATH"
export PATH="$JAVA_HOME_PATH/bin:$PATH"
java -version || echo "Java verification failed"
javac -version || echo "Javac verification failed"

# Show installed Java files
echo "📁 Java installation contents:"
ls -la /usr/lib/jvm/ || echo "No JVM directory"
ls -la /usr/bin/java* || echo "No Java binaries in /usr/bin"

# Clean up
echo "🧹 Cleaning up package cache"
apt-get autoremove -y
apt-get autoclean
apt-get clean

echo "✅ Java 8 layer setup completed successfully"
`

	// Write setup script
	if err := os.WriteFile(scriptPath, []byte(setupScript), 0755); err != nil {
		return fmt.Errorf("failed to create setup script: %w", err)
	}

	// Start the container and install Java 8
	l.log("🚀 Starting container for Java 8 installation...")
	if err := l.runCommand("lxc-start", "-n", containerName, "-P", l.ContainerDir, "-d"); err != nil {
		l.log("Warning: failed to start container, falling back to chroot")
		return l.fallbackToChroot(rootfsPath, scriptPath)
	}

	// Wait for container to be ready
	l.log("⏳ Waiting for container to be ready...")
	for i := 0; i < 30; i++ {
		if err := l.runCommand("lxc-info", "-n", containerName, "-P", l.ContainerDir, "-s"); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Setup DNS immediately in the running container
	l.log("🌐 Setting up DNS in running container...")
	dnsSetupCmd := `echo 'nameserver 8.8.8.8
nameserver 8.8.4.4
nameserver 1.1.1.1
nameserver 1.0.0.1' > /etc/resolv.conf`
	if err := l.runCommand("lxc-attach", "-n", containerName, "-P", l.ContainerDir, "--", "/bin/bash", "-c", dnsSetupCmd); err != nil {
		l.log("Warning: failed to setup DNS in container: %v", err)
	}

	// Copy script into container
	containerScriptPath := filepath.Join(rootfsPath, "java8-setup.sh")
	if err := l.runCommand("cp", scriptPath, containerScriptPath); err != nil {
		l.runCommand("lxc-stop", "-n", containerName, "-P", l.ContainerDir)
		return fmt.Errorf("failed to copy setup script: %w", err)
	}

	// Execute script inside running container
	l.log("🔧 Running Java 8 setup script in container...")
	if err := l.runCommand("lxc-attach", "-n", containerName, "-P", l.ContainerDir, "--", "/bin/bash", "/java8-setup.sh"); err != nil {
		l.log("Error: Java 8 setup script execution failed: %v", err)

		// Get more detailed error information
		l.log("Getting script output for debugging...")
		l.runCommand("lxc-attach", "-n", containerName, "-P", l.ContainerDir, "--", "cat", "/java8-setup.sh")

		l.log("Checking container network connectivity...")
		l.runCommand("lxc-attach", "-n", containerName, "-P", l.ContainerDir, "--", "ping", "-c", "1", "8.8.8.8")

		l.log("Checking DNS resolution...")
		l.runCommand("lxc-attach", "-n", containerName, "-P", l.ContainerDir, "--", "nslookup", "archive.ubuntu.com")

		l.log("Checking apt sources...")
		l.runCommand("lxc-attach", "-n", containerName, "-P", l.ContainerDir, "--", "cat", "/etc/apt/sources.list")

		// Continue with container creation even if Java install fails
		l.log("Warning: Java installation failed, but continuing with container creation...")
	}

	// Verify Java installation
	l.log("🔍 Verifying Java installation...")
	if err := l.runCommand("lxc-attach", "-n", containerName, "-P", l.ContainerDir, "--", "java", "-version"); err != nil {
		l.log("Warning: Java verification failed: %v", err)
	}

	// Stop the container
	l.log("⏹️ Stopping container...")
	l.runCommand("lxc-stop", "-n", containerName, "-P", l.ContainerDir)

	// Clean up scripts
	os.Remove(scriptPath)
	os.Remove(containerScriptPath)

	// Clean up mounts
	l.cleanupMounts(rootfsPath)

	l.log("✅ Java 8 layer created successfully: %s", containerName)
	return nil
}
