package check

import (
	"context"

	libcsv "github.com/opdev/operator-certification/internal/csv"

	"github.com/go-logr/logr"
	"github.com/opdev/knex"
	"github.com/operator-framework/api/pkg/manifests"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
)

var _ knex.Check = &FollowsRestrictedNetworkEnablementGuidelines{}

type FollowsRestrictedNetworkEnablementGuidelines struct{}

func (p FollowsRestrictedNetworkEnablementGuidelines) Validate(ctx context.Context, imgRef knex.ImageReference) (bool, error) {
	return p.validate(ctx, imgRef.ImageFSPath)
}

//nolint:unparam // ctx is unused. Keep for future use.
func (p FollowsRestrictedNetworkEnablementGuidelines) getBundleCSV(ctx context.Context, bundlepath string) (*operatorsv1alpha1.ClusterServiceVersion, error) {
	bundle, err := manifests.GetBundleFromDir(bundlepath)
	if err != nil {
		return nil, err
	}
	return bundle.CSV, nil
}

func (p FollowsRestrictedNetworkEnablementGuidelines) validate(ctx context.Context, bundledir string) (bool, error) {
	logger := logr.FromContextOrDiscard(ctx)
	csv, err := p.getBundleCSV(ctx, bundledir)
	if err != nil {
		return false, err
	}

	// If the CSV does not claim to support disconnected environments, there's no reason to check that it followed guidelines.
	if !libcsv.HasInfrastructureFeaturesAnnotation(csv) {
		logger.Info("this operator does not have the infrastructure-features annotation. You can safely ignore this if you your operator is not intended to be restricted-network aware.")
		return false, nil
	}

	if !libcsv.SupportsDisconnected(csv.Annotations[libcsv.InfrastructureFeaturesAnnotation]) {
		logger.Info("the infrastructure-features enabled for this operator did not include the \"disconnected\" identifier. You can safely ignore this if you your operator is not intended to be restricted-network aware.")
		return false, nil
	}

	// You must have at least one related image (your controller manager) in order to be considered restricted-network ready
	if !libcsv.HasRelatedImages(csv) {
		logger.Info("this operator did not have any related images, and at least one is expected")
		return false, nil
	}

	// All related images must be pinned. No tag references.
	if !libcsv.RelatedImagesArePinned(csv.Spec.RelatedImages) {
		logger.Info("a related image is not pinned to a digest reference of the same image, and this is required.")
		return false, nil
	}

	// Some environment variables should be passed into the environment using the RELATED_IMAGE_ prefix.
	// This isn't the only way to pass this kind of information to the controller, but it is the suggested way in our
	// documentation.
	deploymentSpecs := make([]appsv1.DeploymentSpec, 0, len(csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs))
	for _, ds := range csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
		deploymentSpecs = append(deploymentSpecs, ds.Spec)
	}
	relatedImagesInContainerEnvironment := libcsv.RelatedImageReferencesInEnvironment(deploymentSpecs...)
	if len(relatedImagesInContainerEnvironment) == 0 {
		logger.Info("no environment variables prefixed with \"RELATED_IMAGE_\" were found in your operator's container definitions. These are expected to pass through values into your controller's runtime environment.")
		return false, nil
	}

	return true, nil
}

func (p FollowsRestrictedNetworkEnablementGuidelines) Name() string {
	return "FollowsRestrictedNetworkEnablementGuidelines"
}

func (p FollowsRestrictedNetworkEnablementGuidelines) Metadata() knex.Metadata {
	return knex.Metadata{
		Description: "Checks for indicators that this bundle has implemented guidelines to indicate readiness for running in a disconnected cluster, or a cluster with a restricted network.",
		// TODO: If this check is enforced and no longer optional, we need to identify ways to reduce false failures that may be caused by
		// developers injecting related images in other ways.
		Level:            "optional",
		KnowledgeBaseURL: "https://access.redhat.com/documentation/en-us/red_hat_software_certification/8.45/html/red_hat_openshift_software_certification_policy_guide/assembly-products-managed-by-an-operator_openshift-sw-cert-policy-container-images#con-operator-requirements_openshift-sw-cert-policy-products-managed",
		CheckURL:         "https://access.redhat.com/documentation/en-us/red_hat_software_certification/8.45/html/red_hat_openshift_software_certification_policy_guide/assembly-products-managed-by-an-operator_openshift-sw-cert-policy-container-images#con-operator-requirements_openshift-sw-cert-policy-products-managed",
	}
}

func (p FollowsRestrictedNetworkEnablementGuidelines) Help() knex.HelpText {
	return knex.HelpText{
		Message:    "Check for the implementation of guidelines indicating operator readiness for environments with restricted networking.",
		Suggestion: "If consumers of your operator may need to do so on a restricted network, implement the guidelines outlines in OCP documentation for your cluster version, such as https://docs.openshift.com/container-platform/4.11/operators/operator_sdk/osdk-generating-csvs.html#olm-enabling-operator-for-restricted-network_osdk-generating-csvs for OCP 4.11",
	}
}
