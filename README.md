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
./cluster-service
```


### How to use
```bash
# set env vars for clusters aws key and secret 
export AWS_ACCESS_KEY_ID=<key value>
export AWS_SECRET_ACCESS_KEY=<secret value>
```


```bash
# run the cleanup command in watch mode to delete persistence resources
./cluster-service cleanup <cluster_id> --dry-run=false --watch
# help 
./cluster-service cleanup --help
```

## Testing

To run unit tests, run:

```
make test/unit
```

## Releases

New binaries for a release tag will be created by [GoReleaser](https://goreleaser.com/) automatically.

To try out GoReleaser locally, it can be installed using `make setup/goreleaser`.

## Create image

To create an image and push to you own Quay.io repo 

```
make image/build/push ORG=<quay.io repo name>
```

To create an image and push to Quay.io, run:

```
make image/build/push
```
