

We can reshape the catalog resource to more easily express them, secure them, and claim them.

# Goals
 1.  cluster-admin should control which catalog exist.
 2.  cluster-admin should be able/disable per-catalog exposure on a per-project granularity.
 3.  project-admin should be able to delegate per-catalog exposure.
 4.  project-editor should be able to claim an entry from the catalog.
 5.  ResourceCatalogClaims should be sync-able, but not require syncing.
 6.  project-admin should be able to control the shape of claim realization and prevent unwanted injection.
 7.  ResourceCatalogEntries should not be restricted to specific claim realization mechanism.

To achieve the ACL related goals, we can create namespaced resources that are selected by label at a cluster level.


# API Objects

## ResourceCatalog
ResourceCatalog is a cluster-scoped resource.  It has a cluster-admin controlled spec that is used to drive a controller.
The spec includes a label selector that is used in a list/watch to filter all ResourceCatalogEntries in the cluster.
The controller will find matching entries and copy a reference into the ResourceCatalog.Status field.  The controller
should have a plug-point for deciding if a given entry is accepted into a given catalog: `Accept(catalogName, entryNamespace, entryName)`
is probably sufficient.

```go
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
	// UID  is system controlled.  It is the UID of the ResourceCatalogEntry
	UID string
	// Generation indicates which level of the Entry was accepted
	Generation int

	// Description is client controlled.  It describes the resources being provided.
	Description string
}
```

## ResourceCatalogEntry
ResourceCatalogEntry is namespace scoped resources.  Its labels must match the selector on a ResourceCatalog
to get included in one or more Catalogs.  
The Spec contains a "must pick one" set of pointers to different kinds of "claim satisfiers".  The built-in
ones will be controller managed for later claim resource instantiation.  I've added two as a for instance,
but you could pretty easily envision a remote endpoint and a PodSpec sort of satisfier as well.
The RESTStorage for a ResourceCatalogEntry is pretty complicated.  In the case
of a LocalResourcSpec, it must locate all resources that will be exposed and run secondary ACL checks to be 
sure that the creator has GET rights on the resources in question.  Then the RESTStorage itself creates the 
Status before persisting.  If you don't do it in the RESTStorage, you'll lose track of who requested it and
you'll be unable to run the proper ACL checks.

```go
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
```

## ResoureCatalogClaim
ResoureCatalogClaim is used to create the resources described in the catalog in a local namespace.  Since the 
resources are created by a privileged controller, a service account is impersonated during creation to allow
control of what is created to remain with a project-admin.
A controller is used to satisfy the Spec and create resources in the local namespace based on the claim 
satisfier described by the ResourceCatalogEntry.
The RESTStorage is responsible for running an ACL check to determine if the creator as rights to point to the
given catalog and to confirm referential integrity of the Entry reference overall.

```go
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
	// ResourceCatalogEntry indicates the namespace/name/UID of the entry chosen
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
```