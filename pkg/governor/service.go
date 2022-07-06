package governor

var services = map[string][]string{}

func RegisterService(name string, methods []string) {
	services[name] = methods
}

// Service ...
type Service struct {
	Methods []string `json:"methods,omitempty"`
}
