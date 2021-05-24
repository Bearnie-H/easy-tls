# Easy-TLS
EasyTLS is a batteries-included library for quickly and easily building HTTP(S) applications. Based heavily on the [gorilla](https://github.com/gorilla) toolkit, this library provides an additional layer of abstraction and simplifications for common use-cases.

The primary extension to the gorilla toolkit is the implementation of a dynamically loadable plugin framework. The `plugins` package provides a mechanism for loading [Plugins](https://golang.org/pkg/plugin/) into a framework, allowing you to split up monolithic applications into modular sub-services.

# Install
With a [properly configured](https://golang.org/doc/install#testing) go toolchain:

`go get -u github.com/Bearnie-H/easy-tls`

# tls.Config and TLSBundles
Setting up your application to use proper TLS settings shouldn't be difficult, and you shouldn't have to open security holes (InsecureSkipVerify=true) to get up and running.

A TLSBundle is simply a coherent set of *TLS Resources* which can be used internally to build a valid *tls.Config and get your application up and running with working secure connections.

All you need to build a TLSBundle and get up and running are 3 things:

1.  Client Certificate File (Public Key)
2.  Client Key File (Private Key)
3.  Certificate Authority Certificates (Public Keys)

The CA Certificates you provide will be used to build a whitelist. If you are verifying incoming certificates, they must be signed by one of the CA's whose certificate you allow, or else the connection will fail with a TLS error. If you aren't validating peer certificates, you can simply build a TLSBundle with a Certificate/Key pair and you're ready to go!

``` go
package main

import (
    easytls "github.com/Bearnie-H/easy-tls"
    "github.com/Bearnie-H/easy-tls/client"
    "github.com/Bearnie-H/easy-tls/server"
)

func main() {

    // Build a TLSBundle for a KeyPair for this host
    Bundle := easytls.NewTLSBundle("Certificate.cer", "Key.key")

    // Use this TLSBundle in any of the client, server or proxy packages
    // to run with TLS enabled, it's that easy!

    // Configure a new HTTPS server on the default address ( :8080 )
    Server, err := server.NewServerHTTPS(Bundle)
    if err != nil {
        panic(err)
    }

    // Configure a new HTTPS client
    Client, err := client.NewClientHTTPS(Bundle)
    if err != nil {
        panic(err)
    }

    // ...
}
```

# Client
The `client` package provides an extension to the `*http.Client` of the standard library. This mainly consists of two extensions:

1. Enabling TLS with a TLSBundle, along with toggling TLS on/off safely.
2. Exposing the full set of HTTP methods (Except CONNECT) with optional context.Context values.

The first extension was the original motivator of the `client` package, allowing for simpler and more consistent building and setup of HTTPS Client applications such as crawlers.

The second extension simply provides a consistent mechanism for creating and sending any standard HTTP request without needing to split the process into two steps of building the request then sending it. This functionality comes in very handy when building non-interactive clients to work with RESTful Servers.

## Example GET Request
```go
package main

import (
    easytls "github.com/Bearnie-H/easy-tls"
    "github.com/Bearnie-H/easy-tls/client"
)

func main() {

    // Build a TLSBundle for the client.
    // We don't need a certificate/key pair, since we won't contact servers which
    // require client certificates.
    Bundle := easytls.NewTLSBundle("","")

    // Build a new HTTPS enabled client.
    Client, err := client.NewClientHTTPS(Bundle)
    if err != nil {
        panic(err)
    }

    resp, err := Client.Get("https://www.github.com/Bearnie-H/easy-tls", nil)
    if err != nil {
        panic(err)
    }

    // Do something with resp...
}
```

# Server
The `server` package implements extensions to both the standard library `*http.Server`, as well as the [gorilla/mux](https://www.github.com/gorilla/mux) `*mux.Router`.

Similar to the `client` package, the main extension to the standard `*http.Server` is in allowing for configuring the TLS settings of the server with a TLSBundle. In addition, the interface used to start and stop the server is slightly simplified from the standard library, not requiring different calls depending on HTTP/HTTPS operation.

The primary extension to the `*mux.Router`, involves the `SimpleHandler` type. The `*mux.Router` provides a powerful and easy API for adding routes and route matching to an `*http.Server`, but it can become tedious or complicated to explicitly register every route. This can become impossible if the set of routes to be registered is not fully defined at compile-time (See the Plugins section for a use-case). As such, this package provides a simple mechanism to register an arbitrary set of routes, with corresponding handlers during run-time. As a side benefit, this mechanism also simplifies and provides a consistent way to register any route, even ones which are fully known at compile-time.

## File Server
The `server` package provides the extensions for generic HTTP(S) Server operations, while the `fileserver` package provides a set of SimpleHandlers to allow serving static files from a specified directory tree.

### HTTP File Server Example
``` go
package main

import (
    "fmt"

    "github.com/Bearnie-H/easy-tls/server"
    "github.com/Bearnie-H/easy-tls/server/fileserver"
)

const (
    BaseHRef string = "/"
    ServeBase string = "./"
)

func main() {

    // Create a new HTTP Server on the default address
    Server := server.NewServerHTTP()

    // Add the handlers to implement a File-Server
    // This will serve the current directory on the URLBase, hiding any normally hidden files,
    // and logging any errors to the Server's logger.
    Server.AddHandlers(Server.Router(), fileserver.Handlers(BaseHRef, ServeBase, false, Server.Logger()))

    // Start the server!
    if err := Server.ListenAndServe(); err != nil {
        fmt.Print(err)
    }
}
```

This way you can get a basic file-server up and running in as much time as it takes to copy this example and compile it.

## Single Page Applications
Serving a Single Page Application is such a common use-case the `server` package provides everything you need in a single function:

``` go
package main

import (
    "fmt"

    "github.com/Bearnie-H/easy-tls/server"
)

const (
    BaseHRef string = "/ui/"
    IndexPath string = "./web-app/"
)

func main() {

    Server := server.NewServerHTTP()

    Server.RegisterSPAHandler(BaseHRef, IndexPath)

    if err := Server.ListenAndServe(); err != nil {
        fmt.Println(err)
    }
}
```

This way you can spend the time working on building a beautiful application, and not worrying about how to get it served.

# Proxy
The `proxy` package contains everything needed to extend a `SimpleServer` into being a proxy server, forwarding requests based on a provided set of rules. This proxy server can have different TLS settings applied to the incoming and outgoing traffic, allowing for either TLS termination, or granting access to a TLS-secured channel.

The Rules defining the proxy settings are simplistic but powerful enough for many use-cases. These rules allow translating an incoming URL Prefix to a destination host:port, along with manipulating the matched prefix to anything desired. If these rules are too limiting, any function which satisfies the `ReverseProxyRouterFunc` type can be provided, allowing arbitrary routing rules.

A powerful and handy use-case is to build out a simple proxy server to centralize access to a number of LAN services. Maybe you have a number of web services on a corporate LAN, and don't want to keep track of lots of different hosts to connect to.

## Live Editable Proxy Server
``` go
package main

import (
    "fmt"

    "github.com/Bearnie-H/easy-tls/server"
    "github.com/Bearnie-H/easy-tls/proxy"
)

func main() {

    // Create a new Proxy server, and start listening right away.
    if err := proxy.ConfigureReverseProxy(
        nil, // Create and use a default Server
        nil, // Create and use a default Client
        nil, // Create and use the default logger of the Server
        proxy.LiveFileRouter("Proxy.rules"), // Read proxy rules live from the file "Proxy.rules"
        "/", // Serve the proxy with the base URL, proxying ANYTHING
        ).ListenAndServe(); err != nil {
        fmt.Print(err)
    }
}
```

# Plugins
The `plugins` package is the biggest novel development of this entire library. This provides everything necessary to build a framework application, which can have [plugins](https://golang.org/pkg/plugin/) loaded into it at run-time, allowing applications to be built gradually from small, self-contained modules.

## Plugin
A *plugin* within the context of this library, is anything that has the following properties (satisifes the interface):

1. Is a self-contained piece of code, containing a `main` package (and potentially others), built with the `go build -buildmode=plugin` flag.
2. Contains the following exported symbols: `Init()`, `Version()`, `Status()`, `Name()`, `Stop()`, satisfying the type-requirements described in **plugins/plugin.go** and one of **plugins/client-plugin.go**, **plugins/generic-plugin.go** or **plugins/server-plugin.go**.

The `Version()`, `Status()`, `Name()`, and `Stop()` symbols have common function sigatures across any type of plugin, with the `Init()` symbol expected to vary depending on the exact nature of the plugin itself.

For example, in a **Server** plugin, the `Init()` function must return the set of http.Handlers (server.SimpleHandlers to be exact), which the plugin is providing to the framework to be registered with the server. In contrast, the `Init()` function of a **Client** plugin must only return any errors which occur during startup, as well as *forking* a go-routine to continue executing the main loop of the plugin until the `Stop()` function is called.

To create new plugins for an application, see the included `NewEasyTLSPlugin.sh` administrative script to help copy and prepare one of the plugin templates for use.

## Plugin Agents
A Plugin Agent within the conext of this library is the thing which manages plugins. This involves loading the Shared Object files, extracting the necessary symbols, starting/stopping them, logging their output, and anything else related to the meta-functionality required to let the Plugin logic execute.

# Logging
This library is intended to be exactly that, a library used to build applications, without being a complete application itself. As such, the Client, Server, Plugins, and PluginAgents all allow for injecting of a Logger and will only write to such a logger.

Every type exports at least the `Logger()` and `SetLogger()` functions as standard getters and setters, to help you inject or share logging between components.

## Header
The `header` package provides a very handy feature I've not seen anywhere else; Marshalling and Unmarshalling Go structs into and out of http.Headers, and a corresponding struct tag.

Passing data or parameters between services via HTTP Headers is a relatively common way to attach *out-of-band* data to an HTTP Request/Response without resorting to more complex solutions like multipart requests or responses. The standard library can become tiresome and repetative to encode the fields of a struct or record into an HTTP Header, not to mention the added complexity of guaranteeing a name convention. The problem is even more complex when it comes to parsing an HTTP header into a struct. Any non-string type requires some sort of conversion and corresponding error handling.

This package provides a straightforward and clean mechanism for performing these encoding or decoding operations on any 'stringifiable' type. This includes all numeric, bool, string types, as well as arrays and slices of these, as well as member structs consisting of these. Any unconvertable type are simply ignored by the encoder or decoder.

### Example Encoding and Decoding
``` go
package main

import (
    "fmt"

    "github.com/Bearnie-H/easy-tls/header"
)

type ExampleStruct struct {
    AnInt int       `easytls:"Single-Int"` // Use the struct tag name instead of the field name
    AFloat float32  `easytls:"-"` // Ignore this value during encoding/decoding
    AString string
    ABool bool
    Ints []int
}

func main() {

    // Create a struct to be encoded
    X := ExampleStruct{
        AnInt: 42,
        AFloat: 3.1415,
        AString: "Hello",
        ABool: false,
        Ints: []int{2,3},
    }

    // Attempt to encode the struct, returning an http.Header containing the encoded values
    H, err := header.DefaultEncode(X)
    if err != nil {
        panic(err)
    }

    // Attempt to decode the header into a new struct of the same type
    var X2 ExampleStruct
    if err := header.DefaultDecode(H, &X2); err != nil {
        panic(err)
    }

    // Print out the before and after values, to demonstrate the expected equivalence.
    fmt.Printf("Original: %#v\n", X)
    fmt.Printf("After Conversion: %#v\n", X2)
}
```

# License
Standard MIT License. See LICENSE file for details.