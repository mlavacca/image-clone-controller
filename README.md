# Image-clone-controller

## Problem description

We have a Kubernetes cluster on which we can run applications. These applications will often use
publicly available container images, like official images of popular programs, e.g. Jenkins,
PostgreSQL, and so on. Since the images reside in repositories over which we have no control, it
is possible that the owner of the repo deletes the image while our pods are configured to use it.
In the case of a subsequent node rotation, the locally cached copies of the images would be
deleted and Kubernetes would be unable to re-download them in order to re-provision the
applications.

## Problem solution

The image-clone-controller watches the applications and “caches” the images by re-uploading to your
own registry repository and reconfiguring the applications to use these copies.

The controller watches and patches all the deployments and the daemonsets in all the namespaces
except for kube-system.

## Demo

[![asciicast](https://asciinema.org/a/SrgRSAmIx2JOUs14GazEoiy2n.svg)](https://asciinema.org/a/SrgRSAmIx2JOUs14GazEoiy2n)

## Installation

To use the image-clone-controller, type:
```bash
kubectl apply -f deployments
```

This operation will create:
* a new namespace `images-backup`
* the image-clone-controller deployment
* a serviceAccount
* the rbacs related to the new service account
* a secret containing the docker credentials

### customize the docker config file

A testing docker config file that uses the `index.docker.io/mlvtask` repository has been set in the secret. To use your
own backup repository:
1. customize the secret by setting the base64 of your docker config file in the `config.json` field;
2. customize the controller deployment arguments, by setting the proper registry and repository.

The provided docker configuration in plain text can be found below:
```
{
	"auths": {
		"https://index.docker.io/v1/": {
			"auth": "bWx2dGFzazpkMGY5ZWQ4OC0zZGUwLTQ4MGQtYmRiOC0zZmE1MjcyNTdlZjU="
		}
	}
}
```

## Verify you are using the backup images

Once your controller is deployed, you can check it is correctly working  by simply typing
```bash
kubectl get deploy -A -o=jsonpath='{range .items[*]}{.metadata.namespace}{"\t\t"}{.spec.template.spec.containers[*].image}{"\n"}{end}'
```
for deployments, and
```bash
kubectl get daemonset -A -o=jsonpath='{range .items[*]}{.metadata.namespace}{"\t\t"}{.spec.template.spec.containers[*].image}{"\n"}{end}'
```
for daemonsets. 

All the deployments and the daemonsets images (both initContainers and containers) should be replaced by
images stored in the backup registry (except for those in `kube-system`).
