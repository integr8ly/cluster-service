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

TODO

## Testing

To run unit tests, run:

```
make test/unit
```

## Releases

New binaries for a release tag will be created by [GoReleaser](https://goreleaser.com/) automatically.

To try out GoReleaser locally, it can be installed using `make setup/goreleaser`. 