# Companies API

A simple Go HTTP API to manage companies, with JWT authentication and CDC.

## Getting Started

1. Clone the repository.
2. Launch the app using one of several ways. Definition of tasks definitions from launching the app are provided in the [taskfile](./Taskfile.yml), requires [taskfile binary](https://taskfile.dev/) to be run.
    - Run the full docker compose [file](./build/docker-compose-full.yaml) or `task full`. 
    It launches a Postgres and Kafka, docker container, builds and runs the app container using the config from the **container** [config file](./config/config.container.yaml). 
    - Launch the Postgres and Kafka containers first, then the app.
    In this case, the app uses the **dev** [config file](./config/config.dev.yaml). 
        - `task db-up`, `task kafka-up` [if configured], `task run`, or
        - `task dev` to launch all.
3. The app is set to listen on `localhost:8080`.

## Integration test

The task `integration-test` runs an integration test that verifies a complete flow of operations (Kafka logic is not included), along with some invalid cases.

## Authentication

- Login endpoint returns a JWT token in the JSON response.
- The POST, PATCH, DELETE endpoints are protected, the JWT token should be included in the `Authorization: Bearer <token>`.

## DB

The app uses a PostgreSQL database.
The database schema is defined in [schema.sql](./db/schema.sql), while the migrations applied by the application are located in [migrations](./internal/migrations/migrations/).

## CDC

The app performs CDC (Change Data Capture), communicating the mutating DB changes (Create,Update,Delete).
Currently two methods are supported, controlled by the config `cdc.operator` field.
- `cdc.operator = log`: output in the app logs (default).
- `cdc.operator = kafka` : output to a Kafka topic for downstream ingestion for other services.

  The definition of the Kafka message is as follows:
  - Key: `<op>/<id>, op=c/u/d`
  - Value: a proto-encoded message, [proto message definition](./api/cdc.proto).

## Configuration

The app uses YAML configuration files. Two preset configuration files are currently provided in the [config](./config/) directory, one for
running the service locally and the other for running the service in the docker compose network.

## TODOs

- The user management procedure is very crude, the users are inserted on the SQL migration process and the passwords are stored unencrypted.
    - The app should provide a CRUD user management system and possibly a RBAC / Authorization system for the endpoints.
- More verbose errors to the user.
- Switch to gRPC for communication.
- More tests.

## Endpoints

| Method | Path                  | Description                                | Auth required |
|--------|---------------------- |--------------------------------------------|---------------|
| GET    | /companies            | List all companies                         | No            |
| GET    | /companies/{id}       | Get a single company by id                 | No            |
| POST   | /companies            | Create a new company                       | Yes           |
| PATCH  | /companies/{id}       | Update an existing company by id           | Yes           |
| DELETE | /companies/{id}       | Delete a company by id                     | Yes           |

## Example usage

### Login
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass"}'
```

### Create a company
```bash
curl -X POST http://localhost:8080/companies \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name":"New Company","employees":10,"registered":true,"company_type":"NonProfit"}'
```

### Get its details
```bash
curl -X POST http://localhost:8080/companies/<id>
```

### Patch one of its fields
```bash
curl -X PATCH http://localhost:8080/companies/<id> \
  -H "Content-Type: application/json" \
   -H "Authorization: Bearer <token> \
  -d '{
        "employees": 50
      }'
```

### Delete it
```bash
curl -X DELETE http://localhost:8080/companies/<id>
```