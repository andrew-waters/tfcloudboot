package entities

type Notification struct {
	Name            string   `yaml:"name"`
	Enabled         bool     `yaml:"enabled"`
	DestinationType string   `yaml:"destination_type"`
	Triggers        []string `yaml:"triggers"`
	URL             string   `yam:"url`
}
