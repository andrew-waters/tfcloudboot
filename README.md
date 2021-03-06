# Terraform Bootsrapper

Terraform Boostrapper is a small utility for quickly bootstrapping Terraform Cloud files for Workspaces and generating variable files through a much simpler yaml declaration.

This utility is useful when you have a large number of workspaces with a large number of variables for those workspaces (including environment variables) by dramatically reducing the amount of boilerplate code required. Instead you can declaratively specify the configuration in a much shorter yaml file. Because you don't have enough yaml already.

## Installation

```bash
go install github.com/andrew-waters/tfcloudboot
```

## Usage

**Note:** you need a Terraform Cloud account (which you can get for free) in order to apply your terraform configuration, which is the generated output of this utility. For brevity, that is not outlined here

### Create a manifest yaml file

```yaml
# my-workspace.yaml
kind: Workspace
metadata:
  name: my-terraform-cloud-workspace
  id: my_terraform_cloud_workspace
  shortname: mtcw
  organization: YourOrganisation
spec:
  vcs_repo:
    identifier: org/repo
    branch: master
    ingress_submodules: false
    oauth_token_id: ot-XXXXXXXXXXXXXXXX
  working_directory: path/to/workspace
  auto_apply: false
  file_triggers_enabled: true
  queue_all_runs: true
  terraform_version: 0.12
  resources:
    vars:
    - name: foo
      value: bar
    - name: baz
      type: number
      value: 3
    - name: bat
      type: bool
      value: true
    env:
    - name: foobar
      sensitive: true
      value: babar
    - name: bazbat
      value: batfoo
```

You can (optionally) create a secrets yaml file - which is most useful when encrypted or otherwise omitted from your source control. Values from this file will be merged into your workspace variables and placed in `workspace.auto.tfvars`. The format for this file is:

```yaml
# secrets.yaml
kind: SecretList
spec:
  secrets:
    - name: password
      value: 2S7hprPE84dLxaEa
```

Note that both `env` and `var` resources from the workspace will be substituted by matching the secret name to the resource name.

### Build the terraform files

```bash
tfcloudboot strap -f my-workspace.yaml
```

Or you can output it to a specific directory:

```bash
tfcloudboot strap -f my-workspace.yaml -o output_dir
```

To select a secrets file which contains your secret values, use the `-s` flag to indicate the location of your secrets file:

```bash
tfcloudboot strap -f my-workspace.yaml -s my-secrets.yaml -o output_dir
```

To give your file a distinct name (defaults to `workspace.tf` and `workspace.auto.tfvars`) use the `-n` flag:

```bash
tfcloudboot strap -f my-workspace.yaml -s my-secrets.yaml -o output_dir -n my-workspace
```

After you have executed the `strap` command, you will have two new files in your output location:

```hcl
# workspace.tf
// DO NOT EDIT (this file is automatically generated)
resource "tfe_workspace" "my_terraform_cloud_workspace" {
	organization = "YourOrganisation"
	name         = "my-terraform-cloud-workspace"
}

// variable declarations:

variable "my_terraform_cloud_workspace_var_foo" {}
resource "tfe_variable" "my_terraform_cloud_workspace_var_foo" {
	workspace_id = tfe_workspace.my_terraform_cloud_workspace.id
	key          = "foo"
	value        = var.my_terraform_cloud_workspace_var_foo
	category     = "terraform"
}

variable "my_terraform_cloud_workspace_var_baz" {}
resource "tfe_variable" "my_terraform_cloud_workspace_var_baz" {
	workspace_id = tfe_workspace.my_terraform_cloud_workspace.id
	key          = "baz"
	value        = var.my_terraform_cloud_workspace_var_baz
	category     = "terraform"
}

variable "my_terraform_cloud_workspace_var_bat" {}
resource "tfe_variable" "my_terraform_cloud_workspace_var_bat" {
	workspace_id = tfe_workspace.my_terraform_cloud_workspace.id
	key          = "bat"
	value        = var.my_terraform_cloud_workspace_var_bat
	category     = "terraform"
}



// env variable declarations:

variable "my_terraform_cloud_workspace_env_foobar" {}
resource "tfe_variable" "my_terraform_cloud_workspace_env_foobar" {
	workspace_id = tfe_workspace.my_terraform_cloud_workspace.id
	key          = "foobar"
	value        = var.my_terraform_cloud_workspace_env_foobar
	category     = "env"
	sensitive    = true
}

variable "my_terraform_cloud_workspace_env_bazbat" {}
resource "tfe_variable" "my_terraform_cloud_workspace_env_bazbat" {
	workspace_id = tfe_workspace.my_terraform_cloud_workspace.id
	key          = "bazbat"
	value        = var.my_terraform_cloud_workspace_env_bazbat
	category     = "env"
}
```

```hcl
# workspace.auto.tfvars
// DO NOT EDIT (this file is automatically generated)
// variable values:

my_terraform_cloud_workspace_var_foo = "bar"
my_terraform_cloud_workspace_var_baz = 3
my_terraform_cloud_workspace_var_bat = true


// env variable values:

my_terraform_cloud_workspace_env_foobar = "babar"
my_terraform_cloud_workspace_env_bazbat = "batfoo"
```

## Notes

The output will create a file with your secrets in. You should probably `.gitignore` all `*.auto.tfvars` to be sure you don't accidentally puslish them via SCM.
