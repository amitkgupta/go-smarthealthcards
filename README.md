# go-smarthealthcards

Golang libraries for generating QR codes for [SMART Health Cards](https://smarthealth.cards/en/)
representing COVID-19 Immunizations.

## Demo

[![demo](https://user-images.githubusercontent.com/1580649/144734873-55bf1beb-22bc-4eb4-a3fa-82a634417945.gif)](https://akgupta.ca/smart-health-cards)

## Documentation

[![godoc](https://godoc.org/github.com/amitkgupta/go-smarthealthcards/v2?status.png)](https://godoc.org/github.com/amitkgupta/go-smarthealthcards/v2)

## Example

You can use the libraries in this module independently, or you can compose the libraries in this
module together for a full end-to-end solution:

#### Generate a signing key and set environment variables

```
$ eval `go run utils/keygen.go`

$ env | grep SMART_HEALTH_CARDS
SMART_HEALTH_CARDS_KEY_Y=101429470610882177913719193785842901742785774962016470785491662750285266794880
SMART_HEALTH_CARDS_KEY_X=54331567703018507947599648321661141913001722275227305175319502486118882894610
SMART_HEALTH_CARDS_KEY_D=71127180180681625720019072005809291232785768180646325329981160435676730627285
```

#### Start an example web server

```
$ go run examples/server.go
```

#### Inspect the JSON Web Key Set representation of the signing key's public key

```
$ curl -s http://localhost:8080/.well-known/jwks.json | jq .
{
  "keys": [
    {
      "kty": "EC",
      "kid": "9G2pzRWd-FL4XwNpDuXUHnG5egt38E78hSqMQzL5v3E",
      "use": "sig",
      "alg": "ES256",
      "crv": "P-256",
      "x": "eB6T2wFY60skcvNQAQPS5l_yhCEnrwo5P6yoHIqQYxI",
      "y": "4D8LwoIvKk7di9p83_8oTMvr3VJootJKC6iL1cuJuYA"
    }
  ]
}
```

#### Generate a QR code

```
$ curl -s -X POST http://localhost:8080 \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "\
family_name=Salk\
&given_names=Jonas\
&date_of_birth=1914-10-28\
&first_immunization_performer=MyLocalHospital\
&first_immunization_lot_number=LN01234\
&first_immunization_vaccine_type=Pfizer\
&first_immunization_date=2021-06-01" \
  -o /tmp/qr.png

$ open /tmp/qr.png
```

![](/examples/qr.png)

## Limitations

- This module currently only supports certain COVID-19 immunizations; with minor modifications it
could be generalized to support all COVID-19 immunizations, and even immunizations of other diseases
- This module does not support other types of SMART Health Cards such as those for dianoses or lab
results, only immunizations

## License

[MIT License](/LICENSE)
