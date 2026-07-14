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
	"github.com/google/uuid"
	"github.com/onsi/gomega/format"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/vacation"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var tinyImages struct {
	RunArchive string
	RunImageID string
}

var baseImages struct {
	BuildArchive string
	RunArchive   string
	BuildImageID string
	RunImageID   string
}

var staticImages struct {
	RunArchive string
	RunImageID string
}

var tinyImagesNoStacks struct {
	RunArchive string
	RunImageID string
}

var baseImagesNoStacks struct {
	BuildArchive string
	RunArchive   string
	BuildImageID string
	RunImageID   string
}

var staticImagesNoStacks struct {
	RunArchive string
	RunImageID string
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
	baseImages.BuildImageID = fmt.Sprintf("%s/resolute-base-build-image-%s", RegistryUrl, uuid.NewString())

	baseImages.RunArchive = filepath.Join(root, "builds", "resolute-base-images", "run.oci")
	baseImages.RunImageID = fmt.Sprintf("%s/resolute-base-run-image-%s", RegistryUrl, uuid.NewString())

	tinyImages.RunArchive = filepath.Join(root, "builds", "resolute-tiny-images", "run.oci")
	tinyImages.RunImageID = fmt.Sprintf("%s/resolute-tiny-run-image-%s", RegistryUrl, uuid.NewString())

	staticImages.RunArchive = filepath.Join(root, "builds", "resolute-static-images", "run.oci")
	staticImages.RunImageID = fmt.Sprintf("%s/resolute-static-run-image-%s", RegistryUrl, uuid.NewString())

	baseImagesNoStacks.BuildArchive = filepath.Join(root, "builds", "resolute-base-images-no-stack-id", "build.oci")
	baseImagesNoStacks.BuildImageID = fmt.Sprintf("%s/resolute-base-build-image-no-stacks-%s", RegistryUrl, uuid.NewString())

	baseImagesNoStacks.RunArchive = filepath.Join(root, "builds", "resolute-base-images-no-stack-id", "run.oci")
	baseImagesNoStacks.RunImageID = fmt.Sprintf("%s/resolute-base-run-image-no-stacks-%s", RegistryUrl, uuid.NewString())

	tinyImagesNoStacks.RunArchive = filepath.Join(root, "builds", "resolute-tiny-images-no-stack-id", "run.oci")
	tinyImagesNoStacks.RunImageID = fmt.Sprintf("%s/resolute-tiny-run-image-no-stacks-%s", RegistryUrl, uuid.NewString())

	staticImagesNoStacks.RunArchive = filepath.Join(root, "builds", "resolute-static-images-no-stack-id", "run.oci")
	staticImagesNoStacks.RunImageID = fmt.Sprintf("%s/resolute-static-run-image-no-stacks-%s", RegistryUrl, uuid.NewString())

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
	tmpDir := os.TempDir()

	tarReader, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("unable to open tar: %w", err)
	}

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
