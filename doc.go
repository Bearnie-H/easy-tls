// Package easytls provides a simplification for creating valid tls.Config
// structures for use with the http package.
//
// This package provides a consistent method for generating a tls.Config with
// a whitelisted set of Certificate Authorities to accept, an optional "client"
// certificate/key pair and peer validation policy. This is primarily intended
// to be used with the other packages in this project to simplify building
// simple and secure HTTPS services.
package easytls
