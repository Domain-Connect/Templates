# AI Agent Guidelines for Domain Connect Templates

This document provides guidance for AI tools working with Domain Connect templates in this repository.

## Scope

These rules apply when **adding or modifying templates** in this repository.

## Template Schema

The template file schema is defined in [`template.schema`](template.schema). This schema **MUST be followed** for all template files.

Key schema requirements:
- Templates must include required fields: `providerId`, `providerName`, `serviceId`, `serviceName`, `records`
- Record types supported: A, AAAA, CNAME, NS, TXT, MX, SPFM, SRV, REDIR301, REDIR302, APEXCNAME, and ANY
- Templates must follow the JSON Schema defined in `template.schema`

## Template Validation

Before submitting changes, you **MUST** validate templates using the official linter:

### Installation

```bash
go install github.com/Domain-Connect/dc-template-linter@latest
```

### Usage

Validate a template file with:

```bash
dc-template-linter -loglevel error -tolerate info <template-file.json>
```

**All templates must pass linting without errors before being considered valid.**

> **Note:** If `dc-template-linter` returns a non-zero exit value, the template is invalid and must be corrected before submission.

## Template Naming Convention

Template files must be named following this pattern:

```
providerId.serviceId.json
```

Example: `myprovider.com.website.json`

Place template files in the root directory of the repository.

## Example Templates

For reference, example templates using the provider `exampleservice.domainconnect.org` are available in this repository. These templates demonstrate proper structure and can be used as a starting point when creating new templates.

## Key Quality Rules

When modifying or creating templates, ensure compliance with these critical rules:

1. **Always set `syncPubKeyDomain`** - Enables cryptographic verification (preferred over `warnPhishing`)
2. **Never set `syncPubKeyDomain` and `warnPhishing` together** - They are mutually exclusive
3. **Set `syncRedirectDomain`** when using `redirect_uri` in synchronous flow
4. **Never use TXT records for SPF** - Use the `SPFM` record type instead
5. **Set `txtConflictMatchingMode`** on TXT records that must be unique (e.g., DMARC, verification tokens)
6. **Scope variables narrowly** - Avoid bare variables as full record values; use fixed prefixes
7. **Scope host variables narrowly** - Fix non-variable parts in host fields
8. **Never use variables in `host` to target subdomains** - Use the standard `host` parameter instead
9. **Never write `%host%` explicitly** in the `host` attribute - The protocol handles this automatically
10. **Set `essential`** on records users may need to change independently

For complete details, see the [README.md](README.md).

## References

- [Domain Connect Specification](https://www.ietf.org/archive/id/draft-ietf-dconn-domainconnect-01.html)
- [Template Linter Repository](https://github.com/Domain-Connect/dc-template-linter)
- [PubKey TXT record tester](https://github.com/kerolasa/dc-debug-pubkey)
- [Online Template Editor](https://domainconnect.paulonet.eu/dc/free/templateedit)
