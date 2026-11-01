/*
Copyright © 2024-2026 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package extensions

import (
	"os"
	"path/filepath"

	"github.com/macaroni-os/mark-devkit/pkg/specs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Golang extension go.sum merge", func() {

	Describe("mergeGoSums", func() {

		It("appends only the rows missing from main", func() {
			main := NewGoSum("github.com/a/b v1.0.0 h1:aaa=\n" +
				"github.com/a/b v1.0.0/go.mod h1:bbb=\n")
			extra := NewGoSum("github.com/a/b v1.0.0 h1:aaa=\n" +
				"github.com/c/d v0.2.0 h1:ccc=\n" +
				"github.com/c/d v0.2.0/go.mod h1:ddd=\n")

			added := mergeGoSums(main, extra)

			Expect(added).To(Equal(2))
			Expect(main.Lines).To(HaveLen(4))
			Expect(main.Lines[2]).To(Equal(GoSumRow{
				Module:   "github.com/c/d",
				Version:  "v0.2.0",
				Checksum: "h1:ccc=",
			}))
		})

		It("is idempotent on repeated merges", func() {
			main := NewGoSum("github.com/a/b v1.0.0 h1:aaa=\n")
			extra := NewGoSum("github.com/c/d v0.2.0 h1:ccc=\n")

			Expect(mergeGoSums(main, extra)).To(Equal(1))
			Expect(mergeGoSums(main, extra)).To(Equal(0))
			Expect(main.Lines).To(HaveLen(2))
		})
	})

	Describe("mergeExtraGoSums", func() {

		var dir string

		BeforeEach(func() {
			dir = GinkgoT().TempDir()
			Expect(os.MkdirAll(filepath.Join(dir, "collector"),
				os.ModePerm)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(dir, "go.sum"), []byte(
				"github.com/a/b v1.0.0 h1:aaa=\n"+
					"github.com/a/b v1.0.0/go.mod h1:bbb=\n"),
				0644)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(dir, "collector", "go.sum"), []byte(
				"github.com/a/b v1.0.0 h1:aaa=\n"+
					"github.com/e/f v1.2.3 h1:eee=\n"),
				0644)).To(Succeed())
		})

		newExtension := func(opts string) *ExtensionGolang {
			return &ExtensionGolang{
				ExtensionBase: &ExtensionBase{
					Opts: map[string]string{"extra_go_sums": opts},
				},
			}
		}

		readMainSum := func() *GoSum {
			data, err := os.ReadFile(filepath.Join(dir, "go.sum"))
			Expect(err).ToNot(HaveOccurred())
			return NewGoSum(string(data))
		}

		atom := &specs.AutogenAtom{Name: "test"}

		It("resolves the relative paths against the main go.sum dir", func() {
			merged, err := newExtension("collector/go.sum").mergeExtraGoSums(
				atom, dir, readMainSum())
			Expect(err).ToNot(HaveOccurred())
			Expect(merged.Lines).To(HaveLen(3))
			Expect(merged.Lines[2].Module).To(Equal("github.com/e/f"))
		})

		It("does nothing when the option is not set", func() {
			ext := &ExtensionGolang{
				ExtensionBase: &ExtensionBase{Opts: map[string]string{}},
			}
			merged, err := ext.mergeExtraGoSums(atom, dir, readMainSum())
			Expect(err).ToNot(HaveOccurred())
			Expect(merged.Lines).To(HaveLen(2))
		})

		It("fails on a missing extra go.sum", func() {
			_, err := newExtension("nope/go.sum").mergeExtraGoSums(
				atom, dir, readMainSum())
			Expect(err).To(HaveOccurred())
		})
	})
})
