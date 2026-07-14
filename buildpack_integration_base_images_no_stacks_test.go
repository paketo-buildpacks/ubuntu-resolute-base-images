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

func testBuildpackIntegrationBaseImagesNoStacks(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		mavenBuildpack         string
		jvmBuildpack           string
		syftBuildpack          string
		executableJarBuildpack string

		builderConfigFilepath string

		pack    occam.Pack
		docker  occam.Docker
		source  string
		name    string
		builder string

		image     occam.Image
		container occam.Container
	)

	it.Before(func() {
		var err error

		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()

		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())

		mavenBuildpack = "docker.io/paketobuildpacks/maven"
		jvmBuildpack = "docker.io/paketobuildpacks/sap-machine"
		syftBuildpack = "docker.io/paketobuildpacks/syft"
		executableJarBuildpack = "docker.io/paketobuildpacks/executable-jar"

		source, err = occam.Source(filepath.Join("integration", "testdata", "java_simple_app"))
		Expect(err).NotTo(HaveOccurred())

		builderConfigFile, err := os.CreateTemp("", "builder.toml")
		Expect(err).NotTo(HaveOccurred())
		builderConfigFilepath = builderConfigFile.Name()

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
			baseImagesNoStacks.BuildImageID,
			baseImagesNoStacks.RunImageID,
		)
		Expect(err).NotTo(HaveOccurred())

		Expect(archiveToDaemon(baseImagesNoStacks.BuildArchive, baseImagesNoStacks.BuildImageID)).To(Succeed())
		Expect(archiveToDaemon(baseImagesNoStacks.RunArchive, baseImagesNoStacks.RunImageID)).To(Succeed())

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

		Expect(docker.Image.Remove.Execute(baseImagesNoStacks.BuildImageID)).To(Succeed())
		Expect(docker.Image.Remove.Execute(baseImagesNoStacks.RunImageID)).To(Succeed())

		Expect(os.RemoveAll(source)).To(Succeed())
	})

	it("builds an app with a buildpack", func() {
		var err error
		var logs fmt.Stringer

		image, logs, err = pack.WithNoColor().Build.
			WithPullPolicy("if-not-present").
			WithBuildpacks(
				jvmBuildpack,
				syftBuildpack,
				mavenBuildpack,
				executableJarBuildpack,
			).
			WithEnv(map[string]string{
				"BP_LOG_LEVEL": "DEBUG",
			}).
			WithBuilder(builder).
			Execute(name, source)
		Expect(err).ToNot(HaveOccurred(), logs.String)

		container, err = docker.Container.Run.
			WithEnv(map[string]string{"PORT": "8080"}).
			WithPublish("8080").
			WithPublishAll().
			Execute(image.ID)
		Expect(err).NotTo(HaveOccurred())

		Eventually(container).Should(BeAvailable())
		Eventually(container).Should(Serve(ContainSubstring("Hello World! Java version")).OnPort(8080))
	})
}
