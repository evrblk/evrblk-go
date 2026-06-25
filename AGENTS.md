# Everblack Go SDK Development Guide

This is the main Go SDK for all public APIs for Everblack services: Grackle, Moab, Banyan.
It also contains Everblack Cloud related APIs: IAM, My Account.

- All protobuf definitions live in `./proto` split by services and major version number.
- Each gRPC client is wrapped into a client generated with `./cmd/codegen`.
- Package `authn` contains the reference implementation of Everblack authentication.

## Build & Test Commands

```bash
make build                    # generate all code and fully build the SDK
go test -v ./...              # run all tests with Go directly
```

## Code Style Guidelines

- Follow standard Go formatting (gofmt/goimports)
- Import order: standard lib, external packages (including other `evrblk/*` repositories), then `evrblk/grackle` packages
- Error handling: Always check errors with `if err != nil { return ... }`
- Document all exported functions, types, and variables
- Use table-driven tests when appropriate
- Use `testify/require` for test assertions
- In tests use `EqualValues` when comparing integers instead of `Equal` with a typecast
