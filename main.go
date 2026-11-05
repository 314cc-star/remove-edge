// RemoveEdge - uninstalls Microsoft Edge (Chromium) from Windows.
// Microsoft Edge WebView2 Runtime is left untouched.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"unsafe"
)

const version = "1.1"

func main() {
	// Use UTF-8 in the console
	k32 := syscall.NewLazyDLL("kernel32.dll")
	k32.NewProc("SetConsoleOutputCP").Call(65001)
	k32.NewProc("SetConsoleCP").Call(65001)

	if !isAdmin() {
		fmt.Println("Requesting administrator rights...")
		if err := relaunchAsAdmin(); err != nil {
			fmt.Println("Could not get administrator rights:", err)
			pause()
			os.Exit(1)
		}
		return // the elevated copy continues in a new window
	}

	fmt.Printf("=== RemoveEdge v%s ===\n", version)
	fmt.Println("This tool will uninstall Microsoft Edge from this PC.")
	fmt.Println("(WebView2 Runtime is NOT removed - other apps depend on it.)")
	fmt.Println()
	fmt.Print("Type Y and press Enter to continue: ")
	if !confirm() {
		fmt.Println("Cancelled.")
		pause()
		return
	}

	setups := findEdgeSetups()
	if len(setups) == 0 {
		fmt.Println("\nMicrosoft Edge was not found (it may already be uninstalled).")
		blockReinstall()
		pause()
		return
	}

	step("1/5 Closing running Edge processes")
	for _, p := range []string{"msedge.exe", "MicrosoftEdgeUpdate.exe", "identity_helper.exe"} {
		run("taskkill", "/F", "/IM", p, "/T")
	}

	step("2/5 Unlocking uninstall")
	// lets setup.exe uninstall Edge outside the EEA
	regAdd(`HKLM\SOFTWARE\WOW6432Node\Microsoft\EdgeUpdateDev`, "AllowUninstall", "REG_DWORD", "1")
	regAdd(`HKLM\SOFTWARE\Microsoft\EdgeUpdateDev`, "AllowUninstall", "REG_DWORD", "1")
	// placeholder file some setup.exe versions check for
	legacyDir := filepath.Join(os.Getenv("SystemRoot"), `SystemApps\Microsoft.MicrosoftEdge_8wekyb3d8bbwe`)
	stub := filepath.Join(legacyDir, "MicrosoftEdge.exe")
	createdStub := false
	if _, err := os.Stat(stub); os.IsNotExist(err) {
		if os.MkdirAll(legacyDir, 0755) == nil {
			if f, err := os.Create(stub); err == nil {
				f.Close()
				createdStub = true
			}
		}
	}

	step("3/5 Uninstalling Edge")
	for _, s := range setups {
		fmt.Println("  Using:", s)
		code := run(s, "--uninstall", "--system-level", "--verbose-logging", "--force-uninstall")
		// also remove any per-user install
		run(s, "--uninstall", "--verbose-logging", "--force-uninstall")
		fmt.Println("  exit code:", code)
	}
	if createdStub {
		os.Remove(stub)
	}

	step("4/5 Removing leftover shortcuts")
	for _, lnk := range []string{
		filepath.Join(os.Getenv("PUBLIC"), `Desktop\Microsoft Edge.lnk`),
		filepath.Join(os.Getenv("USERPROFILE"), `Desktop\Microsoft Edge.lnk`),
		filepath.Join(os.Getenv("ProgramData"), `Microsoft\Windows\Start Menu\Programs\Microsoft Edge.lnk`),
		filepath.Join(os.Getenv("APPDATA"), `Microsoft\Internet Explorer\Quick Launch\User Pinned\TaskBar\Microsoft Edge.lnk`),
	} {
		if os.Remove(lnk) == nil {
			fmt.Println("  Removed:", lnk)
		}
	}

	step("5/5 Blocking Edge reinstall via Windows Update")
	blockReinstall()

	fmt.Println()
	if len(findEdgeSetups()) == 0 && !edgeExeExists() {
		fmt.Println("[OK] Microsoft Edge was uninstalled. Please restart your PC.")
	} else {
		fmt.Println("[FAILED] Edge files are still present - this Edge version may block uninstall.")
		fmt.Println("  See the log at:", filepath.Join(os.TempDir(), "msedge_installer.log"), "or C:\\Windows\\Temp\\msedge_installer.log")
	}
	pause()
}

func blockReinstall() {
	regAdd(`HKLM\SOFTWARE\Microsoft\EdgeUpdate`, "DoNotUpdateToEdgeWithChromium", "REG_DWORD", "1")
	regAdd(`HKLM\SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate`, "DoNotUpdateToEdgeWithChromium", "REG_DWORD", "1")
}

// find Edge setup.exe (excludes EdgeWebView)
func findEdgeSetups() []string {
	var out []string
	for _, base := range []string{os.Getenv("ProgramFiles(x86)"), os.Getenv("ProgramFiles"), os.Getenv("LOCALAPPDATA")} {
		if base == "" {
			continue
		}
		matches, _ := filepath.Glob(filepath.Join(base, `Microsoft\Edge\Application\*\Installer\setup.exe`))
		sort.Sort(sort.Reverse(sort.StringSlice(matches))) // newest first
		out = append(out, matches...)
	}
	return out
}

func edgeExeExists() bool {
	for _, base := range []string{os.Getenv("ProgramFiles(x86)"), os.Getenv("ProgramFiles")} {
		if base == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(base, `Microsoft\Edge\Application\msedge.exe`)); err == nil {
			return true
		}
	}
	return false
}

func step(s string) { fmt.Println("\n[" + s + "]") }

func run(name string, args ...string) int {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = nil, nil
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.ExitCode()
		}
		return -1
	}
	return 0
}

func regAdd(key, name, typ, data string) {
	run("reg", "add", key, "/v", name, "/t", typ, "/d", data, "/f")
}

func confirm() bool {
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.EqualFold(strings.TrimSpace(line), "y")
}

func pause() {
	fmt.Print("\nPress Enter to exit...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}

// opening the raw disk device only succeeds with admin rights
func isAdmin() bool {
	f, err := os.Open(`\\.\PHYSICALDRIVE0`)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

func relaunchAsAdmin() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	verb, _ := syscall.UTF16PtrFromString("runas")
	file, _ := syscall.UTF16PtrFromString(exe)
	cwd, _ := syscall.UTF16PtrFromString(filepath.Dir(exe))
	shell32 := syscall.NewLazyDLL("shell32.dll")
	r, _, _ := shell32.NewProc("ShellExecuteW").Call(0,
		uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(file)), 0,
		uintptr(unsafe.Pointer(cwd)), 1)
	if r <= 32 {
		return fmt.Errorf("ShellExecute error %d (UAC prompt may have been declined)", r)
	}
	return nil
}
