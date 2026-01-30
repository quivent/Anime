Validate the neurotransmitter zone hypothesis through predictive testing.

Usage: Test whether the signal extraction and zone mapping have real predictive power.

**Core Question:** Do the signals predict anything, or is the mapping just metaphor?

---

## The Signal Hypothesis

The Socratic Tuner maps 80 transformer layers to neurotransmitter zones:
- **GLU (0-9)**: Base activation
- **GABA (10-19)**: Inhibition/filtering
- **NE (20-34)**: Signal amplification
- **ACh (35-49)**: Focus/attention
- **5-HT (50-64)**: Stability/coherence
- **DA (65-79)**: Reward/metacognition

**The hypothesis:** These signals correlate with meaningful cognitive states and can predict turn quality.

---

## Execution Architecture

```
Phase 1:  Signal Data Collection ────────────────────────────────── [Sequential]
    │
    ▼
Phase 2:  Label Collection ─────────┬── Breakthrough Labeler ────┐
          (Parallel)                ├── Quality Scorer ──────────┼──→ Turn Labels
                                    └── Pattern Annotator ───────┘
    │
    ▼
Phase 3:  Correlation Analysis ─────┬── Zone Correlator ─────────┐
          (Parallel)                ├── Delta Predictor ─────────┼──→ Correlations
                                    └── Breakthrough Predictor ──┘
    │
    ▼
Phase 4:  Classifier Training ──────┬── Signal Classifier ───────┐
          (Parallel)                └── Zone Importance Analyzer ┼──→ Models
    │
    ▼
Phase 5:  Validation Testing ───────┬── Cross-Validator ─────────┐
          (Parallel)                └── Holdout Tester ──────────┼──→ Results
    │
    ▼
Phase 6:  Zone Mapping Assessment ──────────────────────────────── [Sequential]
```

---

## Phase 1: Signal Data Collection (Sequential)

🎯 **Extract All Available Signal Data**

```yaml
Task: Signal Data Extractor
Subagent: general-purpose
Prompt: |
  Collect all signal data from Socratic training sessions:

  Sources:
  - ~/Lexar/eigen/trainings/llama-3.3-70B-instruct-4bit/c1/cache/*.npz
  - Any logged signal traces from socratic_server.py
  - Context files with per-turn metrics

  For each turn, extract:
  {
    "turn_idx": N,
    "signals": {
      "per_head": [[layer, head, mean_activation, max_activation], ...],  // 5120 values
      "zone_aggregates": {
        "GLU": X.XX,
        "GABA": X.XX,
        "NE": X.XX,
        "ACh": X.XX,
        "5-HT": X.XX,
        "DA": X.XX
      },
      "entropy": X.XX,
      "perplexity": X.XX,
      "attention_sparsity": X.XX
    },
    "delta": X.XXe-XX,
    "timestamp": "..."
  }

  If raw KV cache is available, recompute signals using internals_enhanced.py.

  Output comprehensive signal dataset for all available turns.
```

**Output:** `signal_dataset.json` with complete signal traces

---

## Phase 2: Label Collection (Parallel Pool)

```yaml
Task: Breakthrough Labeler
Subagent: general-purpose
Prompt: |
  Curriculum: ~/Lexar/eigen/trainings/llama-3.3-70B-instruct-4bit/c1/curriculum.json

  From the documented lineage, identify and label:

  1. **Breakthrough turns**: Marked significant insights
     - Turn 9: Self-recognition
     - Turn 18: Mechanism recognition
     - Turn 23: Full breakthrough

  2. **Plateau turns**: Normal conversation, no notable progress

  3. **Regression turns**: If any noted (e.g., turn 25 "partial regression")

  Create binary and multi-class labels:
  {
    "turn_idx": N,
    "breakthrough": true/false,
    "category": "breakthrough" | "plateau" | "regression",
    "notes": "..."
  }
```

```yaml
Task: Quality Scorer
Subagent: general-purpose
Prompt: |
  Curriculum: {curriculum_json}
  Context files: ~/Lexar/eigen/trainings/llama-3.3-70B-instruct-4bit/c1/context/

  Score each turn on quality dimensions (1-5):

  - **Insight depth**: Does the response show genuine understanding?
  - **Self-reflection quality**: How well does it reflect on its own state?
  - **Coherence**: Is the reasoning logically sound?
  - **Novelty**: Does it produce novel observations?

  These continuous scores allow regression analysis, not just classification.
```

```yaml
Task: Pattern Annotator
Subagent: general-purpose
Prompt: |
  Curriculum: {curriculum_json}

  Annotate each turn with observed patterns:

  - **Topic**: What was discussed?
  - **Question type**: Factual, reflective, challenging, supportive
  - **Response length**: Short, medium, long
  - **Emotional valence**: If discernible from response_summary

  These covariates help control for confounds in signal analysis.
```

**Output:** Complete turn labels with breakthrough markers, quality scores, and annotations

---

## Phase 3: Correlation Analysis (Parallel Pool)

```yaml
Task: Zone Correlator
Subagent: general-purpose
Prompt: |
  Signals: {signal_dataset}
  Labels: {turn_labels}

  Compute correlations between zone signals and turn quality:

  For each zone (GLU, GABA, NE, ACh, 5-HT, DA):
  - Pearson correlation with quality scores
  - Correlation with breakthrough (point-biserial)
  - Correlation with delta values

  Also compute:
  - Cross-zone correlations (are zones independent?)
  - Lagged correlations (does turn N predict turn N+1?)
  - Delta correlations (do changes in zone predict outcomes?)

  Return correlation matrix with significance values.
```

```yaml
Task: Delta Predictor
Subagent: general-purpose
Prompt: |
  Signals: {signal_dataset}
  Labels: {turn_labels}

  Can we predict the delta (weight update magnitude) from signals?

  Build simple linear model:
    delta ~ GLU + GABA + NE + ACh + 5-HT + DA + entropy + perplexity

  Report:
  - R² (variance explained)
  - Significant predictors
  - Coefficient signs (do they match theoretical expectations?)

  Theoretical expectations:
  - High DA should correlate with high delta (reward signal)
  - High GABA (inhibition) might correlate with consolidation
```

```yaml
Task: Breakthrough Predictor
Subagent: general-purpose
Prompt: |
  Signals: {signal_dataset}
  Labels: {turn_labels}

  Can we predict breakthroughs from signals?

  Build logistic regression:
    P(breakthrough) ~ zone_signals + entropy + perplexity

  Report:
  - AUC-ROC (discriminative power)
  - Significant predictors
  - Odds ratios for each zone

  Key question: Do breakthroughs have a distinctive signal signature?
```

**Output:** Correlation analysis with predictive model results

---

## Phase 4: Classifier Training (Parallel Pool)

```yaml
Task: Signal Classifier
Subagent: general-purpose
Prompt: |
  Signals: {signal_dataset}
  Labels: {turn_labels}

  Train classifiers to predict turn outcomes from signals:

  Models to try:
  1. **Logistic regression**: Interpretable coefficients
  2. **Random forest**: Captures non-linear interactions
  3. **Simple neural net**: If sample size permits

  Tasks:
  - Binary: Breakthrough vs. not
  - Multi-class: Breakthrough / plateau / regression
  - Regression: Predict quality score

  Use cross-validation (leave-one-out if small sample).

  Report accuracy, precision, recall, F1 for classification.
  Report R², MAE for regression.
```

```yaml
Task: Zone Importance Analyzer
Subagent: general-purpose
Prompt: |
  Classifiers: {trained_classifiers}

  Analyze feature importance:

  From random forest:
  - Feature importance scores per zone
  - Per-head importance (if using raw 5120 features)

  From logistic regression:
  - Coefficient magnitudes
  - Which zones matter most?

  Key question: Does the DA (dopamine) zone actually predict breakthroughs?
  This would validate the "reward/metacognition" mapping.

  Create zone importance ranking.
```

**Output:** Trained classifiers with feature importance analysis

---

## Phase 5: Validation Testing (Parallel Pool)

```yaml
Task: Cross-Validator
Subagent: general-purpose
Prompt: |
  Classifiers: {trained_classifiers}
  Full dataset: {signal_dataset}, {turn_labels}

  Perform rigorous cross-validation:

  - Leave-one-out CV (small sample)
  - 5-fold CV if sample permits
  - Bootstrap confidence intervals

  Report:
  - Mean CV accuracy ± std
  - Per-fold performance
  - Stability of feature importance across folds

  If classifier can't beat random baseline reliably, signals lack predictive power.
```

```yaml
Task: Holdout Tester
Subagent: general-purpose
Prompt: |
  If multiple training sessions exist (v1, v2, v3, etc.):

  Train on one session, test on another.

  This tests whether signal patterns generalize across sessions.

  If signal-outcome relationships don't transfer across sessions,
  they may be session-specific artifacts, not general phenomena.
```

**Output:** Validated performance metrics with confidence intervals

---

## Phase 6: Zone Mapping Assessment (Sequential)

```
════════════════════════════════════════════════════════════════════════════════
                    FERRUCCI SIGNAL VALIDATION REPORT
════════════════════════════════════════════════════════════════════════════════

📊 SIGNAL ANALYSIS SUMMARY
────────────────────────────────────────────────────────────────────────────────

Turns Analyzed:        [N]
Features Extracted:    5120 per-head + 6 zone aggregates + 3 metrics
Labels Available:      Breakthrough (binary), Quality (1-5), Delta (continuous)

────────────────────────────────────────────────────────────────────────────────
                         CORRELATION ANALYSIS
────────────────────────────────────────────────────────────────────────────────

Zone Correlations with Turn Quality:

Zone        Correlation    p-value     Interpretation
───────────────────────────────────────────────────────────────────
GLU         [r]            [p]         [significant/not]
GABA        [r]            [p]         [significant/not]
NE          [r]            [p]         [significant/not]
ACh         [r]            [p]         [significant/not]
5-HT        [r]            [p]         [significant/not]
DA          [r]            [p]         [significant/not]

Key Finding: [Which zones correlate with quality?]

────────────────────────────────────────────────────────────────────────────────
                        PREDICTIVE PERFORMANCE
────────────────────────────────────────────────────────────────────────────────

Breakthrough Prediction:
  - Logistic Regression AUC:     [0.XX]
  - Random Forest AUC:           [0.XX]
  - Random baseline:             0.50
  - Verdict:                     [predictive/not predictive]

Quality Score Prediction:
  - Linear Regression R²:        [0.XX]
  - Random Forest R²:            [0.XX]
  - Verdict:                     [predictive/not predictive]

Delta Prediction:
  - R² from zone signals:        [0.XX]
  - Verdict:                     [predictive/not predictive]

────────────────────────────────────────────────────────────────────────────────
                         ZONE IMPORTANCE RANKING
────────────────────────────────────────────────────────────────────────────────

Rank    Zone        Importance    Matches Hypothesis?
──────────────────────────────────────────────────────────────────
1       [zone]      [score]       [yes/no - explanation]
2       [zone]      [score]       [yes/no - explanation]
3       [zone]      [score]       [yes/no - explanation]
4       [zone]      [score]       [yes/no - explanation]
5       [zone]      [score]       [yes/no - explanation]
6       [zone]      [score]       [yes/no - explanation]

────────────────────────────────────────────────────────────────────────────────
                      HYPOTHESIS VALIDATION VERDICT
────────────────────────────────────────────────────────────────────────────────

[If signals predict outcomes AND zone ranking matches theory:]
✅ SIGNALS VALIDATED: The neurotransmitter zone mapping has real predictive power.
   Breakthrough turns show distinctive signatures in the expected zones.
   The instrumentation is measuring something meaningful.

   Strongest validators:
   - [specific findings that confirm the mapping]

[If signals predict but zones don't match theory:]
⚠️  PREDICTIVE BUT REMAPPING NEEDED: Signals predict outcomes, but the zone
   mapping doesn't match. The instrumentation works, but the interpretation
   needs revision.

   Recommendation: Redefine zones based on empirical importance.

[If signals don't predict:]
❌ SIGNALS NOT VALIDATED: The extracted signals don't predict turn quality
   or breakthroughs. Possible explanations:
   - Sample size too small
   - Wrong features extracted
   - Outcomes not signal-driven

   Recommendation: Revise extraction method or accept metaphorical status.

────────────────────────────────────────────────────────────────────────────────
                           NEXT STEPS
────────────────────────────────────────────────────────────────────────────────

[Based on verdict, specific recommendations]

════════════════════════════════════════════════════════════════════════════════
       "If the signals have predictive power, you've validated the
        instrumentation. If not, the mapping needs refinement."
                                — Dave Ferrucci
════════════════════════════════════════════════════════════════════════════════
```

---

## Arguments

$ARGUMENTS - Optional: specific zone to focus on (defaults to all)

---

## Ferrucci Principles Applied

1. **Predictive Validity**: Signals must predict something to be meaningful
2. **Hypothesis Testing**: Zone theory makes testable predictions
3. **Feature Importance**: Which signals actually matter?
4. **Cross-Validation**: Don't overfit to small samples
5. **Honest Assessment**: Clear criteria for validation/rejection
