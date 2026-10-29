# Public OSS release gate

This checklist must be completed **before** changing repository visibility to Public. Checking a box requires evidence; its presence does not imply a completed audit.

## License and ownership
- [ ] Verify the root MIT LICENSE notice, copyright holder, and all third-party licenses, fonts, data, assets, and vendored source.
- [ ] Confirm every contributor has permission to release their contributions; identify any files requiring a separate license or attribution.
- [ ] Review README, CONTRIBUTING.md and SECURITY.md for accurate install, support and vulnerability-reporting instructions.

## Confidentiality and supply-chain review
- [ ] Scan **all reachable Git history**, branches, tags and Git LFS, not just the working tree, for credentials, tokens, private keys, connection strings, private URLs, customer data and internal documents.
- [ ] Check workflows, fixtures, tests, issues/PR attachments and build artifacts for confidential material and excessive permissions.
- [ ] Examine dependencies and license compatibility, known advisories, SBOM and lockfile consistency. Confirm CI does not disclose secrets on forked PRs.
- [ ] If a credential was ever committed, revoke/rotate it and purge sensitive history where appropriate **before** publication.

## Quality and public usability
- [ ] Build and test from a fresh checkout with no private dependencies or credentials.
- [ ] Verify examples with synthetic data; check documented prerequisites and commands.
- [ ] Verify links, project name, prior-repository references, support scope and release notes.
- [ ] Approve the public repository contents, default branch, tags and history after a second-person or independent review.

## Publish
- [ ] Make repository Public only after all above checks pass.
- [ ] Verify anonymous clone and documentation links.
- [ ] Enable private vulnerability reporting, branch protections and appropriate CI permissions.
- [ ] Create the first public release and confirm license display.
- [ ] Only afterward assess whether the prior repository can be archived or deleted.

Do not publish credentials, private customer data, production extracts or private commercial workflows. This file is a gate, not proof of completed verification.
