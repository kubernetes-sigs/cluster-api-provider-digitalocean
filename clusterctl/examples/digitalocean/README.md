# DigitalOcean Example Manifests

This directory contains example manifest files that can be used to create a fully-functional cluster.

## Generation

For convenience, a generation script which populates templates manifests is provided.

1. Run the generation script.
```bash
./generate-yaml.sh
```

If the yaml files already exist, you will see an error like the one below:
```bash
$ ./generate-yaml.sh
File provider-components.yaml already exists. Delete it manually before running this script.
```

In that case, make sure the remove the `out` directory before running the script, or use the `OVERWRITE` environment variable, such as:
```bash
OVERWRITE=1 ./generate-yaml.sh
```

## Defaults and Modifications

You may always manually curate files based on the examples provided. The generation script have default values established, which you can change by setting the appropriate environment variable.

The following environment variables are used by the script:

* `$OUTPUT_DIR` - location where generated manifests are stored. Default `./out`.
* `$REGION` - slug of region where instances will be created. Default `fra1` (Frankfurt).
* `$CLUSTER_NAME` - name of the cluster. Default `test-1`.
* `$MASTER_NAME` - name of the Master instance. The randomly generated suffix is appended to the name by the `machine-controller` on-creation. Default `digitalocean-fra1-master-`.
* `$NODE_NAME` - name of nodes. The randomly generated suffix is appended to the name by the `machine-controller` on-creation. Default `digitalocean-fra1-node-`.

The `$DIGITALOCEAN_ACCESS_TOKEN` environment variable must be set to the [DigitalOcean API Access Token](https://www.digitalocean.com/docs/api/create-personal-access-token/) for the `machine-controller` and `cluster-controller` to work properly.

