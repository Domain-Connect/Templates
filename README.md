# Templates
[![Build Status](https://travis-ci.com/Domain-Connect/Templates.svg?branch=master)](https://travis-ci.com/Domain-Connect/Templates)

Templates for use in the Domain Connect Protocol
These map to the individual service providers for domain connect. See https://www.domainconnect.org/getting-started/

For details on how to constuct a Domain Connect template, refer to section 5.2 and 5.3 of the Domain Connect Spec:

https://github.com/Domain-Connect/spec/blob/master/Domain%20Connect%20Spec%20Draft.adoc#template-definition
https://github.com/Domain-Connect/spec/blob/master/Domain%20Connect%20Spec%20Draft.adoc#template-record

## Template Naming Format

Templates should be named according the following pattern: **providerId.serviceId.json**

For example: **myprovider.com.website.json**

## Template verification

Template can verify for correctness using JSON Schema [template.schema](template.schema).
Passing the schema check is required for the Pull Request to be accepted into the repository.

## Example Template Format

Following is an example of a complete Domain Connect template, with examples of various DNS records included:

```
{
  "providerId": "<Enter providerId>",
  "providerName": "<Enter providerName>",
  "serviceId": "<Enter serviceId>",
  "serviceName": "<Enter serviceName>",
  "version": 1,
  "logoUrl": "<Enter logoUrl>",
  "description": "<Enter description>",
  "variableDescription": "<Enter variableDescription>",
  "syncBlock": false,
  "syncPubKeyDomain": "<Enter syncPubKeyDomain>",
  "syncRedirectDomain": "<Enter syncRedirectDomain>",
  "warnPhishing": true,
  "hostRequired": false,
  "records": [
    {
      "type": "A",
      "host": "@",
      "pointsTo": "1.1.1.1",
      "ttl": 3600
    },
    {
      "type": "A",
      "host": "@",
      "pointsTo": "%a%",
      "ttl": 3600
    },
    {
      "type": "CNAME",
      "host": "www",
      "pointsTo": "@",
      "ttl": 3600
    },
    {
      "type": "CNAME",
      "host": "sub",
      "pointsTo": "%sub%.mydomain.com",
      "ttl": 3600
    },
    {
      "type": "CNAME",
      "host": "%cnamehost%",
      "pointsTo": "%sub%.mydomain.com",
      "ttl": 3600
    },
    {
      "type": "TXT",
      "host": "@",
      "data": "%txt%",
      "ttl": 3600
    },
    {
      "type": "SPFM",
      "host": "@",
      "spfRules": "include:spf.mydomain.com"
    },
    {
      "type": "MX",
      "host": "@",
      "pointsTo": "1.1.1.2",
      "priority": 0,
      "ttl": 3600
    },
    {
      "type": "MX",
      "host": "@",
      "pointsTo": "%mx%",
      "priority": 0,
      "ttl": 3600
    },
    {
      "type": "SRV",
      "service": "_sip",
      "protocol": "_tls",
      "port": "443",
      "weight": "20",
      "priority": 10,
      "name": "@",
      "target": "%target%",
      "ttl": 3600
    }
  ]
}
```
