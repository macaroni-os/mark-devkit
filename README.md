<p align="center">
  <img src="https://github.com/macaroni-os/macaroni-site/blob/master/site/static/images/logo.png">
</p>

# M.A.R.K. Development Kit

[![Build on push](https://github.com/macaroni-os/mark-devkit/actions/workflows/push.yml/badge.svg)](https://github.com/macaroni-os/mark-devkit/actions/workflows/push.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/macaroni-os/mark-devkit)](https://goreportcard.com/report/github.com/macaroni-os/mark-devkit)
[![CodeQL](https://github.com/macaroni-os/mark-devkit/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/macaroni-os/mark-devkit/actions/workflows/codeql-analysis.yml)

The Macaroni OS M.A.R.K. Development tool.

See [mark-stages](https://github.com/macaroni-os/mark-stages) repository for the documentation.


# Kit Merge

The `kit merge` command permits to generate and update a specific kit
with the *atoms* defined in YAML files reading the ebuild from external
repository or different kits and branches.


```
$> mark-devkit kit merge --help
Merge packages between kits.

Usage:
   kit merge [flags]

Aliases:
  merge, m, me

Flags:
      --concurrency int            Define the elaboration concurrency. (default 3)
      --deep int                   Define the limit of commits to fetch. (default 5)
  -h, --help                       help for merge
      --push                       Push commits to origin.
      --signature-email string     Specify the email of the user for the commits.
      --signature-name string      Specify the name of the user for the commits.
      --skip-pull-sources          Skip pull of sources repositories.
      --skip-reposcan-generation   Skip reposcan files generation.
      --specfile string            The specfiles of the jobs.
      --to string                  Override default work directory. (default "workdir")
      --verbose                    Show additional informations.

Global Flags:
  -c, --config string   MARK Devkit configuration file
  -d, --debug           Enable debug output.

```

Example:

```
$> mark-devkit kit merge --config contrib/config.yml --specfile ~/dev/macaroni/kit-fixups/core-kit/merge.kit.d/core.yml \
    --concurrency 10 --verbose --signature-email "mark-bot@macaronios.org" --signature-name "MARK Bot"
```

TODO:
- support automatic bump of a new revision when the md5 of the source is changed.
- 



