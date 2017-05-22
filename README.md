Poseidon is Firmament's (http://www.firmament.io) integration with
Kubernetes.

[![Build Status](https://travis-ci.org/camsas/poseidon.svg)](https://travis-ci.org/camsas/poseidon)

***Note: this repo contains an initial prototype, it may break at any time! :)***

# Getting started

The easiest way to get Poseidon up and running is to use our Docker image:

```
$ docker pull camsas/poseidon:dev
```
Once the image has downloaded, you can start Poseidon as follows:
```
$ docker run camsas/poseidon:dev /usr/bin/poseidon \
    --logtostderr \
    --kubeConfig=<path_kubeconfig_file> \
    --firmamentAddress=<host>:<port> \
    --statsServerAddress=<host>:<port> \
    --kubeVersion=<Major.Minor>
```
Note that Poseidon will try to schedule for Kubernetes even if `kube-scheduler`
is running. If you are using a Kubernetes version older than 1.6 then to avoid
conflicts, shut down `kube-scheduler` first, and make sure your pods are
labeled with `scheduler : poseidon`. If you are using Kubernetes 1.6+ then you
do not have to to shut down `kube-scheduler`, but you must specify
`schedulerName: poseidon` in your pod specs.

You will also need to ensure that the API server is reachable from the Poseidon
container's network (e.g., using `--net=host` if you're running a local
development cluster).

# Building from source

## System requirements

 * Go 1.7+
 * Docker 1.7+
 * Kubernetes v1.1+

A known-good build environment is Ubuntu 14.04.

## Build process

Run :

```
$ go build
$ go install
```

Following, make sure you have a Kubernetes cluster running. If you're running Ubuntu on amd64 then you can execute:

```
./deploy/build_kubernetes.sh
./deploy/run_kubernetes.sh
```

Next, make sure you have a Firmament scheduler service running. You can follow
these [instructions](https://github.com/camsas/firmament/blob/master/README.md)
to build and deploy Firmament.


Finally, to start up Poseidon, run:

```
$ poseidon --logtostderr \
    --kubeConfig=<path_kubeconfig_file> \
    --firmamentAddress=<host>:<port> \
    --statsServerAddress=<host>:<port> \
    --kubeVersion=<Major.Minor>
```

Arguments in `${FIRMAMENT_HOME}/config/firmament_scheduler.cfg` control flow
scheduling features (e.g. to choose a scheduling policy). For more info check
[accepted by Firmament](https://github.com/camsas/firmament/blob/master/README.md).

## Contributing

We always welcome contributions to Poseidon. We use GerritHub for our code
reviews, and you can find the Poseidon review board there:

https://review.gerrithub.io/#/q/project:camsas/poseidon+is:open

In order to do code reviews, you will need an account on GerritHub (you can link
your GitHub account).

The easiest way to submit changes for review is to check out Poseidon from
GerritHub, or to add GerritHub as a remote. Alternatively, you can submit a pull
request on GitHub and we will import it for review on GerritHub.
