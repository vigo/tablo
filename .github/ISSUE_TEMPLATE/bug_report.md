---
name: "🐛 Bug Report"
about: "Report a bug to help me improve"
title: "[BUG]: <title of the bug>"
labels: bug
assignees: vigo
---

## Bug Summary

Briefly describe the issue.

---

## Steps to Reproduce

Provide an example usage to reproduce the issue:

```bash
docker images | tablo -fdc " "
```


What was the output:

```bash
┌──────────────┬ (unexpected behavior) ────────┐  
│ REPOSITORY   │ ???  │ ???  │ ??? │ ??? │ ??? │  
└──────────────┴───────────────────────────────┘  
```

What was the expected output? or missing items?

```bash
┌──────────────┬──────────┐  
│ REPOSITORY   │ IMAGE ID │  
├──────────────┼──────────┤  
│ ubuntu       │ abc123   │  
│ alpine       │ def456   │  
└──────────────┴──────────┘  
```

---

## Expected Behavior

Clearly describe what **should** have happened.

---

## Actual Behavior

Clearly describe what **actually** happened. Include error messages if any.

---

## Environment

- `tablo` version: `tablo -version` value.
- Go Version: \[e.g., go version go1.23.4 darwin/arm64\] (result of `go version`)
- OS: \[e.g., Ubuntu, macOS\]

---

## Additional Context

Any additional information that could help diagnose the issue.
