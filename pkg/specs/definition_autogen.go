/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

const (
	GeneratorBuiltinGitub      = "builtin-github"
	GeneratorBuiltinDirListing = "builtin-dirlisting"

	TmplEngineHelm   = "helm"
	TmplEnginePongo2 = "pongo2"
	TmplEngineJ2cli  = "j2cli"
)

type AutogenSpec struct {
	File string `json:"-" yaml:"-"`

	Version     string                        `json:"version,omitempty" yaml:"version,omitempty"`
	Definitions map[string]*AutogenDefinition `json:"-,inline" yaml:"-,inline"`
}

type AutogenDefinition struct {
	TemplateEngine *AutogenTemplateEngine    `json:"template,omitempty" yaml:"template,omitempty"`
	Generator      string                    `json:"generator,omitempty" yaml:"generator,omitempty"`
	Defaults       *AutogenAtom              `json:"defaults,omitempty" yaml:"defaults,omitempty"`
	Packages       []map[string]*AutogenAtom `json:"packages,omitempty" yaml:"packages,omitempty"`
}

type AutogenArtefact struct {
	Use    string   `json:"use,omitempty" yaml:"use,omitempty"`
	SrcUri []string `json:"src_uri" yaml:"src_uri"`
	Name   string   `json:"name" yaml:"name"`
}

type AutogenAtom struct {
	Name     string                 `json:"-" yaml:"-"`
	Tarball  string                 `json:"tarball,omitempty" yaml:"tarball,omitempty"`
	Github   *AutogenGithubProps    `json:"github,omitempty" yaml:"github,omitempty"`
	Vars     map[string]interface{} `json:"vars,omitempty" yaml:"vars,omitempty"`
	Category string                 `json:"category,omitempty" yaml:"category,omitempty"`

	Template string `json:"template,omitempty" yaml:"template,omitempty"`

	Extensions []string        `json:"extentions,omitempty" yaml:"extentions,omitempty"`
	Assets     []*AutogenAsset `json:"assets,omitempty" yaml:"assets,omitempty"`

	Transforms []*AutogenTransform `json:"transform,omitempty" yaml:"transform,omitempty"`
	Selector   []string            `json:"selector,omitempty" yaml:"selector,omitempty"`

	PythonCompat string `json:"python_compat,omitempty" yaml:"python_compat,omitempty"`
}

type AutogenAsset struct {
	Use     string `json:"use,omitempty" yaml:"use,omitempty"`
	Name    string `json:"name,omitempty" yaml:"name,omitempty"`
	Matcher string `json:"matcher,omitempty" yaml:"matcher,omitempty"`
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

	PerPage  *int `json:"per_page,omitempty" yaml:"per_page,omitempty"`
	Page     *int `json:"page,omitempty" yaml:"page,omitempty"`
	NumPages *int `json:"num_pages,omitempty" yaml:"num_pages,omitempty"`
}
