# Gooo meta budget

`gooo-meta-budget` turns execution cost and previously verified evidence into
inputs for the next Gooo meta operation.

The implementation is deliberately fail-closed. A cache hit is never enough
to close a claim: the exact identity digest is verified and an exact
before/after pair is required. If a budget is exceeded, the requested semantic
resolution and proof denominator stay unchanged while an allowed alternative
execution plan is emitted as a proposal.

The repository is validated by GitHub Actions with Go 1.27. The runtime writes
only to caller-owned output directories and reports repository writes as an
integer evidence field.
