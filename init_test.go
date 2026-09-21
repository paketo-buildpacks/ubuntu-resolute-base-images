package acceptance_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/onsi/gomega/format"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/vacation"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var tinyImages struct {
	RunArchive string
}

var baseImages struct {
	BuildArchive string
	RunArchive   string
}

var staticImages struct {
	RunArchive string
}

var tinyImagesNoStacks struct {
	RunArchive string
}

var baseImagesNoStacks struct {
	BuildArchive string
	RunArchive   string
}

var staticImagesNoStacks struct {
	RunArchive string
}

var RegistryUrl string

func by(_ string, f func()) { f() }

func TestAcceptance(t *testing.T) {

	format.MaxLength = 0
	SetDefaultEventuallyTimeout(30 * time.Second)

	Expect := NewWithT(t).Expect

	RegistryUrl = os.Getenv("REGISTRY_URL")
	Expect(RegistryUrl).NotTo(Equal(""))

	root, err := filepath.Abs(".")
	Expect(err).ToNot(HaveOccurred())

	baseImages.BuildArchive = filepath.Join(root, "builds", "resolute-base-images", "build.oci")
	baseImages.RunArchive = filepath.Join(root, "builds", "resolute-base-images", "run.oci")

	tinyImages.RunArchive = filepath.Join(root, "builds", "resolute-tiny-images", "run.oci")

	staticImages.RunArchive = filepath.Join(root, "builds", "resolute-static-images", "run.oci")

	baseImagesNoStacks.BuildArchive = filepath.Join(root, "builds", "resolute-base-images-no-stack-id", "build.oci")
	baseImagesNoStacks.RunArchive = filepath.Join(root, "builds", "resolute-base-images-no-stack-id", "run.oci")

	tinyImagesNoStacks.RunArchive = filepath.Join(root, "builds", "resolute-tiny-images-no-stack-id", "run.oci")

	staticImagesNoStacks.RunArchive = filepath.Join(root, "builds", "resolute-static-images-no-stack-id", "run.oci")

	suite := spec.New("Acceptance", spec.Report(report.Terminal{}), spec.Parallel())
	suite("MetadataBaseImages", testMetadataBaseImages)
	suite("MetadataTinyImages", testMetadataTinyImages)
	suite("MetadataStaticImages", testMetadataStaticImages)
	suite("MetadataBaseImagesNoStacks", testMetadataBaseImagesNoStacks)
	suite("MetadataTinyImagesNoStacks", testMetadataTinyImagesNoStacks)
	suite("MetadataStaticImagesNoStacks", testMetadataStaticImagesNoStacks)
	suite("BuildpackIntegrationBaseImages", testBuildpackIntegrationBaseImages)
	suite("BuildpackIntegrationTinyImages", testBuildpackIntegrationTinyImages)
	suite("BuildpackIntegrationStaticImages", testBuildpackIntegrationStaticImages)
	suite("BuildpackIntegrationBaseImagesNoStacks", testBuildpackIntegrationBaseImagesNoStacks)
	suite("BuildpackIntegrationTinyImagesNoStacks", testBuildpackIntegrationTinyImagesNoStacks)
	suite("BuildpackIntegrationStaticImagesNoStacks", testBuildpackIntegrationStaticImagesNoStacks)
	suite.Run(t)
}

func createBuilder(config string, name string) (string, error) {
	buf := bytes.NewBuffer(nil)

	pack := pexec.NewExecutable("pack")
	err := pack.Execute(pexec.Execution{
		Stdout: buf,
		Stderr: buf,
		Args: []string{
			"builder",
			"create",
			name,
			fmt.Sprintf("--config=%s", config),
		},
	})
	return buf.String(), err
}

func archiveToDaemon(path, id string) error {
	tmpDir, err := os.MkdirTemp("", "oci-archive-")
	if err != nil {
		return fmt.Errorf("unable to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tarReader, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("unable to open tar: %w", err)
	}
	defer tarReader.Close()

	err = vacation.NewTarArchive(tarReader).Decompress(tmpDir)
	if err != nil {
		return fmt.Errorf("unable to extract files: %w", err)
	}

	pathLayout, err := layout.FromPath(tmpDir)
	if err != nil {
		return fmt.Errorf("unable to load image from path %s: %w", tmpDir, err)
	}

	imageIndex, err := pathLayout.ImageIndex()
	if err != nil {
		return fmt.Errorf("unable to read image index: %w", err)
	}

	ref, err := name.ParseReference(id)
	if err != nil {
		return fmt.Errorf("unable to parse reference from %s: %w", id, err)
	}

	return remote.WriteIndex(ref, imageIndex, remote.WithAuthFromKeychain(authn.DefaultKeychain))
}
