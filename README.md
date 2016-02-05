# Public API for Brands (public-brands-api)
Provides a public API for Brands data

## Build & deployment etc:
*TODO*
_NB You will need to tag a commit in order to build, since the UI asks for a tag to build / deploy_
* [Jenkins view](http://ftjen10085-lvpr-uk-p:8181/view/public-brands-api)
* [Build and publish to forge](http://ftjen10085-lvpr-uk-p:8181/job/public-brands-api-build)
* [Deploy to test or production](http://ftjen10085-lvpr-uk-p:8181/job/public-brands-api-deploy)


## Installation & running locally
* You will need to load some data, see [Brands RW API](https://github.com/Financial-Times/brands-rw-neo4j) for help with that
* `go get -u github.com/Financial-Times/public-brands-api`
* `cd $GOPATH/src/github.com/Financial-Times/public-brands-api`
* `go test ./...`
* `go install`
* `$GOPATH/bin/public-brands-api --neo-url={neo4jUrl} --port={port} --log-level={DEBUG|INFO|WARN|ERROR}`
_Both arguments are optional.
--neo-url defaults to http://localhost:7474/db/data, which is the out of box url for a local neo4j instance.
--port defaults to 8080._
* `curl http://localhost:8080/brands/2d3e16e0-61cb-4322-8aff-3b01c59f4daa | json_pp`


## API definition
* The API only supports HTTP GET requests and only takes one parameter, uuid:
  `http://api.ft.com/brands/{uuid}`
* The an example result structure is shown bellow, _note that when there is no parent brand then we omit the parent attribue_:

```
{
  "id": "http://api.ft.com/things/{uuid}",
  "apiUrl": "http://api.ft.com/brands/{uuid}",
  "types": [
    "http://www.ft.com/ontology/product/Brand"
  ],
  "prefLabel": "Brand Name",
  "description": "A description of the brand, in plain text",
  "descriptionXML": "<body><p>A description of the brand, in <i>bodyXML</i></p></body>",
  "strapline": "A subsidiary heading, caption or advertising slogan",
  "_imageUrl": "http://images.ft.com/tempImageWhilstThisIsResolved.jpg",
  "childBrands": []
}
```

* Brands can have parents, as illustrated bellow when considering Lex Live:

```
{
  "id": "http://api.ft.com/things/e363dfb8-f6d9-4f2c-beba-5162b334272b",
  "apiUrl": "http://api.ft.com/brands/e363dfb8-f6d9-4f2c-beba-5162b334272b",
  "types": [
    "http://www.ft.com/ontology/product/Brand"
  ],
  "prefLabel": "Lex Live",
  "description": "",
  "descriptionXML": "",
  "strapline": "",
  "_imageUrl": "",
  "parentBrand": {
    "id": "http://api.ft.com/things/2d3e16e0-61cb-4322-8aff-3b01c59f4daa",
    "apiUrl": "http://api.ft.com/brands/2d3e16e0-61cb-4322-8aff-3b01c59f4daa",
    "types": [
      "http://www.ft.com/ontology/product/Brand"
    ],
    "prefLabel": "Lex"
  },
  "childBrands": []
}
```

* As well a parents, brands can have children as shown bellow:

```
{
  "id": "http://api.ft.com/things/2d3e16e0-61cb-4322-8aff-3b01c59f4daa",
  "apiUrl": "http://api.ft.com/brands/2d3e16e0-61cb-4322-8aff-3b01c59f4daa",
  "types": [
    "http://www.ft.com/ontology/product/Brand"
  ],
  "prefLabel": "Lex",
  "description": "",
  "descriptionXML": "",
  "strapline": "",
  "_imageUrl": "",
  "parentBrand": {
    "id": "http://api.ft.com/things/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54",
    "apiUrl": "http://api.ft.com/brands/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54",
    "types": [
      "http://www.ft.com/ontology/product/Brand"
    ],
    "prefLabel": "Financial Times"
  },
  "childBrands": [
    {
      "id": "http://api.ft.com/things/e363dfb8-f6d9-4f2c-beba-5162b334272b",
      "apiUrl": "http://api.ft.com/brands/e363dfb8-f6d9-4f2c-beba-5162b334272b",
      "types": [
        "http://www.ft.com/ontology/product/Brand"
      ],
      "prefLabel": "Lex Live"
    }
  ]
}
```

## Healthchecks and other non-functional services
* Healthchecks: [http://localhost:8080/__health](http://localhost:8080/__health)
* Ping: [http://localhost:8080/__ping](http://localhost:8080/__ping) and [http://localhost:8080/ping](http://localhost:8080/ping)
* BuildInfo: [http://localhost:8080/__build-info](http://localhost:8080/__build-info) and [http://localhost:8080/build-info](http://localhost:8080/build-info)

## Todo
* Implement build-info properly
* Documentation for API Gateway
