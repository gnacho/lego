---
title: "DonDominio"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dondominio
dnsprovider:
  since:    "v5.7.0"
  code:     "dondominio"
  url:      "https://www.dondominio.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dondominio/dondominio.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [DonDominio](https://www.dondominio.com/).


<!--more-->

- Code: `dondominio`
- Since: v5.7.0


Here is an example bash command using the DonDominio provider:

```bash
DONDOMINIO_API_USER="xxx" \
DONDOMINIO_API_PASSWORD="yyy" \
lego run --dns dondominio -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DONDOMINIO_API_PASSWORD` | API password |
| `DONDOMINIO_API_USER` | API user |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DONDOMINIO_API_URL` | API URL (Default: https://simple-api.dondominio.net) |
| `DONDOMINIO_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `DONDOMINIO_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 10) |
| `DONDOMINIO_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 300) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

## Description

The API access requires credentials (`apiuser` and `apipasswd`) provided by DonDominio/MrDomain.

The API is not designed to be used for mass requests: check the [usage policy](https://dondominio.dev/en/api/docs/usage-policy/).

The `ttl` parameter of the `service/dnscreate/` action is not exposed and is never sent because the API rejects it.



## More information

- [API documentation](https://dondominio.dev/en/api/docs/api/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dondominio/dondominio.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
