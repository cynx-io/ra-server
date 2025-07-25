package images

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ImageBuilder handles VM image creation
type ImageBuilder struct {
	WorkDir   string
	OutputDir string
	LogFile   *os.File
}

// NewImageBuilder creates a new image builder instance
func NewImageBuilder(workDir, outputDir string) *ImageBuilder {
	logFile, err := os.Create(filepath.Join(workDir, fmt.Sprintf("build-%d.log", time.Now().Unix())))
	if err != nil {
		fmt.Printf("Warning: Failed to create log file: %v\n", err)
	}

	return &ImageBuilder{
		WorkDir:   workDir,
		OutputDir: outputDir,
		LogFile:   logFile,
	}
}

// Close cleans up resources
func (b *ImageBuilder) Close() {
	if b.LogFile != nil {
		b.LogFile.Close()
	}
}

// log writes a message to both stdout and log file
func (b *ImageBuilder) log(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf("[%s] %s", timestamp, fmt.Sprintf(format, args...))

	fmt.Println(message)

	if b.LogFile != nil {
		b.LogFile.WriteString(message + "\n")
		b.LogFile.Sync()
	}
}

// runCommand executes a command and captures output
func (b *ImageBuilder) runCommand(name string, args ...string) error {
	b.log("Running: %s %s", name, strings.Join(args, " "))

	cmd := exec.Command(name, args...)

	// Add PACKER_LOG=1 for packer commands
	if name == "packer" {
		cmd.Env = append(os.Environ(), "PACKER_LOG=1")
	}

	// Capture output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Read and log output in real-time
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				output := string(buf[:n])
				fmt.Print(output)
				if b.LogFile != nil {
					b.LogFile.WriteString(output)
				}
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				output := string(buf[:n])
				fmt.Fprintf(os.Stderr, "%s", output)
				if b.LogFile != nil {
					b.LogFile.WriteString("STDERR: " + output)
				}
			}
			if err != nil {
				break
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		return err
	}

	if b.LogFile != nil {
		b.LogFile.Sync()
	}

	return nil
}

// checkPrerequisites verifies required tools are available
func checkPrerequisites() error {
	requiredTools := []string{"packer", "qemu-system-x86_64"}

	for _, tool := range requiredTools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("required tool not found: %s", tool)
		}
	}

	return nil
}

func RunUbuntu() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get current directory: %v\n", err)
		return
	}

	workDir := filepath.Join(pwd, "sandbox", "build", "ubuntu")
	outputDir := filepath.Join(workDir, "output")

	if err := os.MkdirAll(workDir, 0755); err != nil {
		fmt.Printf("Failed to create work directory: %v\n", err)
		return
	}

	fmt.Printf("🏗️  Work directory: %s\n", workDir)
	fmt.Printf("📁 Output directory: %s\n", outputDir)

	builder := NewImageBuilder(workDir, outputDir)
	defer builder.Close()

	if err := checkPrerequisites(); err != nil {
		fmt.Printf("Prerequisites check failed: %v\n", err)
		return
	}

	if err := createUbuntuImage(builder); err != nil {
		fmt.Printf("Build failed: %v\n", err)
		return
	}

	fmt.Println("✅ Ubuntu image created successfully!")
	imagePath := filepath.Join(outputDir, "ubuntu-base.qcow2")
	fmt.Printf("📁 Image location: %s\n", imagePath)
}

func RunJava8Layer() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get current directory: %v\n", err)
		return
	}

	workDir := filepath.Join(pwd, "sandbox", "build", "java8")
	outputDir := filepath.Join(workDir, "output")
	baseImagePath := filepath.Join(pwd, "sandbox", "build", "ubuntu", "output", "ubuntu-base.qcow2")

	// Check if base image exists
	if _, err := os.Stat(baseImagePath); os.IsNotExist(err) {
		fmt.Printf("❌ Base image not found: %s\n", baseImagePath)
		fmt.Println("Please run RunUbuntu() first to create the base image")
		return
	}

	if err := os.MkdirAll(workDir, 0755); err != nil {
		fmt.Printf("Failed to create work directory: %v\n", err)
		return
	}

	fmt.Printf("🏗️  Work directory: %s\n", workDir)
	fmt.Printf("📁 Output directory: %s\n", outputDir)
	fmt.Printf("🔧 Base image: %s\n", baseImagePath)

	builder := NewImageBuilder(workDir, outputDir)
	defer builder.Close()

	if err := checkPrerequisites(); err != nil {
		fmt.Printf("Prerequisites check failed: %v\n", err)
		return
	}

	if err := createJava8LayeredImage(builder, baseImagePath); err != nil {
		fmt.Printf("Java 8 layer build failed: %v\n", err)
		return
	}

	fmt.Println("✅ Java 8 layer added successfully!")
	imagePath := filepath.Join(outputDir, "ubuntu-java8.qcow2")
	fmt.Printf("📁 Final image location: %s\n", imagePath)
}

func createUbuntuImage(b *ImageBuilder) error {
	b.log("Creating ubuntu base image...")

	// Clean up output directory if it exists
	if _, err := os.Stat(b.OutputDir); err == nil {
		b.log("Removing existing output directory: %s", b.OutputDir)
		if err := os.RemoveAll(b.OutputDir); err != nil {
			return fmt.Errorf("failed to remove existing output directory: %w", err)
		}
	}

	// Create minimal cloud-init for user setup first
	if err := createMinimalCloudInit(b.WorkDir); err != nil {
		return fmt.Errorf("failed to create cloud-init: %w", err)
	}

	// Create simple Packer config
	packerConfig := `packer {
  required_plugins {
    qemu = {
      version = "~> 1"
      source  = "github.com/hashicorp/qemu"
    }
  }
}

source "qemu" "ubuntu" {
  disk_image       = true
  iso_url          = "https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.img"
  iso_checksum     = "none"
  output_directory = "output"
  vm_name          = "ubuntu-base.qcow2"
  format           = "qcow2"
  disk_size        = "20G"
  memory           = 1024
  cpus             = 2
  accelerator      = "kvm"
  net_device       = "virtio-net"
  disk_interface   = "virtio"
  headless         = true
  ssh_username     = "ubuntu"
  ssh_password     = "ubuntu"
  ssh_timeout      = "30s"
  shutdown_command = "sudo shutdown -P now"
  cd_files         = ["user-data", "meta-data"]
  cd_label         = "cidata"
}

build {
  sources = ["source.qemu.ubuntu"]
}`

	packerFile := filepath.Join(b.WorkDir, "simple.pkr.hcl")
	if err := os.WriteFile(packerFile, []byte(packerConfig), 0644); err != nil {
		return fmt.Errorf("failed to write packer config: %w", err)
	}

	// Run packer build
	if err := b.runSimplePackerBuild(); err != nil {
		return fmt.Errorf("packer build failed: %w", err)
	}

	return nil
}

func (b *ImageBuilder) runSimplePackerBuild() error {
	b.log("Installing Packer plugins...")

	packerFile := filepath.Join(b.WorkDir, "simple.pkr.hcl")

	// Change to work directory for relative paths
	oldDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	if err := os.Chdir(b.WorkDir); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			b.log("Warning: failed to restore directory: %v", err)
		}
	}()

	// First, initialize plugins
	if err := b.runCommand("packer", "init", "."); err != nil {
		return fmt.Errorf("failed to initialize Packer plugins: %w", err)
	}

	b.log("Starting Packer build...")
	return b.runCommand("packer", "build", packerFile)
}

func createMinimalCloudInit(workDir string) error {
	// Create user-data for basic ubuntu user
	userData := `#cloud-config
users:
  - name: ubuntu
    plain_text_passwd: ubuntu
    sudo: ALL=(ALL) NOPASSWD:ALL
    lock_passwd: false
    shell: /bin/bash
ssh_pwauth: true
`

	userDataPath := filepath.Join(workDir, "user-data")
	if err := os.WriteFile(userDataPath, []byte(userData), 0644); err != nil {
		return fmt.Errorf("failed to write user-data: %w", err)
	}

	// Create meta-data
	metaData := `instance-id: ubuntu-base
local-hostname: ubuntu-base
`

	metaDataPath := filepath.Join(workDir, "meta-data")
	if err := os.WriteFile(metaDataPath, []byte(metaData), 0644); err != nil {
		return fmt.Errorf("failed to write meta-data: %w", err)
	}

	return nil
}

func createJava8LayeredImage(b *ImageBuilder, baseImagePath string) error {
	b.log("Creating Java 8 layered image from base: %s", baseImagePath)

	// Clean up output directory if it exists
	if _, err := os.Stat(b.OutputDir); err == nil {
		b.log("Removing existing output directory: %s", b.OutputDir)
		if err := os.RemoveAll(b.OutputDir); err != nil {
			return fmt.Errorf("failed to remove existing output directory: %w", err)
		}
	}

	// Create Packer config for Java 8 layer
	packerConfig := fmt.Sprintf(`packer {
  required_plugins {
    qemu = {
      version = "~> 1"
      source  = "github.com/hashicorp/qemu"
    }
  }
}

source "qemu" "java8" {
  disk_image       = true
  iso_url          = "%s"
  iso_checksum     = "none"
  output_directory = "output"
  vm_name          = "ubuntu-java8.qcow2"
  format           = "qcow2"
  disk_size        = "20G"
  memory           = 1024
  cpus             = 2
  accelerator      = "kvm"
  net_device       = "virtio-net"
  disk_interface   = "virtio"
  headless         = true
  ssh_username     = "ubuntu"
  ssh_password     = "ubuntu"
  ssh_timeout      = "30s"
  shutdown_command = "sudo shutdown -P now"
}

build {
  sources = ["source.qemu.java8"]
  
  provisioner "shell" {
    inline = [
      "sudo apt-get update",
      "sudo apt-get install -y openjdk-8-jdk",
      "java -version",
      "echo 'JAVA_HOME=/usr/lib/jvm/java-8-openjdk-amd64' | sudo tee -a /etc/environment",
      "echo 'PATH=$PATH:$JAVA_HOME/bin' | sudo tee -a /etc/environment"
    ]
  }
}`, baseImagePath)

	packerFile := filepath.Join(b.WorkDir, "java8.pkr.hcl")
	if err := os.WriteFile(packerFile, []byte(packerConfig), 0644); err != nil {
		return fmt.Errorf("failed to write packer config: %w", err)
	}

	// Run packer build
	if err := b.runJava8PackerBuild(); err != nil {
		return fmt.Errorf("packer build failed: %w", err)
	}

	return nil
}

func (b *ImageBuilder) runJava8PackerBuild() error {
	b.log("Installing Packer plugins...")

	packerFile := filepath.Join(b.WorkDir, "java8.pkr.hcl")

	// Change to work directory for relative paths
	oldDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	if err := os.Chdir(b.WorkDir); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			b.log("Warning: failed to restore directory: %v", err)
		}
	}()

	// First, initialize plugins
	if err := b.runCommand("packer", "init", "."); err != nil {
		return fmt.Errorf("failed to initialize Packer plugins: %w", err)
	}

	b.log("Starting Packer build for Java 8 layer...")
	return b.runCommand("packer", "build", packerFile)
}
