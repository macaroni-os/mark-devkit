/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs_test

import (
	"os"
	"strings"

	. "github.com/macaroni-os/mark-devkit/pkg/specs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v "github.com/spf13/viper"
)

var _ = Describe("Specs Test", func() {

	Context("Config1", func() {
		os.Setenv("MARKDEVKIT_GENERAL__DEBUG", "true")
		config := NewMarkDevkitConfig(v.New())
		// Set env variable
		config.Viper.SetEnvPrefix(MARKDEVKIT_ENV_PREFIX)
		config.Viper.BindEnv("config")
		config.Viper.SetDefault("config", "")
		config.Viper.SetDefault("etcd-config", false)

		config.Viper.AutomaticEnv()

		// Create EnvKey Replacer for handle complex structure
		replacer := strings.NewReplacer(".", "__")
		config.Viper.SetEnvKeyReplacer(replacer)

		err := config.Unmarshal()

		It("Convert env1", func() {

			Expect(err).Should(BeNil())
			Expect(config.GetGeneral().Debug).To(Equal(true))
		})

	})

})
