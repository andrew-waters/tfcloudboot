package entities

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"gopkg.in/yaml.v1"
)

var (
	defaultVarType = "string"

	workspaceTemplate = `// DO NOT EDIT (this file is automatically generated)
resource "tfe_workspace" "{{ .Metadata.ID }}" {
	organization 					= "{{ .Metadata.Organization }}"
	name         					= "{{ .Metadata.Name }}"
	working_directory 		= "{{ .Spec.WorkingDirectory }}"
	auto_apply 						= {{ .Spec.AutoApply }}
	file_triggers_enabled = {{ .Spec.FileTriggersEnabled }}
	queue_all_runs 				= {{ .Spec.QueueAllRuns }}
	vcs_repo {
		identifier          = "{{ .Spec.VCSRepo.Identifier }}"
		branch              = "{{ .Spec.VCSRepo.Branch }}"
		ingress_submodules  = {{ .Spec.VCSRepo.IngressSubmodules }}
		oauth_token_id      = "{{ .Spec.VCSRepo.OauthTokenID }}"
	}
	trigger_prefixes 		  = [{{ with .Spec.TriggerPrefixes }}{{$n := .}}{{ range $i, $e := . }}"{{ . }}"{{ if last $i $n }}{{ else }}, {{ end }}{{ end }}{{ end }}]
}

{{ with .Spec.Notifications }}{{ range . }}
variable "{{ $.Metadata.ID }}_var_notifications_{{ .Name }}_url" {
	type = string
}

resource "tfe_notification_configuration" "{{ .Name }}" {
  name                      = "{{ .Name }}"
  enabled                   = {{ .Enabled }}
  destination_type          = "{{ .DestinationType }}"
  triggers                  = [{{ with .Triggers }}{{$n := .}}{{ range $i, $e := . }}"{{ . }}"{{ if last $i $n }}{{ else }}, {{ end }}{{ end }}{{ end }}]
  url                       = var.{{ $.Metadata.ID }}_var_notifications_{{ .Name }}_url
  workspace_external_id     = tfe_workspace.{{ $.Metadata.ID }}.external_id
}
{{ end }}{{ end }}

// variable declarations:
{{ with .Spec.Resources.Vars }}
{{ range . }}{{ if ne .Type "map" }}variable "{{ $.Metadata.ID }}_var_{{ .Name }}" {
	type = {{ .Type }}
}{{ end }}

resource "tfe_variable" "{{ $.Metadata.ID }}_var_{{ .Name }}" {
	workspace_id = tfe_workspace.{{ $.Metadata.ID }}.id
	key          = "{{ .Name }}"
	value        = {{ if eq .Type "map" }}"{ {{ range $i, $v := .Value }}{{ range $i2, $v2 := $v }} {{ $i2 }} = \"{{ $v2 }}\" {{ end }}{{ end }} }"{{ else }}var.{{ $.Metadata.ID }}_var_{{ .Name }}{{ end }}
	category     = "terraform"{{ if .Sensitive }}
	sensitive    = true{{ end }}{{ if eq .Type "map" }}
	hcl          = true{{end}}
}

{{ end }}{{ end }}

// env variable declarations:
{{ with .Spec.Resources.Env }}
{{ range . }}variable "{{ $.Metadata.ID }}_env_{{ .Name | ToLower }}" {
	type = {{ .Type }}
}

resource "tfe_variable" "{{ $.Metadata.ID }}_env_{{ .Name | ToLower }}" {
	workspace_id = tfe_workspace.{{ $.Metadata.ID }}.id
	key          = "{{ .Name }}"
	value        = "${var.{{ $.Metadata.ID }}_env_{{ .Name | ToLower }}}"
	category     = "env"{{ if .Sensitive }}
	sensitive    = true{{ end }}
}

{{ end }}{{ end }}`

	varsTemplate = `// DO NOT EDIT (this file is automatically generated)
// variable values:
{{ with .Spec.Resources.Vars }}
{{ range . }}{{ if ne .Type "map" }}{{ $.Metadata.ID }}_var_{{ .Name }} = {{ if eq .Multiline true }}<<EOF
{{ range split .Value "\\n" }}{{ trim . }}
{{ end }}EOF
{{ else }}{{ .Value }}{{end}}{{end}}
{{ end }}{{ end }}

// env variable values:
{{ with .Spec.Resources.Env }}
{{ range . }}{{ $.Metadata.ID }}_env_{{ .Name | ToLower }} = {{ .Value }}
{{ end }}{{ end }}

// notification values:
{{ with .Spec.Resources.Notifications }}
{{ range . }}{{ .Name | ToLower }} = {{ .Value }}
{{ end }}{{ end }}`
)

// Workspace is a Terraform Workspace
type Workspace struct {
	Kind string `yaml:"kind"`

	Metadata struct {
		Name         string `yaml:"name"`
		ID           string `yaml:"id"`
		Shortname    string `yaml:"shortname"`
		Organization string `yaml:"organization"`
	}
	Spec WorkspaceSpec `yaml:"spec"`
}

type WorkspaceSpec struct {
	AutoApply           bool           `yaml:"auto_apply"`
	FileTriggersEnabled bool           `yaml:"file_triggers_enabled"`
	Notifications       []Notification `yaml:"notifications"`
	QueueAllRuns        bool           `yaml:"queue_all_runs"`
	Resources           struct {
		Vars          []Variable
		Env           []Variable
		Notifications []Variable
	} `yaml:"resources"`
	TriggerPrefixes  []string    `yaml:"trigger_prefixes"`
	VCSRepo          VCSRepoSpec `yaml:"vcs_repo"`
	WorkingDirectory string      `yaml:"working_directory"`
}

// NewWorkspace creates a new Workspace from an input yaml file and returns a pointer to it
func NewWorkspace(file string) *Workspace {

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf(err.Error())
	}

	w := Workspace{}

	// unmarshal our input file yaml to the struct
	err = yaml.Unmarshal([]byte(data), &w)
	if err != nil {
		log.Fatalf(err.Error())
	}

	w.Spec.Resources.Vars = initVarMap(w.Spec.Resources.Vars)
	w.Spec.Resources.Env = initVarMap(w.Spec.Resources.Env)

	return &w
}

// Output a workspace to destination files
func (w *Workspace) Output(outputDir string, outputName string, secretsFile string) {

	d := fmt.Sprintf("%s/%s.tf", outputDir, outputName)
	s := fmt.Sprintf("%s/%s.auto.tfvars", outputDir, outputName)

	// subsitute values
	w.substitute(secretsFile)

	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
		"last": func(x int, a interface{}) bool {
			return x == reflect.ValueOf(a).Len()-1
		},
		"split": func(x string, y string) []string {
			return strings.Split(x, y)
		},
		"trim": func(x string) string {
			return strings.TrimSpace(x)
		},
	}

	// create the Terraform stanza's
	wt, err := template.New("workspace").Funcs(funcMap).Parse(workspaceTemplate)
	if err != nil {
		log.Fatalf(err.Error())
	}

	wo := bytes.Buffer{}
	err = wt.Execute(&wo, w)
	if err != nil {
		log.Fatalf(err.Error())
	}

	err = ioutil.WriteFile(d, wo.Bytes(), 0644)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// create the secret values
	vt, err := template.New("vars").Funcs(funcMap).Parse(varsTemplate)
	if err != nil {
		log.Fatalf(err.Error())
	}

	so := bytes.Buffer{}
	err = vt.Execute(&so, w)
	if err != nil {
		log.Fatalf(err.Error())
	}

	err = ioutil.WriteFile(s, so.Bytes(), 0644)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

// substitute var/env values so they're in the struct and available to the templates
func (w *Workspace) substitute(secretsFile string) {

	s := Secrets{}

	if secretsFile != "" {
		secrets, err := ioutil.ReadFile(secretsFile)
		if err != nil {
			log.Fatalf(err.Error())
		}

		// unmarshal our secret file yaml to the secrets struct
		err = yaml.Unmarshal([]byte(secrets), &s)
		if err != nil {
			log.Fatalf(err.Error())
		}
	}

	for i, t := range w.Spec.Resources.Vars {

		if _, ok := t.Value.(int); ok {
			t.Type = "int"
		}

		if t.Value == nil {
			for _, secret := range s.Spec.Secrets {
				if secret.Name == t.Name {
					t.Value = secret.Value
					break
				}
			}
		}

		if t.Type == "string" {
			if t.Multiline {
				v := strings.ReplaceAll(fmt.Sprintf("%s", t.Value), "\n", "-!-_!")
				w.Spec.Resources.Vars[i].Value = v
			} else {
				w.Spec.Resources.Vars[i].Value = fmt.Sprintf("\"%s\"", t.Value)
			}
		}

	}

	for i, t := range w.Spec.Resources.Env {
		if t.Value == nil {
			for _, secret := range s.Spec.Secrets {
				if secret.Name == t.Name {
					t.Value = secret.Value
					break
				}
			}
		}

		if _, ok := t.Value.(int); ok {
			t.Type = "int"
		}

		if t.Type == "string" {
			if t.isJSON() {
				t.Value = strconv.Quote(fmt.Sprintf("%v", t.Value))
			} else {
				t.Value = fmt.Sprintf("\"%s\"", t.Value)
			}
			w.Spec.Resources.Env[i].Value = t.Value
		}
	}

	for _, n := range w.Spec.Notifications {
		for _, sec := range s.Spec.Secrets {
			sub := fmt.Sprintf("notifications_%s_url", n.Name)
			if sec.Name == sub {
				w.Spec.Resources.Notifications = append(w.Spec.Resources.Notifications, Variable{
					Name:  fmt.Sprintf("%s_var_notifications_%s_url", w.Metadata.ID, n.Name),
					Value: fmt.Sprintf("\"%s\"", sec.Value),
				})
			}
		}
	}

}
