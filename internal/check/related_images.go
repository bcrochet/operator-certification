package check

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/opdev/knex"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	mimage "github.com/operator-framework/operator-manifest-tools/pkg/image"
	"github.com/operator-framework/operator-manifest-tools/pkg/pullspec"
	"sigs.k8s.io/yaml"
)

var _ knex.Check = &RelatedImagesCheck{}

type RelatedImagesCheck struct{}

func (p *RelatedImagesCheck) Validate(ctx context.Context, imgRef knex.ImageReference) (bool, error) {
	images, manifests, err := p.dataToValidate(ctx, imgRef.ImageFSPath)
	if err != nil {
		return false, err
	}

	return p.validate(ctx, images, manifests)
}

//nolint:unparam // ctx is unused. Keep for future use.
func (p *RelatedImagesCheck) dataToValidate(ctx context.Context, imagePath string) ([]string, map[string]struct{}, error) {
	operatorManifests, err := pullspec.FromDirectory(filepath.Join(imagePath, "manifests"), pullspec.DefaultHeuristic)
	if err != nil {
		return nil, nil, err
	}

	imageNames, err := mimage.Extract(operatorManifests)
	if err != nil {
		return nil, nil, err
	}

	relatedImages := make(map[string]struct{}, 0)
	for _, manifest := range operatorManifests {
		if !manifest.HasRelatedImages() {
			continue
		}
		csvBytes, err := manifest.ToYaml()
		if err != nil {
			return nil, nil, fmt.Errorf("could not get manifest yaml: %v", err)
		}
		var csv operatorsv1alpha1.ClusterServiceVersion
		err = yaml.Unmarshal(csvBytes, &csv)
		if err != nil {
			return nil, nil, fmt.Errorf("malformed CSV detected: %v", err)
		}

		for _, relatedImage := range csv.Spec.RelatedImages {
			relatedImages[relatedImage.Image] = struct{}{}
		}
	}

	return imageNames, relatedImages, nil
}

//nolint:unparam // ctx is unused. Keep for future use.
func (p *RelatedImagesCheck) validate(ctx context.Context, images []string, relatedImages map[string]struct{}) (bool, error) {
	logger := logr.FromContextOrDiscard(ctx)

	for _, image := range images {
		if _, ok := relatedImages[image]; !ok {
			logger.Info(fmt.Sprintf("warning: image %s is not in relatedImages. This will eventually cause this check to fail", image))
		}
	}
	return true, nil
}

func (p *RelatedImagesCheck) Name() string {
	return "AllImageRefsInRelatedImages"
}

func (p *RelatedImagesCheck) Metadata() knex.Metadata {
	return knex.Metadata{
		Description:      "Check that all images in the CSV are listed in RelatedImages section. Currently, this check is not enforced.",
		Level:            "optional",
		KnowledgeBaseURL: "https://access.redhat.com/documentation/en-us/red_hat_software_certification/8.45/html/red_hat_openshift_software_certification_policy_guide/assembly-products-managed-by-an-operator_openshift-sw-cert-policy-container-images#con-operator-requirements_openshift-sw-cert-policy-products-managed",
		CheckURL:         "https://access.redhat.com/documentation/en-us/red_hat_software_certification/8.45/html/red_hat_openshift_software_certification_policy_guide/assembly-products-managed-by-an-operator_openshift-sw-cert-policy-container-images#con-operator-requirements_openshift-sw-cert-policy-products-managed",
	}
}

func (p *RelatedImagesCheck) Help() knex.HelpText {
	return knex.HelpText{
		Message:    "Check that all images referenced in the CSV are in RelatedImages",
		Suggestion: "Either manually or with a tool, populate the RelatedImages section of the CSV",
	}
}
