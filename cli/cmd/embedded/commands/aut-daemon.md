---
description: Start, stop, or restart the autonomy daemon
argument-hint: start | stop | restart | run-once | status
---

Control the autonomy daemon:

```bash
~/sixth/packages/aut/bin/aut daemon $ARGUMENTS
```

- `start` — launch daemon in background
- `stop` — graceful shutdown
- `restart` — stop + start
- `run-once` — single cycle in foreground
- `status` — daemon process info
