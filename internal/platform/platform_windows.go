//go:build windows

package platform

import (
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

var (
	modkernel32      = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = modkernel32.NewProc("LockFileEx")
	procUnlockFileEx = modkernel32.NewProc("UnlockFileEx")
	procOpenProcess  = modkernel32.NewProc("OpenProcess")
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

func DetachProcess(cmd *exec.Cmd) {
	// CREATE_NO_WINDOW (0x08000000): prevents child process from creating/inheriting a console window
	// DETACHED_PROCESS (0x00000008): detaches from parent's console
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000008}
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
