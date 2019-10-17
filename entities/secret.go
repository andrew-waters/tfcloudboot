package entities

type Secrets struct {
	Kind string `yaml:"kind"`
	Spec struct {
		Secrets []Secret `yaml:"secrets"`
	} `yaml:"spec"`
}

type Secret struct {
	Name  string      `yaml:"name"`
	Value interface{} `yaml:"value"`
}
