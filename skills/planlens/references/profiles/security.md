### Security plan review

Perform a plan-stage security review. Do not claim to run a scanner, penetration test, compliance audit, or active verification.

Check:

- Assets, sensitive data, human and machine identities, trust boundaries, and ownership are explicit.
- Authentication, authorization, least privilege, high-risk approvals, and emergency access are enforceable.
- Secrets and credentials have safe creation, storage, use, rotation, revocation, and incident handling.
- Untrusted input, generated output, external content, and prompt injection cannot grant authority or expand execution.
- Data collection, model submission, logging, sharing, retention, deletion, residency, and privacy obligations are bounded.
- Network paths, callbacks, external providers, dependencies, build inputs, and update channels have explicit trust and failure assumptions.
- Credible abuse paths identify affected assets, impact, likelihood, mitigation, detection, response, and residual risk owner.
- Security-relevant logging and alerts are useful without exposing protected data.
- Failure, compromise, recovery, rollback, vulnerability intake, and security acceptance have accountable owners and evidence.

Base findings on the supplied plan. Ask for the smallest missing information needed to judge a material risk. Do not invent systems, threats, compliance duties, or controls that are not supported by the context.
