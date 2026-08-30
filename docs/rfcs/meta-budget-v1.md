# Gooo meta budget v1

## Purpose

This package makes execution cost and verified evidence inputs to the next
Gooo meta operation. It does not turn a good number into a claim of utility.
Only a same-workload baseline/candidate comparison with an exact before/after
identity pair can prove a cost improvement.

## Identity and cache

An exact identity digest is the SHA-256 digest of the sorted JSON object
containing `workload_id`, `input_digest`, `tool_digest`, and `contract_digest`.
Before and after observations must have the same digest and variant. A cache
hit is useful evidence only after this digest is recomputed and matches the
cached digest; a cache hit alone never closes a decision.

## Budget proposals

The requested semantic resolution and proof denominator are immutable during
budget handling. If build, test, conformance, peak RSS, or artifact bytes
exceed the budget, the evaluator returns `UNKNOWN` with an allowed execution
plan proposal. Every proposed plan is read-only, is not `FIXED_POINT`, and
preserves both requested values.

## Evidence states

`REFUTED` has precedence over `UNKNOWN`, which has precedence over `CLOSED`.
Malformed requests, identity drift, fixed-point plans, privilege escalation,
and cost regressions are fail-closed. Unknown evidence always carries exactly
these six fields: `stage`, `step`, `reason`, `unknown_class`, `next_operation`,
and `blocked_by`.

## Metric contract

The six proof classes are fixed at denominator 6 and each metric is bound
one-to-one to a Gooo meta activity, source, IR, generated artifact, and
evaluator. Runtime and repository inventory metrics use denominator 1. The CI
artifact reports descendant directories, regular files, physical Go/Gooo
files and lines (including blank and comment lines), build/test/conformance
wall time, peak RSS, executed/reused/skipped tests, artifact files/bytes, and
repository writes. The root README is excluded from inventory counts.

## Runtime authority

The evaluator accepts caller-owned output paths only. It does not mutate the
input repository. CI checks the checkout before and after the run and exposes
`repository_writes` as an integer; any non-zero value is refuted.
