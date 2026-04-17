//go:build windows

package platform

import (
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

var (
	modkernel32         = syscall.NewLazyDLL("kernel32.dll")
	moduser32           = syscall.NewLazyDLL("user32.dll")
	procLockFileEx      = modkernel32.NewProc("LockFileEx")
	procUnlockFileEx    = modkernel32.NewProc("UnlockFileEx")
	procOpenProcess     = modkernel32.NewProc("OpenProcess")
	procGetConsoleWindow = modkernel32.NewProc("GetConsoleWindow")
	procShowWindow      = moduser32.NewProc("ShowWindow")
	procFreeConsole     = modkernel32.NewProc("FreeConsole")
)

const (
	lockfileExclusiveLock = 0x00000002
	processQueryInfo      = 0x00000400
)

type overlapped struct {
	Internal     uintptr
	InternalHigh uintptr
	Offset       uint32
	OffsetHigh   uint32
	HEvent       syscall.Handle
}

func FlockExclusive(f *os.File) error {
	ol := new(overlapped)
	r1, _, err := procLockFileEx.Call(
		f.Fd(),
		lockfileExclusiveLock,
		0,
		0xFFFFFFFF, 0xFFFFFFFF,
		uintptr(unsafe.Pointer(ol)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}

func FlockUnlock(f *os.File) error {
	ol := new(overlapped)
	r1, _, err := procUnlockFileEx.Call(
		f.Fd(),
		0,
		0xFFFFFFFF, 0xFFFFFFFF,
		uintptr(unsafe.Pointer(ol)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}

// HideConsoleWindow hides the console window of the current process.
// Used by hook commands (feed, punch) to prevent a visible window flash on Windows.
func HideConsoleWindow() {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd != 0 {
		procShowWindow.Call(hwnd, 7) // SW_SHOWMINNOACTIVE = 7: minimize without stealing focus
	}
}

// FreeConsole detaches the current process from its console.
// This closes the console window if no other process is attached to it.
func FreeConsole() {
	procFreeConsole.Call()
}

func DetachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x00000010, // CREATE_NEW_CONSOLE
	}
}

func IsProcessAlive(pid int) bool {
	handle, err := syscall.OpenProcess(processQueryInfo, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)

	var exitCode uint32
	err = syscall.GetExitCodeProcess(handle, &exitCode)
	if err != nil {
		return false
	}
	return exitCode == 259 // STILL_ACTIVE
}
