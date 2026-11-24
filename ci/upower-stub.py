#!/usr/bin/env python3
"""Owns org.freedesktop.UPower on the system bus while upower.service is stopped.
OnBattery reads true; DisplayDevice Percentage reads the real BAT1 capacity.
Only Properties.Get and GetAll are served. Run as root; SIGTERM ends it."""
import signal
from gi.repository import Gio, GLib

XML = """<node>
 <interface name="org.freedesktop.UPower"><property name="OnBattery" type="b" access="read"/></interface>
 <interface name="org.freedesktop.UPower.Device"><property name="Percentage" type="d" access="read"/></interface>
</node>"""
info = Gio.DBusNodeInfo.new_for_xml(XML)


def capacity():
    return 80.0


PROPS = {
    "/org/freedesktop/UPower": ("org.freedesktop.UPower", lambda: {"OnBattery": GLib.Variant("b", True)}),
    "/org/freedesktop/UPower/devices/DisplayDevice": (
        "org.freedesktop.UPower.Device", lambda: {"Percentage": GLib.Variant("d", capacity())}),
}


def method_call(conn, sender, path, iface, method, params, inv):
    want, values = PROPS[path][0], PROPS[path][1]()
    print(f"call {sender} {path} {iface}.{method} {params}", flush=True)
    if iface == "org.freedesktop.DBus.Properties" and method == "Get":
        i, prop = params.unpack()
        if i == want and prop in values:
            inv.return_value(GLib.Variant("(v)", (values[prop],)))
            return
    if iface == "org.freedesktop.DBus.Properties" and method == "GetAll" and params.unpack()[0] == want:
        inv.return_value(GLib.Variant("(a{sv})", (values,)))
        return
    inv.return_dbus_error("org.freedesktop.DBus.Error.NotSupported", "upower test stub")


def bus_acquired(conn, name):
    conn.register_object("/org/freedesktop/UPower", info.interfaces[0], method_call, None, None)
    conn.register_object("/org/freedesktop/UPower/devices/DisplayDevice", info.interfaces[1], method_call, None, None)


loop = GLib.MainLoop()
Gio.bus_own_name(Gio.BusType.SYSTEM, "org.freedesktop.UPower", Gio.BusNameOwnerFlags.NONE, bus_acquired,
                 lambda c, n: print("stub owns", n, flush=True),
                 lambda c, n: (print("stub lost", n, flush=True), loop.quit()))
GLib.unix_signal_add(GLib.PRIORITY_DEFAULT, signal.SIGTERM, loop.quit)
loop.run()
