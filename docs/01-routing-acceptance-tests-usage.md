---
title: Routing Acceptance Test Usage
expires_at: never
tags: [routing-release,routing-acceptance-tests]
---

# Routing Acceptance Test Usage

## Running Acceptance tests

In order to run tests for this repository, you need to generate a config.json

```json
{
  "addresses": [
    "${CF_TCP_DOMAIN}"
  ],
  "api": "api.${CF_SYSTEM_DOMAIN}",
  "admin_user": "admin",
  "admin_password": "${CF_ADMIN_PASSWORD}",
  "skip_ssl_validation": true,
  "use_http": true,
  "apps_domain": "${CF_SYSTEM_DOMAIN}",
  "include_http_routes": true,
  "default_timeout": 120,
  "cf_push_timeout": 120,
  "tcp_router_group": "default-tcp",
  "tcp_apps_domain": "${CF_TCP_DOMAIN}",
  "oauth": {
    "token_endpoint": "https://uaa.${CF_SYSTEM_DOMAIN}",
    "client_name": "routing_api_client",
    "client_secret": "$(bosh_get_password_from_credhub routing_api_client)",
    "port": 443,
    "skip_ssl_validation": true
  }
}
```

## Description of Config Fields
- `addresses` - contains the IP addresses of the TCP routers and/or the load balancer. IP `10.24.14.2` is the address of the `tcp_router_z1/0` job in routing-release; update this value if your deployment uses a different address. The `addresses` property also accepts a DNS entry for the TCP router, e.g. `tcp.bosh-lite.com`.
- `admin_user` and `admin_password` - the admin credentials used to log in with the CF CLI.
- `skip_ssl_validation` - used for the cf CLI when targeting an environment.
- `include_http_routes` (optional) - a boolean used to run tests for the experimental HTTP routing endpoints of the Routing API.
- `verbose` (optional) - a boolean which allows for the `-v` flag to be passed when running the router acceptance tests errand
- `test_password` (optional) -  by default, users created during the routing acceptance tests are configured with a random name and password. If manually configured, this property enables specifying the password for the user created during the test. `test_password` performs the same function as the manifest property, `user_password`.
- `tcp_router_group` - the router group to use for creating TCP routes.
-  `tcp_apps_domain` - if the property is empty, smoke tests create a temporary shared domain and use the `addresses` field to connect to the TCP application.

