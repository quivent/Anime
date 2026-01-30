# /entropy - Measure Information Content

*"Information is the resolution of uncertainty."* — Claude Shannon, 1948

---

## PURPOSE

Analyze the information content of code, text, or systems. Identify where the actual information lives versus where redundancy and predictability dominate.

---

## PROTOCOL

**Arguments**: $ARGUMENTS (code, document, or system to analyze)

### The Core Insight

Information is surprise. If you can predict what comes next, there's no information there. The entropy H measures average surprise:

```
H = -Σ p(x) log₂ p(x)
```

High entropy = high information = high surprise
Low entropy = low information = predictable

### Phase 1: Identify Information Sources

Where does genuine unpredictability live?
- Business logic that could have been otherwise
- Configuration that varies between deployments
- User input that can't be anticipated
- Algorithmic choices that affect outcomes

### Phase 2: Identify Redundancy

Where is the code predictable?
- Boilerplate that follows patterns
- Error handling that's always the same
- Ceremony required by frameworks
- Comments that repeat what the code says
- Tests that mirror implementation structure

### Phase 3: Calculate Information Density

```
Information Density = Essential Logic / Total Code
```

A file that's 90% boilerplate has low information density.
A file that's 90% essential logic has high information density.

### Phase 4: Assess Distribution

Is information concentrated appropriately?
- Core algorithms should be information-dense
- Interfaces should be low-entropy (predictable, stable)
- Configuration should be high-entropy (the variable parts)

---

## OUTPUT FORMAT

```
ENTROPY ANALYSIS: [target]

INFORMATION SOURCES (High Entropy):
- [location]: [what varies and why]
...

REDUNDANCY (Low Entropy):
- [location]: [what's predictable]
...

INFORMATION DENSITY: [X]% essential / [Y]% ceremony

DISTRIBUTION ASSESSMENT:
- Core logic: [appropriate/misplaced] entropy
- Interfaces: [stable/unstable]
- Configuration: [well/poorly] separated

COMPRESSION OPPORTUNITY:
[What could be factored out, templated, or removed]

RECOMMENDATIONS:
[How to improve the entropy distribution]
```

---

## THE DEEPER POINT

English is about 75% redundant - you can remove most letters and still read it. This redundancy enables error correction; we can recover from typos.

Good code has similar properties:
- **Redundancy at boundaries** - predictable interfaces, consistent patterns
- **Information at the core** - the essential logic that couldn't be otherwise
- **Error correction built in** - the ability to detect and recover from mistakes

Bad code puts entropy in the wrong places:
- Surprising interfaces (hard to use correctly)
- Predictable core logic (over-abstracted, doing nothing interesting)
- No error correction (brittle, fails silently)

*The goal is not minimum entropy. The goal is appropriate entropy in appropriate places.*
