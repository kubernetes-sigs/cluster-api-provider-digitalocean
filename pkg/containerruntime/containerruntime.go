package containerruntime

// RuntimeInfo contains name and version of runtime to be used.
type RuntimeInfo struct {
	// docker, rkt, containerd, ...
	Name string `json:"name"`

	// Semantic version of the container runtime to use
	Version string `json:"version"`
}
