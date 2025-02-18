# dictionaryGO
Service to store and manipulate translations from Polish to English with examples.


## Running App
In order to run the app you must create `.env` file in project's root directory based on `.env.example`. You need to replace DB_USER and DB_PASSWORD for yours. Then in the same directory run:

```bash
docker compose up -d
```

This will start Postresql database server on localhost:5432.

To run the app use:

```bash
go run main.go
```
App will start on localhost:8080. You can open it in a browser to use GraphQL playground. Example usage of queires and mutations is provided in `example_usage.md` file.

To run the tests use:

```bash
go test ./...
```
