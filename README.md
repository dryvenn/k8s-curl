# k8s-curl

This is program is a basic controller for Kubernetes that will fetch webpages
specified in a ConfigMap's annotation and put the result in its data.

The annotation format is `"x-k8s.io/curl-me-that": <key>=<url>...`

For example this:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: example
  annotations:
    x-k8s.io/curl-me-that: joke=curl-a-joke.herokuapp.com jest=curl-a-joke.herokuapp.com poke=curl-a-poke.herokuapp.com
data:
  existing: field
```

Will result in this:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: example
  annotations:
    x-k8s.io/curl-me-that: joke=curl-a-joke.herokuapp.com jest=curl-a-joke.herokuapp.com poke=curl-a-poke.herokuapp.com
data:
  existing: field
  joke: |
    What gets wetter the more it dries? A towel!
  jest: |
    What do you get when you cross a snowman with a vampire? Frostbite!
```

Errors resulting from user input, that is the annotation, are sent to the
ConfigMap's events (see `kubectl describe configmap`).

## How-to

```shell
# To test
go test .

# To build
go build .

# To start locally
KUBECONFIG=~/.kube/config ./k8s-curl

# To start in-cluster
./k8s-curl
```

## Questions

> How would you deploy your controller to a Kubernetes cluster?

I would write a Helm chart and deploy my controller as a Deployment, to make
sure the Pod stays available and to be able to easily re-deploy it. I would
also set the update strategy to `Recreate` to avoid potential bugs due to
concurrency. The chart would also need proper RBAC for its queries to the
apiserver.

> Kubernetes is level-based as opposed to edge-triggered. What benefits does it
> bring in the context of your controller?

With edge triggering, the controller would only process ConfigMaps when it is
ready to do so. If the controller crashes, or if it is deployed after
ConfigMaps are changed, it would not process them.
Level triggering also means it's possible to aggregate several updates together
to reduce processing.

> Kubernetes being a distributed system based on eventual consistency, how do
> reconcialiation loops cope with concurrent read and writes that might
> interfere with each other?

The loops themselves are part of the eventual consistency. Each loop watches
for inconsistencies between the observable state and the desired state of the
section it is managing. When the propagation of the desired state is delayed
because of eventual consistency, this will in turn lead to a propagation delay
of the observable state. But hopefully after enough time has passed, each loop
gets the same view of the desired state and can do its part to converge to a
balance where the observable matches the desired state.
