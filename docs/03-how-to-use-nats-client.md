---
title: How To Use NATS Client
expires_at: never
tags: [routing-release]
---

# How To Use NATS Client

## What is it?

Gorouter is loosely coupled with user apps via the NATS message bus. They share common subjects like `router.register` and `router.unregister`; apps publish their routes while Gorouter subscribes to them.

To debug a specific scenario, developers typically need to set up applications and NATS to reproduce it within Gorouter. This is cumbersome, as it requires a full Cloud Foundry deployment, pushing apps, scaling them up, and potentially adjusting NATS or the IaaS provider to simulate network errors or other issues.

To mitigate this problem, `nats_client` was created. It has the following features:

- Subscribe to NATS subjects and stream them to the shell to see what Gorouter receives
- Publish messages to NATS subjects to inject events such as `router.register` and `router.unregister`
- Save Gorouter's current route table to a JSON file for inspection or modification
- Load a previously saved route table into Gorouter via NATS

## Where is it?

After you have deployed `routing`, you will find two files on a Gorouter VM:
- `/var/vcap/packages/routing_utils/bin/nats_client` (the actual NATS client binary)
- `/var/vcap/jobs/gorouter/bin/nats_client` (a wrapper script that calls the binary with Gorouter's config file)

## How to use it
Show the usage help by running:
```shell
/var/vcap/jobs/gorouter/bin/nats_client --help

Usage:
/var/vcap/jobs/gorouter/bin/nats_client [COMMAND]

COMMANDS:
  sub          [SUBJECT] [MESSAGE]
               (Default) Streams NATS messages from server with provided SUBJECT. Default SUBJECT is 'router.*'
               Example: /var/vcap/jobs/gorouter/bin/nats_client sub 'router.*'

  pub          [SUBJECT] [MESSAGE]
               Publish the provided message JSON to SUBJECT subscription. SUBJECT and MESSAGE are required
               Example: /var/vcap/jobs/gorouter/bin/nats_client pub router.register '{"host":"172.217.6.68","port":80,"uris":["bar.example.com"]}'

  save         <FILE>
               Save this Gorouter's route table to a JSON file.
               Example: /var/vcap/jobs/gorouter/bin/nats_client save routes.json

  load         <FILE>
               Load routes from a JSON file into this Gorouter.
               Example: /var/vcap/jobs/gorouter/bin/nats_client load routes.json
```

### Streaming NATS Messages
By default, `nats_client` will subscribe to `router.*` subjects:
```shell
/var/vcap/jobs/gorouter/bin/nats_client

Subscribing to router.*
new message with subject: router.register
{"uris":["some-app.cf.mydomain.com"],"host":"10.1.3.7","port":3000,"tags":null,"private_instance_id":"abea5c4c-4c91-4827-7156-2e9496512903"}
new message with subject: router.register
{"uris":["another-app.cf.mydomain.com","*.another-app.cf.my-domain.com"],"host":"10.1.1.73","port":8083,"tls_port":8083,"tags":null,"private_instance_id":"efcc4e10-f705-423c-6ec4-b25e9d4fa327","server_cert_domain_san":"another-app.cf.my-domain.com"}
(...)
```

### Publishing NATS Messages
You may use the `nats_client` to publish messages such as `router.register` to simulate a CF app starting up:
```shell
/var/vcap/jobs/gorouter/bin/nats_client pub router.register '{"host":"httpstat.us","tls_port":443,"server_cert_domain_san":"httpstat.us", "uris":["httpstat.us"]}'

Publishing message to router.register
Done
```

You can then test the new route and see if the backend can be reached using:
```shell
curl http://localhost:8081/200 -H "Host: httpstat.us"
200 OK%
```
(the above example assumes you have Gorouter running locally without TLS)


### Saving the Route Table to Disk
The `save` command stores the current route table as a JSON file.
```shell
/var/vcap/jobs/gorouter/bin/nats_client save routes.json

Saving route table to routes.json
Done
```
You can then view and edit the route table to your needs.

### Loading a Route Table from Disk
Once you have prepared a route table JSON file, you can load it using the `load` command
```shell
/var/vcap/jobs/gorouter/bin/nats_client load routes.json

Loading route table from routes.json
Done
```
The routes will not be loaded directly; instead, the contents of `routes.json` will be transformed into `router.register` messages and published to Gorouter via NATS in order.

> [!NOTE]
> Non-TLS routes that are not continuously refreshed will be pruned.

## When to use it
There are many scenarios where you may use `nats_client` to debug Gorouter issues:
- Debug retries of failing endpoints
- Test different kinds of backend errors (e.g. dial timeout, TLS handshake issues, app misbehaving etc.)
- Debug load balancing algorithms
- Set up large deployments with hundreds of apps and thousands of routes, without having to actually deploy all of them
- Simulate outages where large numbers of backends no longer respond (e.g. AZ outages)
- Simulate NATS outages where apps have moved elsewhere but Gorouter did not receive the proper `router.unregister` message
- etc.
