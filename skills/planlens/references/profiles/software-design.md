### Software design review

Review an architecture or technical design against its stated outcome, constraints, scale, and operating context. Do not perform line-level code review.

Check:

- System boundaries, component responsibilities, authoritative state, and ownership are unambiguous.
- Interfaces define relevant inputs, outputs, invariants, errors, timeouts, and compatibility expectations.
- Important data and control flows can be traced across trust and failure boundaries.
- Dependency direction, shared state, coupling, and coordination are justified and maintainable.
- Reliability, latency, capacity, consistency, security, privacy, and cost assumptions are explicit where material.
- Failure, retry, idempotency, degradation, recovery, observability, and test seams cover the main risks.
- Migration, coexistence, rollback, and schema or contract evolution protect existing users and data.
- Credible alternatives and expensive-to-reverse choices are acknowledged.

Tie every finding to a concrete correctness, compatibility, operational, delivery, or evolution consequence. Do not prescribe a fashionable architecture without evidence that it fits this plan.
