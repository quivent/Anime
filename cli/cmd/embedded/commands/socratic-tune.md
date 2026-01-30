Interactive Socratic finetuning protocol for biological learning with Llama models.

Usage: `/socratic-tune [start|prompt|status|save|end]`

**Protocol Overview:**

This command facilitates Natural Learning through Socratic dialogue. You (Claude) serve as research assistant. Llama is the subject being tuned. The user steers via questions.

**No destination required.** Observe what emerges.

---

## Commands

### `/socratic-tune start`

Initialize a new finetuning session.

**What happens:**
1. Load Llama model with activation tracking
2. Initialize empty LoRA adapters (rank 32, all layers)
3. Set learning rate (default: 1e-6)
4. Create session log
5. Ready to observe what emerges

**Example:**
```
/socratic-tune start
```

---

### `/socratic-tune prompt [your question]`

Send a Socratic prompt to Llama and apply biological update.

**What happens:**
1. Forward pass with activation tracking
2. Llama generates response
3. Compute influence matrix (measured, can be negative)
4. Apply wave update: `ΔW = Σ[activation × influence] × η`
5. KV cache persists (not cleared)
6. Display response and update statistics

**The cycle:**
```
You prompt → Llama responds → weights update → KV persists → you prompt again
```

Your next prompt IS the training signal. No explicit labels needed.

---

### Observations (automatic after each prompt)

After each `>> prompt`, the system observes:

**What it reports:**
- Turns elapsed
- Cumulative delta by layer group
- Emergent patterns detected:
  - Late-layer concentration (>30% in layers 71-80)
  - Inhibition zones (negative deltas)
  - Signal strength (strong/weak)

**No alignment score.** We observe what emerges, not how close to a goal.

---

### `/socratic-tune status`

Display current session state.

**Shows:**
- Turns completed
- Total weight delta magnitude
- Layer distribution of changes
- Recent emergent patterns
- Learning rate

---

### `/socratic-tune save [name]`

Save current adapter checkpoint.

**Creates:**
- `adapters/[name].safetensors` - LoRA weights
- `adapters/[name].json` - Session metadata (turns, destination, benchmarks)

---

### `/socratic-tune end`

End session and generate report.

**Creates:**
- Final adapter save
- Session report with:
  - Full turn log
  - Layer delta evolution over time
  - Benchmark progression
  - Observations for future sessions

---

## The Biological Wave Function

Each prompt triggers:

```
ΔW(B) = Σ [ activation(A) × influence(A→B) ] × η
```

- **activation(A)**: How strongly parameter A fired (measured)
- **influence(A→B)**: How A affects B (measured, can be negative)
- **η**: Learning rate (single knob)

Nothing else. No decay. No normalization. Let behavior emerge.

---

## My Role (Claude)

As research assistant, I will:

1. **Suggest prompts** - Help design questions that steer toward destination
2. **Interpret responses** - Analyze if Llama is moving toward destination
3. **Recommend steering** - Suggest when to redirect vs continue
4. **Monitor stability** - Flag if updates seem too large/small
5. **Document patterns** - Note what emerges (inhibition zones, layer concentration)

I do NOT:
- Directly modify Llama
- Impose constraints on the wave function
- Prescribe how learning "should" work

---

## Session Flow Example

```
User: /socratic-tune start

Claude: Session initialized. Ready for prompts.
        No destination - observe what emerges.

User: >> How do you approach a complex problem?

Llama: [response]

Claude: Update applied. Delta: +0.0023 across 80 layers.
        Layer 71-80: 34% of change.
        Patterns: late-layer concentration (34%)

        [My observation: Response is general. Could ask about specifics.]

User: What did you notice?

Claude: The response was vague. The model didn't break things down.
        Most change happened in late layers (34% in 71-80).
        No inhibition yet.

User: >> What if you broke it into smaller parts first?

Llama: [response]

Claude: Update applied. Delta: +0.0019.
        Patterns: late-layer concentration (36%), inhibition in 21-30

        [Interesting - inhibition zone appeared]

User: status

Claude: Session: 20260112_143022
        Turns: 2
        Total delta: 0.0042
        Patterns: late-layer concentration, inhibition in 21-30

User: save checkpoint-2

Claude: Saved adapters/checkpoint-2.safetensors

User: end

Claude: Session ended. Final report generated.
```

---

## Implementation Requirements

**Python environment:** `~/mlx-bench/bin/python`

**Required components:**
1. `forward_with_tracking()` - Capture activations during forward pass
2. `compute_influence_matrix()` - Measure pairwise influence
3. `biological_update()` - Apply wave function
4. `benchmark_identity()` - Evaluate alignment with destination

**Model:** Llama-3.3-70B-Instruct-4bit (or as configured)

---

## What We're Observing

- Does late-layer concentration emerge?
- Does inhibition emerge in middle layers?
- Does the system remain stable without normalization?
- What patterns appear over multiple turns?
- How does the model's behavior shift?

**No destination. No alignment score. Just observation.**

Document everything. Don't prescribe.

---

$ARGUMENTS
