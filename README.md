<p align="center">
  <img src="https://github.com/macaroni-os/macaroni-site/blob/master/site/static/images/logo.png">
</p>

# M.A.R.K. Development Kit

[![Build on push](https://github.com/macaroni-os/mark-devkit/actions/workflows/push.yml/badge.svg)](https://github.com/macaroni-os/mark-devkit/actions/workflows/push.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/macaroni-os/mark-devkit)](https://goreportcard.com/report/github.com/macaroni-os/mark-devkit)
[![CodeQL](https://github.com/macaroni-os/mark-devkit/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/macaroni-os/mark-devkit/actions/workflows/codeql-analysis.yml)

The Macaroni OS M.A.R.K. Development tool.


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

It permits to defined additional mirrors where download tarballs if the main URLs are no more valid
with different layouts: *flat* or *content-hash*.

```yaml
fallback_mirrors:
  - alias: macaroni-cdn-hash
    uri:
      - https://distfiles.macaronios.org/distfiles
      - https://distfiles-flat.macaronios.org/
    layout:
      modes:
        - type: "content-hash"
          hash: "SHA512"
          hash_mode: "8:8:8"

  - alias: macaroni-cdn-flat
    uri:
      - https://distfiles-flat.macaronios.org/distfiles
```

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

# Kit clone

The command `kit clone` permits to clone a list of kits and optionally generate:

* a YAML file with the list of the kits and the hash of the last commit

* generate the *reposcan* JSON files through the *anise-portage-converter* tool.

```
mark-devkit kit clone --help
Clone/Sync kits locally from a YAML specs rules file.

Usage:
   kit clone [flags]

Aliases:
  clone, c, cl, sync

Flags:
      --concurrency int             Define the elaboration concurrency. (default 3)
      --deep int                    Define the limit of commits to fetch. (default 5)
      --generate-reposcan-files     Generate reposcan files of the pulled kits.
  -h, --help                        help for clone
      --kit-cache-dir string        Directory where generate reposcan files. (default "kit-cache")
      --show-summary                Show YAML summary results
      --single-branch               Pull only the used branch. (default true)
      --specfile string             The specfiles of the jobs.
      --to string                   Target dir where sync kits. (default "output")
      --verbose                     Show additional informations.
      --write-summary-file string   Write the sync summary to the specified file in YAML format.

Global Flags:
  -c, --config string   MARK Devkit configuration file
  -d, --debug           Enable debug output.
```

# Autogen

The `autogen` command is used by M.A.R.K. workflow to autogen new ebuilds in a similar way at
metatools before but with a better integration for our CD/CI chain and for a CDN / Object Store
integration.

These the major differences:

* it doesn't follow a monolitic process on autogen ebuilds. It permits defined different YAML
  files with a specific target Kit. The defined YAML could be elaborated in parallel on different
  distributed nodes.

* it doesn't need a specific path for the YAML files that could be called with different names.

* it manages the autogen process like a step to use together with the process defined in the
  `kit merge` command. In particular, it generates the defined ebuilds in a source directoty
  that is later merged to the target kit as direct commit or as PR but as additional version
  to existing version. It doesn't replace the existing version.

* the *generator* component doesn't defined a complete process that describe the full
  generation process, but it used to define the searching data step about availables versions
  of a specific package and later describe how define artefacts to use after the core logic
  (transforms, selector) is been executed. This permits to implement easily new generator and
  to use existing core software for common operations (fetching, hashing, sync, etc).

* it permits to use different render engines:

  - `helm`: golang engine used by K8S and based on sprig
  - `pongo2`: golang engine that uses a django/jinja similar syntax
  - `j2cli`: python jinja2 engine that uses the `j2cli` tool and that permits to define custom
     macro to use on template.

* it uses a `flat` distfiles paradigm to store tarballs. It uses the feature available
  in the `distfiles-sync` command too, to deploy to a specific directory or to a specific S3
  Object Storage like Minio, Ceph, etc.

* inside the specfile used to define the package to autogen it permits to use Helm engine too
  for specific fields in order to render the string based on the metadata retrieved.
  These fields are:
    - *asset name*: The asset name is the name used as target distfiles.
    - *asset url*: used on *builtin-noop* generator to defined the url of the artefact
    - *asset prefix*: used on *builtin-dirlisting* generator to define the prefix path to use
      on download the tarball of the package.
    - *tarball*: used to override the name of the distfiles file when there aren't assets.

The `autogen` command needs at least two specific flags: `--specfile <file>` to pass the name
of the file with YAML specs of the packages to autogen, `--kitfile <file>` used to retrieve
the metadata of the target kit.

## Generators

The generator availables at the moment are:

* `builtin-github`: This generator permits to use GitHub API to retrieve tags or releases
  available for a specific repository and get the last and/or a specific version.

* `builtin-dirlisting`: this generator permits to using HTML indexes page and retrieve
  tarballs available for a specific package.

* `builtin-noop`: this generator permits to define static versions and/or snapshot.

## Definitions

In the *autogen* language every block of YAML is called *definition* and is managed
by a map that has as values the following informations:

* `generator`: the name of the generator to use.
* `template`: the name of the template engine to use: *helm*, *pongo2*, *j2cli*.
* `defaults`: the defaults section permits to define attributes to apply to all
  packages to autogen for the specific *definition*.
* `packages`: the list of the packages to autogen as a map where the key is the name
  of the package and the values are the configuration options with custom options based
  on the used generator.
* `transform`: as a value of a specific atom, it permits to define multiple transform
  rules (at the moment only of kind *string*) to clean the existing version from data
  not usable for version string parsing.
* `selector`: as a value of a specific atom, it permits to define condition (in and)
  about the version to select.

Example:

```yaml
macaroni:
  generator: builtin-github
  template:
    engine: helm

  defaults:
    category: dev-util
    github:
      user: macaroni-os
      query: releases
  packages:
    - mark-devkit:
        template: templates/mark-devkit-custom.tmpl
```

By default if the `template` attribute is not defined the tools search in the
directory of the specfiles following the convention: `templates/<aton-name>.tmpl`.

```
$> mark-devkit autogen --help
Executes Autogen elaboration.

Usage:
   autogen [flags]

Aliases:
  autogen, a

Flags:
      --backend string             Set the fetcher backend to use: dir|s3. (default "dir")
      --concurrency int            Define the elaboration concurrency. (default 3)
      --deep int                   Define the limit of commits to fetch. (default 5)
      --download-dir string        Override the default ${workdir}/downloads directory.
      --github-user string         Override the default Github user used for PR.
  -h, --help                       help for autogen
      --keep-workdir               Avoid to remove the working directory.
  -k, --kitfile string             The YAML with the target kit definition.
      --minio-bucket string        Set minio bucket to use or set env MINIO_BUCKET.
      --minio-endpoint string      Set minio endpoint to use or set env MINIO_URL.
      --minio-keyid string         Set minio Access Key to use or set env MINIO_ID.
      --minio-prefix string        Set the prefix path to use or set env MINIO_PREFIX. Note: The path is without initial /.
      --minio-region string        Optionally define the minio region.
      --minio-secret string        Set minio Access Key to use or set env MINIO_SECRET.
      --pr                         Push commit over specific branch and as Pull Request.
      --push                       Push commits to origin.
      --show-values                For debug purpose print generated values for any elaborated package in YAML format.
      --signature-email string     Specify the email of the user for the commits.
      --signature-name string      Specify the name of the user for the commits.
      --skip-merge                 Just generate the ebuild without merge to target kit. To use with --keep-workdir.
      --skip-pull-sources          Skip pull of sources repositories.
      --skip-reposcan-generation   Skip reposcan files generation.
      --specfile string            The specfile with the rules of the packages to autogen.
      --sync                       Sync artefacts to S3 backend server. (default true)
      --to string                  Override default work directory. (default "workdir")
      --verbose                    Show additional informations.

Global Flags:
  -c, --config string   MARK Devkit configuration file
  -d, --debug           Enable debug output.
```

# Autogen Thin a.k.a. `doit`

To help developers and contributors this command permits to execute a subset of the
steps of the `autogen` command. In particular, it used to generates locally the new ebuilds
without the other steps (merging, sync, etc.).

```
$> mark-devkit doit --help
Executes minimal Autogen elaboration for testing purpose.

Usage:
   autogen-thin [flags]

Aliases:
  autogen-thin, doit

Flags:
      --concurrency int       Define the elaboration concurrency. (default 3)
      --deep int              Define the limit of commits to fetch. (default 5)
      --download-dir string   Override the default ${workdir}/downloads directory.
  -h, --help                  help for autogen-thin
  -k, --kitfile string        The YAML with the target kit definition.
      --show-values           For debug purpose print generated values for any elaborated package in YAML format.
      --specfile string       The specfile with the rules of the packages to autogen.
      --to string             Override default work directory. (default "workdir")
      --verbose               Show additional informations.

Global Flags:
  -c, --config string   MARK Devkit configuration file
  -d, --debug           Enable debug output.
```

Considering the following examples of spec file:

```yaml
# file myspec.yml
x11libs:
  generator: builtin-noop
  template:
    engine: helm

  defaults:
    category: x11-libs
  packages:
    - libX11:
        vars:
          desc: "X.Org X11 library"
          versions:
            - "1.8.1"

        assets:
          - name: "libX11-{{ .Values.version}}.tar.xz"
            prefix: https://www.x.org/releases/individual/lib/

    - libX11:
        vars:
          desc: "X.Org X11 library"
          versions:
            - "1.7.0"

        assets:
          - name: "libX11-{{ .Values.version}}.tar.bz2"
            prefix: https://www.x.org/releases/individual/lib/

    - metatools:
        category: sys-apps
        vars:
          github_user: macaroni-os
          github_repo: funtoo-metatools
          homepage: https://github.com/macaroni-os/funtoo-metatools
          license: Apache-2.0
          desc: "M.A.R.K. metatools -- autogeneration framework."
          versions:
            - 1.3.8_pre20240818
          snapshot: 3ca77d7dac6b9256e161d0bce47c7f7eb85e6e96

        assets:
          - name: "{{ .Values.pn }}-{{ .Values.version }}-{{ substr 0 7 .Values.snapshot }}.zip"
            url: "https://github.com/{{ .Values.github_user }}/{{ .Values.github_repo }}/archive/{{ .Values.snapshot }}.zip"
```

and the following kit file:

```yaml
# cat merge.kit.d/test.yml
sources:
# We need this to inherit the eclass used for kit cache JSON generation
- name: "core-kit"
  url: "https://github.com/macaroni-os/core-kit"
  branch: "mark-testing"

target:
  name: test-kit
  url: "https://github.com/macaroni-os/test-kit"
  branch: mark-labs

  fixups:
    include:
      - file: ../../LICENSE
        to: LICENSE

      - file: ../../.github/FUNDING.yml
        to: .github/FUNDING.yml

  atoms_defaults:
    max_versions: 2

  atoms:

    - pkg: sys-apps/metatools
    - pkg: x11-libs/libX11
```

The command:

```bash
$> mark-devkit doit --specfile myspec.yml --kitfile merge.kit.d/macaroni.yml --show-values
😷 Loading specfile contrib/autogen/noop_example.yml
🏰 Work directory:	workdir
🚀 Target Kit:		mark-kit
🏭 [x11libs] Processing definition ...
🏭 [x11libs] Processing atom libX11...
🍕 [libX11] For package x11-libs/libX11 selected version 1.8.1
👀 [libX11] Values:
category: x11-libs
desc: X.Org X11 library
original_version: 1.8.1
pn: libX11
version: 1.8.1
versions:
    - 1.8.1

🏭 [x11libs] Processing atom libX11...
🍕 [libX11] For package x11-libs/libX11 selected version 1.7.0
👀 [libX11] Values:
category: x11-libs
desc: X.Org X11 library
original_version: 1.7.0
pn: libX11
version: 1.7.0
versions:
    - 1.7.0

🏭 [x11libs] Processing atom metatools...
🍕 [metatools] For package sys-apps/metatools selected version 1.3.8_pre20240818
👀 [metatools] Values:
category: sys-apps
desc: M.A.R.K. metatools -- autogeneration framework.
github_repo: funtoo-metatools
github_user: macaroni-os
homepage: https://github.com/macaroni-os/funtoo-metatools
license: Apache-2.0
original_version: 1.3.8_pre20240818
pn: metatools
snapshot: 3ca77d7dac6b9256e161d0bce47c7f7eb85e6e96
version: 1.3.8_pre20240818
versions:
    - 1.3.8_pre20240818

🎉 All done
```

This command generates under the *workdir* (managed by the `--to` flag) the following files:

```
$> tree workdir/
workdir/
├── downloads
│   ├── libX11-1.7.0.tar.bz2
│   ├── libX11-1.8.1.tar.xz
│   └── metatools-1.3.8_pre20240818-3ca77d7.zip
└── sources
    └── mark-kit
        ├── sys-apps
        │   └── metatools
        │       ├── Manifest
        │       └── metatools-1.3.8_pre20240818.ebuild
        └── x11-libs
            └── libX11
                ├── Manifest
                ├── libX11-1.7.0.ebuild
                └── libX11-1.8.1.ebuild

7 directories, 8 files
```

Under `downloads` the tarballs of the artefacts elaborated and under `sources` the directory
of the target kit with the new ebuilds.

The flag `--show-values` is a good debugging tool to show what versions are been retrieved
and later elaborated.

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

See [mark-stages](https://github.com/macaroni-os/mark-stages) repository for the documentation.

