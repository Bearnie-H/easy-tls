// Package plugins expands on the standard library "plugin" package to provide
// the basis for building modular, plugin-based applications.
//
// This package implements both Plugin and PluginAgent functionality for both
// Server and Client type applications.
//
// Plugins
//
// A "Plugin" or module within this package refers to a set of Go packages
// which implements a public interface matching the basic common interface
// described in the "PluginAPI" struct. This provides all of the necessary
// functionality to work with a plugin except for starting it. Each of the
// separate "ClientPluginAPI" and "ServerPluginAPI" structs provide extensions
// to the basic interface to include an "Init()" method. This Init() method
// is the standard way to start a plugin.
//
// The methods of the PluginAPI are expected to function in a common manner
// across any type of plugin, with only the Init() method being customized to
// specific plugin types.
//
// Server Plugins
//
// Server plugins are primarily a mechanism to allow building up a Server-side
// application with the easy-tls/server package, and generating the routing
// tree dynamically. The Init() function returns an array of SimpleHandler
// objects, which provide all of the necessary information to be registered
// with the incorporated HTTP Router. As such, the Init function must perform
// any internal initialization, and spawn any concurrent go-routines which may
// need to keep running after Init() returns, while also preparing the array of
// handlers to actually be returned.
//
// Once Init() returns for a ServerPlugin, only go-routines spawned during Init
// will continue to execute, with any incoming requests being passed in to the
// specified handler.
//
// Client Plugins
//
// Client plugins are substantially different from server plugins in structure.
// These are best modelled as "stand-alone" applications, with the Init() call
// equivalent to calling their "main()" function. Arguments can be provided,
// but must be entirely parsed from the interface array within Init(). Init()
// is expected to return after "forking" off a go-routine to perform the main
// body of the plugin, typically a main loop.
//
// Plugin Agents
//
// In addition to plugins and the necessary functionality to work with them,
// this package includes Agents to oversee and work with sets of plugins in a
// simple fashion. These handle the process of loading plugins, extracting the
// functions forming their public interface, starting, stopping, and logging
// the Plugins. When building a modular application with this library, the
// Agents are expected to be the primary driver of the modular functionality.
//
// The only real difference between the Server and Client agents are the Init()
// functions they look for.
//
package plugins
