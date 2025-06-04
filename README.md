# Everblack Go SDK

[![Go](https://github.com/evrblk/evrblk-go/actions/workflows/go.yml/badge.svg)](https://github.com/evrblk/evrblk-go/actions/workflows/go.yml)

The official Go SDK for Everblack services. Also, package `authn` is the reference implementation of authentication 
mechanism, which is used internally in Everblack Cloud to verify signatures.

## Versioning

Each service is versioned independently of each other. Packages are organized by major versions. For example, Preview
version of Moab API is available with `import moab "github.com/evrblk/evrblk-go/moab/preview"`, V1 version will be 
available with `import moab "github.com/evrblk/evrblk-go/moab/v1"`, and so forth. It is guaranteed that all minor changes
are backward compatible with old SDKs.

Currently available versions:

* Moab
    * `preview`
* Grackle
    * `preview`
* IAM
    * `preview`
* My Account
    * `preview`

## Installing

Use go get to install the latest version of the library.

```
go get -u github.com/evrblk/evrblk-go@latest
```

## Example

```go
import (
    evrblk "github.com/evrblk/evrblk-go"
    moab "github.com/evrblk/evrblk-go/moab/preview"
)

apiKeyId := "key_alfa_z141pKeFzfmGGyYlUyPsbF"
privatePem := "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEIN33cCNGxsuxwMaJ2jWvWcgxBSVr8HV7WUUSKGc71/BtoAoGCCqGSM49\nAwEHoUQDQgAE0m8+ZVijytLp01dsupG7QF8ZpjX5UmP20wj/sluPdoHW3BgiiyCn\n/pMwYptUs0yJUtUZ/0wzEyp8PgAWWhxglw==\n-----END EC PRIVATE KEY-----"

signer := evrblk.NewAlfaRequestSigner(apiKeyId, privatePem)

moabClient := moab.NewMoabGrpcClient("moab.us-east-2.api.evrblk.com", signer)

createQueueResp, err := moabClient.CreateQueue(context.Background(), &moab.CreateQueueRequest{
    Name:                      "my_queue_1",
    Description:               "Some description",
    KeepaliveTimeoutInSeconds: 15,
    ExpiresInSeconds:          86400,
})
```

## License

Everblack Go SDK is released under the [MIT License](https://opensource.org/licenses/MIT).
