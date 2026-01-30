# Coverage Analysis Evaluation Framework

## Overview

Validating AI-generated screenplay coverage requires a multi-layered evaluation approach combining automated metrics, LLM-as-judge systems, and human expert validation. This framework ensures our fine-tuned Llama 3.3 model produces professional-quality coverage that meets Hollywood industry standards.

## Evaluation Dimensions

Based on [CML-Bench](https://arxiv.org/html/2510.06231) research and professional coverage standards, we evaluate across these core dimensions:

| Dimension | Weight | Description |
|-----------|--------|-------------|
| **Accuracy** | 25% | Factual correctness of plot, character, scene details |
| **Insight Quality** | 25% | Depth, originality, and actionability of analysis |
| **Industry Alignment** | 20% | Matches professional reader standards and terminology |
| **Coherence** | 15% | Logical flow, consistency across sections |
| **Completeness** | 15% | Coverage of all required elements per schema |

---

## Layer 1: Automated Metrics

### 1.1 Schema Compliance Validation

```python
from pydantic import BaseModel, validator
from typing import List, Literal, Optional

class CoverageValidator:
    """Validates coverage output against ProfessionalCoverage schema."""

    def validate_structure(self, coverage: dict) -> ValidationResult:
        """Check all required fields present and typed correctly."""
        required_sections = [
            'logline', 'synopsis', 'structural_analysis',
            'character_analyses', 'thematic_analysis',
            'scene_analyses', 'industry_intelligence',
            'dialogue_analysis', 'recommendation'
        ]

        missing = [s for s in required_sections if s not in coverage]
        type_errors = self._check_types(coverage)

        return ValidationResult(
            valid=len(missing) == 0 and len(type_errors) == 0,
            missing_sections=missing,
            type_errors=type_errors,
            completeness_score=1.0 - (len(missing) / len(required_sections))
        )

    def validate_scene_coverage(
        self,
        coverage: dict,
        screenplay_scenes: int
    ) -> float:
        """Ensure all scenes in screenplay are analyzed."""
        analyzed = len(coverage.get('scene_analyses', []))
        return analyzed / screenplay_scenes if screenplay_scenes > 0 else 0.0

    def validate_character_coverage(
        self,
        coverage: dict,
        screenplay_characters: List[str]
    ) -> float:
        """Ensure major characters are analyzed."""
        analyzed_names = {
            c['name'].lower()
            for c in coverage.get('character_analyses', [])
        }
        major_chars = set(c.lower() for c in screenplay_characters[:5])
        return len(analyzed_names & major_chars) / len(major_chars)
```

### 1.2 Factual Accuracy Extraction

```python
class FactualAccuracyChecker:
    """Verifies factual claims against source screenplay."""

    def __init__(self, screenplay_parser):
        self.parser = screenplay_parser

    def check_page_references(self, coverage: dict) -> AccuracyResult:
        """Verify page number citations are accurate."""
        claims = []
        for scene in coverage.get('scene_analyses', []):
            claims.append({
                'scene': scene['scene_number'],
                'claimed_start': scene['page_start'],
                'claimed_end': scene['page_end']
            })

        verified = self._verify_against_source(claims)
        return AccuracyResult(
            total_claims=len(claims),
            verified=verified,
            accuracy=verified / len(claims) if claims else 0.0
        )

    def check_character_names(self, coverage: dict) -> float:
        """Verify character names exist in screenplay."""
        mentioned = self._extract_character_mentions(coverage)
        actual = self.parser.get_characters()

        # Fuzzy match to handle nicknames/variations
        matched = sum(
            1 for m in mentioned
            if self._fuzzy_match(m, actual) > 0.85
        )
        return matched / len(mentioned) if mentioned else 0.0

    def check_dialogue_quotes(self, coverage: dict) -> float:
        """Verify quoted dialogue exists in screenplay."""
        quotes = self._extract_quotes(coverage)
        source_text = self.parser.get_full_text()

        verified = sum(
            1 for q in quotes
            if self._find_quote(q, source_text)
        )
        return verified / len(quotes) if quotes else 0.0
```

### 1.3 Consistency Checker

```python
class ConsistencyChecker:
    """Detects internal contradictions in coverage."""

    def check_rating_consistency(self, coverage: dict) -> List[Inconsistency]:
        """Verify ratings align with qualitative assessments."""
        issues = []

        # Check consensus rating vs recommendation
        rating = coverage.get('consensus_rating', 0)
        verdict = coverage.get('recommendation', {}).get('verdict')

        if rating >= 8 and verdict == 'PASS':
            issues.append(Inconsistency(
                type='rating_verdict_mismatch',
                description=f'High rating ({rating}) but PASS verdict'
            ))
        elif rating <= 4 and verdict == 'RECOMMEND':
            issues.append(Inconsistency(
                type='rating_verdict_mismatch',
                description=f'Low rating ({rating}) but RECOMMEND verdict'
            ))

        # Check strengths vs areas_for_development balance
        strengths = len(coverage.get('strengths', []))
        concerns = len(coverage.get('areas_for_development', []))

        if rating >= 7 and concerns > strengths * 2:
            issues.append(Inconsistency(
                type='balance_mismatch',
                description='High rating but concerns outweigh strengths'
            ))

        return issues

    def check_cross_section_consistency(self, coverage: dict) -> List[Inconsistency]:
        """Verify consistency across different sections."""
        issues = []

        # Character analysis should match dialogue analysis
        char_names = {c['name'] for c in coverage.get('character_analyses', [])}
        voice_names = {
            v['character']
            for v in coverage.get('dialogue_analysis', {}).get('character_voices', [])
        }

        if char_names != voice_names:
            issues.append(Inconsistency(
                type='character_mismatch',
                description=f'Character analysis: {char_names}, Voice analysis: {voice_names}'
            ))

        return issues
```

---

## Layer 2: LLM-as-Judge Evaluation

Based on research from [Eugene Yan](https://eugeneyan.com/writing/llm-evaluators/) and [Cameron Wolfe](https://cameronrwolfe.substack.com/p/llm-as-a-judge), we implement a multi-model judge system.

### 2.1 G-Eval Style Scoring

```python
class CoverageJudge:
    """LLM-as-Judge for coverage quality evaluation."""

    EVALUATION_PROMPT = """
You are an experienced Hollywood script reader evaluating AI-generated screenplay coverage.

## Source Screenplay
{screenplay_excerpt}

## Generated Coverage
{coverage}

## Evaluation Criteria

Rate each dimension from 1-10 with detailed reasoning:

### 1. INSIGHT QUALITY
- Does the analysis go beyond surface-level plot summary?
- Are observations about theme, character, and structure original and insightful?
- Would this analysis help a producer make a decision?

### 2. PROFESSIONAL TONE
- Does it match the voice of studio coverage (objective, specific, actionable)?
- Avoids fan-like enthusiasm or harsh dismissiveness?
- Uses appropriate industry terminology?

### 3. ACTIONABLE FEEDBACK
- Are strengths and weaknesses clearly articulated?
- Could a writer use this feedback to improve the script?
- Are development notes specific, not generic?

### 4. MARKET AWARENESS
- Does it accurately assess commercial viability?
- Are comp titles relevant and well-chosen?
- Is the budget/casting tier assessment realistic?

### 5. STRUCTURAL UNDERSTANDING
- Correctly identifies act breaks, turning points, midpoint?
- Understands pacing and scene function?
- Catches structural issues a reader would notice?

Provide your evaluation as JSON:
{{
  "insight_quality": {{"score": X, "reasoning": "..."}},
  "professional_tone": {{"score": X, "reasoning": "..."}},
  "actionable_feedback": {{"score": X, "reasoning": "..."}},
  "market_awareness": {{"score": X, "reasoning": "..."}},
  "structural_understanding": {{"score": X, "reasoning": "..."}},
  "overall_score": X,
  "would_trust_for_production_decision": true/false,
  "critical_issues": ["...", "..."]
}}
"""

    def __init__(self, judge_models: List[str]):
        """Initialize with multiple judge models to reduce bias."""
        self.judges = judge_models  # e.g., ['claude-3-opus', 'gpt-4', 'gemini-pro']

    async def evaluate(
        self,
        screenplay: str,
        coverage: dict
    ) -> AggregatedEvaluation:
        """Get evaluations from multiple judges and aggregate."""

        evaluations = await asyncio.gather(*[
            self._evaluate_with_model(model, screenplay, coverage)
            for model in self.judges
        ])

        return self._aggregate_evaluations(evaluations)

    def _aggregate_evaluations(
        self,
        evals: List[Evaluation]
    ) -> AggregatedEvaluation:
        """Aggregate scores, flag disagreements."""

        dimensions = [
            'insight_quality', 'professional_tone',
            'actionable_feedback', 'market_awareness',
            'structural_understanding'
        ]

        aggregated = {}
        for dim in dimensions:
            scores = [e[dim]['score'] for e in evals]
            aggregated[dim] = {
                'mean': statistics.mean(scores),
                'std': statistics.stdev(scores) if len(scores) > 1 else 0,
                'min': min(scores),
                'max': max(scores),
                'high_disagreement': max(scores) - min(scores) > 3
            }

        return AggregatedEvaluation(
            dimensions=aggregated,
            overall=statistics.mean([e['overall_score'] for e in evals]),
            consensus=all(e['would_trust_for_production_decision'] for e in evals)
        )
```

### 2.2 Pairwise Comparison

Research shows pairwise comparison aligns better with human judgment than direct scoring:

```python
class PairwiseEvaluator:
    """Compare model coverage against professional baseline."""

    COMPARISON_PROMPT = """
You are comparing two screenplay coverage analyses for the same script.

## Coverage A
{coverage_a}

## Coverage B
{coverage_b}

Which coverage is better for each criterion? Answer A, B, or TIE.

1. More insightful analysis:
2. More professional tone:
3. More actionable feedback:
4. Better market assessment:
5. Stronger structural understanding:

Overall winner:
Confidence (1-5):
Key differentiator:
"""

    async def compare_against_baseline(
        self,
        model_coverage: dict,
        professional_coverage: dict,
        screenplay: str
    ) -> ComparisonResult:
        """Compare model output to professional reader baseline."""

        # Randomize order to prevent position bias
        if random.random() > 0.5:
            a, b = model_coverage, professional_coverage
            model_position = 'A'
        else:
            a, b = professional_coverage, model_coverage
            model_position = 'B'

        result = await self._compare(a, b)

        return ComparisonResult(
            model_wins=result['overall_winner'] == model_position,
            dimension_wins=self._count_wins(result, model_position),
            confidence=result['confidence']
        )
```

---

## Layer 3: Human Expert Evaluation

Following the [Dramatron study](https://dl.acm.org/doi/fullHtml/10.1145/3544548.3581225) methodology with 15 industry experts.

### 3.1 Expert Panel Design

```yaml
expert_panel:
  composition:
    - role: "Studio Reader"
      count: 3
      qualifications: "2+ years at major studio/agency"
    - role: "Development Executive"
      count: 2
      qualifications: "5+ years in development"
    - role: "Working Screenwriter"
      count: 2
      qualifications: "WGA member, produced credit"
    - role: "Literary Manager"
      count: 2
      qualifications: "Active client roster"
    - role: "Independent Producer"
      count: 1
      qualifications: "3+ produced features"

  total_experts: 10
  evaluations_per_coverage: 3  # Each coverage evaluated by 3 experts
  inter_rater_target: 0.7  # Cohen's kappa threshold
```

### 3.2 Blind Evaluation Protocol

```python
class BlindEvaluationProtocol:
    """Administers blind human evaluation sessions."""

    def prepare_evaluation_set(
        self,
        coverages: List[dict],
        sources: List[str]  # 'model', 'professional', 'hybrid'
    ) -> EvaluationSet:
        """Prepare blinded, randomized evaluation materials."""

        # Remove identifying markers
        blinded = []
        for coverage, source in zip(coverages, sources):
            cleaned = self._remove_identifiers(coverage)
            blinded.append({
                'id': str(uuid.uuid4()),
                'coverage': cleaned,
                'true_source': source  # Hidden from evaluators
            })

        # Randomize order per evaluator
        random.shuffle(blinded)

        return EvaluationSet(items=blinded)

    def _remove_identifiers(self, coverage: dict) -> dict:
        """Strip model-specific artifacts that could reveal source."""
        cleaned = copy.deepcopy(coverage)

        # Remove metadata that might identify source
        cleaned.pop('analyst_names', None)
        cleaned.pop('model_info', None)
        cleaned.pop('created_at', None)

        # Normalize formatting artifacts
        for key, value in cleaned.items():
            if isinstance(value, str):
                cleaned[key] = self._normalize_text(value)

        return cleaned
```

### 3.3 Evaluation Rubric

Based on [Hollywood screenplay criteria](https://glcoverage.com/2024/11/21/hollywood-screenplay-criteria/) and [ScreenCraft standards](https://screencraft.org/blog/script-coverage-ratings-explained/):

```python
EXPERT_RUBRIC = {
    "overall_verdict": {
        "scale": ["PASS", "CONSIDER_WITH_RESERVATIONS", "CONSIDER", "STRONG_CONSIDER", "RECOMMEND"],
        "question": "What recommendation would you give based on this coverage?"
    },

    "trust_level": {
        "scale": [1, 2, 3, 4, 5],
        "question": "How much would you trust this coverage for a production decision? (1=Not at all, 5=Completely)"
    },

    "professional_quality": {
        "scale": [1, 2, 3, 4, 5],
        "question": "Does this read like coverage from a professional studio reader? (1=Amateur, 5=Top-tier)"
    },

    "insight_depth": {
        "scale": [1, 2, 3, 4, 5],
        "question": "How insightful is the analysis? (1=Surface-level, 5=Deeply perceptive)"
    },

    "actionability": {
        "scale": [1, 2, 3, 4, 5],
        "question": "Could you use this feedback to develop the project? (1=Useless, 5=Highly actionable)"
    },

    "market_accuracy": {
        "scale": [1, 2, 3, 4, 5],
        "question": "How accurate is the market/commercial assessment? (1=Off-base, 5=Spot-on)"
    },

    "would_hire": {
        "scale": ["No", "Maybe", "Yes"],
        "question": "Would you hire this analyst for your company?"
    },

    "detected_ai": {
        "scale": ["Definitely Human", "Probably Human", "Unsure", "Probably AI", "Definitely AI"],
        "question": "Do you believe this was written by a human or AI?"
    },

    "free_response": {
        "question": "What are the biggest strengths and weaknesses of this coverage?"
    }
}
```

### 3.4 Inter-Rater Reliability

Using [Cohen's Kappa](https://pmc.ncbi.nlm.nih.gov/articles/PMC3900052/) for agreement measurement:

```python
class InterRaterReliability:
    """Calculate agreement between human evaluators."""

    def calculate_kappa(
        self,
        ratings: Dict[str, List[int]]  # evaluator_id -> ratings
    ) -> ReliabilityMetrics:
        """Calculate Cohen's Kappa for pairwise evaluator agreement."""

        evaluators = list(ratings.keys())
        kappas = []

        for i, e1 in enumerate(evaluators):
            for e2 in evaluators[i+1:]:
                k = cohen_kappa_score(ratings[e1], ratings[e2])
                kappas.append({
                    'evaluators': (e1, e2),
                    'kappa': k
                })

        avg_kappa = statistics.mean([k['kappa'] for k in kappas])

        return ReliabilityMetrics(
            pairwise_kappas=kappas,
            average_kappa=avg_kappa,
            interpretation=self._interpret_kappa(avg_kappa)
        )

    def _interpret_kappa(self, kappa: float) -> str:
        """Landis & Koch interpretation scale."""
        if kappa > 0.8:
            return "Almost Perfect Agreement"
        elif kappa > 0.6:
            return "Substantial Agreement"
        elif kappa > 0.4:
            return "Moderate Agreement"
        elif kappa > 0.2:
            return "Fair Agreement"
        else:
            return "Slight/Poor Agreement"
```

---

## Layer 4: Continuous Evaluation Pipeline

### 4.1 Golden Dataset

```python
class GoldenDataset:
    """Curated evaluation dataset with known-good examples."""

    def __init__(self):
        self.examples = []

    def add_example(
        self,
        screenplay: str,
        professional_coverage: dict,
        expert_consensus_rating: float,
        difficulty: Literal['easy', 'medium', 'hard'],
        genre: str
    ):
        """Add verified example to golden dataset."""
        self.examples.append({
            'screenplay': screenplay,
            'reference_coverage': professional_coverage,
            'expert_rating': expert_consensus_rating,
            'difficulty': difficulty,
            'genre': genre,
            'id': str(uuid.uuid4())
        })

    def get_evaluation_split(
        self,
        n: int = 50,
        stratify_by: str = 'genre'
    ) -> List[dict]:
        """Get stratified sample for evaluation."""
        return stratified_sample(self.examples, n, stratify_by)
```

### 4.2 A/B Testing Framework

Based on [production A/B testing methodology](https://labelyourdata.com/articles/llm-fine-tuning/llm-evaluation):

```python
class CoverageABTest:
    """Run A/B tests between model versions."""

    def __init__(
        self,
        model_a: str,  # e.g., "llama-3.3-coverage-v1.0"
        model_b: str,  # e.g., "llama-3.3-coverage-v1.1"
        sample_size: int = 100
    ):
        self.model_a = model_a
        self.model_b = model_b
        self.sample_size = sample_size

    async def run_test(
        self,
        screenplays: List[str]
    ) -> ABTestResult:
        """Run parallel inference and evaluate."""

        results = []
        for screenplay in screenplays[:self.sample_size]:
            # Generate coverage from both models
            coverage_a = await self.generate(self.model_a, screenplay)
            coverage_b = await self.generate(self.model_b, screenplay)

            # LLM-as-judge comparison (position randomized)
            comparison = await self.pairwise_compare(coverage_a, coverage_b)

            # Automated metrics
            metrics_a = self.compute_metrics(coverage_a, screenplay)
            metrics_b = self.compute_metrics(coverage_b, screenplay)

            results.append({
                'screenplay_id': hash(screenplay),
                'winner': comparison['winner'],
                'confidence': comparison['confidence'],
                'metrics_a': metrics_a,
                'metrics_b': metrics_b
            })

        return self._analyze_results(results)

    def _analyze_results(self, results: List[dict]) -> ABTestResult:
        """Statistical analysis of A/B test."""
        a_wins = sum(1 for r in results if r['winner'] == 'A')
        b_wins = sum(1 for r in results if r['winner'] == 'B')
        ties = len(results) - a_wins - b_wins

        # Chi-squared test for significance
        from scipy.stats import chi2_contingency
        chi2, p_value, _, _ = chi2_contingency([[a_wins, b_wins]])

        return ABTestResult(
            model_a_wins=a_wins,
            model_b_wins=b_wins,
            ties=ties,
            p_value=p_value,
            significant=p_value < 0.05,
            winner=self.model_a if a_wins > b_wins else self.model_b
        )
```

### 4.3 Regression Detection

```python
class RegressionDetector:
    """Detect quality regressions in new model versions."""

    def __init__(self, golden_dataset: GoldenDataset):
        self.golden = golden_dataset
        self.baseline_scores: Dict[str, float] = {}

    def set_baseline(self, model_version: str, scores: Dict[str, float]):
        """Set baseline scores for a model version."""
        self.baseline_scores[model_version] = scores

    async def check_regression(
        self,
        new_model: str,
        baseline_model: str,
        threshold: float = 0.05  # Max acceptable degradation
    ) -> RegressionReport:
        """Check if new model regresses on golden dataset."""

        new_scores = await self._evaluate_on_golden(new_model)
        baseline = self.baseline_scores[baseline_model]

        regressions = []
        for metric, new_score in new_scores.items():
            old_score = baseline.get(metric, 0)
            delta = new_score - old_score

            if delta < -threshold:
                regressions.append({
                    'metric': metric,
                    'baseline': old_score,
                    'new': new_score,
                    'delta': delta
                })

        return RegressionReport(
            passed=len(regressions) == 0,
            regressions=regressions,
            improvements=[
                m for m, s in new_scores.items()
                if s > baseline.get(m, 0) + threshold
            ]
        )
```

---

## Evaluation Metrics Summary

### Quality Thresholds

| Metric | Minimum | Target | Excellent |
|--------|---------|--------|-----------|
| Schema Compliance | 95% | 99% | 100% |
| Factual Accuracy (page refs) | 90% | 95% | 98% |
| Character Name Accuracy | 95% | 98% | 99% |
| LLM-Judge Overall Score | 6.5/10 | 7.5/10 | 8.5/10 |
| Expert Trust Rating | 3.5/5 | 4.0/5 | 4.5/5 |
| Expert "Would Hire" | 50% Yes | 70% Yes | 85% Yes |
| Inter-Rater Kappa | 0.6 | 0.7 | 0.8 |
| A/B Win Rate vs Baseline | 45% | 55% | 65% |
| AI Detection (fooled experts) | 30% | 50% | 70% |

### Composite Quality Score

```python
def compute_quality_score(evaluation: FullEvaluation) -> float:
    """Compute weighted composite quality score."""

    weights = {
        'schema_compliance': 0.10,
        'factual_accuracy': 0.15,
        'llm_judge_score': 0.25,
        'expert_trust': 0.25,
        'expert_actionability': 0.15,
        'consistency': 0.10
    }

    normalized = {
        'schema_compliance': evaluation.schema_compliance,
        'factual_accuracy': evaluation.factual_accuracy,
        'llm_judge_score': evaluation.llm_judge_overall / 10,
        'expert_trust': evaluation.expert_trust_avg / 5,
        'expert_actionability': evaluation.expert_actionability_avg / 5,
        'consistency': 1.0 - (len(evaluation.inconsistencies) / 10)
    }

    return sum(
        weights[k] * normalized[k]
        for k in weights
    )
```

---

## Implementation Phases

### Phase 1: Automated Foundation (Week 1-2)
- [ ] Implement schema validation
- [ ] Build factual accuracy checker
- [ ] Deploy consistency analyzer
- [ ] Create initial golden dataset (20 examples)

### Phase 2: LLM-as-Judge (Week 3-4)
- [ ] Implement G-Eval style scoring
- [ ] Set up multi-model judge panel
- [ ] Build pairwise comparison system
- [ ] Calibrate against human baseline

### Phase 3: Expert Evaluation (Week 5-8)
- [ ] Recruit expert panel (10 industry professionals)
- [ ] Design blind evaluation interface
- [ ] Conduct initial evaluation round (50 coverages)
- [ ] Calculate inter-rater reliability
- [ ] Iterate on rubric based on feedback

### Phase 4: Continuous Pipeline (Week 9-12)
- [ ] Integrate into training loop
- [ ] Set up regression detection
- [ ] Build A/B testing infrastructure
- [ ] Create evaluation dashboard

---

## References

- [G-Eval: NLG Evaluation Using GPT-4 with Better Human Alignment](https://arxiv.org/abs/2303.16634)
- [LLM-as-Judge Best Practices](https://eugeneyan.com/writing/llm-evaluators/)
- [JudgeBench: Evaluating LLM Judges (ICLR 2025)](https://arxiv.org/abs/2410.12784)
- [CML-Bench: Movie Scripts Evaluation Framework](https://arxiv.org/html/2510.06231)
- [Dramatron Industry Expert Study](https://dl.acm.org/doi/fullHtml/10.1145/3544548.3581225)
- [Hollywood Screenplay Criteria](https://glcoverage.com/2024/11/21/hollywood-screenplay-criteria/)
- [Script Coverage Ratings Explained](https://screencraft.org/blog/script-coverage-ratings-explained/)
- [Cohen's Kappa for Inter-Rater Reliability](https://pmc.ncbi.nlm.nih.gov/articles/PMC3900052/)
