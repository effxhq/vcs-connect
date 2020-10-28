package model

// Repository represents a source potentially containing effx.yaml files
type Repository struct {
	// CloneURL defines a target used to pull down source code.
	CloneURL string
	// Tags common to both teams and services discovered by this integration.
	Tags map[string]string
	// Annotations common to both teams and services discovered by this integration.
	Annotations map[string]string
}
