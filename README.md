# updatey

Brings Semantic versioning constraints to Kubernetes resources!

# Background

Many Kubernetes clusters rely on third party software to aid your own applications. For example, FileBeat collects logs and ships them to Elastic search. Once deployed into the cluster, either you never update the version or actively have to keep up to date with version releases. Wouldn't it be great if you didn't have to worry about this?
This is where **updatey** comes in, by specifying semantic versioning constraints in your image versions you can let updatey do the work.

# Example on Semantic version constraints
 
 For more information: https://github.com/Masterminds/semver#checking-version-constraints  

**Results may vary due to nginx creating new releases**
## Patch update
Lets say we have a pod manifest:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mypod
  labels:
    app: mypod
spec:
  containers:
  - name: nginx
    image: "nginx:1.14.1"
```

And you want to benefit from updating to the latest patch version:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mypod
  labels:
    app: mypod
spec:
  containers:
  - name: nginx
    image: "nginx:~1.14"
```

Resulting as of **24.03.19**
```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: mypod
  name: mypod
spec:
  containers:
  - image: nginx:1.14.2
    name: nginx
```

## Minior update
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mypod
  labels:
    app: mypod
spec:
  containers:
  - name: nginx
    image: "nginx:1.15.8-alpine"
```

And you want to benefit from updating to the latest minor version:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mypod
  labels:
    app: mypod
spec:
  containers:
  - name: nginx
    image: "nginx:^1.15"
```

Resulting as of **24.03.19**
```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: mypod
  name: mypod
spec:
  containers:
  - image: nginx:1.15.9
    name: nginx
```

# Installation

1. Replace the `ca` and `key` fields in the helm chart with your own.
2. `helm install -n updatey --namespace=<YOUR_NAMESPACE> helm/updatey`

# Caveats

* The current [semantic versioning implementation](https://github.com/Masterminds/semver) doesn't respect pre releases. So `-alpine` won't be respected, this will be fixed in later versions. 

* Updates have to be triggered by updating the resources, which will cause the resource to go through the admission controller, this will be fixed by providing another component responsible for handling this automatically.