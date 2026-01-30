# /noise - Identify Signal Corruption

*"Noise is anything that corrupts the intended signal."*

---

## PURPOSE

Identify what's corrupting the signal in a system. Noise is not the opposite of information - noise is what prevents information from being received correctly.

---

## TYPES OF NOISE

### 1. Syntactic Noise
Corruption of the symbols themselves:
- Typos in code
- Malformed data
- Encoding errors
- Transmission corruption

### 2. Semantic Noise
Corruption of meaning:
- Misleading names
- Wrong abstractions
- Outdated documentation
- Implicit assumptions

### 3. Pragmatic Noise
Corruption of intent:
- Solving the wrong problem
- Optimizing the wrong metric
- Missing the actual use case
- Cargo culting patterns

### 4. Temporal Noise
Corruption over time:
- Bit rot
- Dependency drift
- Context loss
- Knowledge decay

### 5. Cognitive Noise
Corruption in human processing:
- Complexity overload
- Attention fragmentation
- Inconsistent conventions
- Surprising behavior

---

## PROTOCOL

**Arguments**: $ARGUMENTS (system, code, or communication to analyze for noise)

### Phase 1: Map the Signal Path

What is the intended signal? Where does it originate? Where must it arrive?

### Phase 2: Identify Noise Sources

At each stage of transmission:
- What could corrupt the signal here?
- What IS corrupting the signal here?
- How severe is the corruption?

### Phase 3: Measure Signal-to-Noise Ratio

```
SNR = Signal Power / Noise Power
SNR (dB) = 10 × log₁₀(Signal/Noise)
```

In code terms:
- Signal = essential logic, clear intent, correct behavior
- Noise = bugs, confusion, misdirection, waste

### Phase 4: Noise Reduction Strategies

For each noise source:
- Can it be eliminated? (remove the source)
- Can it be filtered? (add validation/checks)
- Can it be corrected? (add redundancy for error recovery)
- Must it be tolerated? (accept and document)

---

## OUTPUT FORMAT

```
NOISE ANALYSIS: [target]

INTENDED SIGNAL:
[What should be transmitted/understood]

NOISE SOURCES:

Syntactic:
- [noise]: [severity] - [impact]

Semantic:
- [noise]: [severity] - [impact]

Pragmatic:
- [noise]: [severity] - [impact]

Temporal:
- [noise]: [severity] - [impact]

Cognitive:
- [noise]: [severity] - [impact]

SIGNAL-TO-NOISE ESTIMATE: [qualitative or quantitative]

NOISE REDUCTION PLAN:
1. [highest impact noise]: [mitigation strategy]
2. [next highest]: [mitigation strategy]
...

IRREDUCIBLE NOISE:
[What noise must be accepted and why]
```

---

## EXAMPLES

### Noisy Codebase
- **Syntactic**: Inconsistent formatting, deprecated syntax
- **Semantic**: Functions that don't do what their names suggest
- **Pragmatic**: Solving problems users don't have
- **Temporal**: Comments describing code that's changed
- **Cognitive**: 1000-line files, deep nesting, global state

### Noisy API
- **Syntactic**: Malformed responses, encoding issues
- **Semantic**: Field names that mislead
- **Pragmatic**: Endpoints that don't match workflows
- **Temporal**: v1 behavior leaking into v2
- **Cognitive**: Inconsistent conventions across endpoints

---

## THE DEEPER POINT

You cannot eliminate noise entirely. The universe is noisy. Channels are imperfect.

But you can:
1. **Reduce noise at the source** - write clearer code
2. **Add redundancy for error correction** - tests, types, validation
3. **Increase signal power** - make the essential logic prominent
4. **Match encoding to channel** - use the right abstraction level

*The goal is not zero noise. The goal is reliable transmission despite noise.*
