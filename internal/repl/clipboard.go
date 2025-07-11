package repl

import (
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
)

func copyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("clip")
	case "linux":
		// Try xclip first
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("no clipboard utility found (xclip or xsel)")
		}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := io.Copy(in, strings.NewReader(text)); err != nil {
		in.Close()
		return err
	}

	if err := in.Close(); err != nil {
		return err
	}

	return cmd.Wait()
}
