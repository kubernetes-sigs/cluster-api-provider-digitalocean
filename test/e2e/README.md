# E2E Test

This document is to help developers understand how to run e2e tests for CAPDO.

## Requirements

In order to run the e2e tests the following requirements must be met:

* [Ginkgo](https://github.com/onsi/ginkgo)
* [Docker](https://www.docker.com/)
* [Kind](https://github.com/kubernetes-sigs/kind) v0.7.0+

### Environment variables

The first step to running the e2e tests is setting up the required environment variables:

| Environment variable              | Description                                                                                           |
| --------------------------------- | ----------------------------------------------------------------------------------------------------- |
| `DIGITALOCEAN_ACCESS_TOKEN`       | The DigitalOcean API V2 access token                                                                  |
| `DO_CONTROL_PLANE_MACHINE_IMAGE`  | The DigitalOcean Image id or slug                                                                     |
| `DO_NODE_MACHINE_IMAGE`           | The DigitalOcean Image id or slug                                                                     |
| `DO_SSH_KEY_FINGERPRINT`          | The ssh key id or fingerprint (Should be already registered in the DigitalOcean Account)

### Running e2e test

In the root project directory run:

```
make test-e2e
```

### Running Conformance test

In the root project directory run:

```
make test-conformance
```
