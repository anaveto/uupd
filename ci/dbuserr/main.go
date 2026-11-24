package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/godbus/dbus/v5"
)

func main() {
	label, prop := os.Args[1], "ActiveProfile"
	if len(os.Args) > 2 {
		prop = os.Args[2]
	}
	conn, err := dbus.SystemBus()
	if err != nil {
		fmt.Printf("%s: no bus: %v\n", label, err)
		os.Exit(1)
	}
	done := make(chan struct{})
	start := time.Now()
	go func() {
		v, err := conn.Object("org.freedesktop.UPower.PowerProfiles", "/org/freedesktop/UPower/PowerProfiles").
			GetProperty("org.freedesktop.UPower.PowerProfiles." + prop)
		var e dbus.Error
		switch {
		case err == nil:
			fmt.Printf("RESULT %s: value %v (%.1fs)\n", label, v, time.Since(start).Seconds())
		case errors.As(err, &e):
			fmt.Printf("RESULT %s: name=%s message=%q type=%T (%.1fs)\n", label, e.Name, e.Error(), err, time.Since(start).Seconds())
		default:
			fmt.Printf("RESULT %s: non-dbus error %T %v (%.1fs)\n", label, err, err, time.Since(start).Seconds())
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(150 * time.Second):
		fmt.Printf("RESULT %s: no reply in 150 s\n", label)
	}
}
