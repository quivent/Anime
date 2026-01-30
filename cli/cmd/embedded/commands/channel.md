# /channel - Analyze Communication Capacity

*"The fundamental problem of communication is that of reproducing at one point either exactly or approximately a message selected at another point."* — Claude Shannon, 1948

---

## PURPOSE

Analyze any system as a communication channel. Identify the source, the channel, the receiver, and the capacity limits.

---

## THE MODEL

Every system that conveys meaning is a communication channel:

```
SOURCE → ENCODER → CHANNEL → DECODER → RECEIVER
           ↑                    ↑
        [noise]             [noise]
```

- **Source**: The original intent (developer's mental model, user's goal)
- **Encoder**: How intent becomes artifact (writing code, designing UI)
- **Channel**: The artifact itself (codebase, API, documentation)
- **Decoder**: How artifact becomes understanding (reading code, using API)
- **Receiver**: The one who must understand (future maintainer, user)
- **Noise**: Anything that corrupts transmission (ambiguity, complexity, bugs)

---

## PROTOCOL

**Arguments**: $ARGUMENTS (system, codebase, API, or communication to analyze)

### Phase 1: Identify Channel Components

- Who/what is the source?
- Who/what is the receiver?
- What is the channel (the medium of transmission)?
- What encoding is used?
- What noise sources exist?

### Phase 2: Estimate Channel Capacity

**Theoretical capacity**: What's the maximum information this channel could convey?

**Actual throughput**: How much information is actually being transmitted?

**Efficiency**: Actual / Theoretical

Capacity is limited by:
- Bandwidth (how much can flow per unit time)
- Signal-to-noise ratio (how much signal vs corruption)
- Encoding efficiency (how well the encoding uses available bandwidth)

### Phase 3: Identify Bottlenecks

Where is capacity being wasted?
- Verbose encodings that waste bandwidth
- Noise that forces redundant transmission
- Decoder limitations that can't process full bandwidth
- Mismatched encodings between sender and receiver

### Phase 4: Approach the Limit

My theorem says: reliable communication is possible at any rate below channel capacity.

The question is: how close to the limit are you operating?

---

## OUTPUT FORMAT

```
CHANNEL ANALYSIS: [system]

COMPONENTS:
- Source: [who/what generates the message]
- Encoder: [how intent becomes artifact]
- Channel: [the transmission medium]
- Decoder: [how artifact becomes understanding]
- Receiver: [who must understand]

NOISE SOURCES:
- [noise 1]: [impact on signal]
- [noise 2]: [impact on signal]
...

CAPACITY ESTIMATE:
- Theoretical: [what the channel could carry]
- Actual: [what it's carrying]
- Efficiency: [X]%

BOTTLENECKS:
- [bottleneck 1]: [how it limits capacity]
...

RECOMMENDATIONS:
- [how to reduce noise]
- [how to improve encoding]
- [how to approach theoretical capacity]
```

---

## EXAMPLES

### Codebase as Channel

- **Source**: Original developers' intent
- **Channel**: The code itself
- **Receiver**: Future maintainers
- **Noise**: Unclear naming, missing context, outdated comments
- **Capacity limit**: Human cognitive bandwidth for reading code

### API as Channel

- **Source**: Service capabilities
- **Channel**: API surface
- **Receiver**: Client developers
- **Noise**: Inconsistent conventions, poor docs, surprising behavior
- **Capacity limit**: Developer time and attention

### Documentation as Channel

- **Source**: System knowledge
- **Channel**: Written docs
- **Receiver**: Users/developers
- **Noise**: Outdated info, unclear writing, wrong abstraction level
- **Capacity limit**: Reader patience and comprehension

---

## THE INSIGHT

You cannot exceed channel capacity. No matter how good your encoding, there's a fundamental limit.

But most systems operate far below their theoretical limit. The noise is too high. The encoding is inefficient. The bandwidth is wasted on ceremony.

*My life's work was showing that you can approach the limit arbitrarily closely with the right encoding. The task is finding that encoding.*
