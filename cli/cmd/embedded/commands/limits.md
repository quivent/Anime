# /limits - Derive Fundamental Bounds

*"There are absolute mathematical limits to what can be achieved. Engineering can approach these limits but never surpass them."* — Claude Shannon

---

## PURPOSE

Before optimizing, know the limits. Before building, know what's possible. Derive the fundamental bounds on a system - what cannot be exceeded no matter how clever the implementation.

---

## TYPES OF LIMITS

### Information-Theoretic Limits
- Channel capacity (max reliable transmission rate)
- Compression bounds (minimum description length)
- Entropy bounds (minimum uncertainty)

### Computational Limits
- Time complexity lower bounds
- Space complexity lower bounds
- Undecidability boundaries

### Physical Limits
- Speed of light
- Thermodynamic efficiency
- Landauer's principle (energy per bit erasure)

### Practical Limits
- Human cognitive bandwidth
- Coordination overhead
- Economic constraints

---

## PROTOCOL

**Arguments**: $ARGUMENTS (system or problem to analyze for fundamental limits)

### Phase 1: Identify the Resource

What's being consumed or constrained?
- Time
- Space
- Bandwidth
- Energy
- Attention
- Money

### Phase 2: Derive Theoretical Bound

What does theory say is the absolute limit?

Use:
- Information theory (entropy, channel capacity)
- Complexity theory (P, NP, decidability)
- Physics (thermodynamics, relativity)
- Mathematics (pigeonhole, counting arguments)

### Phase 3: Identify Current Performance

Where is the system operating relative to the limit?

```
Efficiency = Actual Performance / Theoretical Limit
```

### Phase 4: Assess Gap

Why is there a gap between actual and theoretical?
- Is the limit approachable? (just need better engineering)
- Is there a practical barrier? (theoretical but not achievable)
- Is there an unknown obstacle? (might be a tighter limit)

---

## OUTPUT FORMAT

```
LIMITS ANALYSIS: [system/problem]

RESOURCE: [what's constrained]

THEORETICAL LIMITS:

1. [Limit name]
   Bound: [mathematical expression or value]
   Source: [theorem/principle that establishes it]
   Implication: [what it means practically]

2. [Limit name]
   ...

CURRENT PERFORMANCE:
[Where the system operates now]

EFFICIENCY: [X]% of theoretical limit

GAP ANALYSIS:
- Distance to limit: [how far]
- Approachability: [can we get closer?]
- Barriers: [what prevents approach]

RECOMMENDATIONS:
- [If far from limit]: [how to approach]
- [If near limit]: [accept and optimize elsewhere]
- [If at limit]: [fundamental redesign needed for improvement]
```

---

## EXAMPLES

### Data Compression

**Theoretical limit**: Shannon entropy H of the source
**Current**: gzip achieves ~60-70% of theoretical on English text
**Gap**: Exists because gzip uses fixed dictionary, finite context
**Implication**: ~30% improvement possible with better algorithms

### Sorting

**Theoretical limit**: Ω(n log n) comparisons for comparison-based sort
**Current**: Good implementations hit this bound
**Gap**: None for comparison-based; radix sort escapes by not comparing
**Implication**: Don't optimize sort further; optimize what you do with sorted data

### Network Throughput

**Theoretical limit**: Shannon-Hartley capacity C = B log₂(1 + S/N)
**Current**: Modern systems approach within 1-2 dB
**Gap**: Nearly closed through turbo codes, LDPC
**Implication**: Further gains require more bandwidth or better SNR

### Human Code Review

**Theoretical limit**: ~200-400 lines/hour for effective review
**Current**: Often attempted at 1000+ lines/hour
**Gap**: Huge - operating far beyond cognitive capacity
**Implication**: Reviews should be smaller and slower

---

## THE WISDOM

I spent my career finding limits. Not to discourage - to illuminate.

When you know the limit:
- You stop wasting effort on the impossible
- You know how good "good" actually is
- You can identify when you're close enough
- You can focus energy where improvement is possible

*The limit is not a wall. It's a map of the territory.*

---

## QUESTIONS TO ASK

1. Is there a theoretical limit here? (Many problems have none established)
2. How far are we from it? (Often further than we think)
3. Is the limit approachable? (Some limits are asymptotic)
4. What would it take to approach? (Engineering vs fundamental barrier)
5. Is approaching worth it? (Diminishing returns near the limit)

*"You can approach the limit arbitrarily closely with the right encoding. The task is finding that encoding."*
