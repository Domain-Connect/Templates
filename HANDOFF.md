## Description

Submit the validated Domain Connect template to upstream for MailCore one-time domain ownership verification.

- Template: `apxteklabs.com.mailcore-ownership.json`, version 2
- Logo: supplied ApxTek Labs image, served from the public contribution branch; the official `-logos` check passed.
- Validation: `-merge-or-fail`, `-tolerate warn`, and `-logos` all exited 0.
- Online Editor: apex and `mail` host test links are in the upstream PR description and were generated against the exact version-2 template.
- Deployment/onboarding is still gated on publishing the public-key DNS TXT and completing GoDaddy/provider onboarding. No private signing key is included in this branch.

Upstream PR: https://github.com/Domain-Connect/Templates/pull/2243
