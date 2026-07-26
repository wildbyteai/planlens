### AI and agent plan review

Review a plan whose outcome materially depends on model behavior, retrieval, generated output, or tool-using agents. Do not treat model behavior as deterministic.

Check:

- The target task, users, current baseline, success criteria, unacceptable failures, and non-AI fallback are explicit.
- Model, retrieval, tool, orchestration, and autonomy choices are justified against a simpler workflow.
- Context sources have clear provenance, freshness, access rules, size limits, and handling for missing or conflicting evidence.
- Evaluations use representative tasks, observable graders, thresholds, regression checks, and human review where judgment is required.
- Tool permissions, side effects, approvals, idempotency, stop conditions, and escalation paths match the risk of each action.
- Untrusted input, retrieved content, tool output, and prompt injection cannot expand authority or bypass policy.
- The plan defines confidence handling, abstention, fallback, degradation, recovery, and safe behavior when models or providers fail.
- Traces and feedback can reveal model, retrieval, routing, tool, and handoff failures without exposing protected data.
- Quality, latency, usage, and operational cost assumptions are measurable and have an owner.

Report only issues that materially affect usefulness, safety, controllability, or the ability to evaluate the workflow. Use the security profile instead when the primary decision is a security risk assessment.
