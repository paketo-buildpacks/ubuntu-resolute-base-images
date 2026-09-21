package acceptance_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"

	"github.com/paketo-buildpacks/occam"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testBuildpackIntegrationStaticImages(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		goDistBuildpack  string
		goBuildBuildpack string

		builderConfigFilepath string

		pack    occam.Pack
		docker  occam.Docker
		source  string
		name    string
		builder string

		image     occam.Image
		container occam.Container

		buildImageID      string
		builderRunImageID string
		runImageID        string
	)

	it.Before(func() {
		var err error

		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()

		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())

		goDistBuildpack = "docker.io/paketobuildpacks/go-dist"
		goBuildBuildpack = "docker.io/paketobuildpacks/go-build"

		source, err = occam.Source(filepath.Join("integration", "testdata", "go_simple_app"))
		Expect(err).NotTo(HaveOccurred())

		builderConfigFile, err := os.CreateTemp("", "builder.toml")
		Expect(err).NotTo(HaveOccurred())
		builderConfigFilepath = builderConfigFile.Name()

		buildImageID = fmt.Sprintf("%s/resolute-base-build-image-%s", RegistryUrl, uuid.NewString())
		builderRunImageID = fmt.Sprintf("%s/resolute-base-run-image-%s", RegistryUrl, uuid.NewString())
		runImageID = fmt.Sprintf("%s/resolute-static-run-image-%s", RegistryUrl, uuid.NewString())

		_, err = fmt.Fprintf(builderConfigFile, `
[build]
  image = "%s:latest"

[run]

  [[run.images]]
    image = "%s:latest"

[[targets]]
  arch = "amd64"
  os = "linux"

[[targets]]
  arch = "arm64"
  os = "linux"
`,
			buildImageID,
			builderRunImageID,
		)
		Expect(err).NotTo(HaveOccurred())

		Expect(archiveToDaemon(baseImages.BuildArchive, buildImageID)).To(Succeed())
		Expect(archiveToDaemon(baseImages.RunArchive, builderRunImageID)).To(Succeed())
		Expect(archiveToDaemon(staticImages.RunArchive, runImageID)).To(Succeed())

		builder = fmt.Sprintf("builder-%s", uuid.NewString())
		logs, err := createBuilder(builderConfigFilepath, builder)
		Expect(err).NotTo(HaveOccurred(), logs)
	})

	it.After(func() {
		Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
		Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
		Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())

		Expect(docker.Image.Remove.Execute(builder)).To(Succeed())
		Expect(os.RemoveAll(builderConfigFilepath)).To(Succeed())

		Expect(docker.Image.Remove.Execute(buildImageID)).To(Succeed())
		Expect(docker.Image.Remove.Execute(builderRunImageID)).To(Succeed())
		Expect(docker.Image.Remove.Execute(runImageID)).To(Succeed())

		Expect(os.RemoveAll(source)).To(Succeed())
	})

	it("builds an app with a buildpack", func() {
		var err error
		var logs fmt.Stringer
		image, logs, err = pack.WithNoColor().Build.
			WithBuildpacks(
				goDistBuildpack,
				goBuildBuildpack,
			).
			WithEnv(map[string]string{
				"BP_LOG_LEVEL":      "DEBUG",
				"CGO_ENABLED":       "0",
				"BP_GO_BUILD_FLAGS": "-buildmode=default",
			}).
			WithPullPolicy("if-not-present").
			WithRunImage(runImageID).
			WithBuilder(builder).
			Execute(name, source)
		Expect(err).ToNot(HaveOccurred(), logs.String)

		Expect(logs.String()).To(ContainSubstring("Using provided run-image '%s'", runImageID))

		container, err = docker.Container.Run.
			WithEnv(map[string]string{"PORT": "8080"}).
			WithPublish("8080").
			WithPublishAll().
			Execute(image.ID)
		Expect(err).NotTo(HaveOccurred())

		Eventually(container).Should(BeAvailable())
		Eventually(container).Should(Serve(MatchRegexp(`go1.*`)).OnPort(8080))
	})
}
