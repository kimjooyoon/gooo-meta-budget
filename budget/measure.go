package budget

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

func MeasureCommand(phase, output string, command []string) (PhaseMeasurement, error) {
	measurement := PhaseMeasurement{Phase: phase, ExitCode: -1}
	if phase == "" || output == "" || len(command) == 0 {
		return measurement, fmt.Errorf("phase, output, and command are required")
	}
	if deniedCommand(command[0]) {
		return measurement, fmt.Errorf("privilege escalation command is not allowed: %s", command[0])
	}
	process := exec.Command(command[0], command[1:]...)
	process.Stdout = os.Stdout
	process.Stderr = os.Stderr
	start := time.Now()
	err := process.Run()
	measurement.WallMS = time.Since(start).Milliseconds()
	if measurement.WallMS < 1 {
		measurement.WallMS = 1
	}
	if process.ProcessState != nil {
		measurement.ExitCode = process.ProcessState.ExitCode()
		if usage, ok := process.ProcessState.SysUsage().(*syscall.Rusage); ok {
			measurement.PeakRSSKiB = usage.Maxrss
			if runtime.GOOS == "darwin" {
				measurement.PeakRSSKiB /= 1024
			}
		}
	}
	if writeErr := WriteJSON(output, measurement); writeErr != nil {
		return measurement, writeErr
	}
	return measurement, err
}

func deniedCommand(command string) bool {
	name := filepath.Base(command)
	switch name {
	case "sudo", "su", "doas", "pkexec", "setpriv":
		return true
	default:
		return false
	}
}

func ArtifactSize(root string) (int64, int64, error) {
	info, err := os.Stat(root)
	if err != nil {
		return 0, 0, err
	}
	if !info.IsDir() {
		return 0, 0, fmt.Errorf("artifact root is not a directory: %s", root)
	}
	var files, bytes int64
	err = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.Mode().IsRegular() {
			files++
			bytes += info.Size()
		}
		return nil
	})
	return files, bytes, err
}
