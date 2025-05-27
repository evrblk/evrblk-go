# Everblack Go SDK


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

moabClient := moab.NewMoabPreviewGrpcClient("moab.us-east-2.api.evrblk.com", signer)

createQueueResp, err := moabClient.CreateQueue(context.Background(), &moab.CreateQueueRequest{
    Name:                      "my_queue_1",
    Description:               "Some description",
    KeepaliveTimeoutInSeconds: 15,
    ExpiresInSeconds:          86400,
})
```

## License

Everblack Go SDK is released under the [MIT License](https://opensource.org/licenses/MIT).
