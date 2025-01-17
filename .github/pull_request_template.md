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
