package log

import (
	"golang.org/x/sys/windows"
	"syscall"
	"log"
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
		log.Println("Error getting console handle:", err)
		return
	} 

	var mode uint32
	err = windows.GetConsoleMode(windows.Handle(console_handle), &mode)
	if err != nil {
		log.Println("Error getting console mode:", err)
		return
	}
	err = windows.SetConsoleMode(windows.Handle(console_handle), mode | windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	if err != nil {
		log.Println("Error setting console mode:", err)
		return
	}
}
