package session

import (
	"syscall"

	"github.com/godbus/dbus/v5"
)

const (
	login1Dest    = "org.freedesktop.login1"
	login1Path    = "/org/freedesktop/login1"
	login1Manager = "org.freedesktop.login1.Manager"
)

var inhibitTypes = []string{"sleep", "handle-lid-switch"}

// InhibitSuspend holds the locks it managed to take and reports the first failure.
func InhibitSuspend() (func(), error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	var fds []dbus.UnixFD
	var firstErr error
	for _, what := range inhibitTypesFor(activeDesktop(conn)) {
		var fd dbus.UnixFD
		err := conn.Object(login1Dest, login1Path).
			Call(login1Manager+".Inhibit", 0, what, "uupd", "System update in progress", "block").
			Store(&fd)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		fds = append(fds, fd)
	}
	if len(fds) == 0 {
		return nil, firstErr
	}

	return releaseInhibitors(fds), firstErr
}

func inhibitTypesFor(desktop string) []string {
	if desktop == "gamescope" {
		return []string{"handle-lid-switch"}
	}
	return inhibitTypes
}

func activeDesktop(conn *dbus.Conn) string {
	active, err := conn.Object(login1Dest, login1Path+"/seat/seat0").GetProperty("org.freedesktop.login1.Seat.ActiveSession")
	if err != nil {
		return ""
	}
	var session struct {
		ID   string
		Path dbus.ObjectPath
	}
	if err := active.Store(&session); err != nil {
		return ""
	}
	desktop, err := conn.Object(login1Dest, session.Path).GetProperty("org.freedesktop.login1.Session.Desktop")
	if err != nil {
		return ""
	}
	name, _ := desktop.Value().(string)
	return name
}

func releaseInhibitors(fds []dbus.UnixFD) func() {
	return func() {
		for _, fd := range fds {
			_ = syscall.Close(int(fd))
		}
	}
}
