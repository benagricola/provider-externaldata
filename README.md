# provider-externaldata

`provider-externaldata` is a minimal [Crossplane](https://crossplane.io/) Provider
that can be used to retrieve data from external sources. This provider does not *create* any
'real' resources, but fakes the creation, update and deletion of a resource as it loads
data from external sources. It currently supports retrieving data from:

- A `ConfigMap` within the current Kubernetes cluster. This requires
  a `namespace` value to be configured on the `ProviderConfig`, and allows the cluster admin
  to control where data can be retrieved from.
- A URI containing JSON, retrieved using `go-resty`. Note: this will be retrieved at least _once_ per reconciliation loop of the resource, and the request must take less than one second.

## Usage

```yaml

apiVersion: external.crossplane.io/v1alpha1
kind: ProviderConfig
metadata:
  name: default
spec:
  # Configure namespace where ConfigMaps can be looked up.
  namespace: test
---
apiVersion: datasource.external.crossplane.io/v1alpha1
kind: DataSource
metadata:
  name: cm-example
spec:
  forProvider:
    type: configmap
    # Retrieve all values from 'Data' section of the 'my-values' ConfigMap.
    configMapName:
      name: my-values
---
apiVersion: datasource.external.crossplane.io/v1alpha1
kind: DataSource
metadata:
  name: uri-example
spec:
  forProvider:
    type: uri
    # Retrieve four cheese margherita pizza information from URI.
    uri: https://raw.githubusercontent.com/elastic/examples/master/Search/recipe_search_java/data/four-cheese-margherita-pizza.json
```

## Developing

Run against a Kubernetes cluster:

```console
make run
```

Build, push, and install:

```console
make all
```

Build image:

```console
make image
```

Push image:

```console
make push
```

Build binary:

```console
make build
```
