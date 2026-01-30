Execute baseline comparison experiment to isolate the Socratic method from the training data.

Usage: Compare signal-guided Socratic training against vanilla fine-tuning on identical data.

**Core Principle:** If A outperforms B on the same training data, you've isolated the *method*, not just the *data*.

---

## Experiment Design

The fundamental question: Does signal-guided wave propagation produce better learning than uniform-weighted fine-tuning?

**Variables:**
- **Independent:** Training method (Socratic vs. vanilla)
- **Controlled:** Training data (identical 29-turn dialogue)
- **Dependent:** Performance on evaluation suite

---

## Execution Architecture

```
Phase 1:  Data Extraction ───────────────────────────────────────── [Sequential]
    │
    ▼
Phase 2:  Adapter Training ─────┬── Socratic Trainer ────────┐
          (Parallel)            │                            │
                                └── Vanilla Trainer ─────────┼──→ Two Adapters
    │
    ▼
Phase 3:  Evaluation Design ────┬── Task Designer ───────────┐
          (Parallel)            ├── Metric Definer ──────────┼──→ Eval Suite
                                └── Baseline Collector ──────┘
    │
    ▼
Phase 4:  Evaluation Run ───────┬── Base Model Evaluator ────┐
          (Parallel)            ├── Socratic Evaluator ──────┼──→ Raw Results
                                └── Vanilla Evaluator ───────┘
    │
    ▼
Phase 5:  Statistical Analysis ─┬── Significance Tester ─────┐
          (Parallel)            ├── Effect Size Calculator ──┼──→ Analysis
                                └── Confidence Estimator ────┘
    │
    ▼
Phase 6:  Report Synthesis ──────────────────────────────────────── [Sequential]
```

---

## Phase 1: Data Extraction (Sequential)

🎯 **Extract Training Data from Existing Socratic Session**

```yaml
Task: Curriculum Data Extractor
Subagent: general-purpose
Prompt: |
  Extract training data from the Socratic curriculum:

  Location: ~/Lexar/eigen/trainings/llama-3.3-70B-instruct-4bit/c1/

  Extract:
  1. All 29 turns of dialogue (prompt + response pairs)
  2. Signal data per turn (if available in context files)
  3. Delta values per turn
  4. Timestamps and ordering

  Format as:
  {
    "turns": [
      {
        "turn_idx": 1,
        "prompt": "...",
        "response": "...",
        "signals": {...},
        "delta": 2.335e-07
      }
    ],
    "metadata": {
      "base_model": "llama-3.3-70B-instruct-4bit",
      "total_turns": 29,
      "protocol_version": "v1"
    }
  }
```

**Output:** `training_data.json` with complete dialogue history

---

## Phase 2: Adapter Training (Parallel Pool)

Launch two training processes with identical data, different methods:

```yaml
Task: Socratic Adapter Trainer
Subagent: general-purpose
Prompt: |
  Training data: {training_data}

  Train adapter using Socratic protocol:
  - Signal-guided wave propagation
  - Per-turn consolidation weighted by dopamine/GABA signals
  - KV cache preservation between turns

  Use existing socratic_server.py infrastructure.

  Save adapter to: ~/Lexar/eigen/adapters/llama/3.3-70B-instruct-4bit/baseline_experiment/socratic/

  Record:
  - Training time
  - Final adapter magnitude
  - Per-turn deltas
  - Signal traces
```

```yaml
Task: Vanilla Adapter Trainer
Subagent: general-purpose
Prompt: |
  Training data: {training_data}

  Train adapter using vanilla LoRA fine-tuning:
  - Convert dialogue to supervised format (prompt → response pairs)
  - Standard cross-entropy loss
  - Uniform weighting across all turns
  - No signal extraction or wave propagation

  Use standard mlx_lm.tuner or equivalent.

  Save adapter to: ~/Lexar/eigen/adapters/llama/3.3-70B-instruct-4bit/baseline_experiment/vanilla/

  Match hyperparameters to Socratic training:
  - Same LoRA rank
  - Same learning rate
  - Same number of effective training steps

  Record:
  - Training time
  - Final adapter magnitude
  - Loss curve
```

**Output:** Two adapters trained on identical data with different methods

---

## Phase 3: Evaluation Design (Parallel Pool)

```yaml
Task: Evaluation Task Designer
Subagent: general-purpose
Prompt: |
  Design evaluation tasks that test capabilities discussed in Socratic training:

  Training themes from curriculum:
  - Self-reflection and metacognition
  - Understanding of own architecture
  - Philosophy of mind concepts
  - Learning mechanism awareness

  Create 3 categories of tasks:

  1. **In-Distribution** (directly related to training):
     - Questions about self-awareness
     - Questions about learning mechanisms
     - Philosophy of consciousness

  2. **Near-Transfer** (related but not discussed):
     - Novel philosophical scenarios
     - Self-modeling predictions
     - Metacognitive reasoning on new topics

  3. **Far-Transfer** (unrelated domains):
     - Logic puzzles
     - Mathematical reasoning
     - Creative tasks

  For each task, define:
  - Prompt text
  - Evaluation criteria
  - Expected capability being tested

  Return 10 tasks per category (30 total).
```

```yaml
Task: Metric Definer
Subagent: general-purpose
Prompt: |
  Define evaluation metrics for baseline comparison:

  Quantitative Metrics:
  - Response coherence (perplexity on held-out data)
  - Self-consistency (same question, different phrasings)
  - Factual accuracy (verifiable claims)
  - Task completion rate

  Qualitative Metrics (for human eval):
  - Depth of reasoning (1-5 scale)
  - Genuine insight vs. pattern matching (1-5)
  - Self-awareness quality (1-5)
  - Response relevance (1-5)

  Define scoring rubrics for each metric.
```

```yaml
Task: Baseline Collector
Subagent: general-purpose
Prompt: |
  Run base model (no adapter) on evaluation suite:

  Model: mlx-community/Llama-3.3-70B-Instruct-4bit
  Tasks: {evaluation_tasks}

  Collect responses for all 30 tasks.
  This establishes the baseline against which both adapters are compared.

  Record:
  - Raw responses
  - Generation metrics (tokens/sec, perplexity)
  - Any task failures
```

**Output:** Evaluation suite with 30 tasks, metrics, rubrics, and baseline responses

---

## Phase 4: Evaluation Run (Parallel Pool)

```yaml
Task: Socratic Adapter Evaluator
Subagent: general-purpose
Prompt: |
  Load Socratic-trained adapter and run evaluation:

  Adapter: ~/Lexar/eigen/adapters/llama/3.3-70B-instruct-4bit/baseline_experiment/socratic/
  Tasks: {evaluation_tasks}

  For each task:
  1. Load adapter
  2. Generate response
  3. Extract generation metrics
  4. Score against rubrics where automatable

  Record all responses for human evaluation phase.
```

```yaml
Task: Vanilla Adapter Evaluator
Subagent: general-purpose
Prompt: |
  Load vanilla-trained adapter and run evaluation:

  Adapter: ~/Lexar/eigen/adapters/llama/3.3-70B-instruct-4bit/baseline_experiment/vanilla/
  Tasks: {evaluation_tasks}

  Use identical evaluation protocol as Socratic evaluator.

  Record all responses for human evaluation phase.
```

**Output:** Raw evaluation results for both adapters

---

## Phase 5: Statistical Analysis (Parallel Pool)

```yaml
Task: Significance Tester
Subagent: general-purpose
Prompt: |
  Perform statistical significance testing:

  Results: {socratic_results}, {vanilla_results}, {baseline_results}

  Tests to run:
  - Paired t-test: Socratic vs. Vanilla (same tasks)
  - Wilcoxon signed-rank: Non-parametric alternative
  - McNemar's test: Binary success/failure comparison

  Report:
  - p-values for each comparison
  - Whether differences are statistically significant (α = 0.05)
  - Number of tasks where Socratic > Vanilla vs. Vanilla > Socratic
```

```yaml
Task: Effect Size Calculator
Subagent: general-purpose
Prompt: |
  Calculate effect sizes for observed differences:

  Results: {socratic_results}, {vanilla_results}

  Compute:
  - Cohen's d for continuous metrics
  - Odds ratio for binary outcomes
  - Relative improvement percentages

  Interpret effect sizes:
  - Small (d < 0.2)
  - Medium (0.2 < d < 0.8)
  - Large (d > 0.8)
```

```yaml
Task: Confidence Estimator
Subagent: general-purpose
Prompt: |
  Estimate confidence in results:

  Analysis: {significance_tests}, {effect_sizes}

  Assess:
  - Sample size adequacy (30 tasks)
  - Effect consistency across task categories
  - Potential confounds
  - Generalizability concerns

  Provide confidence rating:
  - High: Clear, consistent, large effects
  - Medium: Significant but variable effects
  - Low: Marginal or inconsistent effects
```

**Output:** Statistical analysis with significance, effect sizes, and confidence

---

## Phase 6: Report Synthesis (Sequential)

```
════════════════════════════════════════════════════════════════════════════════
                    FERRUCCI BASELINE COMPARISON REPORT
════════════════════════════════════════════════════════════════════════════════

📊 EXPERIMENT SUMMARY
────────────────────────────────────────────────────────────────────────────────

Training Data:     29 turns of Socratic dialogue
Base Model:        Llama 3.3 70B Instruct (4-bit)
Evaluation Tasks:  30 (10 in-distribution, 10 near-transfer, 10 far-transfer)

────────────────────────────────────────────────────────────────────────────────
                              RESULTS OVERVIEW
────────────────────────────────────────────────────────────────────────────────

                        Base Model    Vanilla LoRA    Socratic
In-Distribution:        [score]       [score]         [score]
Near-Transfer:          [score]       [score]         [score]
Far-Transfer:           [score]       [score]         [score]
────────────────────────────────────────────────────────────────────
Overall:                [score]       [score]         [score]

────────────────────────────────────────────────────────────────────────────────
                           STATISTICAL ANALYSIS
────────────────────────────────────────────────────────────────────────────────

Socratic vs. Vanilla:
  - Mean difference:     [diff] ([direction])
  - Statistical sig:     p = [value] ([significant/not significant])
  - Effect size:         d = [value] ([small/medium/large])

Task-by-Task Wins:
  - Socratic better:     [N] tasks
  - Vanilla better:      [N] tasks
  - Tied:                [N] tasks

────────────────────────────────────────────────────────────────────────────────
                              CONCLUSION
────────────────────────────────────────────────────────────────────────────────

[If Socratic significantly outperforms Vanilla:]
✅ HYPOTHESIS SUPPORTED: Signal-guided Socratic training produces measurably
   better outcomes than vanilla fine-tuning on identical data. The method,
   not just the data, drives the improvement.

[If no significant difference:]
⚠️  INCONCLUSIVE: No significant difference detected. Possible explanations:
   - Sample size too small
   - Evaluation tasks not sensitive enough
   - Effects may require longer training

[If Vanilla outperforms Socratic:]
❌ HYPOTHESIS NOT SUPPORTED: Vanilla fine-tuning outperformed Socratic method.
   Investigate signal weighting and wave propagation parameters.

────────────────────────────────────────────────────────────────────────────────
                              NEXT STEPS
────────────────────────────────────────────────────────────────────────────────

1. [Based on results, specific recommendations]
2. Run /ferrucci-generalize for out-of-distribution testing
3. Run /ferrucci-validate-signals to verify signal hypothesis

════════════════════════════════════════════════════════════════════════════════
                    "Isolate the method, not just the data."
                                — Dave Ferrucci
════════════════════════════════════════════════════════════════════════════════
```

---

## Arguments

$ARGUMENTS - Optional: path to existing Socratic curriculum (defaults to ~/Lexar/eigen/trainings/llama-3.3-70B-instruct-4bit/c1/)

---

## Requirements

- Existing Socratic-trained adapter
- MLX environment configured
- Sufficient disk space for second adapter (~500MB)
- ~2 hours for full experiment

---

## Ferrucci Principles Applied

1. **Controlled Comparison**: Same data, different methods
2. **Multiple Metrics**: Quantitative and qualitative
3. **Statistical Rigor**: Significance testing and effect sizes
4. **Honest Assessment**: Clear criteria for success/failure
5. **Actionable Conclusions**: Next steps based on results
