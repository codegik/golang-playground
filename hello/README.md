# First Steps with GO Language

## Creating module

```shell
go mod init codegik.com/hello
go run .
```

## Redirecting dependencies
```shell
go mod edit -replace codegik.com/greetings=../greetings
```

## Download dependencies
```shell
go mod tidy
```

## Running

```shell
go run .
```
