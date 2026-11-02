# what-to-eat-api
## generate
```
go run github.com/99designs/gqlgen generate
```
and
```
go generate ./...
```

## Run unit tests
Run all tests:
```
go test ./...
```
Run with verbose output:
```
go test -v ./...
```
Run a specific test:
```
go test -v -run TestCustomValidator ./...
```

## Start application
1. Install nodemon
   ```
   npm i -g nodemon
   ```
2. Run
    ```
    sh run.sh
    ```