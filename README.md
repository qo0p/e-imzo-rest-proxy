# Build

## Windows OS

```
go mod tidy
GOOS=windows GOARCH=386 go build -o e-imzo-rest-proxy.exe main.go
```

# Test with CURL

## Get E-IMZO Version

```
curl -v -X POST -d '{"name":"version"}' http://localhost:64647
```

## Install ApiKey

```
curl -v -X POST -d '{"name":"apikey","arguments":["localhost","96D0C1491615C82B9A54D9989779DF825B690748224C2B04F500F370D51827CE2644D8D4A82C18184D73AB8530BB8ED537269603F61DB0D03D2104ABF789970B"]}' http://localhost:64647
```

## Get List of all PFX keys

```
curl -v -X POST -d '{"plugin":"pfx","name":"list_all_certificates"}' http://localhost:64647
```

_See http://127.0.0.1:64646/apidoc.html_