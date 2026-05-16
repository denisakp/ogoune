// Package port defines the interface contracts (ports) for the application's
// adapters. It centralizes all repository, scheduler, notifier, and monitoring
// interfaces that form the boundaries between layers.
//
// This package imports only the domain package and the standard library.
// Implementations (adapters) live in their respective packages and satisfy
// these interfaces via compile-time checks.
package port
