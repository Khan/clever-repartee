### Clever Repartee

The [Clever API](https://dev.clever.com/) has [OpenAPI specs that are available](https://github.com/Clever/swagger-api).

This repository uses the OpenAPI specs to autogenerate a client to access the Clever API.

The command line application requires some environment variable to function:

```
CLEVER_ID
CLEVER_SECRET
MAP_CLEVER_ID
MAP_CLEVER_SECRET
FROM_EMAIL //gmail account
TO_EMAIL
GMAIL_PASSWORD
```

The Gmail password should be an [App Passwords for GMAIL](https://support.google.com/accounts/answer/185833?p=InvalidSecondFactor&visit_id=637336409852469141-2997794709&rd=1) so you must [Add an App Password to Gmail](https://myaccount.google.com/apppasswords).

You are limited to 99 emails per 24 hours using this mechanism.

### Sample Usage
```
make build
clever-repartee diff -district=${DISTRICT_ID} -json
```
The `-json` flag will attempt to write the summary report out to a local json file.

### Background
At Khan Academy, we use the [OpenAPIv2 spec file here](https://github.com/Clever/swagger-api/blob/master/full-v2.yml), convert it to OpenAPI **v3** format, and use [oapi-codegen](https://github.com/deepmap/oapi-codegen) to autogenerate API-contract compliant golang clients for the V2.1 Clever API.

We believe that API First (or Document Driven Design) is an engineering and architecture best practice. API First involves establishing an API contract, separate from the code. This allows us to more clearly track the evolution of that API contract, separate from the evolution of the implementation of that contract. API contracts can be specified following the OpenAPI Specification (previously Swagger).

We know that there is a [clever-go](https://github.com/Clever/clever-go) library, but we prefer to track API specification updates both more closely and more proactively. We also inject some fault tolerance by injecting an [http client that will retry with exponential backoff](https://github.com/sethgrid/pester) to provide resilience against temporary network failures and exceeding rate limits.

### OpenAPI Document Driven Process

1. We downloaded the [Clever V2.1 OpenAPIv2 spec file here](https://github.com/Clever/swagger-api/blob/master/full-v2.yml).

2. The Clever spec document is in swagger (OpenAPIv2) format, so we use `swagger2openapi` to convert it to OpenAPI **v3** (and patch a few minor issues).
  + Other alternatives are to use [https://editor.swagger.io/](https://editor.swagger.io/) or `api-spec-converter`, but [swagger2openapi](https://github.com/Mermade/oas-kit/tree/master/packages/swagger2openapi) is less prone to failures and it's "patch minor issues" feature helps improve specs.
```
swagger2openapi --patch -y "${swaggerfile}" -o "./oas3/${oapi_output}.oas3.yaml"
```  
3. Lint the resulting spec file for warnings using [Speccy](https://github.com/wework/speccy):
```
speccy lint -v ./oas3/full-v2.oas3.yaml
```
4. Run [stoplight prism](https://github.com/stoplightio/prism) Mock server for integration testing
```
prism mock -h 0.0.0.0 -p 4010 ./oas3/full-v2.oas3.yaml
```
5. Generate Golang client from spec using [Oapi-codegen](https://github.com/deepmap/oapi-codegen):
```
go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --generate types,client --package=clever -o ./clever.gen.go ./oas3/full-v2.oas3.yml
```

[OpenAPI Client and Server Code Generator](https://github.com/deepmap/oapi-codegen) is the most popular Go tool for this purpose, but [most languages have similar support](https://openapi.tools/) and we've adopted it as a best practice to help save on costly API client maintenance and tedious contract testing. 