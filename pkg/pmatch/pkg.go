// Package pmatch implements image patch matching and search.
// There are different implementations of the same functionality:
// 1. Naive Go implementation - slow (due to the Go compiler being bad at optimization)
// 2. Slightly optimized Go version
// 3. Cgo version - fastest
package pmatch
