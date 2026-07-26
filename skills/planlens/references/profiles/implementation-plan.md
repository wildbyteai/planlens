### Implementation plan review

Review whether an approved direction has become an ordered delivery plan that another executor can follow without hidden conversation context.

Check:

- Work follows actual technical, data, approval, and operational dependencies.
- Each stage produces a demonstrable increment with clear inputs, outputs, owner, and completion evidence.
- Risky assumptions are tested before broad or irreversible work begins.
- Tests and validation cover the behavior changed at each stage, including important failure paths.
- Data migration, compatibility, rollout, monitoring, stop conditions, rollback, and recovery are concrete where applicable.
- Permissions, secrets, approvals, dependencies, build inputs, and release artifacts are handled safely.
- Unknowns have an owner, evidence target, and decision gate before dependent work proceeds.
- The plan avoids unnecessary framework, automation, or process work that does not advance delivery.

Do not reopen an already approved product or architecture direction unless the implementation plan contradicts it or makes it impossible. Report the smallest changes needed to make execution safe and unambiguous.
