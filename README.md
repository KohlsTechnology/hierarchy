# Hierarchy

[![Build Status](https://travis-ci.com/KohlsTechnology/prometheus_bigquery_remote_storage_adapter.svg?branch=master)](https://travis-ci.com/KohlsTechnology/hierarcbhy)
[![Go Report Card](https://goreportcard.com/badge/github.com/KohlsTechnology/prometheus_bigquery_remote_storage_adapter)](https://goreportcard.com/report/github.com/KohlsTechnology/hierarcbhy)

Hierarchy is a simple utility to merge a set of yaml or json files, based on a defined hierarchy. It is inspired by Hiera from Puppet.

## Installation

You can either download a binary file from GitHub releases or [compile from source](#compiling-from-source).

## Documentation

The main goal of `hierarchy` is to prevent the configuration data duplication and allow for a fine-grained control over the GitOps process. The output YAML can be used to generate a `values.yml` file for GitOps tools, like [Eunomia](https://github.com/KohlsTechnology/eunomia), [Helm](https://helm.sh), and others.

### Usage

```
./hierarchy -h
usage: hierarchy [<flags>]

Hierarchy

Flags:
  -h, --help                    Show context-sensitive help (also try --help-long and --help-man).
  -f, --file="./hierarchy.lst"  Path and name of the hierarchy file.
  -o, --output="./output.yaml"  Path and name of the output file.
  -i, --filter="(.yaml|.yml|.json)$"
                                Regex for file extension of files being merged
      --trace                   Prints a diff after processing each file. This generates A LOT of output
  -m, --failmissing             Fail if a directory in the hierarchy is missing
  -V, --version                 Print version and build information, then exit
  -d, --debug                   Print debug output

```

### Merging

The `hierarchy` utility processes the YAML structure as a deep merge, with the exception of lists. Lists are completely overwritten, and therefore, it is important to keep that in mind when using them.

### Hierarchy

The hierarchy is defined in the file `hierarchy.lst`. This is a simple text file that lists one include folder per line and supports comments prefixed with `#`. The directories listed can be relative or absolute (try to avoid) paths. You can have directories included that are higher or lower in the structure to control their precedence. You can look at examples [here](https://github.com/KohlsTechnology/hierarchy/blob/master/testdata/).

#### Example content
```
../defaults    #this is the first ... lowest priority
../marketing   #this is the second
../development #this is the third ... highest priority
```

In this case, it will load all yaml files from `../defaults`, then merge it with everything in `../marketing`, and lastly merge it with everything in `../development`. You can also use the relative path `./`, which means that it will also load variables defined in contextDir directly (same folder level as `hierarchy.lst`). You can insert `./` in any desired order in the `hierarchy.lst`, thus determining its priority.

#### Example

Let's assume you have multiple applications that get deployed over different cloud providers. This application also has development, QA, and production environments. You can decide in whichever priority (order) the configuration files are merged.

```
defaults             # all applications will have this
└── cloud            # settings specific to a cloud
  └── environment    # settings specific to the environment level (e.g. development or production)
    └── application  # settings specific to an application
```

In order to generate the final configuration for an application running in the development environment on cloud A, the `hierarchy.lst` could look like the below example.

```
# Hierarchy file for application "demo" running in "cloud A".
# location `..../applications/demo/dev/clouda1.hierarchy.lst
../../defaults
../clouds/A
./environments/dev
./
```

## Developing

See [CONTRIBUTING.md](.github/CONTRIBUTING.md) for details.


### Dependencies
Go 1.15+

### Compiling From Source
```
make build
```

### Testing

To run the tests, simply execute:
```
make test
```

### Releasing

This project is using [goreleaser](https://goreleaser.com). GitHub release creation is automated using Travis
CI. New releases are automatically created when new tags are pushed to the repo.
```
$ TAG=v0.0.2 make tag
```

How to manually create a release without relying on Travis CI.
```
$ TAG=v0.0.2 make tag
$ GITHUB_TOKEN=xxx make clean release
```

## License

See [LICENSE](LICENSE) for details.

## Code of Conduct

See [CODE_OF_CONDUCT.md](.github/CODE_OF_CONDUCT.md)
for details.
