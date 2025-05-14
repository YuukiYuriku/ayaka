# :green_book: Documentation

## About

_Pakuningratan_ is skeleton service that can be used as a guidance when creating new service. The structure of this code is following clean architecture's best practice.

Maintainer: dev@runsystem.id, ...

## Features

Here are some features/standarizations that has been implemented on this service:

- code structure (clean architecture)
- graceful shutdown
- logger using golog
- custom middleware
- loading configuration file
  - environment variable
  - configuration yaml variable
- Makefile
- Dockerfile
- gorm
- cache (redis) client
- easy bootstraping
- container and dependency injection
- support cmd custom variable on running

The next roadmap will include:

- endpoint metrics using statsd
- timeout configuration per external client
- integration test
- load test using k6

## Setting Up

**Dependency and requirements:**

- Go 1.21.3 or later
- [cosmtrek/air](https://github.com/cosmtrek/air) (to run watch mode)
- Postgres
- Redis

### Installation

1. Copy env file.

```bash
cp .env.sample .env
```

2. Update `.env` file with your configuration.

3. Download dependency library.

```bash
go mod tidy
```

### Migration

Database migration handled by Gorm

### Quick start

1. To run in watch mode: `make watch` (for development purpose)

2. To build in local arch: `make build`

### Using docker compose

You can run the service using docker compose. Make sure you have docker installed and environment variables are set the `.env` file.

```bash
docker-compose up -d
```

To stop the service, run:

```bash
docker-compose stop
```

If you want to remove the service, run:

```bash
docker-compose down
```

## Template Endpoints

- `GET http://localhost:8000/ping`

- `GET http://localhost:8000/ready`

- `POST http://localhost:8000/v1/templates`

- `GET http://localhost:8000/v1/templates`

- `PUT http://localhost:8000/v1/templates/:id`

- `GET http://localhost:8000/v1/templates/:id`

## Diagram

![diagram-31-03-2022](/assets/arch_diagram.png)

## Philosophies and Rules

- Keep things flat. _Resist the temptation to create subdirectory/subfiles unless we're creating submarines. Usually most of those problems can be solved with good naming._
- If it's hard to understand, it is a valid concern! Start by the assumptions that the code is not good enough.
- Keep things simple and easy. _Don't worry, hardships will come from every direction sooner or later, no need to invite it earlier._
- Try to not make people think when reading a code.
- Fast understandability over fast performance. _Golang is fast enough. And lately clouds are cheap enough to let us spin multiple instances for a service. So yeah, push for common sense but not too far to sacrifice simplicity._
- If it's good, it's copasable. If it's really copasable, then it's perfect.
