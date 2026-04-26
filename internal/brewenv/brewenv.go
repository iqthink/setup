package brewenv

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func Prefix() string {
	if runtime.GOARCH == "arm64" {
		return "/opt/homebrew"
	}
	return "/usr/local"
}

func BrewPath() string {
	return Prefix() + "/bin/brew"
}

func Installed() bool {
	if _, err := exec.LookPath("brew"); err == nil {
		return true
	}
	if _, err := os.Stat(BrewPath()); err == nil {
		return true
	}
	return false
}

func PackageInstalled(pkg string) bool {
	cmd := exec.Command(BrewPath(), "list", "--versions", pkg)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

func CaskInstalled(cask string) bool {
	cmd := exec.Command(BrewPath(), "list", "--cask", "--versions", cask)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

// AddToPath ensures the Homebrew bin directory is on PATH for this process.
func AddToPath() {
	binDir := Prefix() + "/bin"
	sbinDir := Prefix() + "/sbin"
	path := os.Getenv("PATH")
	parts := strings.Split(path, ":")
	has := func(p string) bool {
		for _, x := range parts {
			if x == p {
				return true
			}
		}
		return false
	}
	prefix := ""
	if !has(binDir) {
		prefix = binDir + ":"
	}
	if !has(sbinDir) {
		prefix = prefix + sbinDir + ":"
	}
	if prefix != "" {
		os.Setenv("PATH", prefix+path)
	}
}
