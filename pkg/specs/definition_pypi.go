/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	gentoo "github.com/geaaru/pkgs-checker/pkg/gentoo"
)

// curl https://pypi.org/pypi/<pypi_name>/json

type PypiMetadata struct {
	Info       *PypiPackageInfo              `json:"info,omitempty" yaml:"info,omitempty"`
	LastSerial *int                          `json:"last_serial,omitempty" yaml:"last_serial,omitempty"`
	Releases   map[string][]*PypiReleaseFile `json:"releases,omitempty" yaml:"releases,omitempty"`
	// Files of the last releases
	Urls []*PypiReleaseFile `json:"urls,omitempty" yaml:"urls,omitempty"`
}

type PypiPackageInfo struct {
	Author                 string             `json:"author,omitempty" yaml:"author,omitempty"`
	AuthorEmail            string             `json:"author_email,omitempty" yaml:"author_email,omitempty"`
	BugtrackUrl            string             `json:"bugtrack_url,omitempty" yaml:"bugtrack_url,omitempty"`
	Classifiers            []string           `json:"classifiers,omitempty" yaml:"classifiers,omitempty"`
	Description            string             `json:"description,omitempty" yaml:"description,omitempty'`
	DescriptionContentType string             `json:"description_content_type,omitempty" yaml:"description_content_type,omitempty"`
	DocsUrl                string             `json:"docs_url,omitempty" yaml:"docs_url,omitempty"`
	DownloadUrl            string             `json:"download_url,omitempty,omitempty" yaml:"download_url,omitempty"`
	Downloads              *PypiDownloadsInfo `json:"downloads,omitempty" yaml:"downloads,omitempty"`
	Homepage               string             `json:"homepage,omitempty" yaml:"homepage,omitempty"`
	Keywords               string             `json:"keywords,omitempty" yaml:"keywords,omitempty"`
	License                string             `json:"license,omitempty" yaml:"license,omitempty"`
	LicenseExpr            string             `json:"license_expression,omitempty" yaml:"license_expression,omitempty"`
	Maintainer             string             `json:"maintainer,omitempty" yaml:"maintainer,omitempty"`
	MaintainerEmail        string             `json:"maintainer_email,omitempty" yaml:"maintainer_email,omitempty"`
	Name                   string             `json:"name,omitempty" yaml:"name,omitempty"`
	PackageUrl             string             `json:"package_url,omitempty" yaml:"package_url,omitempty"`
	Platform               string             `json:"platform,omitempty" yaml:"platform,omitempty"`
	ProjectUrl             string             `json:"project_url,omitempty" yaml:"project_url,omitempty"`
	ProjectUrls            map[string]string  `json:"project_urls,omitempty" yaml:"project_urls,omitempty"`
	ReleaseUrl             string             `json:"release_url,omitempty" yaml:"release_url,omitempty"`
	RequiresDist           []string           `json:"requires_dist,omitempty" yaml:"requires_dist,omitempty"`
	RequiresPython         string             `json:"requires_python,omitempty" yaml:"requires_python,omitempty"`
	Summary                string             `json:"summary,omitempty" yaml:"summary,omitempty"`
	Version                string             `json:"version,omitempty" yaml:"version,omitempty"`
	Yanked                 bool               `json:"yanked,omitempty" yaml:"yanked,omitempty"`
	YankedReason           string             `json:"yanked_reason,omitempty" yaml:"yanked_reason,omitempty"`
}

type PypiDownloadsInfo struct {
	LastDay   *int `json:"last_day,omitempty" yaml:"last_day,omitempty"`
	LastMonth *int `json:"last_month,omitempty" yaml:"last_month,omitempty"`
	LastWeek  *int `json:"last_week,omitempty" yaml:"last_week,omitempty"`
}

type PypiReleaseFile struct {
	CommentText string            `json:"comment_text,omitempty" yaml:"comment_text,omitempty"`
	Digests     map[string]string `json:"digests,omitempty" yaml:"digests,omitempty"`
	Downloads   *int              `json:"downloads,omitempty" yaml:"downloads,omitempty"`
	Filename    string            `json:"filename,omitempty" yaml:"filename,omitempty"`
	HasSig      bool              `json:"has_sig,omitempty" yaml:"has_sig,omitempty"`
	Md5Digest   string            `json:"md5_digest,omitempty" yaml:"md5_digest,omitempty"`
	// Type: sdist (source)|bdist_wheel
	PackageType       string `json:"packagetype,omitempty" yaml:"packagetype,omitempty"`
	PythonVersion     string `json:"python_version,omitempty" yaml:"python_version,omitempty"`
	RequiresPython    string `json:"requires_python,omitempty" yaml:"requires_python,omitempty"`
	Size              int    `json:"size,omitempty" yaml:"size,omitempty"`
	UploadTime        string `json:"upload_time,omitempty" yaml:"upload_time,omitempty"`
	UploadTimeIso8601 string `json:"upload_time_iso_8901,omitempty" yaml:"upload_time_iso_8901,omitempty"`
	Url               string `json:"url,omitempty" yaml:"url,omitempty'`
	Yanked            bool   `json:"yanked,omitempty" yaml:"yanked,omitempty"`
	YankedReason      string `json:"yanked_reason,omitempty" yaml:"yanked_reason,omitempty"`
}

type PythonDep struct {
	Use        string                `json:"use,omitempty" yaml:"use,omitempty"`
	PyVersions []string              `json:"py_versions,omitempty" yaml:"py_versions,omitempty"`
	Dependency *gentoo.GentooPackage `json:"dep,omitempty" yaml:"dep,omitempty"`
}
