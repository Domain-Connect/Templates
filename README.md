# Templates
[![Build Status](https://travis-ci.com/Domain-Connect/Templates.svg?branch=master)](https://travis-ci.com/Domain-Connect/Templates)

Templates for use in the Domain Connect Protocol
These map to the individual service providers for domain connect. See https://www.domainconnect.org/getting-started/

For details on how to constuct a Domain Connect template, refer to section 5.2 and 5.3 of the Domain Connect Spec:

https://github.com/Domain-Connect/spec/blob/master/Domain%20Connect%20Spec%20Draft.adoc#template-definition
https://github.com/Domain-Connect/spec/blob/master/Domain%20Connect%20Spec%20Draft.adoc#template-record

## Online Editor

An [Online Editor](https://domainconnect.paulonet.eu/dc/free/templateedit) is available, with features like syntax check, testing of variables replacement and groups.

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
    "providerId": "exampleProvider",
    "providerName": "Example Provider Name",
    "serviceId": "exampleService",
    "serviceName": "Example Service Name",
    "version": 1,
    "logoUrl": "https://example.com/logo.svg",
    "description": "Example description explaining overall purpose of the record updates",
    "variableDescription": "%a%: domain apex IP; %sub%: sub record destination; %cnamehost%: host pointing to sub destination; %txt%: domain apex text; %mx%: domain apex mail destination; %target%: domain apex service record target; %ttlvar%: variable TTL for SRV record; %srvport%: variable port for SRV record; %srvproto%: variable ptotocol of SRV record; %srvservice%: variable service of SRV record",
    "syncPubKeyDomain": "keys.example.com",
    "syncRedirectDomain": "www.example.com, www.example.net",
    "warnPhishing": true,
    "records": [
        {
            "type": "A",
            "host": "@",
            "pointsTo": "192.0.2.1",
            "ttl": "3600"
        },
        {
            "type": "A",
            "host": "@",
            "pointsTo": "%a%",
            "ttl": "3600"
        },
        {
            "type": "CNAME",
            "host": "www",
            "pointsTo": "@",
            "ttl": "3600"
        },
        {
            "type": "CNAME",
            "host": "sub",
            "pointsTo": "%sub%.mydomain.com",
            "ttl": "3600"
        },
        {
            "type": "CNAME",
            "host": "%cnamehost%",
            "pointsTo": "%sub%.mydomain.com",
            "ttl": "3600"
        },
        {
            "type": "TXT",
            "host": "@",
            "ttl": "3600",
            "data": "%txt%"
        },
        {
            "type": "SPFM",
            "host": "@",
            "ttl": "3600",
            "spfRules": "include:spf.mydomain.com"
        },
        {
            "type": "MX",
            "host": "@",
            "pointsTo": "%mx%",
            "ttl": "3600",
            "priority": "5"
        },
        {
            "type": "MX",
            "host": "@",
            "pointsTo": "192.0.2.2",
            "ttl": "3600",
            "priority": "%mxprio%"
        },
        {
            "type": "SRV",
            "name": "@",
            "ttl": "3600",
            "priority": "10",
            "weight": "20",
            "port": "443",
            "protocol": "_tls",
            "service": "_sip",
            "target": "%target%"
        },
        {
            "type": "CAA",
            "host": "@",
            "ttl": "1800",
            "data": "0 issuewild \"ca2.example.\""
        },
        {
            "type": "SRV",
            "name": "@",
            "ttl": "%ttlvar%",
            "priority": "0",
            "weight": "20",
            "port": "%srvport%",
            "protocol": "%srvproto%",
            "service": "_%srvservice%",
            "target": "srv.example.com"
        }
    ]
}
```

## Template validation tool

Please see https://github.com/Domain-Connect/dc-template-linter
