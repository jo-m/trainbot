// Package pmatch implements image patch matching and search.
// For most functionality, there are three different implementations:
//
// 1. Naive Go implementation - rather slow, but hopefully correct
// 2. Slightly optimized Go version
// 3. Cgo version - fastest
// 4. Vulkan - even faster
package pmatch
