Poseidon is Firmament's (http://www.firmament.io) integration with
Kubernetes.

***Note: nothing to see here at present -- we will push our initial prototype
work here, but do not expect working code! :)***

## System requirements

 * CMake 2.8+
 * Docker 1.7+
 * Kubernetes v1.1+

The build process will install local version of Poseidon's dependencies, which
currently include:

 * [swagger-codegen](https://github.com/swagger-api/swagger-codegen/) v2.1.4
 * the [Microsoft C++ REST SDK](https://github.com/Microsoft/cpprestsdk) v2.7.0

# Getting started

First, generate the build infrastructure:

```
$ cmake .
```

Then, build Poseidon:

```
$ make
```

## Contributing

We always welcome contributions to Poseidon. We use GerritHub for our code
reviews, and you can find the Poseidon review board there:

https://review.gerrithub.io/#/q/project:ms705/poseidon+is:open

In order to do code reviews, you will need an account on GerritHub (you can link
your GitHub account).

The easiest way to submit changes for review is to check out Poseidon from
GerritHub, or to add GerritHub as a remote. Alternatively, you can submit a pull
request on GitHub and we will import it for review on GerritHub.
