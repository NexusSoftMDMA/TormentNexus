package buildinfo

// Version is set at build time via -ldflags "-X internal/buildinfo.Version=...".
// If not overridden, it falls back to the value below (keep in sync with /VERSION file).
var Version = "1.0.0-alpha.128"

// ProductName is the display name of the product.
const ProductName = "TormentNexus"
