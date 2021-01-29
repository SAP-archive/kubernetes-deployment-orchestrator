package renderer

type Entry struct {
	Deprecated bool     `yaml:"deprecated,omitempty"`
	URLs       []string `yaml:"urls,omitempty"`
}

type Index struct {
	Entries map[string][]Entry `yaml:"entries,omitempty"`
}
