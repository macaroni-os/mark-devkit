
# Release 0.30.0 - May 12, 2026

* Upgrade vendor github.com/go-git/go-git/v5 v5.17.2 => v5.19.0 (CVE-2026-45022, CVE-2026-41506)

* Add `builtin-forgejo` generator. The host token could be configured through the
  *authorization* section of the config file.

# Release 0.29.0 - April 6, 2026

* Go 1.25 is now required

* Upgrade vendor github.com/go-git/go-git/v5 v5.16.4 => v5.17.2 (CVE-2026-34165)

* `builting-dirlisting` generator now permits to set `ignore_artefacts` to avoid
  download of the artefacts generated from URL. For example on using *custom*
  extension we can elaborate and generate artefacts later.

  Here a little example:

```yaml
    dir:
      url: '{{ .Values.base_url }}'
      matcher: 'version_[0-9.]+.json'
      ignore_artefacts: true
```


