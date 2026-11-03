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

var _ = Describe("Golang extension gosum_paths", func() {

	var unpackDir, repoDir string

	BeforeEach(func() {
		// Unpack layout of a GitHub tarball: a single top level directory
		// hosting the main module and a nested collector one, each with
		// its own go.sum.
		unpackDir = GinkgoT().TempDir()
		repoDir = filepath.Join(unpackDir, "alloy-1.21.3")
		Expect(os.MkdirAll(filepath.Join(repoDir, "collector"),
			os.ModePerm)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoDir, "go.sum"), []byte(
			"github.com/a/b v1.0.0 h1:aaa=\n"+
				"github.com/a/b v1.0.0/go.mod h1:bbb=\n"),
			0644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoDir, "collector", "go.sum"), []byte(
			"github.com/a/b v1.0.0 h1:aaa=\n"+
				"github.com/e/f v1.2.3 h1:eee=\n"),
			0644)).To(Succeed())
	})

	newExtension := func(opts map[string]string) *ExtensionGolang {
		return &ExtensionGolang{
			ExtensionBase: &ExtensionBase{Opts: opts},
		}
	}

	atom := &specs.AutogenAtom{Name: "test"}

	Describe("collectGoSums", func() {

		It("keeps the default go.sum discovery when gosum_paths is unset", func() {
			goSums, err := newExtension(map[string]string{}).collectGoSums(
				atom, unpackDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(goSums).To(HaveLen(1))
			Expect(goSums[0].Lines).To(HaveLen(2))
		})

		It("builds one GoSum object per defined file, in order", func() {
			goSums, err := newExtension(map[string]string{
				"gosum_paths": "go.sum, collector/go.sum",
			}).collectGoSums(atom, unpackDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(goSums).To(HaveLen(2))
			Expect(goSums[0].Lines).To(HaveLen(2))
			Expect(goSums[1].Lines).To(HaveLen(2))
			Expect(goSums[1].Lines[1].Module).To(Equal("github.com/e/f"))
		})

		It("resolves the paths against the prefixed repository dir", func() {
			goSums, err := newExtension(map[string]string{
				"gosum_paths":          "collector/go.sum",
				"unpack_srcdir_prefix": "alloy-",
			}).collectGoSums(atom, unpackDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(goSums).To(HaveLen(1))
			Expect(goSums[0].Lines[1].Module).To(Equal("github.com/e/f"))
		})

		It("fails when a defined go.sum is missing", func() {
			_, err := newExtension(map[string]string{
				"gosum_paths": "go.sum,nope/go.sum",
			}).collectGoSums(atom, unpackDir)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("downloadBundles", func() {

		It("skips the download when the bundle file is already present", func() {
			// restGuard is nil on purpose: any download attempt would
			// blow up, so the spec passes only when every row is
			// recognised as already bundled.
			bundlesDir := filepath.Join(GinkgoT().TempDir(), "bundles")
			Expect(os.MkdirAll(bundlesDir, os.ModePerm)).To(Succeed())
			for _, bundle := range []string{
				"github.com%2Fa%2Fb%2F@v%2Fv1.0.0.zip",
				"github.com%2Fa%2Fb%2F@v%2Fv1.0.0.mod",
			} {
				Expect(os.WriteFile(filepath.Join(bundlesDir, bundle),
					[]byte("already here"), 0644)).To(Succeed())
			}

			goSum := NewGoSum("github.com/a/b v1.0.0 h1:aaa=\n" +
				"github.com/a/b v1.0.0/go.mod h1:bbb=\n")

			Expect(newExtension(map[string]string{}).downloadBundles(
				nil, atom, goSum, bundlesDir)).To(Succeed())
		})
	})
})
