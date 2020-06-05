# RHMI Cluster Tooling and Service

This repo is intended to store the services and tooling related to the
management of RHMI clusters.

## CLI

This repo contains a CLI which can be used to teardown remaining cloud
resources in a cloud provider for a specific RHMI cluster.

### Building

To build the CLI, run from the root of this repo:

```
make build/cli
```

A binary will be created in the root directory of the repo, which can be run:

```
./cli
```


### How to use

The cli expects a `cluster id`, flags and env vars for AWS access and secret keys used by the cluster in order to run. 

This info can be fetched from the OSD cluster itself if your signed in as a cluster-admin via `oc`. 

```
# Fetch the AWS keys used for resource provisioning
export AWS_ACCESS_KEY_ID=$(oc get secret aws-creds -n kube-system -o jsonpath={.data.aws_access_key_id} | base64 --decode)        
export AWS_SECRET_ACCESS_KEY=$(oc get secret aws-creds -n kube-system -o jsonpath={.data.aws_secret_access_key} | base64 --decode)

# Run cleanup in dry run (default) by fetching the cluster id and region via oc 
# To perform resource deletion, add --dry-run=false to command
./cli cleanup $(oc get infrastructure cluster -o jsonpath='{.status.infrastructureName}') --region=$(oc get infrastructure cluster -o jsonpath='{.status.platformStatus.aws.region}') --verbose
```

## Testing

To run unit tests, run:

```
make test/unit
```

## Releases

New binaries for a release tag will be created by [GoReleaser](https://goreleaser.com/) automatically.

To try out GoReleaser locally, it can be installed using `make setup/goreleaser`. 