---
title: How To Limit Trusted CAs for Gorouter
expires_at: never
tags: [routing-release]
---

# How to Limit Trusted CAs for Gorouter

This doc is for operators who want to use the "only trust client CA certs" feature for Gorouter to limit the CA certs that Gorouter trusts.

## Version
This feature is available in [0.210.0](https://github.com/cloudfoundry/routing-release/releases/tag/0.210.0)

## Context
Operators already had the ability to add custom CAs to Gorouter using `router.ca_certs`, but they did not have the ability to prevent Gorouter from trusting the default CAs provided with the stemcell.

## Feature Description 

This feature introduced two BOSH properties:

* `router.client_ca_certs` (`optional; default: ""`) allows operators to specify CA certs for Gorouter to trust for client requests.
* `router.only_trust_client_ca_certs` (`default: false`) allows the operator to decide if Gorouter should _only_ trust the above CA certs, or concatenate them with those in `router.ca_certs` and those provided by the stemcell.
  * When `true`, only the certs configured in `router.client_ca_certs` are loaded as trusted client certs
  * When `false`, all the certs in `router.ca_certs`, `router.client_ca_certs`, plus the local system store are trusted client certificates.  **This maintains backward compatibility.**
  
## Example Usage

These examples assume that the load balancer is not terminating TLS. 

### Scenario 1: `only_trust_client_ca_certs: false`

With `only_trust_client_ca_certs: false`, all the certs in `router.ca_certs`, `router.client_ca_certs`, plus the local system store are trusted client certificates. This is the backwards compatible option.

```
router:
  ca_certs:
    - a-cert-named-apple
  client_ca_certs: |
   a-cert-named-cucumber
  only_trust_client_ca_certs: false
  client_cert_validation: require
```

```
# Using cert in ca_certs
curl --cert apple.crt --key apple.key https://GOROUTER_IP -H "Host: dora.example.com"
# OK

# Using cert in client_ca_certs
curl --cert cucumber.crt --key cucumber.key https://dora.example.com/
# OK

# Using cert in the local store
curl --cert some-stemcell-trusted-cert.crt --key some-stemcell-trusted-cert.key https://dora.example.com/
# OK

# Using other cert
curl --cert melon.crt --key melon.key https://dora.example.com/
# FAIL
```

### Scenario 2: `only_trust_client_ca_certs: true`

With `only_trust_client_ca_certs: true`, _only_ the certs configured in `router.client_ca_certs` are loaded as trusted client certs.

```
router:
  ca_certs:
   - a-cert-named-apple
  client_ca_certs: |
   a-cert-named-cucumber
  only_trust_client_ca_certs: true
  client_cert_validation: require
```

```
# Using cert in ca_certs
curl --cert apple.crt --key apple.key https://dora.example.com/
# FAIL

# Using cert in client_ca_certs
curl --cert cucumber.crt --key cucumber.key https://dora.example.com/
# OK

# Using cert in the local store
curl --cert some-stemcell-trusted-cert.crt --key some-stemcell-trusted-cert.key https://dora.example.com/
# FAIL

# Using other cert
curl --cert melon.crt --key melon.key https://dora.example.com/
# FAIL
```


