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

```
make build
clever-repartee -district=${DISTRICT_ID} -json
```