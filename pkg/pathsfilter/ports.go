package pathsfilter

// ChangedFile represents a single changed file.
// This is part of the core domain to avoid dependencies on adapter-specific types.
type ChangedFile struct {
	Path   string
	Status string
}

// ChangedFilesProvider is a port for an adapter that can provide a list of changed files.
type ChangedFilesProvider interface {
	GetChangedFiles(base, head string) ([]ChangedFile, error)
}

// ResultWriter is a port for an adapter that writes the result of a filtering operation.
type ResultWriter interface {
	Write(result Result) error
}
