---
title: High Availability & Scaling
expires_at: never
tags: [routing-release]
---

# High Availability & Scaling

The TCP Router and Routing API are stateless and horizontally scalable.
TCP Routers must be placed behind a load balancer for high availability. 
The Routing API depends on a database that can be clustered for high availability. 
For high availability, deploy multiple instances of each job, distributed across regions
of your infrastructure.
