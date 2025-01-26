/*
	Copyright © 2024-2025 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/macaroni-os/mark-devkit/pkg/autogen"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/spf13/cobra"
)

func autogenCmdCommand(config *specs.MarkDevkitConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "autogen",
		Aliases: []string{"a"},
		Short:   "Autogen specs.",
		Long:    `Executes Autogen elaboration.`,
		PreRun: func(cmd *cobra.Command, args []string) {
			log := logger.GetDefaultLogger()
			specfile, _ := cmd.Flags().GetString("specfile")
			kitfile, _ := cmd.Flags().GetString("kitfile")

			if specfile == "" {
				log.Fatal("No specfile param defined.")
			}

			if kitfile == "" {
				log.Fatal("No kitfile param defined.")
			}

		},
		Run: func(cmd *cobra.Command, args []string) {
			log := logger.GetDefaultLogger()
			specfile, _ := cmd.Flags().GetString("specfile")
			kitfile, _ := cmd.Flags().GetString("kitfile")
			deep, _ := cmd.Flags().GetInt("deep")
			concurrency, _ := cmd.Flags().GetInt("concurrency")
			to, _ := cmd.Flags().GetString("to")
			downloadDir, _ := cmd.Flags().GetString("download-dir")
			backend, _ := cmd.Flags().GetString("backend")
			skipMerge, _ := cmd.Flags().GetBool("skip-merge")
			skipPullSources, _ := cmd.Flags().GetBool("skip-pull-sources")
			skipGenReposcan, _ := cmd.Flags().GetBool("skip-reposcan-generation")
			signatureName, _ := cmd.Flags().GetString("signature-name")
			signatureEmail, _ := cmd.Flags().GetString("signature-email")
			verbose, _ := cmd.Flags().GetBool("verbose")
			keepWorkdir, _ := cmd.Flags().GetBool("keep-workdir")
			push, _ := cmd.Flags().GetBool("push")
			githubUser, _ := cmd.Flags().GetString("github-user")
			pr, _ := cmd.Flags().GetBool("pr")
			sync, _ := cmd.Flags().GetBool("sync")
			showValues, _ := cmd.Flags().GetBool("show-values")
			forceMergeCheck, _ := cmd.Flags().GetBool("force-merge-check")

			minioBucket, _ := cmd.Flags().GetString("minio-bucket")
			minioAccessId, _ := cmd.Flags().GetString("minio-keyid")
			minioSecret, _ := cmd.Flags().GetString("minio-secret")
			minioEndpoint, _ := cmd.Flags().GetString("minio-endpoint")
			minioRegion, _ := cmd.Flags().GetString("minio-region")
			minioPrefix, _ := cmd.Flags().GetString("minio-prefix")

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

			log.InfoC(log.Aurora.Bold(
				fmt.Sprintf(":mask:Loading specfile %s", specfile)),
			)

			autogenOpts := autogen.NewAutogenBotOpts()
			autogenOpts.Concurrency = concurrency
			autogenOpts.GitDeepFetch = deep
			autogenOpts.Push = push
			autogenOpts.PullRequest = pr
			autogenOpts.Verbose = verbose
			autogenOpts.SyncFiles = sync
			autogenOpts.SignatureName = signatureName
			autogenOpts.SignatureEmail = signatureEmail
			autogenOpts.CleanWorkingDir = !keepWorkdir
			autogenOpts.PullSources = !skipPullSources
			autogenOpts.GenReposcan = !skipGenReposcan
			autogenOpts.MergeAutogen = !skipMerge
			autogenOpts.MergeForced = forceMergeCheck
			autogenOpts.ShowGeneratedValues = showValues

			if githubUser != "" {
				autogenOpts.GithubUser = githubUser
			}

			autogenBot := autogen.NewAutogenBot(config)
			autogenBot.SetWorkDir(to)
			if downloadDir != "" {
				autogenBot.SetDownloadDir(downloadDir)
			}
			err := autogenBot.SetupFetcher(backend, backendOpts)
			if err != nil {
				log.Fatal(err.Error())
			}

			err = autogenBot.Run(specfile, kitfile, autogenOpts)
			if err != nil {
				log.Fatal(err.Error())
			}

			log.InfoC(log.Aurora.Bold(":party_popper:All done"))
		},
	}

	flags := cmd.Flags()
	flags.String("specfile", "", "The specfile with the rules of the packages to autogen.")
	flags.StringP("kitfile", "k", "", "The YAML with the target kit definition.")
	flags.String("to", "workdir", "Override default work directory.")
	flags.String("download-dir", "", "Override the default ${workdir}/downloads directory.")
	flags.Bool("verbose", false, "Show additional informations.")
	flags.Int("deep", 5, "Define the limit of commits to fetch.")
	flags.Int("concurrency", 3, "Define the elaboration concurrency.")
	flags.Bool("skip-reposcan-generation", false,
		"Skip reposcan files generation.")
	flags.Bool("skip-pull-sources", false,
		"Skip pull of sources repositories.")
	flags.Bool("skip-merge", false,
		"Just generate the ebuild without merge to target kit. To use with --keep-workdir.")
	flags.Bool("push", false, "Push commits to origin.")
	flags.Bool("sync", true, "Sync artefacts to S3 backend server.")
	flags.Bool("keep-workdir", false, "Avoid to remove the working directory.")
	flags.Bool("pr", false, "Push commit over specific branch and as Pull Request.")
	flags.Bool("show-values", false,
		"For debug purpose print generated values for any elaborated package in YAML format.")
	flags.Bool("force-merge-check", false,
		"Force merge comparision for package with the same version.")

	flags.String("signature-name", "", "Specify the name of the user for the commits.")
	flags.String("signature-email", "", "Specify the email of the user for the commits.")
	flags.String("github-user", "", "Override the default Github user used for PR.")

	// Sync S3 / Minio backend flags
	flags.String("backend", "dir", "Set the fetcher backend to use: dir|s3.")
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
