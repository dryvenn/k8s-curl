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
