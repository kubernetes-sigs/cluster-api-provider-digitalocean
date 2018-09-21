#!/bin/sh
set -e

OVERWRITE=0
OUTPUT_DIR=${OUTPUT_DIR:-out}
CONTROLLER_VERSION=${CONTROLLER_VERSION:-0.0.1}

PROVIDERCOMPONENT_TEMPLATE_FILE=provider-components.yaml.template
PROVIDERCOMPONENT_GENERATED_FILE=${OUTPUT_DIR}/provider-components.yaml
MACHINES_TEMPLATE_FILE=machines.yaml.template
MACHINES_GENERATED_FILE=${OUTPUT_DIR}/machines.yaml
CLUSTER_TEMPLATE_FILE=cluster.yaml.template
CLUSTER_GENERATED_FILE=${OUTPUT_DIR}/cluster.yaml
ADDONS_TEMPLATE_FILE=addons.yaml.template
ADDONS_GENERATED_FILE=${OUTPUT_DIR}/addons.yaml

REGION=${REGION:-fra1}
CLUSTER_NAME=${CLUSTER_NAME:-test-1}
MASTER_NAME=${MASTER_NAME:-digitalocean-fra1-master-}
NODE_NAME=${NODE_NAME:-digitalocean-fra1-node-}
DIGITALOCEAN_ACCESS_TOKEN=${DIGITALOCEAN_ACCESS_TOKEN}

SSH_KEY_GENERATED_FILE=${OUTPUT_DIR}/${CLUSTER_NAME}_rsa

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

if [ $OVERWRITE -ne 1 ] && [ -f $MACHINES_GENERATED_FILE ]; then
  echo "File $MACHINES_GENERATED_FILE already exists. Delete it manually before running this script."
  exit 1
fi

if [ $OVERWRITE -ne 1 ] && [ -f $CLUSTER_GENERATED_FILE ]; then
  echo "File $CLUSTER_GENERATED_FILE already exists. Delete it manually before running this script."
  exit 1
fi

if [ $OVERWRITE -ne 1 ] && [ -f $ADDONS_GENERATED_FILE ]; then
  echo "File $ADDONS_GENERATED_FILE already exists. Delete it manually before running this script."
  exit 1
fi

if [ $OVERWRITE -ne 1 ] && [ -f $SSH_KEY_GENERATED_FILE ]; then
  echo "File $SSH_KEY_GENERATED_FILE already exists. Delete it manually before running this script."
  exit 1
fi

mkdir -p ${OUTPUT_DIR}

# This command generates new SSH key to be used by the machine controller to communicate with the cluster.
# The SSH private/public keypair is saved in `out` directory, so it can be used by the user if needed.
# The key doesn't have passphrase as locked SSH keys are not supported by the upstream API: https://github.com/kubernetes-sigs/cluster-api/issues/160
ssh-keygen -f $SSH_KEY_GENERATED_FILE -N "" -C $CLUSTER_NAME -t rsa
echo "Done generating SSH key $SSH_KEY_GENERATED_FILE"

SSH_PUBLIC_KEY="$(cat ${SSH_KEY_GENERATED_FILE}.pub | base64 | tr -d '\r\n')"
SSH_PRIVATE_KEY=$(cat ${SSH_KEY_GENERATED_FILE} | base64 | tr -d '\r\n')

cat $PROVIDERCOMPONENT_TEMPLATE_FILE \
  | sed -e "s/\$DIGITALOCEAN_ACCESS_TOKEN/$DIGITALOCEAN_ACCESS_TOKEN/" \
  | sed -e "s/\$SSH_PRIVATE_KEY/$SSH_PRIVATE_KEY/" \
  | sed -e "s/\$SSH_PUBLIC_KEY/$SSH_PUBLIC_KEY/" \
  | sed -e "s/\$CONTROLLER_VERSION/$CONTROLLER_VERSION/" \
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

cat $ADDONS_TEMPLATE_FILE \
    | sed -e "s/\$DIGITALOCEAN_ACCESS_TOKEN/$DIGITALOCEAN_ACCESS_TOKEN/" \
  > $ADDONS_GENERATED_FILE
echo "Done generating $ADDONS_GENERATED_FILE"
