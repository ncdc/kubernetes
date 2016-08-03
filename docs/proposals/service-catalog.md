<!-- BEGIN MUNGE: UNVERSIONED_WARNING -->

<!-- BEGIN STRIP_FOR_RELEASE -->

<img src="http://kubernetes.io/img/warning.png" alt="WARNING"
     width="25" height="25">
<img src="http://kubernetes.io/img/warning.png" alt="WARNING"
     width="25" height="25">
<img src="http://kubernetes.io/img/warning.png" alt="WARNING"
     width="25" height="25">
<img src="http://kubernetes.io/img/warning.png" alt="WARNING"
     width="25" height="25">
<img src="http://kubernetes.io/img/warning.png" alt="WARNING"
     width="25" height="25">

<h2>PLEASE NOTE: This document applies to the HEAD of the source tree</h2>

If you are using a released version of Kubernetes, you should
refer to the docs that go with that version.

<strong>
The latest release of this document can be found
[here](http://releases.k8s.io/release-1.1/docs/proposals/service-catalog.md).

Documentation for other releases can be found at
[releases.k8s.io](http://releases.k8s.io).
</strong>
--

<!-- END STRIP_FOR_RELEASE -->

<!-- END MUNGE: UNVERSIONED_WARNING -->

# Abstract

A new catalog concept is proposed for sharing reusable recipes for services,
the containers that back them, and configuration data associated with them.
Users will be able to publish recipes to the catalog and browse the catalog
for recipes to use.

# Motivation

Users don’t like reinventing the wheel. Most users would prefer to be able to
search for and use something like a database template to run their own
database over doing the work necessary to create a custom solution.  There are
a number of pieces attendant to such an effort: a Service to provide a stable
network endpoint for the database, a Deployment to back that endpoint with
database containers, and possibly a ConfigMap and or Secret to store
configuration information and credentials about the database.

If you assume that most or all namespaces and the resources in them are
private, we need a way for users to be able to share resources with others.
Role-based access control (RBAC) could help, as it allows users to control who
can access resources in a namespace, but that pertains to existing resources.
If you want create a pre-canned way of running something like a database and
share that with others, that’s not RBAC against existing resources; that’s
something more akin to publishing a "recipe" to a searchable catalog.

This document describes a new “Catalog” concept for sharing reusable
resources.

# Constraints and assumptions

# Use cases

1.  Advertising and discovering services and recipes:
    1.  As a service operator, I want to be able to label my service offerings
        and recipes, so that users can search for my services and recipes
        according to labels, without knowing about my service to begin with
    2.  As a user, I want to be able to search for services that are shared that
        I can consume in order to locate the right service to use in my
        application
2.  Recipes for running software systems in Kubernetes:
    1.  As someone who created a recipe for running a software system in
        Kubernetes, I want to share this recipe with others so that they can
        easily stand up their own copy
    2.  As someone who wants to run a particular software system in Kubernetes,
        I want to be able to search for and use recipes that others may have
        already created, so I can avoid spending time getting it to run myself
3.  Sharing resources for a service:
    1.  As an operator of a software system, I want to share the resources that
        are required to use the system so that my users can easily consume
        them in their own namespaces
    2.  As a user of a software system running in Kubernetes, I want to consume
        the shared resources associated with that system in my own namespace so
        that I can use the system in my application
4.  Sharing unique resources for a service:
    1.  As an operator of a software system, I want to be able to generate a
        unique resource for each user that wants to use the system so that I can
        manage permissions granularly
    2.  As a user of a software system, I want to get a unique set of resources
        for system in my namespace so that I can use the system in my
        application
5.  Policy for viewing and using services
    1.  As a service operator, I want to be able to describe a policy for who is
        able to view my service, so that I can ensure that only the right users
        have access to see that my service exists
    2.  As a service operator, I want to be able to describe a policy for who is
        able to use my service, so that I can ensure that users have the right
        degree of autonomy when using my service
6.  Consuming services - visibility
    1.  As a service operator, I want to be able to see the users consuming my
        service and track their usage of my service, so that I can be aware of
        the consumers of the service and charge them according to their usage
    2.  As an developer consuming services, I want to be able to see the
        services are being consumed, so that I can ensure that I am consuming
        only services that I need
7.  Resource provisioning
    1.  As a service operator, I want to be able to provision new resources for
        a user when they begin to consume my service so that the user will have
        their own resources to use for my service

### Use case write-ups

# Analysis

## Related Work

### Helm

[Helm](https://github.com/kubernetes/helm) uses “charts” (e.g. resource
templates) that can instantiated as “releases”. Charts are stored in a HTTP
repository as compressed tarballs.  The tarball contains, among other files:

- `chart.yaml` - metadata about the chart
- `values.yaml` -  default variable values to use when processing the resource
   template files
- `templates/` - Kubernetes resource templates processed at `helm install` time

Workflow:

- `helm init` -  starts Tiller in the cluster and initializes CLI environment.
   Tiller runs in the cluster and manages the running Releases.
- `helm install <chartname>` - downloads the chart, processes the resource
   templates, and instantiates the resources in a release

Helm addresses use case 2.1: it provides a way to both create templates and
search for/consume them.  Charts can also incorporate secrets, however, secrets
created as part of a Helm release are not updated when the secret in the chart
is modified.

It is also possible to address use case 4.2; a user could take advantage of
Helm’s support for pre hooks to e.g. run a job that requests that a new
username/password be provisioned. The job, however, would need to have
sufficient privileges to create a username/password, which means the
administrative credentials necessary to provision the new account would need to
be included in the chart itself, which is a security risk and presumably not a
viable solution. Alternatively, the job would need to make an unprivileged
request to some other service that does have sufficient privileges.

TODO
Also: 7.1, use pre-hooks to provision resources

### Docker Compose

Docker Compose is most like a pod spec file.  It can contain information about
container(s) to be deployed together (similar to a pod) and the volumes,
environment variables, and ports that the combined composition will use.

Compose is similar to Helm in that it can bring up and take down a templated set
of resources as a unit.

### OpenShift Templates

Openshift templates offer a feature set similar to Helm.  Templates are
resources created as yaml files and imported as a resource of type "Template"
into Openshift.  Templates are processed with the "oc process" command, filling
in fields from parameter list and can even dynamically generate parameters like
passwords, for example. Once processed, the template becomes a single yaml file
output that can be fed to `oc create -f` to create the resources in Openshift.

The use case application is the same as that of Helm.  Once resources created by
the template, they are not tracked any further as a member of the template.  It
is up to the users to apply labels such that all resources created by a template
can be selected in the cluster.

### Cloud Foundry Service Broker

The CF Service Broker is more full featured than the previous examples, in that
it implements a catalog of services (CCDB) provided by any number of service
brokers and claimable by cluster users.  The cloud controller can discover
services from service brokers that implementto toss it. the Service Broker API
using the "catalog" call.  Other calls that happen in response to a user
claiming the service from the catalog are "provision", "bind", "unbind", and
"deprovision".  The separation of the provisioning stage from the binding stage
allows for asynchronous dynamic provisioning of a service for a particular user.

This satisfies all the use cases except for the syncing of changed credentials.
There is no mechanism for the Service Broker to initiate a deprovision or
resync.

## The Thing

## The Prototype Thing

<!--- -->

### Shared development database

A development team is working on an application that uses a database. The IT department manages the
database (i.e., it lives off-cluster). All developers share the same credentials to access the
database, but these credentials are managed by IT. Rather than having each developer create his or
her own `Service` and `Secret` to connect to the database, IT creates a "db-app-xyz" `Service` and a
"db-app-xyz" `Secret` in the "info-tech" namespace. IT also has lots of other `Service` resources in
their namespace and they don't want to expose all of them to the development team. Therefore, they
publish "db-app-xyz" to a service catalog. To use this service, a developer searches for it in the
service catalog and adds it to their namespace.

### Easy linking of a Service to a Deployable resource

A user has developed an application that uses a database. The user doesn't want to hard-code
the URL to the database, because that would be brittle and require rebuilding the application if the
database coordinates change. Instead, the user wants to be able to create the application and link
it to a database service. A typical sequence for this use case in a Platform as a Service (PaaS) might
look like:

1. User asks PaaS to create a new application A
2. User asks PaaS to add a database to application A
3. User pushes code to application A
4. PaaS builds and deploys application A
5. Application A starts up, connecting to the database by looking up a well-known, predefined
   environment variable for the database's URL which was injected and set by the PaaS

## Service configuration data

The only data currently associated with a service is a list of zero or more endpoints. Many services
also have configuration data required to use them. These include things such as logins, passwords,
and additional connection parameters. To tie configuration data to a service, we propose adding
information to the `Service` type to be able to express a relationship between the `Service` and its
relevant configuration data. This could be accomplished by adding new field(s) or by using
annotations; these would contain references to the relevant associated resources.  Potential
configuration data target reference types include `Secrets` and `ConfigMap` (but to avoid limiting
to just those types, we can use `LocalObjectReferences`).

Note: some of the associated data such as protocol and path (a.k.a. context root) are probably
better suited as annotations or fields on the `Service` itself, and not as references. These items
are most likely used by the cluster itself in some way, or potentialy by a UI; e.g., to show a link
to the service if the UI knows it supports HTTP.

Specifying references to `Secrets`, `ConfigMaps`, etc. by itself does not accomplish much; to be
useful, that data needs to be available to the processes executing in containers. More on this
process is below in the sections on service claims and linking.

## Service catalogs

A service catalog is a listing of published service entries. For Phase 1, the only valid type of
entry is a `Service`. This allows us to support a service that points to an external database that
is deployed outside of the cluster, as well as services that act as load balancers to a selected set
of backend pods via the kube-proxy.

The catalog is not meant to include every service in the cluster. Instead, it should contain those
services that users wish to highlight and make available to other users. For example, your namespace
might contain "etcd", "etcd-discovery", and "postgresql" services, and the only one you want to
share with others is the postgresql service.

One way of implementing a service catalog could be to create a namespace to represent a shared
catalog. This could be feasible, but it has some limitations:

- If we eventually add support for additional types of entries in the catalog (templates, service
  brokers), users/UIs would have to query multiple resource types to retrieve the entire catalog
- Assuming we add functionality to associate `Secrets` with a `Service`, a `Pod` running in
  namespace "foo" wouldn't be able to access the secret in namespace "servicecatalog"
- Users have to know the name of the namespace(s) that contain shared services

In light of this, we believe an actual `ServiceCatalog` resource, combined with the claiming and
provisioning mechanisms described below, offers richer functionality to end users than a shared
namespace.

Because the service catalog is meant to span namespaces, it should not be a namespaced resource. We
should support multiple service catalogs, as different groups using the cluster might want to offer
their own catalogs. We may also want to consider a "default" catalog that is displayed in the
absence of multiple catalogs, to simplify how users interact with service catalogs.

### Publishing to a catalog

Users should be able to publish entries to a service catalog. We have considered two options for
this.

#### Option 1: annotate the `Service`

If you want to have a `Service` included in a service catalog, add the following annotations:

- kubernetes.io/catalog.destination = "default"
- kubernetes.io/catalog.entry.name = "awesome-etcd"
- kubernetes.io/catalog.entry.description = "an etcd service"

This example publishes a `Service` into the "default" `ServiceCatalog` with the entry name
"awesome-etcd".

In order to appear in the listing of a service catalog's entries, we most likely will want to use a
reflector/cache of some sort. Any time a `Service` is modified, the reflector inspects the updated
resource and adds/updates/removes it from the cache. The implementation of this cache is potentially
not trivial, as it needs to take into account security policy decisions such as "is UserX allowed to
publish to this catalog?".

**Pros**

- does not require an API change to the core `Service` type
- information about the catalog entry is stored on the `Service` itself, making it easy to see if
  the service has been published or not

**Cons**

- requires a cache for efficiency
- checking security policy decisions could be difficult
- users won't receive any immediate indication that publishing an entry to a catalog was denied
  (but they could potentially see the denial in an annotation)

#### Option 2: add `ServiceCatalogEntry` resource

If you want to have a `Service` included in a service catalog, create a new `ServiceCatalogEntry`
resource, such as:

```yaml
apiVersion: catalog/v1beta1
kind: ServiceCatalogEntry
metadata:
  name: awesome-etcd
catalog: default
description: an etcd service
targetObjectReference:
  apiVersion: v1
  kind: Service
  namespace: foo
  name: etcd
```

This example publishes a `Service` called "etcd" from the namespace "foo" into the "default"
`ServiceCatalog` with the entry name "awesome-etcd".

**Pros**

- is a strongly-typed resource, with specific fields
- users can get immediate feedback that publishing an entry was accepted or denied

**Cons**

- checking security policy decisions could be difficult

### Viewing a catalog

Users should be able to list the entries in a service catalog. Users should be able to select an
entry and "consume" it. We call this consumption "claiming" a service from a service catalog.

### Claiming a service

When you want to use an entry from the service catalog, you create a `ServiceClaim` that references
the desired entry. A controller processes new claims for admission. This determines if the user who
created the claim is allowed to consume the entry from the catalog. This decision can be flexible:
it could be automated based on policy, or it could support manual intervention and workflow.

Once the claim has been admitted, a controller performs the provisioning process. The controller
creates a new service in the user's namespace that "points" to the original service.  Additionally,
the controller clones each of the original service's referenced resources to the user's namespace.

#### Communicating with the claimed service

##### Option 1: CNAME

Given an original service "foo.bob.svc.cluster.local", and a service created via claim "bar" in
namespace "alice", DNS requests for "bar.alice.svc.cluster.local" resolve via CNAME to
"foo.bob.svc.cluster.local".

**Cons**:

- Doesn't work for TLS hostname verification. A request to "bar.alice.svc.cluster.local" will
  receive a certificate for "foo.bob.svc.cluster.local", which will cause hostname verification to
  fail.

##### Option 2: ????

(I need some assistance from the community on this topic)

## Linking services

Claiming a service catalog entry only creates resources in the user's namespace. If all you need is
a service and its DNS entry, this may be sufficient for your pods to function. But if you need the
configuration data injected into your pod, it would be nice to make that easier to do.

We want to add the ability to link a service to a deployable resource such as a Deployment. This
could look something like:

    kubectl link svc/postgresql deployment/web
    service "postgresql" linked to deployemnt "web"
    configdata "postgresql-options" linked to deployment "web" as a volume
    secret "postgresql-credentials" linked to deployment "web" as a volume

This command would automatically inject the ConfigData and Secret objects as volumes into the
Deployment. This could be flexible as well, allowing you instead to expose these items as
environment variables.

In the example above, the volumes could potentially be mounted as:

/var/run/kubernetes.io/links/configdata/postgresql-options
/var/run/kubernetes.io/links/secrets/postgresql-credentials

Note: linking is orthogonal to a service catalog and service claims.

There are several outstanding questions in this area:

- How do we best represent the intent that a resource such as a `Deployment` is linked to a `Service`?
	- 1 suggestion is to add a `ServiceLinks []ObjectReference` to `PodSpec`
- When do we process the link to create the environment variables and/or volumes?
- Is it acceptable to define naming conventions for these volumes?
	- How can an application/container author develop for these conventions most easily?
- How can we easily support a use case where we need to:
	1. Mount a file
	2. Create an environment variable with a specific name that points to that file (e.g., to access the GCP APIs via standard client libraries; see https://developers.google.com/identity/protocols/application-default-credentials)


## Cross-namespace networking

If the cluster has multi-tenant network isolation enabled, then a pod in namespace A won't be able
to talk to a service in namespace B. We should look into ways to automatically manipulate the
isolation rules to "poke holes" when claims are provisioned, to make the desired connectivity work,
and to "unpoke" when the connectivity is no longer needed.

## Proposed API changes

- Add annotation/field to `Service` to specify associated resources
- Add `ServiceCatalog` types in a separate `catalog` API group

```go
type ServiceCatalog struct {
  unversioned.TypeMeta
  ObjectMeta

  // Not sure exactly what we want/need here
  // Could include who is allowed to publish
}

type ServiceCatalogList struct {
  unversioned.TypeMeta
  ListMeta

  Items []ServiceCatalog
}
```

- Add `ServiceCatalogEntry` types

```go
type ServiceCatalogEntry struct {
  unversioned.TypeMeta
  ObjectMeta

  // The ServiceCatalog to which this entry belongs
  Catalog string
  // This entry's description
  Description string
  // The resource to which this entry refers
  TargetObjectReference ObjectReference
  // The entry's type for provisioning. Different types may be handled by different provisioners to support distinct functionality. Defaults to "reference" as that is all that phase 1 supports, but needed to support phase 2.
  Type string
}

type ServiceCatalogEntryList struct {
  unversioned.TypeMeta
  ListMeta

  Items []ServiceCatalogEntry
}
```

- Add `ServiceClaim` types

```go
type ServiceClaim struct {
  unversioned.TypeMeta
  ObjectMeta

  Spec ServiceClaimSpec
  Status ServiceClaimStatus
}

type ServiceClaimSpec struct {
  // Specifies the desired service catalog
  ServiceCatalogName string
  // Specifies the entry to claim
  Entry ObjectReference
}

type ServiceClaimStatus struct {
  State ServiceClaimState
  // An array of the items created when this claim was provisioned
  ProvisionedItems []LocalObjectReference
}

type ServiceClaimState string

const (
  ServiceClaimState ServiceClaimStateNew = "New"
  ServiceClaimState ServiceClaimStateAdmitted = "Admitted"
  ServiceClaimState ServiceClaimStateRejected = "Rejected"
  ServiceClaimState ServiceClaimStateProvisioned = "Provisioned"
)

type ServiceClaimList struct {
  unversioned.TypeMeta
  ListMeta

  Items []ServiceClaim
}
```

# Phase 2

## Use cases

### Template-based provisioning

A user creates a "template" (e.g. if something similar to [OpenShift
Templates](https://docs.openshift.org/latest/dev_guide/templates.html) exists) that makes it easy to
create everything needed to spin up a new PostgreSQL database (customizable username/password,
`Service`, `Deployment`, etc.). The user wants to share only this template in a service catalog so
others can find it and use it, while keeping other templates in the namespace private.

### Custom provisioning

The IT department manages an off-cluster database. Each developer wanting to access the database is
required to use a unique username and password to access the database. Additionally, each developer
accesses a unique tablespace that no other team members can access. The IT department used to create
database accounts and tablespaces by hand in response to individual developer requests.

Moving forward, IT wants to automate the process. They create a web service that implements a
"service broker" API (something similar to [Cloud
Foundry's](http://docs.cloudfoundry.org/services/api.html#api-overview)). They add an entry to the
service catalog for their database service, pointing at their web service. When a developer consumes
the entry from the catalog, IT's web service is contacted, resulting in a new username, password,
and tablespace.

## Types of service catalog entries

An entry in the service catalog has a "type", which indicates the behavior that occurs when it is
consumed by a user. We have thought of the following potential types:

- Reference: "I want to use this entry as-is, including its configuration data"
	- This is what Phase 1 implements; namely, provisioning a `Service` and its related resources in the destination namespace
- Template: "I want to create items from the specified template"
- ServiceBroker: "I want the creation to be goverened by some other entity that implements the 'service broker' HTTP interface"

We see the type as an arbitrary `string`; one or more controllers could run to process each type, performing whatever logic is appropriate to fulfill the specific type in question.

Additional fulfillment types are possible as long as there is a controller that is handling them.

# TODOs

- Figure out security
	- How do we determine who can publish a service to a given catalog?
	- How do we determine who can see/use a specific service from a given catalog?
- How do you keep the claimed services in sync with the sources?
	- e.g. if the source service references a secret, and the secret's content changes
- What does it mean to "unlink"?
- What happens if you delete a claim - does it cascade?

# Prototype

We have implemented a prototype to demonstrate a possible workflow for Catalogs in a Kubernetes
cluster.  The implementation is [here](https://github.com/sjenning/kubernetes/tree/catalog-apiserver).

This prototype allows for the following type of workflow:

1. A user (the “provider” or “publisher”) creates one or more resources that he/she wants to share
   with others. For this example, let’s imagine that I want to share a secret I created with other
   users.
1. The publisher creates a catalog posting that references the secret.
1. Another user (the “consumer”) searches for and locates the entry in the catalog for the secret.
1. The consumer instantiates the catalog entry (we’re currently calling this a claim, but the name
   is open for discussion)
1. The end result is a new secret in the consumer’s namespace.
1. If the publisher changes the contents of the secret, all consumers’ secrets are automatically
   updated with the latest data.


We added a new executable that runs an apiserver and controllers related to catalogs. This process
handles requests to a new API group, servicecatalog/v1alpha1. It also starts the controllers that
monitor the resources in the new API group and performs actions such as catalog entry generation,
catalog entry claim binding, and synchronization of resources created by catalog entry claims.

The types in the new API group are:

- CatalogPosting
  - Namespace-scoped references to a collection of resources to be made available for claiming
- CatalogEntry
    - Cluster-scoped reference to a CatalogPosting with information relevant to consumers; displays
      sufficient level of details to consumers without exposing sensitive information
- CatalogClaim
  - A resource that expresses intent to consume (i.e. import into a destination namespace) the
    resources offered by a CatalogPosting via a CatalogEntry

Here is a more detailed description of the workflow above:

1. The publisher creates one or more resources in his/her namespace that are to be shared with
   others
1. The publisher creates a CatalogPosting that references the resources from step 1
1. A controller creates a CatalogEntry that corresponds to the new CatalogPosting that contains
   information about the shared resources relevant to consumers
1. The consumer lists the available entries in the catalog
1. The consumer creates a CatalogClaim for a specific CatalogEntry expressing intent to consume that
   entry
1. A controller processes the CatalogClaim and provisions the appropriate resources in the
   consumer’s namespace
1. A controller monitors resources associated with CatalogPostings for change and keeps all
   resources provisioned from CatalogClaims in sync

The controller to sync all possible resources types that a catalog posting could reference would be
a large task. For demonstration purposes, the claim sync controller in this prototype only syncs
secrets, since those are a resource a publisher could likely change and would need to sync down to
the claimed secrets in order for pods to continue functioning.

For example, if a publisher shares a secret and a consumer claims it, when the publisher changes the
secret, the copy in the consumer’s namespace is updated. The consumer may need to restart/redeploy
any pods referencing that secret, if the secret’s contents are referenced as environment variables,
or if the application isn’t able to react to changes to secrets mounted as files.

A video demo of the prototype is [here](https://www.youtube.com/watch?v=Jbi19qk79bo).

<!-- BEGIN MUNGE: GENERATED_ANALYTICS -->
[![Analytics](https://kubernetes-site.appspot.com/UA-36037335-10/GitHub/docs/proposals/service-catalog.md?pixel)]()
<!-- END MUNGE: GENERATED_ANALYTICS -->
