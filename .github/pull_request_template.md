# Description

<!-- short description of the template(s) and/or reason for update -->

## Type of change

Please mark options that are relevant.

- [ ] New template
- [ ] Bug fix (non-breaking change which fixes an issue in the template)
- [ ] New feature (non-breaking change which adds functionality to the template)
- [ ] Breaking change (fix or feature that would cause existing template behavior to be not backward compatible)

# How Has This Been Tested?

Please mark the following checks done
- [ ] Template functionality checked using [Online Editor](https://domainconnect.paulonet.eu/dc/free/templateedit)
- [ ] Template file name follows the pattern `<providerId>.<serviceId>.json`
- [ ] resource URL provided with `logoUrl` is actually served by a webserver

# Checklist of common problems

Mark all the checkboxes after conducting the check. Comment on any point which is not fulfilled.
See [Template Quality Guidelines](../README.md#template-quality-guidelines) for details and rationale on each rule.

- [ ] `syncPubKeyDomain` is set — **this is mandatory**; omitting it requires explicit justification in the PR description or the PR will be rejected
- [ ] `warnPhishing` is **not** set alongside `syncPubKeyDomain` — the two must not appear together
- [ ] `syncRedirectDomain` is set whenever the template uses `redirect_uri` in the synchronous flow
- [ ] no TXT record contains SPF content (`"v=spf1 ..."`) — use the `SPFM` record type instead
- [ ] `txtConflictMatchingMode` is set on every TXT record that must be unique per label or content prefix (e.g. DMARC)
- [ ] no variable is used as a bare full record value (e.g. `@ TXT "%foo%"`) unless necessary — prefer `@ TXT "service-foo=%foo%"`; if bare, justify in the PR description
- [ ] no variable is used in the `host` field to create a subdomain — use the `host` parameter or `multiInstance` instead
- [ ] `%host%` does not appear explicitly in any `host` attribute
- [ ] `essential` is set to `OnApply` on records the end user may need to modify or remove without breaking the template (e.g. DMARC)

## Online Editor test results

<!-- 
  Required. Follow these steps in the Online Editor (https://domainconnect.paulonet.eu/dc/free/templateedit):
    1. Load your template and use "Check template" to perform extended schema and consistency check
    2. Fill in domain/host as well as variable values
    3. Click "Test apply template"
    4. Click "Copy Markdown" and paste the link below
    5. If necessary repeat steps 2-4 with different group setups and/or domain/host configuration
-->

**Editor test link(s):** 
<!-- paste the links from "Copy Markdown" here -->
<!-- paste multiple links if more test conducted. At least 1 per template file included in the PR -->
