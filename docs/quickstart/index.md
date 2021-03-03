---
title: Getting started
---

# Getting started

The easiest way to create a simple web app, is to use the [gofer-skeleton](https://github.com/patrickascher/gofer-skeleton) app. 
The skeleton app should give you an idea how to create a web app with the `gofer` framework.

## Install for development

gofer-skeleton offers two ways to install a dev env.
Both ways offer a backend and frontend hot reload.
For performance reasons I would suggest the local env.

### Docker-compose

Requirements:

* [Docker](https://www.docker.com/products/docker-desktop)

Run this command in the application folder.

```
docker-compose up
```

### local

Requirements:

* [GO](https://golang.org/dl/)
* [Node.JS](https://nodejs.org/en/download/)
* DB instance (mysql)

1) import the sql dump file to your mysql instance and define your mysql server in the `config.json` file (databases).

2) Install fresh. Which is a file monitoring tool for GO.

```go 
go get github.com/gravityblast/fresh
```

3) Run the following command in the application folder:

Will start the backend(go) hot reload.

```go 
fresh 
```

switch to the frontend

```go 
cd frontend/
npm install
npm run serve
```

The application is now running under localhost:8080
