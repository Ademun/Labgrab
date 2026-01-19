package config

type PollingServiceConfig struct {
	ParserConfig ParserConfig `yaml:"parser"`
}

type ParserConfig struct {
	NumberRegexpPattern     string `yaml:"number_pattern"`
	AuditoriumRegexpPattern string `yaml:"auditorium_pattern"`
	SpotRegexpPattern       string `yaml:"spot_pattern"`
	TopicRegexpPattern      string `yaml:"topic_pattern"`

	NamePrefix string `yaml:"name_prefix"`

	Timezone string `yaml:"timezone"`

	TopicMap    map[string]string `yaml:"topic_map"`
	TypeMap     map[string]string `yaml:"type_map"`
	DefaultType string            `yaml:"default_type"`
}
