package pathsfilter

// ChangedFile represents a single changed file.
type ChangedFile struct {
	Path   string
	Status string
}

// ChangedFilesProvider is a port for an adapter that can provide a list of changed files.
type ChangedFilesProvider interface {
	GetChangedFiles(base string, head string) ([]string, error)
}

// ResultWriter is a port for an adapter that writes the result of a filtering operation.
type ResultWriter interface {
	Write(results map[string]bool) error
}
