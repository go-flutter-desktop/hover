package log

import (
	"golang.org/x/sys/windows"
	"syscall"
)

// EnableColoredConsoleOutput sets flag that enables VT escape sequences
// in windows console. See link for Win API documentation.
//
// https://docs.microsoft.com/en-us/windows/console/setconsolemode
//
func EnableColoredConsoleOutput() {
	var console_handle syscall.Handle
	var err error

	console_handle, err = syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	if err != nil {
		Errorf("Error getting console handle: %v", err)
		return
	} 

	var mode uint32
	err = windows.GetConsoleMode(windows.Handle(console_handle), &mode)
	if err != nil {
		Errorf("Error getting console mode: %v", err)
		return
	}
	err = windows.SetConsoleMode(windows.Handle(console_handle), mode | windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	if err != nil {
		Errorf("Error setting console mode: %v", err)
		return
	}
}
