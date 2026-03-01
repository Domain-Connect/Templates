# Templates

Templates for use in the Domain Connect Protocol
These map to the individual service providers for domain connect. See https://www.domainconnect.org/getting-started/

For details on how to constuct a Domain Connect template, refer to section 5.2 and 5.3 of the Domain Connect Spec:

https://github.com/Domain-Connect/spec/blob/master/Domain%20Connect%20Spec%20Draft.adoc#template-definition
https://github.com/Domain-Connect/spec/blob/master/Domain%20Connect%20Spec%20Draft.adoc#template-record

## Submitting a Pull Request

Follow these steps carefully when contributing a new or updated template:

### 1. Create and Test Your Template

1. Open the [Online Editor](https://domainconnect.paulonet.eu/dc/free/templateedit).
2. Define your template and verify it passes the syntax check.
3. Test all variable replacements and groups using the editor's built-in testing tools.
4. Perform tests with domain Apex and subdomain using `host` field.
5. Click the **"Copy Markdown"** button in the editor to copy a shareable link to your test results.

> ⚠️ **Testing with the Online Editor is required.** Pull Requests that do not include a link to editor test results will not be reviewed.

> ⚠️ **Test results MUST match the submitted template.** Pull Requests with outdated tests will not be reviewed.

### 2. Name Your File Correctly

Ensure your template file follows the required naming convention and is placed in the root folder of the repository:

```
providerId.serviceId.json
```

For example: `myprovider.com.website.json`

### 3. Open a Pull Request

When opening your Pull Request:

- **Always use the [PR Template](.github/pull_request_template.md)** — it is loaded automatically when you open a new PR on GitHub.
- **Follow the PR template exactly** — fill in every section and do not alter or remove any part of the template structure.
- **Paste the Markdown link** from the Online Editor (copied in step 1) into the designated field in the PR template to share your test results.

> ⚠️ Pull Requests that do not use the PR template, skip sections, or modify the template structure will be asked to revise before review begins.

## Online Editor

An [Online Editor](https://domainconnect.paulonet.eu/dc/free/templateedit) is available, with features like syntax check, testing of variables replacement and groups.
To create the PR define the template and test with the editor and provide the link to test results in PR ("Copy Markdown" button).

<img width="112" height="45" alt="image" src="https://github.com/user-attachments/assets/53493ca4-8458-4209-a658-0e4948c4bbe1" />

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
