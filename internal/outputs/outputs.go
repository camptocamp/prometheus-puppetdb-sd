package outputs

import (
	"fmt"
)

// Options stores options that might be used by the different output types
type Options struct {
	Name string
	// Used by File
	FilePath string
	// Used by Kubernetes Configmap
	ConfigMapName string
	Namespace     string
}

// Output is an abstraction to the different output types
type Output interface {
	WriteOutput(data interface{}) (err error)
}

// Setup returns an output type
func Setup(options *Options) (Output, error) {
	switch options.Name {
	case "stdout":
		return &OutputStdout{}, nil
	case "file":
		return setupOutputFile(options.FilePath)
	case "configmap":
		return setupOutputK8SConfigMap(options.Namespace, options.ConfigMapName)
	case "":
		return nil, fmt.Errorf("no output defined")
	default:
		return nil, fmt.Errorf("unknown output: `%s'", options.Name)
	}
}
