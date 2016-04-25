package resourcecatalog

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
)

// ResourceCatalog is a cluster-scoped resource with a specified spec and a controller maintained status
type ResourceCatalog struct {
	unversioned.TypeMeta
	api.ObjectMeta

	Spec   ResourceCatalogSpec
	Status ResourceCatalogStatus
}

type ResourceCatalogSpec struct {
	// Selector is the selector to use when gathering the namespace scope `ResourceCatalogEntries` for
	// the `Status`
	Selector map[string]string
}

type ResourceCatalogStatus struct {
	// Entries is dynamically retrieved by selecting all matching ResourceCatalogEntries from all namespaces
	// This can be done by a controller using a list/watch on all ResourceCatalogEntries based on the spec.selector.
	Entries []ResourceCatalogInfo
}

// ResourceCatalogInfo exists so that the details of how the ResourceCatalogEntry statisfies the request
// can be properly abstracted from whoever claims it.  It also adds a degree of privacy for the
// ResourceCatalogEntry
type ResourceCatalogInfo struct {
	// Namespace is system controlled.  It is the namespace of the originating ResourceCatalogEntry
	Namespace string
	// Name  is system controlled.  It is the name of the ResourceCatalogEntry
	Name string
	// Generation indicates which level of the Entry was accepted
	Generation int

	// Description is client controlled.  It describes the resources being provided.
	Description string
}

// ResourceCatalogEntry is a namespace scoped resource.  In order to be added to a ResourceCatalog the labels
// on the Entry must match the selector of the Catalog
type ResourceCatalogEntry struct {
	unversioned.TypeMeta
	api.ObjectMeta

	Spec   ResourceCatalogEntrySpec
	Status ResourceCatalogEntryStatus
}

// ResourceCatalogEntrySpec gives information about the Entry and describes how a claim would be satisfied.
// Only one of the pointers maybe specified.  Only the Description is exposed.
// An Update can be used to re-detect the exposed resources.  This could ripple through and kick synchronization
// across all claims.  Otherwise, new claims get new resource and old ones keep their existing ones.
type ResourceCatalogEntrySpec struct {
	// Description says what this does
	Description string

	// LocalResources describes the case where resources exist in this namespace that will be expose via a claim
	// to another namespace
	LocalResources *LocalResourceSpec

	// ResourceList describes the case where we have list that we want to create in the claim-namespace, but the
	// resources don't actually exist in this namespace
	ResourceList *ResourceListSpec

	// Extension is a spot for a generic spec.  This would need a custom controller to make use of.  Imagine this
	// as your placeholder for an openshift template.
	Extension runtime.Object
}

// LocalResourceSpec describes the case where resources exist in this namespace that will be expose via a claim
// to another namespace.
type LocalResourceSpec struct {
	// Transitive indicates whether "depends-on" annotations in the linked objects are chased and added to Status
	Transitive bool

	// ExposedResources is the list of resources (name or name/UID) to be included in the exposure
	ExposedResources []api.LocalObjectReference
}

// LocalResourceStatus indicates which resources are actually exposed.
type LocalResourceStatus struct {
	// ExposedResources is the list of resources complete with name/UID matching to be included in the exposure
	// This will include all transitive links based on "depends-on" if chosen in Spec.  The UID MUST be specified
	// so that we don't accidentally expose new resources with the same name as the old resources.
	ExposedResources []api.LocalObjectReference
}

// ResourceListSpec describes the case where we have list that we want to create in the claim-namespace, but the
// resources don't actually exist in this namespace.  You can imagine this being useful for cases where you don't
// want the resources in your namespace.  RC or DaemonSet maybe.
type ResourceListSpec struct {
	// Resource is the list of resources to create to satisfy a claim.
	Resources api.List
}

// ResourceListStatus is for congruence with Spec.  This one ought not need anything
type ResourceListStatus struct {
	// Resource is the list of resources to create to satisfy a claim.
	Resources api.List
}

// ResourceCatalogEntrySpec gives information about the Entry and describes how a claim would be satisfied.
// Only one of the pointers maybe specified.  Only the Description is exposed
type ResourceCatalogEntryStatus struct {
	// Catalogs indicates which catalogs this entry has been selected into
	Catalogs []string
	// Errors indicate any failures during the selection process
	Errors []string
	// Generation gives you enough information to know which one of your status entries has been accepted into
	// the Catalog
	Generation int

	// LocalResources describes the case where resources exist in this namespace that will be expose via a claim
	// to another namespace
	LocalResources *LocalResourceStatus

	// ResourceList describes the case where we have list that we want to create in the claim-namespace, but the
	// resources don't actually exist in this namespace
	ResourceList *ResourceListStatus
}

// ResourceCatalogClaim is used to get resources created in your namespace
type ResourceCatalogClaim struct {
	unversioned.TypeMeta
	api.ObjectMeta

	Spec   ResourceCatalogClaimSpec
	Status ResourceCatalogClaimStatus
}

type ResourceCatalogClaimSpec struct {
	// CatalogName indicates the name of the catalog
	CatalogName string
	// ResourceCatalogEntry indicates the namespace/name of the entry chosen
	ResourceCatalogEntry ObjectReference

	// KeepSynchronized indicates whether or not the claim controller should force your local objects to match
	// the upstream objects.  If this is false, then you'll need to kick `Generation` to force a resync.  If this
	// is true, then you object can change without warning
	KeepSynchronized bool

	// ServiceAccountName specifies which service account the claim controller should act-as when creating resources
	// in your namespace.  This gives project-admins the power to control what a claim can create in their namespace
	ServiceAccountName string

	// AdditionalLabels tells the claim controller to add (and overwrite any existing values) for labels on objects
	// it creates.  This makes it easy to select objects that have been created for a given claim
	AdditionalLabels map[string]string

	// NamePrefix tells the claim controller to add prefix to the name of objects it creates.  Be aware that this can
	// cause resources with long names to fail creation
	NamePrefix string

	// Generation indicates whether spec and status match
	Generation int
}

type ResourceCatalogClaimStatus struct {
	// Errors indicate any failures during the instantiation process.  No cleanup is attempted
	Errors []string

	// CreatedResources indicates objects that were created by this claim.  It is name/UID bound.
	CreatedResources []api.LocalObjectReference

	// EntryGeneration indicates which level of the Entry was used for creation
	EntryGeneration int

	// Generation indicates whether spec and status match
	Generation int
}
