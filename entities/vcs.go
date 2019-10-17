package entities

type VCSRepoSpec struct {
	Identifier        string `yaml:"identifier"`
	Branch            string `yaml:"branch"`
	IngressSubmodules bool   `yaml:"ingress_submodules"`
	OauthTokenID      string `yaml:"oauth_token_id"`
}
