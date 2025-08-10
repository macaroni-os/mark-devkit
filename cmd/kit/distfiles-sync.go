/*
	Copyright © 2024-2025 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmdkit

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/macaroni-os/mark-devkit/pkg/kit"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type DistfilesReport struct {
	Stats         *kit.AtomsStats  `json:"stats,omitempty" yaml:"stats,omitempty"`
	AtomsInErrors []*kit.AtomError `json:"atoms_errors,omitempty" yaml:"atoms_errors,omitempty"`
}

func (r *DistfilesReport) Yaml() ([]byte, error) {
	return yaml.Marshal(r)
}

func (r *DistfilesReport) Json() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

func (r *DistfilesReport) WriteJsonFile(f string) error {
	data, err := r.Json()
	if err != nil {
		return err
	}

	return os.WriteFile(f, data, 0644)
}

func (r *DistfilesReport) WriteYamlFile(f string) error {
	data, err := r.Yaml()
	if err != nil {
		return err
	}

	return os.WriteFile(f, data, 0644)
}

func KitDistfilesSyncCommand(config *specs.MarkDevkitConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "distfiles-sync",
		Aliases: []string{"ds", "distfiles"},
		Short:   "Sync distfiles for a list of kits.",
		PreRun: func(cmd *cobra.Command, args []string) {
			log := logger.GetDefaultLogger()
			specfile, _ := cmd.Flags().GetString("specfile")

			if specfile == "" {
				log.Fatal("No specfile param defined.")
			}

		},
		Run: func(cmd *cobra.Command, args []string) {
			log := logger.GetDefaultLogger()

			specfile, _ := cmd.Flags().GetString("specfile")
			to, _ := cmd.Flags().GetString("to")
			downloadDir, _ := cmd.Flags().GetString("download-dir")
			backend, _ := cmd.Flags().GetString("backend")
			verbose, _ := cmd.Flags().GetBool("verbose")
			skipGenReposcan, _ := cmd.Flags().GetBool("skip-reposcan-generation")
			concurrency, _ := cmd.Flags().GetInt("concurrency")
			showSummary, _ := cmd.Flags().GetBool("show-summary")
			writeSummaryFile, _ := cmd.Flags().GetString(
				"write-summary-file")
			summaryFormat, _ := cmd.Flags().GetString("summary-format")
			keepWorkdir, _ := cmd.Flags().GetBool("keep-workdir")
			checkOnlySize, _ := cmd.Flags().GetBool("check-only-size")
			atoms, _ := cmd.Flags().GetStringArray("pkg")

			minioBucket, _ := cmd.Flags().GetString("minio-bucket")
			minioAccessId, _ := cmd.Flags().GetString("minio-keyid")
			minioSecret, _ := cmd.Flags().GetString("minio-secret")
			minioEndpoint, _ := cmd.Flags().GetString("minio-endpoint")
			minioRegion, _ := cmd.Flags().GetString("minio-region")
			minioPrefix, _ := cmd.Flags().GetString("minio-prefix")

			if showSummary {
				config.GetLogging().Level = "error"
			}

			log.InfoC(log.Aurora.Bold(
				fmt.Sprintf(":mask:Loading specfile %s", specfile)),
			)

			backendOpts := make(map[string]string, 0)

			if backend == "s3" {
				if minioEndpoint != "" {
					backendOpts["minio-endpoint"] = minioEndpoint
				} else {
					backendOpts["minio-endpoint"] = os.Getenv("MINIO_URL")
				}

				if minioBucket != "" {
					backendOpts["minio-bucket"] = minioBucket
				} else {
					backendOpts["minio-bucket"] = os.Getenv("MINIO_BUCKET")
				}

				if minioAccessId != "" {
					backendOpts["minio-keyid"] = minioAccessId
				} else {
					backendOpts["minio-keyid"] = os.Getenv("MINIO_ID")
				}

				if minioSecret != "" {
					backendOpts["minio-secret"] = minioSecret
				} else {
					backendOpts["minio-secret"] = os.Getenv("MINIO_SECRET")
				}

				backendOpts["minio-region"] = minioRegion

				if minioPrefix != "" {
					backendOpts["minio-prefix"] = minioPrefix
				} else if os.Getenv("MINIO_PREFIX") != "" {
					backendOpts["minio-prefix"] = os.Getenv("MINIO_PREFIX")
				}
			}

			fetchOpts := kit.NewFetchOpts()
			fetchOpts.Concurrency = concurrency
			fetchOpts.Verbose = verbose
			fetchOpts.GenReposcan = !skipGenReposcan
			fetchOpts.CleanWorkingDir = !keepWorkdir
			fetchOpts.CheckOnlySize = checkOnlySize
			fetchOpts.Atoms = atoms

			fetcher, err := kit.NewFetcher(config, backend, backendOpts)
			if err != nil {
				log.Fatal(err.Error())
			}

			fetcher.SetWorkDir(to)
			if downloadDir != "" {
				fetcher.SetDownloadDir(downloadDir)
			}
			err = fetcher.Sync(specfile, fetchOpts)
			if err != nil {
				log.Fatal(err.Error())
			}

			report := &DistfilesReport{
				Stats:         fetcher.GetStats(),
				AtomsInErrors: *fetcher.GetAtomsInError(),
			}

			if writeSummaryFile != "" {
				if summaryFormat == "json" {
					err = report.WriteJsonFile(writeSummaryFile)
				} else {
					err = report.WriteYamlFile(writeSummaryFile)
				}
				if err != nil {
					log.Fatal(err.Error())
				}
			}

			if showSummary {
				var data []byte

				if summaryFormat == "json" {
					data, err = report.Json()
				} else {
					data, err = report.Yaml()
				}
				if err != nil {
					log.Fatal(err.Error())
				}

				fmt.Println(string(data))
			} else {
				log.InfoC(log.Aurora.Bold(":party_popper:All done"))
			}
		},
	}

	flags := cmd.Flags()
	flags.String("specfile", "", "The specfiles of the jobs.")
	flags.String("to", "workdir", "Override default work directory.")
	flags.String("download-dir", "", "Override the default ${workdir}/downloads directory.")
	flags.String("backend", "dir", "Set the fetcher backend to use: dir|s3.")
	flags.Int("concurrency", 3, "Define the elaboration concurrency.")
	flags.Bool("skip-reposcan-generation", false,
		"Skip reposcan files generation.")
	flags.Bool("verbose", false, "Show additional informations.")
	flags.Bool("show-summary", false, "Show YAML/JSON summary results")
	flags.String("write-summary-file", "",
		"Write the sync summary to the specified file in YAML/JSON format.")
	flags.String("summary-format", "yaml", "Specificy the summary format: json|yaml")
	flags.Bool("keep-workdir", false, "Avoid to remove the working directory.")

	flags.StringArray("pkg", []string{}, "Sync only specified packages.")
	flags.Bool("check-only-size", false,
		"Just compare file size without MD5 checksum.")

	// Fetcher S3 / Minio backend flags
	flags.String("minio-bucket", "",
		"Set minio bucket to use or set env MINIO_BUCKET.")
	flags.String("minio-endpoint", "",
		"Set minio endpoint to use or set env MINIO_URL.")
	flags.String("minio-keyid", "",
		"Set minio Access Key to use or set env MINIO_ID.")
	flags.String("minio-secret", "",
		"Set minio Access Key to use or set env MINIO_SECRET.")
	flags.String("minio-region", "", "Optionally define the minio region.")
	flags.String("minio-prefix", "",
		"Set the prefix path to use or set env MINIO_PREFIX. Note: The path is without initial /.")

	return cmd
}
