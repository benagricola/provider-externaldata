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

Output from the above `DataSource` resources:

````
Name:         cm-example
Namespace:
Labels:       <none>
Annotations:  crossplane.io/external-name: cm-example
API Version:  datasource.external.crossplane.io/v1alpha1
Kind:         DataSource
Metadata:
  UID:               fbb5aefc-c04d-42ae-94e9-bd283d9240fc
Spec:
  For Provider:
    Config Map Name:  my-values
    Type:             configmap
  Provider Config Ref:
    Name:  default
Status:
  At Provider:
    player_initial_lives:     2
    ui_properties_file_name:  user-interface.properties
    Wee Woo Test:             NINE
  Conditions:
    Last Transition Time:  2021-07-24T14:39:19Z
    Reason:                Creating
    Status:                False
    Type:                  Ready
    Last Transition Time:  2021-07-24T14:39:19Z
    Reason:                ReconcileSuccess
    Status:                True
    Type:                  Synced
Events:
  Type    Reason                   Age   From                                                  Message
  ----    ------                   ----  ----                                                  -------
  Normal  CreatedExternalResource  8s    managed/datasource.datasource.external.crossplane.io  Successfully requested creation of external resource


Name:         url-example
Namespace:
Labels:       <none>
Annotations:  crossplane.io/external-name: url-example
API Version:  datasource.external.crossplane.io/v1alpha1
Kind:         DataSource
Metadata:
  UID:               d3b0ef41-d6c0-4f89-8bc4-d341747cd3e7
Spec:
  For Provider:
    Type:  url
    URL:   https://raw.githubusercontent.com/elastic/examples/master/Search/recipe_search_java/data/four-cheese-margherita-pizza.json
  Provider Config Ref:
    Name:  default
Status:
  At Provider:
    Author:
      Name:         Michelle
      URL:          http://allrecipes.com/cook/18668259/profile.aspx
    cook_time_min:  10
    Description:    This is a fantastic version of an Italian classic. The feta cheese adds a rich flavor that brings this dish to life. Incredibly easy and incredibly delicious!
    Directions:
      Stir together olive oil, garlic, and salt; toss with tomatoes, and allow to stand for 15 minutes.
      Preheat oven to 400 degrees F (200 degrees C).
      Brush each pizza crust with some of the tomato marinade. Sprinkle the pizzas evenly with Mozzarella and Fontina cheeses. Arrange tomatoes overtop, then sprinkle with shredded basil, Parmesan, and feta cheese.
      Bake in preheated oven until the cheese is bubbly and golden brown, about 10 minutes.
    Ingredients:
      1/4 cup olive oil
      1 tablespoon minced garlic
      1/2 teaspoon sea salt
      8 Roma tomatoes, sliced
      2 (12 inch) pre-baked pizza crusts
      8 ounces shredded Mozzarella cheese
      4 ounces shredded Fontina cheese
      10 fresh basil leaves, washed, dried
      1/2 cup freshly grated Parmesan cheese
      1/2 cup crumbled feta cheese
    prep_time_min:  15
    Servings:       8
    source_url:     http://allrecipes.com/recipe/four-cheese-margherita-pizza
    Tags:
      main dish
    Title:  Four Cheese Margherita Pizza
  Conditions:
    Last Transition Time:  2021-07-24T14:39:19Z
    Reason:                Creating
    Status:                False
    Type:                  Ready
    Last Transition Time:  2021-07-24T14:39:19Z
    Reason:                ReconcileSuccess
    Status:                True
    Type:                  Synced
Events:
  Type    Reason                   Age              From                                                  Message
  ----    ------                   ----             ----                                                  -------
  Normal  CreatedExternalResource  8s               managed/datasource.datasource.external.crossplane.io  Successfully requested creation of external resource
  Normal  UpdatedExternalResource  7s (x2 over 8s)  managed/datasource.datasource.external.crossplane.io  Successfully requested update of external resource
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
