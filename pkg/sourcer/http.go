/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package sourcer

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	rguard "github.com/geaaru/rest-guard/pkg/guard"
	rgspecs "github.com/geaaru/rest-guard/pkg/specs"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

func (s *SourceHandler) fetchttp(source *specs.JobSource) error {
	u, err := url.Parse(source.Uri)
	if err != nil {
		return err
	}

	if utils.Exists(source.Target) {
		s.Logger.Info(
			fmt.Sprintf(
				":gem_stone:File %s is already present. Nothing to do.",
				source.Target,
			))
		return nil
	}

	rconfig := rgspecs.NewConfig()
	guard, err := rguard.NewRestGuard(rconfig)
	if err != nil {
		return err
	}

	service := rgspecs.NewRestService(u.Host)
	service.Retries = 2

	ssl := false
	if u.Scheme == "https" {
		ssl = true
	}
	node := rgspecs.NewRestNode(u.Host, u.Host, ssl)

	guard.AddService(service.GetName(), service)
	guard.AddRestNode(service.GetName(), node)

	t := service.GetTicket()
	defer t.Rip()

	_, err = guard.CreateRequest(t, "GET", u.Path)
	if err != nil {
		return err
	}

	s.Logger.Info(
		fmt.Sprintf(":truck:Downloading file %s...",
			filepath.Base(u.Path),
		))

	err = guard.Do(t)
	if err != nil {
		httpStatusCode := t.GetResponseStatusCode(500)
		s.Logger.Error(fmt.Sprintf(
			"Error on download source %s - (%v): %s",
			source.Uri, httpStatusCode, err.Error()))
		return err
	}

	targetdir := filepath.Dir(source.Target)
	if !utils.Exists(targetdir) {
		err := os.MkdirAll(targetdir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Create target file
	dst, err := os.Create(source.Target)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy the response to file
	if _, err = io.Copy(dst, t.Response.Body); err != nil {
		return err
	}

	return nil
}
