/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

const (
	GeneratorBuiltinGitub      = "builtin-github"
	GeneratorBuiltinDirListing = "builtin-dirlisting"
	GeneratorBuiltinNoop       = "builtin-noop"
	GeneratorBuiltinPypi       = "builtin-pypi"

	GeneratorCustom = "custom"

	TmplEngineHelm   = "helm"
	TmplEnginePongo2 = "pongo2"
	TmplEngineJ2cli  = "j2cli"

	ExtensionCustom = "custom"
	ExtensionGolang = "golang"
)

type AutogenSpec struct {
	File string `json:"-" yaml:"-"`

	Version     string                        `json:"version,omitempty" yaml:"version,omitempty"`
	Definitions map[string]*AutogenDefinition `json:"-,inline" yaml:"-,inline"`
}

type AutogenDefinition struct {
	TemplateEngine *AutogenTemplateEngine    `json:"template,omitempty" yaml:"template,omitempty"`
	Generator      string                    `json:"generator,omitempty" yaml:"generator,omitempty"`
	GeneratorOpts  map[string]string         `json:"generator_opts,omitempty" yaml:"generator_opts,omitempty"`
	Defaults       *AutogenAtom              `json:"defaults,omitempty" yaml:"defaults,omitempty"`
	Packages       []map[string]*AutogenAtom `json:"packages,omitempty" yaml:"packages,omitempty"`

	Extensions map[string]*AutogenExtension `json:"extensions_defs,omitempty" yaml:"extensions_defs,omitempty"`
}

type AutogenExtension struct {
	Name    string            `json:"name" yaml:"name"`
	Options map[string]string `json:"opts,omitempty" yaml:"opts,omitempty"`
}

type AutogenArtefact struct {
	Use    string            `json:"use,omitempty" yaml:"use,omitempty"`
	SrcUri []string          `json:"src_uri" yaml:"src_uri"`
	Name   string            `json:"name" yaml:"name"`
	Hashes map[string]string `json:"hashes,omitempty" yaml:"hashes,omitempty"`
	Local  *bool             `json:"local,omitempty" yaml:"local,omitempty"`
}

type AutogenAtom struct {
	Name     string                  `json:"-" yaml:"-"`
	Tarball  string                  `json:"tarball,omitempty" yaml:"tarball,omitempty"`
	Github   *AutogenGithubProps     `json:"github,omitempty" yaml:"github,omitempty"`
	Dir      *AutogenDirlistingProps `json:"dir,omitempty" yaml:"dir,omitempty"`
	Python   *AutogenPythonOpts      `json:"-,inline" yaml:"-,inline"`
	Vars     map[string]interface{}  `json:"vars,omitempty" yaml:"vars,omitempty"`
	Category string                  `json:"category,omitempty" yaml:"category,omitempty"`

	Template string `json:"template,omitempty" yaml:"template,omitempty"`
	FilesDir string `json:"files_dir,omitempty" yaml:"files_dir,omitempty"`

	Extensions []string        `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	Assets     []*AutogenAsset `json:"assets,omitempty" yaml:"assets,omitempty"`

	Transforms []*AutogenTransform `json:"transform,omitempty" yaml:"transform,omitempty"`
	Excludes   []string            `json:"exclude,omitempty" yaml:"exclude,omitempty"`
	Selector   []string            `json:"selector,omitempty" yaml:"selector,omitempty"`

	IgnoreArtefacts *bool `json:"ignore_artefacts,omitempty" yaml:"ignore_artefacts,omitempty"`
}

type AutogenPythonOpts struct {
	PythonCompat         string              `json:"python_compat,omitempty" yaml:"python_compat,omitempty"`
	PythonRequiresIgnore string              `json:"python_requires_ignore,omitempty" yaml:"python_requires_ignore,omitempty"`
	DepsIgnore           []string            `json:"pydeps_ignore,omitempty" yaml:"pydeps_ignore,omitempty"`
	PypiName             string              `json:"pypi_name,omitempty" yaml:"pypi_name,omitempty"`
	Pydeps               map[string][]string `json:"pydeps,omitempty" yaml:"pydeps,omitempty"`
}

type AutogenAsset struct {
	Use     string `json:"use,omitempty" yaml:"use,omitempty"`
	Name    string `json:"name,omitempty" yaml:"name,omitempty"`
	Matcher string `json:"matcher,omitempty" yaml:"matcher,omitempty"`
	Prefix  string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Url     string `json:"url,omitempty" yaml:"url,omitempty"`
	Local   *bool  `json:"local,omitempty" yaml:"local,omitempty"`
}

type AutogenTransform struct {
	Kind    string `json:"kind,omitempty" yaml:"kind,omitempty"`
	Match   string `json:"match,omitempty" yaml:"match,omitempty"`
	Replace string `json:"replace,omitempty" yaml:"replace,omitempty"`
}

type AutogenTemplateEngine struct {
	Engine string   `json:"engine,omitempty" yaml:"engine,omitempty"`
	Opts   []string `json:"opts,omitempty" yaml:"opts,omitempty"`
}

type AutogenGithubProps struct {
	User  string `json:"user,omitempty" yaml:"user,omitempty"`
	Repo  string `json:"repo,omitempty" yaml:"repo,omitempty"`
	Query string `json:"query,omitempty" yaml:"query,omitempty"`
	Match string `json:"match,omitempty" yaml:"match,omitempty"`

	PerPage  *int `json:"per_page,omitempty" yaml:"per_page,omitempty"`
	Page     *int `json:"page,omitempty" yaml:"page,omitempty"`
	NumPages *int `json:"num_pages,omitempty" yaml:"num_pages,omitempty"`
}

type AutogenDirlistingProps struct {
	Url             string `json:"url,omitempty" yaml:"url,omitempty"`
	Matcher         string `json:"matcher,omitempty" yaml:"matcher,omitempty"`
	ExcludesMatcher string `json:"exclude,omitempty" yaml:"exclude,omitempty"`
}
