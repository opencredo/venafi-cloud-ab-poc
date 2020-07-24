# Using Venafi Cloud and CockroachDB Together

The aim of the document is to describe how to use Venafi Cloud to issue TLS
certificates for use by CockroachDB. In this guide we will be using:

 * cert-manager
 * CockroachDB
 * Venafi Cloud
 * Kubernetes (though the existence of a accessible cluster is assumed)
 * Helm
 * helmfile

## Overview

### CockroachDB

From Cockroachlabs site:

> CockroachDB is an elastic, indestructible SQL database for developers building
> modern applications.

CockroachDB is a global-scale NewSQL database based on ideas from Google Spanner
but suited to environments run by "normal" people, rather than the very strict
requirements needed by Spanner.

It is wire-compatible with PostgreSQL and, for the most part, can be a drop-in
replacement with a few gotchas:

 * The SQL flavour is different
 * There are a few areas where behaviour is different and primary key best-practices
   are different
 * The read/write performance is different as CockroachDB is doing more

It is extremely useful when building global-scale applications where strongly
consistent, relational database is invovled.

### Venafi Cloud

Venafi Cloud is a product that allows organisation to enforce TLS policies
in a DevOps environment. It allows security teams to build zones and templates
for certificates to ensure certificate standards and provides APIs and integrations
to allow engineers to source TLS certificates that meet these standards.

As CockroachDB uses TLS certificates for node-to-node in-flight encryption and user
authentication is seems ideal to use these together.

## Assumptions

This guide makes the following assumptions:

 * You already have access to a Kubernetes cluster and the relevant access and
   permissions (i.e. doing `kubectl get pods -A` shows _all_ the pods in your
   cluster).
 * You're Kubernetes cluster is configured with storage and the appropriate storage
   class.
 * You already have a Venafi Cloud account, an issuing template created and a
   project and zone available.

For later in the documentation you will need:

 * A Venafi Cloud API key (a UUID that can be found by clicking on your user
   in the top-right of the Venafi Cloud web console).
 * The zone ID (a UUID that can be found by clicking on the zone in the project
   screen, it is the UUID part of the ACME URL).
 * The CA chain for the CA configured in the Venafi Cloud project zone.

## Installing `cert-manager`

Installing `cert-manager` is pretty simple and more detailed information can be
found in their documentation.

From the command-line:

```bash
$ kubectl create namespace cert-manager

namespace/cert-manager created

$ helm repo add jetstack https://charts.jetstack.io

"jetstack" has been added to your repositories

$ helm install cert-manager jetstack/cert-manager --namespace cert-manager --set installCRDs=true

NAME: cert-manager
LAST DEPLOYED: Fri Jul 24 07:24:48 2020
NAMESPACE: cert-manager
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
cert-manager has been deployed successfully!

In order to begin issuing certificates, you will need to set up a ClusterIssuer
or Issuer resource (for example, by creating a 'letsencrypt-staging' issuer).

More information on the different types of issuers and how to configure them
can be found in our documentation:

https://cert-manager.io/docs/configuration/

For information on how to configure cert-manager to automatically provision
Certificates for Ingress resources, take a look at the `ingress-shim`
documentation:

https://cert-manager.io/docs/usage/ingress/
```
If you're using helmfile then this would be the equivalent configuration:

```YAML
repositories:
- name: jetstack
  url: https://charts.jetstack.io
releases:
- name: cert-manager
  labels:
    cert-manager: "true"
  version: "v0.16.0"
  chart: jetstack/cert-manager
  namespace: cert-manager
  values:
  - installCRDs: true
```

Running:

```bash
$ helmfile apply
Adding repo jetstack https://charts.jetstack.io
"jetstack" has been added to your repositories

Updating repo
Hang tight while we grab the latest from your chart repositories...
...Successfully got an update from the "hashicorp" chart repository
...Successfully got an update from the "cockroachdb" chart repository
...Successfully got an update from the "nginx-stable" chart repository
...Successfully got an update from the "linkerd" chart repository
...Successfully got an update from the "ingress-nginx" chart repository
...Successfully got an update from the "jetstack" chart repository
Update Complete. ⎈ Happy Helming!⎈

Comparing release=cert-manager, chart=jetstack/cert-manager
...
Lots of manifests
...
Upgrading release=cert-manager, chart=jetstack/cert-manager
Release "cert-manager" does not exist. Installing it now.
NAME: cert-manager
LAST DEPLOYED: Fri Jul 24 07:29:33 2020
NAMESPACE: cert-manager
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
cert-manager has been deployed successfully!

In order to begin issuing certificates, you will need to set up a ClusterIssuer
or Issuer resource (for example, by creating a 'letsencrypt-staging' issuer).

More information on the different types of issuers and how to configure them
can be found in our documentation:

https://cert-manager.io/docs/configuration/

For information on how to configure cert-manager to automatically provision
Certificates for Ingress resources, take a look at the `ingress-shim`
documentation:

https://cert-manager.io/docs/usage/ingress/

Listing releases matching ^cert-manager$
cert-manager    cert-manager    1               2020-07-24 07:29:33.4404155 +0100 BST   deployed        cert-manager-v0.16.0    v0.16.0


UPDATED RELEASES:
NAME           CHART                   VERSION
cert-manager   jetstack/cert-manager   v0.16.0
```

The installation of `cert-manager` is complete at this point.

## Configuring Venafi Cloud `ClusterIssuer`

To make use of Venafi Cloud's TLS certificates in our cluster we need to create
either a `Issuer` or `ClusterIssuer` resource. These are `cert-manager` custom
resources. As we are expecting to use our CockroachDB across all namespaces we
will choose a `ClusterIssuer` as, unlike `Issuer`, it is not restricted to a 
single namespace. We will also need to create a `Secret` with our Venafi Cloud API
key.

Depending on your deployment setup you may want to wrap this up either in Terraform,
a Helm chart or customise the CockroachDB Helm chart.

This is the manifest we need to apply:

```YAML
---
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: cloud-venafi-secret
  namespace: cert-manager
data:
  apikey: <Venafi Cloud API Key>
---
apiVersion: cert-manager.io/v1alpha3
kind: ClusterIssuer
metadata:
  name: cloud-venafi-issuer
spec:
  venafi:
    cloud:
      apiTokenSecretRef:
        key: apikey
        name: cloud-venafi-secret
    zone: <Venafi Cloud Zone ID>
```

Running:

```bash
$ kubectl apply -f venafi-issuer.yaml
secret/cloud-venafi-secret created
clusterissuer.cert-manager.io/cloud-venafi-issuer created
```

We can check to see if this all worked as expected:

```bash
$ kubectl get clusterissuer -A

NAME                  READY   AGE
cloud-venafi-issuer   True    47s
```

This is a good indicator it has worked correctly. We can go a step further:

```bash
$ kubectl describe clusterissuer cloud-venafi-issuer

Name:         cloud-venafi-issuer
Namespace:
Labels:       <none>
Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                {"apiVersion":"cert-manager.io/v1alpha3","kind":"ClusterIssuer","metadata":{"annotations":{},"name":"cloud-venafi-issuer"},"spec":{"venafi..."
API Version:  cert-manager.io/v1beta1
Kind:         ClusterIssuer
Metadata:
  Creation Timestamp:  2020-07-24T06:44:08Z
  Generation:          1
  Resource Version:    6544
  Self Link:           /apis/cert-manager.io/v1beta1/clusterissuers/cloud-venafi-issuer
  UID:                 7ec2de63-779a-4ea1-941e-dbd50eaad6b3
Spec:
  Venafi:
    Cloud:
      API Token Secret Ref:
        Key:   apikey
        Name:  cloud-venafi-secret
    Zone:      f00a6cb0-c680-11ea-b46e-630913383a67
Status:
  Conditions:
    Last Transition Time:  2020-07-24T06:44:09Z
    Message:               Venafi issuer started
    Reason:                Venafi issuer started
    Status:                True
    Type:                  Ready
Events:
  Type    Reason  Age   From          Message
  ----    ------  ----  ----          -------
  Normal  Ready   98s   cert-manager  Verified issuer with Venafi server
```

This shows that our `ClusterIssuer` has been able to authenticated with Venafi Cloud.

## Issuing CockroachDB certificates

Before we deploy CockroachDB to our cluster we should issue the certificates it
requires. This can be done after the CockroachDB deploy but CockroachDB cluster
initialisation may need manual intervention if done that way.

Again, depending on your deployment setup you may want to wrap this up either in Terraform,
a Helm chart or customise the CockroachDB Helm chart.

This is the manifest we need to apply:

```YAML
---
apiVersion: v1
kind: Namespace
metadata:
  labels:
    name: cockroachdb
  name: cockroachdb
---
apiVersion: cert-manager.io/v1alpha3
kind: Certificate
metadata:
  name: cockroachdb-node
  namespace: cockroachdb
spec:
  commonName: node
  dnsNames:
  - cockroachdb-public.cockroachdb.svc.cluster.local
  - cockroachdb-0.cockroachdb.cockroachdb.svc.cluster.local
  - cockroachdb-1.cockroachdb.cockroachdb.svc.cluster.local
  - cockroachdb-2.cockroachdb.cockroachdb.svc.cluster.local
  - cockroachdb-0.cockroachdb.svc.cluster.local
  - cockroachdb-1.cockroachdb.svc.cluster.local
  - cockroachdb-2.cockroachdb.svc.cluster.local
  - cockroachdb-0.cockroachdb
  - cockroachdb-1.cockroachdb
  - cockroachdb-2.cockroachdb
  - cockroachdb-0
  - cockroachdb-1
  - cockroachdb-2
  duration: 24h0m0s
  issuerRef:
    kind: ClusterIssuer
    name: cloud-venafi-issuer
  keyAlgorithm: rsa
  renewBefore: 1h0m0s
  secretName: cockroachdb-node
  usages:
  - server auth
  - client auth
---
apiVersion: cert-manager.io/v1alpha3
kind: Certificate
metadata:
  name: cockroachdb-root
  namespace: cockroachdb
spec:
  commonName: root
  duration: 24h0m0s
  issuerRef:
    kind: ClusterIssuer
    name: cloud-venafi-issuer
  keyAlgorithm: rsa
  renewBefore: 1h0m0s
  secretName: cockroachdb-root
  usages:
  - server auth
  - client auth
```

There are 3 parts to this manifest:

 1. Create the CockroachDB namespace
 2. Create a `cockroachdb-node` certificate (this will result in a
    `cockroachdb-node` secret in the `cockroachdb` namespace). This will be
    used for node-to-node encryption.
 3. Create a `cockroachdb-root` certificate (this will result in a
    `cockroachdb-root` secret in the `cockroachdb` namespace). This will be
    used for root user-to-node encryption and authentication.

We can apply this:

```bash
$ kubectl apply -f cockroach-certs.yaml

namespace/cockroachdb created
certificate.cert-manager.io/cockroachdb-node created
certificate.cert-manager.io/cockroachdb-root created
```

And we can check to see if this worked:

```bash
$ kubectl get secret cockroachdb-node -n cockroachdb -o yaml

apiVersion: v1
data:
  tls.crt: <base64 encoded data>
  tls.key: <base64 encoded key>
kind: Secret
metadata:
  annotations:
    cert-manager.io/alt-names: cockroachdb-public.cockroachdb.svc.cluster.local,cockroachdb-0.cockroachdb.cockroachdb.svc.cluster.local,cockroachdb-1.cockroachdb.cockroachdb.svc.cluster.local,cockroachdb-2.cockroachdb.cockroachdb.svc.cluster.local,cockroachdb-0.cockroachdb.svc.cluster.local,cockroachdb-1.cockroachdb.svc.cluster.local,cockroachdb-2.cockroachdb.svc.cluster.local,cockroachdb-0.cockroachdb,cockroachdb-1.cockroachdb,cockroachdb-2.cockroachdb,cockroachdb-0,cockroachdb-1,cockroachdb-2,node
    cert-manager.io/certificate-name: cockroachdb-node
    cert-manager.io/common-name: node
    cert-manager.io/ip-sans: ""
    cert-manager.io/issuer-kind: ClusterIssuer
    cert-manager.io/issuer-name: cloud-venafi-issuer
    cert-manager.io/uri-sans: ""
  creationTimestamp: "2020-07-24T06:57:49Z"
  name: cockroachdb-node
  namespace: cockroachdb
  resourceVersion: "8848"
  selfLink: /api/v1/namespaces/cockroachdb/secrets/cockroachdb-node
  uid: e0dced8e-0c49-4ce1-bd09-8c8019ef8143
type: kubernetes.io/tls
```

We can see this worked and we can do the same for `cockroachdb-root` _but_ there
is no `ca.crt` entry this will cause us problems later as CockroachDB expects to
be able to use that `ca.crt` to verify the certificates we've issued. There is
no clean way to sideload this other than to patch the secret. I'll discuss how to
automate that later but for the time being we'll `kubectl patch` both of these
secrets with the CA chain:

```YAML
data:
  ca.crt: |
    <base64 encoded CA chain PEM>
```

To apply the patch:

```bash
$  kubectl patch secret cockroachdb-root -n cockroachdb --patch "$(cat patch.yaml)"

secret/cockroachdb-root patched

$  kubectl patch secret cockroachdb-node -n cockroachdb --patch "$(cat patch.yaml)"

secret/cockroachdb-node patched
```

## Installing CockroachDB

Installing CockroachDB is also pretty simple using their Helm charts more 
information can be found in their documentation. We are going to installing
CockroachDB in secure mode.

From the command-line:

```bash
$ helm repo add cockroachdb https://charts.cockroachdb.com

"cockroachdb" has been added to your repositories
```

We need to set a number of configuration options. We can specify these on the
Helm command-line but there are quite a few so we will put them into a YAML file:

```YAML
statefulset:
  resources:
    requests:
      memory: 2Gi
    limits:
      memory: 2Gi
conf:
  cache: 500Mi
  max-sql-memory: 500Mi
storage:
  persistentVolume:
    size: 10Gi
tls:
  enabled: true
  certs:
    provided: true
    tlsSecret: true
```

Installing CockroachDB using Helm:

```bash
$ helm install cockroachdb -n cockroachdb --values my-values.yaml cockroachdb/cockroachdb

NAME: cockroachdb
LAST DEPLOYED: Fri Jul 24 10:16:14 2020
NAMESPACE: cockroachdb
STATUS: deployed
REVISION: 1
NOTES:
CockroachDB can be accessed via port 26257 at the
following DNS name from within your cluster:

cockroachdb-public.cockroachdb.svc.cluster.local

Because CockroachDB supports the PostgreSQL wire protocol, you can connect to
the cluster using any available PostgreSQL client.

Note that because the cluster is running in secure mode, any client application
that you attempt to connect will either need to have a valid client certificate
or a valid username and password.

Finally, to open up the CockroachDB admin UI, you can port-forward from your
local machine into one of the instances in the cluster:

    kubectl port-forward cockroachdb-0 8080

Then you can access the admin UI at https://localhost:8080/ in your web browser.

For more information on using CockroachDB, please see the project's docs at:
https://www.cockroachlabs.com/docs/
```

If you are using helmfile then your configuration would look like:

```YAML
repositories:
- name: jetstack
  url: https://charts.jetstack.io
- name: cockroachdb
  url: https://charts.cockroachdb.com
releases:
- name: cert-manager
  version: "v0.15.2"
  chart: jetstack/cert-manager
  namespace: cert-manager
  values:
  - installCRDs: true
- name: cockroachdb
  version: "4.1.0"
  chart: cockroachdb/cockroachdb
  namespace: cockroachdb
  values:
  - statefulset:
      resources:
        requests:
          memory: 2Gi
        limits:
          memory: 2Gi
  - conf:
      cache: 500Mi
      max-sql-memory: 500Mi
  - storage:
      persistentVolume:
        size: 10Gi
  - tls:
      enabled: true
      certs:
        provided: true
        tlsSecret: true
```

Given a bit of time `kubectl get pods -n cockroachdb` should show all containers
within all 3 pods up and running.

At this point you're done but you probably want to allow some services to access
your CockroachDB database.

## Service User provisioning

This is also pretty simple. First, we need to create the appropriate certificates.
If you are deploying your services with Helm you would most likely add a cert-manager
`Certificate` resource to your service's Helm chart. It would look something like:

```YAML
apiVersion: cert-manager.io/v1alpha3
kind: Certificate
metadata:
  name: cockroachdb-<service name>
  namespace: <service namespace>
spec:
  commonName: <name of database user>
  duration: 24h0m0s
  issuerRef:
    kind: ClusterIssuer
    name: cloud-venafi-issuer
  keyAlgorithm: rsa
  renewBefore: 1h0m0s
  secretName: cockroachdb-<service name>
  usages:
  - client auth
```

Running:

```bash
$ kubectl apply -f cockroachdb-service.yaml

certificate.cert-manager.io/cockroachdb-<service name>
```

Will create the certificate. The generate `Secret` resource will also be missing
the `ca.crt` field and we will need to patch it. We can use exactly the same
patch as before:

```bash
$  kubectl patch secret cockroachdb-<service name> -n <service namespace> --patch "$(cat patch.yaml)"

secret/cockroachdb-<service name> patched
```

You can then configure your pod or deployment to mount this secret as a volume and
use it in your PostgreSQL connection string:

```YAML
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: <service name>
  name: <service name>
  namespace: <service namespace>
spec:
  replicas: 1
  selector:
    matchLabels:
      app: <service name>
    spec:
      containers:
      - env:
        - name: POSTGRESQL_CONNECTION_STRING
          value: user=<service name> host=cockroachdb-public.cockroachdb.svc.cluster.local
            port=26257 dbname=<database name> sslmode=verify-full sslcert=/cockroachdb-certs/tls.crt
            sslkey=/cockroachdb-certs/tls.key sslrootcert=/cockroachdb-certs/ca.crt
        image: <repo>/<service name>:<tag>
        imagePullPolicy: IfNotPresent
        name: <service name>
        ports:
        - containerPort: 8080
          protocol: TCP
        volumeMounts:
        - mountPath: /cockroachdb-certs
          name: cockroachdb-secret
          readOnly: true
      volumes:
      - name: cockroachdb-secret
        secret:
          defaultMode: 420
          secretName: cockroachdb-<service name>
```

## Secrets Workaround

Finally, to use the `Secret`s created by cert-manager we've had to do some `kubectl
patching`. Can we do away with that? Apart from changing cert-manager to populate
`ca.crt` we can also use a Mutating Webhook to do this patch for us automatically.

An example of this can be found [here](https://github.com/opencredo/venafi-cloud-ab-poc/tree/master/go/cmd/secretfixer).