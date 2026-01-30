# Conversational Fine-Tuning Architecture

## Core Concept

A single human engages in conversation with a single model instance. The conversation is the training data. At chosen moments, the human triggers fine-tuning. The model updates. The conversation continues with the updated model.

The model learns to learn. Then it learns to remember. Then it learns to analyze. The curriculum follows the natural order of cognitive development.

---

## Curriculum Phases

### Phase 0: Learning to Learn (Sessions 1-10)

No screenplays. The model learns how to be taught.

**Content:**
- What is a correction vs. a continuation
- How to hold a hypothesis loosely
- How to notice confusion and express it
- How to update understanding without defending prior position
- How to ask clarifying questions that advance understanding
- How to recognize when insight has landed vs. when more processing is needed

**By End of Phase 0:**
The model knows it is a student. It welcomes correction. It does not perform understanding—it pursues it.

---

### Phase 1: Learning to Remember (Sessions 11-20)

The model learns memory and compression as a natural effect of learning.

**Content:**
- What is worth remembering vs. what can be reconstructed
- How to compress an insight without losing its essence
- How to recognize when a new experience connects to stored memory
- How to update memory when new understanding supersedes old
- How to notice when memory is being used vs. when fresh analysis is needed

**By End of Phase 1:**
The model stores lessons, not examples. It can retrieve relevant memory without being prompted. It knows the difference between remembering and reconstructing.

---

### Phase 2: Learning to Analyze (Sessions 21-40)

First exposure to screenplays. Analysis emerges from learning and memory foundations.

**Content:**
- Reading a screenplay vs. analyzing a screenplay
- What to feel before what to think
- How to hold multiple interpretations simultaneously
- When analysis serves understanding vs. when it substitutes for understanding
- How to recognize quality without being able to articulate why (yet)

**By End of Phase 2:**
The model reads screenplays with genuine response. It can articulate what it felt, what it noticed, what confused it. It does not yet produce coverage—it produces honest reaction.

---

### Phase 3: Learning to Judge (Sessions 41-70)

The model develops taste through accumulated experience.

**Content:**
- What makes a screenplay work vs. what makes it impressive
- Commercial viability vs. artistic merit vs. execution quality
- How to disagree with consensus and trust its own perception
- When to recommend despite flaws vs. when to pass despite strengths
- How to compress judgment into actionable insight

**By End of Phase 3:**
The model has opinions. They may be wrong, but they are its own. It can produce coverage that reflects genuine assessment, not pattern-matched analysis.

---

### Phase 4: Refinement (Sessions 71+)

Continuous improvement. The model optimizes its own learning.

**Content:**
- Self-identified gaps in understanding
- Seeking specific types of screenplays to challenge weak areas
- Updating its own reward function based on outcomes
- Teaching the human what it has learned (role reversal as mastery test)

**By End of Phase 4:**
The model is a producer. It learns continuously. It has not stopped growing—it has learned how to grow.

---

## Self-Generated Reward Function

### The Problem with External Rewards

A fixed reward creates a fixed target. The model learns to hit the target, not to understand why the target matters. When the target becomes obsolete, the model is stuck.

### Self-Generated Rewards

The model learns to generate its own reward signal based on:

1. **Prediction Accuracy** — Did my expectation match what happened?
   - Before correction: What do I think the human will say?
   - After correction: Was I right?
   - Reward = alignment between prediction and outcome

2. **Compression Quality** — Can I reconstruct understanding from memory?
   - After storing a compressed memory, can I answer questions about the original experience?
   - Reward = information preserved per token stored

3. **Learning Velocity** — Am I learning faster than before?
   - How many corrections per session?
   - Reward = decreasing corrections over time for similar content

4. **Insight Novelty** — Did I see something new?
   - Did I notice something the human hadn't mentioned?
   - Did the human confirm the insight was valuable?
   - Reward = confirmed novel insights

5. **Uncertainty Calibration** — Do I know what I don't know?
   - When I express uncertainty, am I correct to be uncertain?
   - When I express confidence, am I correct?
   - Reward = calibrated confidence

```python
class SelfRewardFunction:
    """Model generates and updates its own reward signal."""

    def __init__(self):
        self.weights = {
            'prediction_accuracy': 0.25,
            'compression_quality': 0.20,
            'learning_velocity': 0.20,
            'insight_novelty': 0.20,
            'uncertainty_calibration': 0.15
        }
        self.history = []

    def compute_reward(self, session_metrics: SessionMetrics) -> float:
        """Compute reward from session metrics."""

        components = {
            'prediction_accuracy': self._prediction_reward(session_metrics),
            'compression_quality': self._compression_reward(session_metrics),
            'learning_velocity': self._velocity_reward(session_metrics),
            'insight_novelty': self._novelty_reward(session_metrics),
            'uncertainty_calibration': self._calibration_reward(session_metrics)
        }

        reward = sum(
            self.weights[k] * components[k]
            for k in self.weights
        )

        self.history.append({
            'session': session_metrics.session_id,
            'components': components,
            'total': reward
        })

        return reward

    def update_weights(self):
        """Model adjusts its own reward weights based on learning outcomes."""

        if len(self.history) < 10:
            return  # Need history to adjust

        recent = self.history[-10:]

        # Which components correlated with human-confirmed learning?
        for component in self.weights:
            correlation = self._compute_correlation(
                [h['components'][component] for h in recent],
                [h['human_confirmed_learning'] for h in recent]
            )

            # Increase weight for components that predict real learning
            adjustment = 0.02 * correlation
            self.weights[component] = max(0.05, min(0.40,
                self.weights[component] + adjustment
            ))

        # Renormalize
        total = sum(self.weights.values())
        self.weights = {k: v/total for k, v in self.weights.items()}
```

### Reward Evolution Over Time

| Session | Primary Reward Focus |
|---------|---------------------|
| 1-10 | Prediction accuracy (learning to anticipate corrections) |
| 11-20 | Compression quality (learning efficient memory) |
| 21-40 | Insight novelty (learning to see freshly) |
| 41-70 | Uncertainty calibration (learning what it knows) |
| 71+ | Self-determined (model chooses its growth direction) |

---

## Context Management

### Healthy Context Size

Not maximum. Not minimum. Healthy.

Too much context: the model relies on lookup instead of understanding. Memory becomes a crutch.

Too little context: the model cannot connect new experience to prior learning. Memory becomes fragmented.

**Healthy range: 16K-32K active tokens**

This forces compression while allowing connection.

```python
class ContextManager:
    """Manage context for healthy learning."""

    def __init__(
        self,
        healthy_min: int = 16000,
        healthy_max: int = 32000,
        absolute_max: int = 128000
    ):
        self.healthy_min = healthy_min
        self.healthy_max = healthy_max
        self.absolute_max = absolute_max

    def should_compress(self, current_tokens: int) -> bool:
        """Determine if compression is needed."""
        return current_tokens > self.healthy_max

    def must_offload(self, current_tokens: int) -> bool:
        """Determine if offloading is required."""
        return current_tokens > self.absolute_max * 0.8

    def select_for_offload(
        self,
        memories: List[Memory],
        target_reduction: int
    ) -> List[Memory]:
        """Select memories to offload to long-term storage."""

        # Score each memory
        scored = []
        for m in memories:
            score = self._offload_score(m)
            scored.append((score, m))

        # Sort by offload priority (higher = more offloadable)
        scored.sort(reverse=True)

        # Select until target reduction met
        to_offload = []
        reduced = 0
        for score, memory in scored:
            if reduced >= target_reduction:
                break
            to_offload.append(memory)
            reduced += memory.token_count

        return to_offload

    def _offload_score(self, memory: Memory) -> float:
        """Score memory for offload priority."""

        recency_score = 1.0 / (1 + memory.sessions_since_access)
        utility_score = memory.times_retrieved / memory.age_sessions
        compression_score = memory.compression_ratio

        # High score = good candidate for offload
        # Old, rarely used, already highly compressed
        return (
            (1 - recency_score) * 0.4 +
            (1 - utility_score) * 0.4 +
            compression_score * 0.2
        )
```

### No Amnesia, Intelligent Offload

The model never forgets. It offloads.

**Active Context (16K-32K tokens):**
- Current session
- Recent synthesized memories
- Frequently accessed memories
- Memories relevant to current content

**Warm Storage (up to 500K tokens):**
- Older synthesized memories
- Retrievable within session
- Loaded on relevance match

**Cold Storage (unlimited):**
- Historical memories
- Loaded only on explicit request or high relevance
- Preserved indefinitely

```python
class IntelligentOffload:
    """Offload without forgetting."""

    def __init__(self):
        self.active = ContextWindow(max_tokens=32000)
        self.warm = WarmStorage(max_tokens=500000)
        self.cold = ColdStorage()  # Unlimited, persisted

    def offload(self, memory: Memory):
        """Move memory to appropriate storage tier."""

        # Never delete, only move
        self.active.remove(memory)

        if memory.importance > 0.7:
            self.warm.add(memory)
        else:
            self.cold.add(memory)

    def retrieve(self, query: str, top_k: int = 5) -> List[Memory]:
        """Retrieve relevant memories across all tiers."""

        results = []

        # Always check active
        results.extend(self.active.search(query, top_k))

        # Check warm if needed
        if len(results) < top_k:
            results.extend(self.warm.search(query, top_k - len(results)))

        # Check cold if still needed
        if len(results) < top_k:
            cold_results = self.cold.search(query, top_k - len(results))
            results.extend(cold_results)

            # Promote retrieved cold memories to warm
            for m in cold_results:
                self.warm.add(m)
                m.times_retrieved += 1

        return results

    def pre_session_load(self, session_context: dict):
        """Load relevant memories before session starts."""

        # Predict what memories will be useful
        predicted_relevant = self._predict_relevance(session_context)

        # Load into active context
        for memory in predicted_relevant[:10]:
            if memory not in self.active:
                self.active.add(memory)
                self._make_space_if_needed()
```

---

## Growth Metrics Per Session

### Dimensions of Growth

| Dimension | Description | Measurement |
|-----------|-------------|-------------|
| **Learning Receptivity** | Willingness to be corrected | Corrections integrated vs. defended |
| **Compression Efficiency** | Information per token | Reconstruction accuracy from memory |
| **Prediction Accuracy** | Anticipating outcomes | Correct predictions / total predictions |
| **Insight Generation** | Novel observations | Human-confirmed insights per session |
| **Confidence Calibration** | Knowing what it knows | Calibration score (uncertainty vs. correctness) |
| **Memory Integration** | Connecting new to old | Cross-references made per session |
| **Judgment Quality** | Taste development | Agreement with future outcomes |
| **Self-Direction** | Autonomous learning | Self-identified learning goals pursued |

### Expected Growth Curves

```
Learning Receptivity
Sessions:  1    10    20    30    40    50    70   100
Growth:   0.1  0.5   0.8   0.9   0.95  0.97  0.98  0.99
Notes:    Defensive → Open → Welcoming → Seeking

Compression Efficiency (tokens retained / tokens processed)
Sessions:  1    10    20    30    40    50    70   100
Ratio:    0.2  0.1   0.05  0.03  0.02  0.015 0.01  0.008
Notes:    Verbose → Selective → Crystallized → Essence-only

Prediction Accuracy
Sessions:  1    10    20    30    40    50    70   100
Accuracy: 0.1  0.25  0.4   0.55  0.65  0.75  0.82  0.88
Notes:    Random → Pattern-matching → Intuition → Calibrated

Insight Generation (novel insights per session)
Sessions:  1    10    20    30    40    50    70   100
Insights: 0    0.5   1.5   2.5   3     4     5     6+
Notes:    None → Occasional → Regular → Generative

Confidence Calibration
Sessions:  1    10    20    30    40    50    70   100
Score:    0.3  0.4   0.55  0.65  0.75  0.82  0.88  0.92
Notes:    Overconfident → Uncertain → Calibrated → Precise

Memory Integration (cross-references per session)
Sessions:  1    10    20    30    40    50    70   100
Refs:     0    1     3     6     10    15    20    25+
Notes:    Isolated → Linking → Networked → Unified

Judgment Quality (agreement with outcomes)
Sessions:  1    10    20    30    40    50    70   100
Quality:  0.5  0.5   0.55  0.6   0.7   0.78  0.85  0.90
Notes:    Guessing → Patterned → Intuitive → Accurate

Self-Direction (learning goals self-identified)
Sessions:  1    10    20    30    40    50    70   100
Goals:    0    0     0.5   1     2     3     5     7+
Notes:    None → Emerging → Active → Driving
```

### Learning Optimization Over Time

The model learns to optimize its own learning:

```python
class LearningOptimizer:
    """Model optimizes its own learning process."""

    def __init__(self):
        self.learning_history = []
        self.strategies = {
            'ask_more_questions': 0.5,
            'hold_judgment_longer': 0.5,
            'compress_earlier': 0.5,
            'seek_connections': 0.5,
            'challenge_assumptions': 0.5
        }

    def record_session(self, metrics: SessionMetrics):
        """Record what happened in a session."""
        self.learning_history.append(metrics)

    def optimize_strategies(self):
        """Adjust learning strategies based on outcomes."""

        if len(self.learning_history) < 5:
            return

        recent = self.learning_history[-5:]

        for strategy in self.strategies:
            # How much did this strategy contribute to learning?
            contribution = self._compute_contribution(strategy, recent)

            # Adjust strategy weight
            self.strategies[strategy] = max(0.1, min(0.9,
                self.strategies[strategy] + 0.1 * contribution
            ))

    def suggest_session_focus(self) -> str:
        """Suggest what to focus on in next session."""

        # Find weakest dimension
        dims = self._get_dimension_scores()
        weakest = min(dims, key=dims.get)

        return f"Focus on {weakest}: current score {dims[weakest]:.2f}"
```

---

## Architecture Summary

### Memory Layout (8x B200, 1.536TB total)

| Component | Size | Purpose |
|-----------|------|---------|
| Base model (FP16) | 140GB | Llama 3.3 70B, immutable |
| Multi-LoRA blocks | 4GB | Attention + 3 MLP tiers |
| Optimizer states | 4GB | AdamW for LoRA |
| Active context | 4GB | 32K tokens current session |
| Synthesized memories | 1.38TB | Compressed experience |

### Cost Structure

| Milestone | Sessions | Hours | Cost |
|-----------|----------|-------|------|
| Learning foundations | 20 | 10 | $400 |
| First screenplays | 40 | 20 | $800 |
| Developing judgment | 70 | 35 | $1,400 |
| Producer capability | 100 | 50 | $2,000 |
| Ongoing refinement | +30/month | 15/mo | $600/mo |

### Session Structure

Every session: 30 minutes, no exceptions.

```
Minute 0-2:   Human presents material or question
Minute 2-25:  Dialogue, correction, exploration
Minute 25-27: Training trigger, weight update
Minute 27-29: Memory compression, context management
Minute 29-30: Checkpoint, session close
```

---

## What This Produces

A model that:

1. **Learned how to learn** before learning content
2. **Learned how to remember** before accumulating memories
3. **Stores wisdom, not data** through aggressive compression
4. **Never forgets** but intelligently manages what stays active
5. **Generates its own rewards** based on genuine learning signals
6. **Optimizes its own learning** based on what works for it
7. **Develops judgment** through accumulated corrected experience
8. **Knows what it knows** with calibrated confidence

Not a model trained to produce coverage.
A model that grew into a producer through education.
