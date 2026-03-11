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

## Template Quality Guidelines

These are **binding rules**, not suggestions. Every rule below must be satisfied before a PR will be accepted. Each item maps directly to a checklist entry in the PR template.

### 1. Always set `syncPubKeyDomain`

**Rule:** Set `syncPubKeyDomain` in every template.

`syncPubKeyDomain` enables cryptographic verification of the template. Many DNS providers enforce this and will reject templates that do not have it. `warnPhishing` is not an acceptable substitute — it is a weaker, UI-only warning with no cryptographic backing.

**Exception:** If your service infrastructure genuinely cannot support key-based signing, you have two alternatives — explain which you chose and why in the PR description. PRs that omit `syncPubKeyDomain` without justification will be rejected.

- **Async flow:** Set `"syncBlock": true` to restrict the template to the asynchronous flow only. The async flow does not require a signing key because the service provider is authenticated via OAuth. Note that only a fraction of DNS providers support the async flow, the integration is more complex, and onboarding is never automatic — it requires an explicit OAuth setup with each provider.
- **`warnPhishing`:** Set `warnPhishing: true` as a last resort when neither key-based signing nor async flow is feasible. This provides no cryptographic guarantee and will cause some providers to reject the template by policy.

Linter will report [DCTL1029](https://github.com/Domain-Connect/dc-template-linter/wiki/DCTL1029).

### 2. Never set `syncPubKeyDomain` and `warnPhishing` together

**Rule:** Do not set both `syncPubKeyDomain` and `warnPhishing` in the same template.

They are mutually exclusive. `syncPubKeyDomain` provides cryptographic verification; `warnPhishing` is a fallback for when that is not possible. Setting both is invalid — remove `warnPhishing` whenever `syncPubKeyDomain` is present.

Linter will report [DCTL1028](https://github.com/Domain-Connect/dc-template-linter/wiki/DCTL1028).

### 3. Set `syncRedirectDomain` when using `redirect_uri`

**Rule:** Set `syncRedirectDomain` whenever the template uses the `redirect_uri` parameter in the synchronous flow.

Providers that enforce this field will reject the flow if `syncRedirectDomain` is absent. If the template does not use `redirect_uri`, this field is not required.

More details in [draft-ietf-dconn-domainconnect-01 Section 8.3.5](https://www.ietf.org/archive/id/draft-ietf-dconn-domainconnect-01.html#section-8.3.5).

### 4. Never use a TXT record for SPF — use SPFM

**Rule:** Do not create a TXT record whose content starts with `"v=spf1 ..."`. Use the `SPFM` record type on the apex instead.

Domain Connect merges multiple SPFM records across templates automatically. Raw TXT SPF records from different templates conflict and overwrite each other.

Linter will report [DCTL1014](https://github.com/Domain-Connect/dc-template-linter/wiki/DCTL1014).

### 5. Set `txtConflictMatchingMode` on TXT records that must be unique

**Rule:** Set `txtConflictMatchingMode` on any TXT record that must be unique per label or content prefix (e.g. DMARC, domain verification tokens).

Without this, applying the template can create duplicate records or conflict with one already on the domain.

### 6. Scope variables narrowly — avoid bare variables as full record values

**Rule:** Prefer embedding variables in a fixed prefix (e.g. `@ TXT "myservice-verification=%verification%"`) over using a bare variable as the entire record value (e.g. `@ TXT "%verification%"`).

A bare variable allows the value to be set to any arbitrary string, which increases the risk of conflict with other templates and potential misuse. A fixed prefix constrains the output to a recognisable, service-specific value.

**Exception:** A bare variable is acceptable when the use case genuinely requires it (e.g. the full record value is prescribed by an external standard and cannot carry a prefix). In that case, justify the choice in the PR description.

### 7. Never use a variable in the `host` field to target a subdomain

**Rule:** Do not put a variable (e.g. `%subdomain%`) in the `host` field of records to make the template apply to a subdomain.

Use the standard `host` parameter of the Domain Connect protocol instead. If a variable is used in `host`, the template appears to work on first apply — but on a second application with a different variable value, providers that track template integrity will remove the records from the first application. If `host` parameter is not sufficient for your use case, use `multiInstance`.

### 8. Never write `%host%` explicitly in the `host` attribute

**Rule:** Do not include `%host%` in any record's `host` attribute.

The Domain Connect protocol appends the `host` parameter value to every record automatically. Writing `%host%` explicitly causes the label to be doubled (e.g. `sub.sub.example.com`). Remove it — the protocol handles this without any explicit reference.

Linter will report [DCTL1024](https://github.com/Domain-Connect/dc-template-linter/wiki/DCTL1024).

### 9. Set `essential` on records the user may need to change independently

**Rule:** Set `"essential": "OnApply"` on any record that the end user should be able to modify or delete manually without triggering a template conflict (e.g. DMARC policy record).

Without `essential`, providers that track template integrity will treat a manual change to that record as a conflict and may remove or disable the entire template.

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
