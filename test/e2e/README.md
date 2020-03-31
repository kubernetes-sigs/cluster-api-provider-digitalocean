# E2E Test

This document is to help developers understand how for run e2e test CAPDO.

## Requirements

In order to run the e2e tests the following requirements must be met:

* Ginkgo
* Docker
* Kind v0.6.0+ (*Kind v0.7.0+ not working due upstream dependencies*)

### Environment variables

The first step to running the e2e tests is setting up the required environment variables:

| Environment variable          | Description                                                                                           |
| ----------------------------- | ----------------------------------------------------------------------------------------------------- |
| `DIGITALOCEAN_ACCESS_TOKEN`   | The DigitalOcean API V2 access token                                                                  |
| `MACHINE_TYPE`                | The DigitalOcean Droplet size                                                                         |
| `MACHINE_IMAGE`               | The DigitalOcean Image id or slug                                                                     |
| `MACHINE_SSHKEY`              | The ssh key id or fingerprint (Should already registered in the DigitalOcean Account)                 |

### Running e2e test

In the root project directory run:

```
make test-e2e
```