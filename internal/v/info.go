package v

// Info contains metadata about the current version of the application.
type Info struct {
	Version string
	Commit  string
	Date    string
	OS      string
	Arch    string
}
