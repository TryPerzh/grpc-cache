## Implementation of a simple cache via gRPC.
This package implements a client and server to create a cache that exchanges data via gRPC.

### Example Usage Server
```
import gS "github.com/TryPerzh/grpc-cache/server"

var server *gS.Server

server = gS.NewWithConfig(gS.Config{
     Port:                   "8080",
     DefaultCacheExpiration: 24 * time.Hour,
     CleanupCacheInterval:   10 * time.Minute,
})
server.Tokens.AddUser("testlogin", "testpassword")
server.RunServer()
```
### Example Usage Client
```
import gC "github.com/TryPerzh/grpc-cache/client"

var client *gC.CacheClient

client = gC.New("localhost", "8080", "testlogin", "testpassword")
client.Connect()

client.Set("1", "data", time.Hour) //key, value, storage time

value, f := client.Get("1")
if f {
	fmt.Println(value)
}
```
