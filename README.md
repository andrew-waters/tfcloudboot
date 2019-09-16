# Terraform Boostrapper

Terraform Boostrapper is a small utility for quickly bootstrapping Terraform Cloud files for Workspaces and generating variable files through a much simpler yaml declaration.


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
  resources:
    vars:
    - name: foo
      val: bar
    - name: baz
      type: number
      val: 3
    - name: bat
      type: bool
      val: true
    env:
    - name: foobar
      sensitive: true
    - name: bazbat
```

### Build the terraform files

```bash
tfcloudboot strap -f my-workspace.yaml
```

Or you can output it to a specific directory:

```bash
tfcloudboot strap -f my-workspace.yaml -o output_dir
```

## Notes

The output will create a file with your secrets in. You should probably `.gitignore` all `*.auto.tfvars` to be sure you don't accidentally puslish them via SCM.
