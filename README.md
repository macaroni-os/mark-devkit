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

The Github Token is read from `GITHUB_TOKEN` environment variable.

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
      --github-user string         Override the default Github user used for PR.
  -h, --help                       help for merge
      --keep-workdir               Avoid to remove the working directory.
      --pr                         Push commit over specific branch and as Pull Request.
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

These the main features:

* generation and update of the profiles directory of the target kit: (categories, repo_name, thirdpartymirrors, etc.)
* generation and update of the metadata directory of the target kit
* generation and update of the eclass directory of the target kit
* automatic merge of the defined packages in the YAML file (optionally with specific conditions)
  from sources kit to the target kit. The last version ebuild is merged.
* automatic bump with a new revision of existing ebuild with changes
* generate and update static directories defined in the YAML file.
* permits to merge changes as direct commits in the target kit branch or through Github PR.

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

# Kit clean

The `kit clean` command permits to purge old ebuilds.

```
$> mark-devkit kit clean --help
Clean old packages from kits.

Usage:
   kit clean [flags]

Aliases:
  clean, purge, c

Flags:
      --concurrency int            Define the elaboration concurrency. (default 3)
      --deep int                   Define the limit of commits to fetch. (default 5)
      --github-user string         Override the default Github user used for PR.
  -h, --help                       help for clean
      --keep-workdir               Avoid to remove the working directory.
      --pr                         Push commit over specific branch and as Pull Request.
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

The Github Token is read from `GITHUB_TOKEN` environment variable.

The number of ebuilds maintained depends on *default* versions defined in the YAML
as `atoms_defaults`:

```yaml
target:
  atoms_defaults:
    max_versions: 2
```

or directly as custom field *max_versions* of the defined package:

```yaml
target:
  atoms:
    - pkg: dev-util/mark-devkit
      max_versions: 2
      #versions:
      #  - "0.8.0"
```

and if it's been pinned a specific version:

```yaml
target:
  atoms:
    - pkg: dev-util/mark-devkit
      versions:
        - "0.8.0"
```

Example:

```
$> mark-devkit kit clean --config contrib/config.yml --specfile ~/dev/macaroni/kit-fixups/mark-kit/merge.kit.d/macaroni.yml \
    --concurrency 10 --verbose --signature-email "mark-bot@macaronios.org" --signature-name "MARK bot" --pr --push
```

# Kit distfiles-sync

The `kit distfiles-sync` permits to generate and/or sync the distfiles mirror as *flat* mode.

```
$> mark-devkit kit distfiles-sync --help
Sync distfiles for a list of kits.

Usage:
   kit distfiles-sync [flags]

Aliases:
  distfiles-sync, ds, distfiles

Flags:
      --backend string              Set the fetcher backend to use: dir|s3. (default "dir")
      --check-only-size             Just compare file size without MD5 checksum.
      --concurrency int             Define the elaboration concurrency. (default 3)
  -h, --help                        help for distfiles-sync
      --keep-workdir                Avoid to remove the working directory.
      --minio-bucket string         Set minio bucket to use or set env MINIO_BUCKET.
      --minio-endpoint string       Set minio endpoint to use or set env MINIO_URL.
      --minio-keyid string          Set minio Access Key to use or set env MINIO_ID.
      --minio-prefix string         Set the prefix path to use or set env MINIO_PREFIX. Note: The path is without initial /.
      --minio-region string         Optionally define the minio region.
      --minio-secret string         Set minio Access Key to use or set env MINIO_SECRET.
      --pkg stringArray             Sync only specified packages.
      --show-summary                Show YAML/JSON summary results
      --skip-reposcan-generation    Skip reposcan files generation.
      --specfile string             The specfiles of the jobs.
      --summary-format string       Specificy the summary format: json|yaml (default "yaml")
      --to string                   Override default work directory. (default "workdir")
      --verbose                     Show additional informations.
      --write-summary-file string   Write the sync summary to the specified file in YAML/JSON format.

Global Flags:
  -c, --config string   MARK Devkit configuration file
  -d, --debug           Enable debug output.
```

It supports two backends:

* `dir`: It syncs tarballs to a filesystem directory

* `s3`: It syncs tarballs to an S3/Minio/Ceph Object Store. Considering that the S3 doesn't share
  a specific hash of the target object `mark-devkit` sets custom Metadata attributes with the
  tarball BLAKE2B and/or SHA512 hashes.

# Kit bump-release

The command `kit bump-release` permits to bump a revision to a specific `meta-repo`.

The `meta-repo` is used by *ego* to sync kits.

```
$> mark-devkit kit bump-release --help
Bump a new kits release.

Usage:
   kit bump-release [flags]

Aliases:
  bump-release, br, bump, release

Flags:
      --deep int                 Define the limit of commits to fetch. (default 5)
  -h, --help                     help for bump-release
      --push                     Push commits to origin.
      --signature-email string   Specify the email of the user for the commits.
      --signature-name string    Specify the name of the user for the commits.
      --specfile string          The specfiles of the jobs.
      --to string                Override default work directory. (default "workdir")
      --verbose                  Show additional informations.

Global Flags:
  -c, --config string   MARK Devkit configuration file
  -d, --debug           Enable debug output.
```

# Metro run

The command `metro run` replace the Funtoo *metro* tool and permit to generate Stage tarballs.

```
$> mark-devkit metro run --help
Run one or more job.

Usage:
   metro run [flags]

Aliases:
  run, r, j

Flags:
      --cleanup             Cleanup rootfs directory. (default true)
      --cpu string          Specify specific CPU type for QEMU to use
      --fchroot-debug       Enable debug on fchroot.
  -h, --help                help for run
      --job string          The job to diagnose.
      --quiet               Avoid to see the hooks command output.
      --skip-hooks-phase    Skip hooks executions. For development only.
      --skip-packer-phase   Skip packer phase.
      --skip-source-phase   Skip source phase.
      --specfile string     The specfiles of the jobs.

Global Flags:
  -c, --config string   MARK Devkit configuration file
  -d, --debug           Enable debug output.
```
