# Description

<-- short description of the template(s) and/or reason for update -->

## Type of change

Please mark options that are relevant.

- [ ] New template
- [ ] Bug fix (non-breaking change which fixes an issue in the template)
- [ ] New feature (non-breaking change which adds functionality to the template)
- [ ] Breaking change (fix or feature that would cause existing template behavior to be not backward compatible)

# How Has This Been Tested?

Please mark the following checks done
- [ ] Schema validated using JSON Schema [template.schema](./template.schema)
- [ ] Template functionality checked using [Online Editor](https://domainconnect.paulonet.eu/dc/free/templateedit)
- [ ] Template is checked using [template linter](https://github.com/Domain-Connect/dc-template-linter)
- [ ] Template file name follows the pattern `<providerId>.<serviceId>.json`
- [ ] resource URL provided with `logoUrl` is actually served by a webserver

# Checklist of common prolems of issues (mark all the checkboxes after conducting the check). Comment on any point which is not fulfilled.
- [ ] digital signatures are used and `syncPubKeyDomain` specified (yes, `warnPhishing` is an option, but some providers reject such templates by policy, so signing shall be a default)
- [ ] `syncRedirectDomain` is specified when intended to use `redirect_uri` parameter in the synchronous flow
- [ ] no TXT record with SPF content (i.e. `"v=spf1 ..."`) instead of using SPFM record type on APEX
- [ ] `txtConflictMatchingMode` is set on TXT records which shall be unique on a label (like DMARC)
- [ ] variables are set to the smallest scope needed (i.e. limit possibility to be misused to set any arbitrary record and conflict with other template). Too broad scope example: @ TXT "%verification%". Better usage: @ TXT "foo-verification=%verification%".
- [ ] no variables as a host name to apply template on subdomain instead of standard `host` parameter
- [ ] no explicit usage of `%host%` variable in `host` attribute 
- [ ] `essential` setting is used on records, which the user shall be able to change or remove manually later without dropping the whole template (like DMARC)    

# Example variable values
<-- to make review process easier please provide example set of variable values for this template -->

<-- Example: -->

```
var1: aaa
var2: foo.com
```

<-- Or provide the whole `testData` object from the [Online Editor](https://domainconnect.paulonet.eu/dc/free/templateedit) after testing and using "Add as test" button -->
```
"testData": {
    "testset": {
      "variables": {
        "domain": "example.com",
        "host": "foo",
        "example": "bar"
      },
      "results": [
        {
          "type": "TXT",
          "name": "foo",
          "ttl": 86400,
          "data": "\"bar\""
        }
      ]
    }
  }
```
