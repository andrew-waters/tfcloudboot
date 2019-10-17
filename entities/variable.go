package entities

import (
	"encoding/json"
	"fmt"
)

type Variable struct {
	Name      string      `yaml:"name"`
	Type      string      `yaml:"type"`
	Value     interface{} `yaml:"value"`
	Sensitive bool        `yaml:"sensitive"`
}

// initialise variables (can be normal terraform variables of environment variables)
func initVarMap(vars []Variable) []Variable {
	for i, v := range vars {
		if v.Type == "" {
			vars[i].Type = defaultVarType
		}
	}
	return vars
}

func (v Variable) isJSON() bool {
	var x struct{}
	err := json.Unmarshal([]byte(fmt.Sprintf("%v", v.Value)), &x)
	return err == nil
}
