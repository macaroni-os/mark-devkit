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
	"github.com/macaroni-os/mark-devkit/pkg/sourcer"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/macaroni-os/macaronictl/pkg/utils"
)

type RunOpts struct {
	CleanupRootfs bool
	Opts          *executor.FchrootOpts
}

func (m *Metro) RunJob(job *specs.JobRendered, opts *RunOpts) error {
	// Consume source
	sourceHandler := sourcer.NewSourceHandler(m.Config)
	err := sourceHandler.Produce(&job.Source)
	if err != nil {
		return err
	}

	// Prepare rootfs
	rootfsdir := filepath.Join(job.WorkspaceDir, "rootfs")
	if !utils.Exists(rootfsdir) {
		err := helpers.EnsureDir(rootfsdir, 0, 0, 0700)
		if err != nil {
			return err
		}
	}
	if opts.CleanupRootfs {
		defer os.RemoveAll(rootfsdir)
	}
	err = sourceHandler.Consume(&job.Source, rootfsdir)
	if err != nil {
		return err
	}

	// Prepare Host executor
	hostExecutor := executor.NewHostExecutor(m.Config)
	// Prepare fchroot executor
	fchrootExecutor := executor.NewFchrootExecutor(m.Config,
		opts.Opts)

	// Prepare Stdout, Stderr writer
	stdOutWriter := executor.NewExecutorWriter("stdout")
	stdErrWriter := executor.NewExecutorWriter("stderr")

	// Prepare env map from job options
	envMap := job.GetOptionsEnvsMap()

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
			if h.Type != specs.HookOuterPostChroot &&
				h.Type != specs.HookOuterPreChroot {

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

	// Generate tarball

	return nil
}

func (m *Metro) runHook(job *specs.JobRendered, hook *specs.Hook,
	rootfsdir string,
	hostExecutor *executor.HostExecutor,
	fchrootExecutor *executor.FchrootExecutor,
	stdOutWriter, stdErrWriter *executor.ExecutorWriter,
	envMap *map[string]string) error {

	if hook.Type == specs.HookOuterPostChroot ||
		hook.Type == specs.HookOuterPreChroot {

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

	return nil
}
