# Prow Addons Controller Manager

## Overview

The Prow Addons Controller Manager embeds all custom controller extension for Prow infrastructure.    

## Prerequisites

Use the following tools to set up the project:

* [Go distribution](https://golang.org)
* [Docker](https://www.docker.com/)
* [Mockery](https://github.com/vektra/mockery) 

## Usage

### Run a local version

To run the controller outside the cluster, run this command:

```bash
make run
```

### Build a production version

To build the production Docker image, run this command:

```bash
IMG={image_name}:{image_tag} make docker-build
```

The variables are:

* `{image_name}` which is the name of the output image.
* `{image_tag}` which is the tag of the output image.


## Development

### Install dependencies

This project uses `dep` as a dependency manager. To install all required dependencies, use the following command:
```bash
make resolve
```

### Run tests

To run all unit tests, use the following command:

```bash
make test
```
