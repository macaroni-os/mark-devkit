/*
Copyright © 2024-2026 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package extensions

import (
	"testing"

	"github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

func TestGolangExtension(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Golang extension Suite")
}

var _ = BeforeSuite(func() {
	// The extension helpers log through the default logger; give them an
	// initialized instance to work on.
	logger.NewMarkDevkitLogger(
		specs.NewMarkDevkitConfig(viper.New())).SetAsDefault()
})
