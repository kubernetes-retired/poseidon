Poseidon is Firmament's (http://www.firmament.io) integration with
Kubernetes.

[![Build Status](https://travis-ci.org/camsas/poseidon.svg)](https://travis-ci.org/camsas/poseidon)
[![Coverage Status](https://coveralls.io/repos/github/camsas/poseidon/badge.svg?branch=master)](https://coveralls.io/github/camsas/poseidon?branch=master)

***Note: this repo contains an initial prototype, it may break at any time! :)***

# Getting started

The easiest way to get Poseidon up and running is to use our Docker image:

```
$ docker pull camsas/poseidon:dev
```
Once the image has downloaded, you can start Poseidon as follows:
```
$ docker run camsas/poseidon:dev poseidon \
    --logtostderr \
    --kubeConfig=<path_kubeconfig_file> \
    --firmamentAddress=<host>:<port> \
    --statsServerAddress=<host>:<port> \
    --kubeVersion=<Major.Minor>
```
Note that Poseidon will try to schedule for Kubernetes even if `kube-scheduler`
is running. How you best avoid conflicts depends on the version of Kubernetes
you are running:
 1. If you are using Kubernetes <1.6, shut down `kube-scheduler` first, and make
    sure your pods are labeled with `scheduler: poseidon`.
 2. If you are using Kubernetes 1.6+, you do not have to to shut down
    `kube-scheduler`, but you must specify `schedulerName: poseidon` in your pod
    specs.

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

Run:

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
these [instructions](https://github.com/camsas/firmament/blob/master/README.md#running-the-firmament-scheduler-service)
to build and deploy Firmament.


Finally, to start up Poseidon, run:

```
$ poseidon --logtostderr \
    --kubeConfig=<path_kubeconfig_file> \
    --firmamentAddress=<host>:<port> \
    --statsServerAddress=<host>:<port> \
    --kubeVersion=<Major.Minor>
```

The configuration of the Firmament scheduler service is controlled by the
configuration file in `${FIRMAMENT_HOME}/config/firmament_scheduler.cfg`. You
can modify this file to control scheduling features (e.g. choose a scheduling
policy). To apply changes, you need to restart (but not recompile) the
Firmament scheduler service. For more info, check
[the arguments accepted by Firmament](https://github.com/camsas/firmament/blob/master/README.md#using-the-flow-scheduler).
All of these arguments can be set via the configuration file.

## Contributing

We always welcome contributions to Poseidon. We use GerritHub for our code
reviews, and you can find the Poseidon review board there:

https://review.gerrithub.io/#/q/project:camsas/poseidon+is:open

In order to do code reviews, you will need an account on GerritHub (you can link
your GitHub account).

The easiest way to submit changes for review is to check out Poseidon from
GerritHub, or to add GerritHub as a remote. Alternatively, you can submit a pull
request on GitHub and we will import it for review on GerritHub.
