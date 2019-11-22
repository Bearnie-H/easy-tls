package main

import easytls "github.com/Bearnie-H/easy-tls"

// Main is the top-level ACTION performed by this plugin.  Returning or exiting from main is equivalent to stopping the plugin.
//
// Check the status of the "KillChannel" channel at routine times during the execution of the main action, to better allow this plugin to exit in a timely fashion.
//
// Set the StatusMonitor string and error values before and after important states, as this represents the logging information to be tracked.
func Main(Client *easytls.SimpleClient) {

}
