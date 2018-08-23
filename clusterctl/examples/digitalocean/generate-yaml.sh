#!/bin/sh
set -e

OVERWRITE=0
OUTPUT_DIR=out

PROVIDERCOMPONENT_TEMPLATE_FILE=provider-components.yaml.template
PROVIDERCOMPONENT_GENERATED_FILE=${OUTPUT_DIR}/provider-components.yaml
MACHINES_TEMPLATE_FILE=machines.yaml.template
MACHINES_GENERATED_FILE=${OUTPUT_DIR}/machines.yaml
CLUSTER_TEMPLATE_FILE=cluster.yaml.template
CLUSTER_GENERATED_FILE=${OUTPUT_DIR}/cluster.yaml

REGION=fra1
CLUSTER_NAME=test-1
MASTER_NAME=digitalocean-fra1-master-
NODE_NAME=digitalocean-fra1-node-

SCRIPT=$(basename $0)
while test $# -gt 0; do
        case "$1" in
          -h|--help)
            echo "$SCRIPT - generates input yaml files for Cluster API on openstack"
            echo " "
            echo "$SCRIPT [options]"
            echo " "
            echo "options:"
            echo "-h, --help                show brief help"
            echo "-f, --force-overwrite     if file to be generated already exists, force script to overwrite it"
            exit 0
            ;;
          -f)
            OVERWRITE=1
            shift
            ;;
          --force-overwrite)
            OVERWRITE=1
            shift
            ;;
          *)
            break
            ;;
        esac
done

if [ $OVERWRITE -ne 1 ] && [ -f $PROVIDERCOMPONENT_GENERATED_FILE ]; then
  echo "File $PROVIDERCOMPONENT_GENERATED_FILE already exists. Delete it manually before running this script."
  exit 1
fi

mkdir -p ${OUTPUT_DIR}

cat $PROVIDERCOMPONENT_TEMPLATE_FILE \
  > $PROVIDERCOMPONENT_GENERATED_FILE
echo "Done generating $PROVIDERCOMPONENT_GENERATED_FILE"

cat $MACHINES_TEMPLATE_FILE \
    | sed -e "s/\$REGION/$REGION/" \
    | sed -e "s/\$MASTER_NAME/$MASTER_NAME/" \
    | sed -e "s/\$NODE_NAME/$NODE_NAME/" \
  > $MACHINES_GENERATED_FILE
echo "Done generating $MACHINES_GENERATED_FILE"

cat $CLUSTER_TEMPLATE_FILE \
    | sed -e "s/\$CLUSTER_NAME/$CLUSTER_NAME/" \
  > $CLUSTER_GENERATED_FILE
echo "Done generating $CLUSTER_GENERATED_FILE"

