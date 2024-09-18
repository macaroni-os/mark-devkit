/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package metro

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/macaroni-os/mark-devkit/pkg/executor"
	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/packer"
	"github.com/macaroni-os/mark-devkit/pkg/sourcer"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/macaroni-os/macaronictl/pkg/utils"
)

type RunOpts struct {
	CleanupRootfs bool
	SkipSource    bool
	SkipPacker    bool
	SkipHooks     bool
	Quiet         bool
	Opts          *executor.FchrootOpts
}

func (m *Metro) RunJob(job *specs.JobRendered, opts *RunOpts) error {
	var err error

	m.Logger.Info(fmt.Sprintf(
		":rocket:Starting job %s...",
		job.Name))

	// Prepare rootfs
	rootfsdir := filepath.Join(job.WorkspaceDir, "rootfs")
	m.Logger.Info(fmt.Sprintf(":screwdriver:Preparing rootfs %s...",
		rootfsdir))

	if !utils.Exists(rootfsdir) {
		err = helpers.EnsureDir(rootfsdir, 0, 0, 0700)
		if err != nil {
			return err
		}
	} else if opts.CleanupRootfs {

		m.Logger.Info(fmt.Sprintf(":knife:Cleanup existing rootfs %s...",
			rootfsdir))
		err = os.RemoveAll(rootfsdir)
		if err != nil {
			return err
		}

		err = helpers.EnsureDir(rootfsdir, 0, 0, 0700)
		if err != nil {
			return err
		}

	}
	if opts.CleanupRootfs {
		defer os.RemoveAll(rootfsdir)
	}

	// Ensure that job chroot binds exists
	for _, bind := range job.ChrootBinds {
		if !utils.Exists(bind.Source) {
			err := os.MkdirAll(bind.Source, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	// Prepare Host executor
	hostExecutor := executor.NewHostExecutor(m.Config)
	// Prepare fchroot executor
	fchrootExecutor := executor.NewFchrootExecutor(m.Config,
		opts.Opts)
	if opts.Quiet {
		hostExecutor.Quiet = true
		fchrootExecutor.Quiet = true
	}

	// Prepare Stdout, Stderr writer
	stdOutWriter := executor.NewExecutorWriter("stdout", opts.Quiet)
	stdErrWriter := executor.NewExecutorWriter("stderr", opts.Quiet)

	// Prepare env map from job options
	envMap := job.GetOptionsEnvsMap()

	// Add special envs
	envMap["MARKDEVKIT_VERSION"] = specs.MARKDEVKIT_VERSION
	envMap["MARKDEVKIT_WORKSPACE"] = job.WorkspaceDir
	envMap["MARKDEVKIT_ROOTFS"] = rootfsdir

	// Get TERM and COLORTERM from env
	// TODO: I haven't find a complete workaround to
	//       simulate an interactive shell with colors
	//       without pty
	envMap["TERM"] = os.Getenv("TERM")
	if os.Getenv("COLORTERM") != "" {
		envMap["COLORTERM"] = os.Getenv("COLORTERM")
	}
	if os.Getenv("LS_COLORS") != "" {
		envMap["LS_COLORS"] = os.Getenv("LS_COLORS")
	}

	m.Logger.Info(fmt.Sprintf(
		":wrench:Rootfs ready for hooks!"))

	if !opts.SkipHooks {

		// Run pre sourcer hooks
		preSourcerHooks := job.GetPreSourcesHooks()
		if len(*preSourcerHooks) > 0 {

			for _, hook := range *preSourcerHooks {
				err := m.runHook(job, hook, rootfsdir,
					hostExecutor, fchrootExecutor,
					stdOutWriter, stdErrWriter,
					&envMap,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	// Consume source
	sourceHandler := sourcer.NewSourceHandler(m.Config)
	if !opts.SkipSource {
		err = sourceHandler.Produce(&job.Source)
		if err != nil {
			return err
		}
	}

	if !opts.SkipSource {
		err = sourceHandler.Consume(&job.Source, rootfsdir)
		if err != nil {
			return err
		}
	}

	if !opts.SkipHooks {
		// Run pre chroot hooks
		preChrootHooks := job.GetPreChrootHooks()
		if len(*preChrootHooks) > 0 {

			for _, hook := range *preChrootHooks {
				err := m.runHook(job, hook, rootfsdir,
					hostExecutor, fchrootExecutor,
					stdOutWriter, stdErrWriter,
					&envMap,
				)
				if err != nil {
					return err
				}
			}

		}

		// Run inner-chroot and/or outer-chroot
		for _, hf := range job.HookFile {
			for _, h := range hf.Hooks {
				if h.Type == specs.HookInnerChroot ||
					h.Type == specs.HookOuterChroot {

					err := m.runHook(job, &h, rootfsdir,
						hostExecutor, fchrootExecutor,
						stdOutWriter, stdErrWriter,
						&envMap,
					)
					if err != nil {
						return err
					}
				}
			}
		}

		// Run post chroot hooks
		postChrootHooks := job.GetPostChrootHooks()
		if len(*postChrootHooks) > 0 {
			for _, hook := range *postChrootHooks {
				err := m.runHook(job, hook, rootfsdir,
					hostExecutor, fchrootExecutor,
					stdOutWriter, stdErrWriter,
					&envMap,
				)
				if err != nil {
					return err
				}
			}

		}
	}

	if !opts.SkipPacker {
		// Generate tarball
		packerHandler := packer.NewPacker(m.Config)
		err = packerHandler.Produce(rootfsdir, &job.Output)
		if err != nil {
			return err
		}

		// Run post chroot hooks
		postOutputHooks := job.GetPostOutputHooks()
		if len(*postOutputHooks) > 0 {
			for _, hook := range *postOutputHooks {
				err := m.runHook(job, hook, rootfsdir,
					hostExecutor, fchrootExecutor,
					stdOutWriter, stdErrWriter,
					&envMap,
				)
				if err != nil {
					return err
				}
			}

		}

	}

	return nil
}

func (m *Metro) runHook(job *specs.JobRendered, hook *specs.Hook,
	rootfsdir string,
	hostExecutor *executor.HostExecutor,
	fchrootExecutor *executor.FchrootExecutor,
	stdOutWriter, stdErrWriter *executor.ExecutorWriter,
	envMap *map[string]string) error {

	m.Logger.Info(
		fmt.Sprintf(":magic_wand:>>> Running hook %s",
			hook.Name))

	if hook.Type == specs.HookOuterPostChroot ||
		hook.Type == specs.HookOuterPreChroot ||
		hook.Type == specs.HookOuterPreSourcer ||
		hook.Type == specs.HookOuterPostOutput {

		for _, command := range hook.Commands {
			res, err := hostExecutor.RunCommandWithOutput(
				command, *envMap,
				stdOutWriter, stdErrWriter,
				hook.Entrypoint,
			)
			if err != nil {
				return err
			}

			if res > 0 {
				return fmt.Errorf("Hook %s exiting with %d",
					hook.Name, res)
			}
		}

	} else {
		binds := job.GetBindsMap()

		if len(hook.Binds) > 0 {
			for _, b := range hook.Binds {
				binds[b.Source] = b.Target
			}
		}

		for _, command := range hook.Commands {

			res, err := fchrootExecutor.RunCommandWithOutput(
				command, *envMap,
				stdOutWriter, stdErrWriter,
				hook.Entrypoint,
				rootfsdir,
				binds)
			if err != nil {
				return err
			}

			if res > 0 {
				return fmt.Errorf("Hook %s exiting with %d",
					hook.Name, res)
			}
		}

	}

	m.Logger.Info(
		fmt.Sprintf(":magic_wand:>>> Completed hook %s :check_mark:",
			hook.Name))

	return nil
}
