package provider

type Preset struct {
	Name         string
	Endpoint     string
	DefaultModel string
}

var presets = []Preset{
	{"claude", "", ""},
	{"groq", "https://api.groq.com/openai/v1", "llama-3.3-70b-versatile"},
	{"gemini", "https://generativelanguage.googleapis.com/v1beta/openai", "gemini-2.0-flash"},
	{"openai", "https://api.openai.com/v1", "gpt-4o-mini"},
	{"ollama", "http://localhost:11434/v1", "llama3"},
}

// Presets returns all known provider presets.
func Presets() []Preset {
	return presets
}

// FindPreset returns the preset for a given provider name, or nil if not found.
func FindPreset(name string) *Preset {
	for i := range presets {
		if presets[i].Name == name {
			return &presets[i]
		}
	}
	return nil
}
