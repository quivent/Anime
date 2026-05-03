---
description: Deep health check of the autonomy system
---

Run comprehensive health checks:

```bash
~/sixth/packages/aut/bin/aut health
```

Checks: daemon alive, DB readable, logs writable, Claude CLI present, disk space, config valid, workers not stuck.
