Poseidon is Firmament's (http://www.firmament.io) integration with
Kubernetes.

***Note: nothing to see here at present -- we will push our initial prototype
work here, but do not expect working code! :)***

## System requirements

 * CMake 2.8+
 * Docker 1.7+
 * Kubernetes v1.1+
 * Boost 1.54+
 * libssl 1.0.0+

The build process will install local version of Poseidon's dependencies, which
currently include:

 * the [Microsoft C++ REST SDK](https://github.com/Microsoft/cpprestsdk) v2.7.0
 * the [Firmament scheduler](https://github.com/camsas/firmament) (HEAD)

and their dependencies.

A known-good build environment is Ubuntu 16.04 with gcc 5.1.3.

# Getting started

First, generate the build infrastructure:

```
$ cmake .
```

Then, build Poseidon:

```
$ make
```

To start up, run

```
$ build/poseidon --k8s_api_server_host=<hostname> --k8s_api_server_port=8080
```

Additional arguments (e.g. to choose a scheduling policy) follow those
[accepted by Firmament](https://github.com/camsas/firmament/README.md).


## Contributing

We always welcome contributions to Poseidon. We use GerritHub for our code
reviews, and you can find the Poseidon review board there:

https://review.gerrithub.io/#/q/project:ms705/poseidon+is:open

In order to do code reviews, you will need an account on GerritHub (you can link
your GitHub account).

The easiest way to submit changes for review is to check out Poseidon from
GerritHub, or to add GerritHub as a remote. Alternatively, you can submit a pull
request on GitHub and we will import it for review on GerritHub.
