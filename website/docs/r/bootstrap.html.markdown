---
layout: "rancher2"
page_title: "Rancher2: rancher2_bootstrap"
sidebar_current: "docs-rancher2-resource-bootstrap"
description: |-
  Provides a Rancher v2 bootstrap resource. This can be used to bootstrap rancher v2 environments and output information.
---

# rancher2\_bootstrap

Provides a Rancher v2 bootstrap resource. This can be used to bootstrap rancher v2 environments and output information.

This resource is indeed to bootstrap a rancher system doing following tasks:
- Update default admin password, provided by `password` or generating a random one.
- Set `server-url` setting, based on `api_url`.
- Set `telemetry-opt` setting.
- Create a token for admin user with concrete TTL.

It just works if `bootstrap = true` is added to the provider configuration or exporting env variable `RANCHER_BOOTSTRAP=true`. In this mode, `token_key` or `access_key` and `secret_key` can not be provided.

Rancher2 admin password could be updated setting `password` field and applying this resource again. Login to rancher2 is done using `token` first and if fails, using admin `current_password`. If admin password has been changed from other methods and terraform token is expired, `current_password` field could be especified to recover terraform configuration and reset admin password and token.

## Example Usage

```hcl
# Provider config
provider "rancher2" {
  api_url   = "https://rancher.my-domain.com"
  bootstrap = true
}

# Create a new rancher2 Bootstrap
resource "rancher2_bootstrap" "admin" {
  password = "blahblah"
  telemetry = true
}
```

## Argument Reference

The following arguments are supported:

* `current_password` - (Optional/computed/sensitive) Current password for Admin user. Just needed for recover if admin password has been changed from other resources and token is expired.
* `password` - (Optional/computed/sensitive) Password for Admin user or random generated if empty.
* `telemetry` - (Optional) Send telemetry anonymous data. Default: `false`
* `token_ttl` - (Optional) TTL (seconds) for generated admin token. Default: `0` 
* `token_update` - (Optional) Update generated admin token. Default: `false` 

## Attributes Reference

The following attributes are exported:

* `id` - (Computed) The ID of the resource.
* `token` - (Computed) Generated API token for Admin User.
* `token_id` - (Computed) Generated API token id for Admin User.
* `url` - (Computed) URL set as server-url.
* `user` - (Computed) Admin username.