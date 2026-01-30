# Screenplay Analysis AI Architecture

## Fine-Tuning Llama 3.3 70B + RAG System for Professional Script Coverage

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [System Architecture Overview](#system-architecture-overview)
3. [Fine-Tuning Pipeline](#fine-tuning-pipeline)
4. [RAG System Design](#rag-system-design)
5. [Data Models](#data-models)
6. [Infrastructure Requirements](#infrastructure-requirements)
7. [API Specifications](#api-specifications)
8. [Integration with Existing Systems](#integration-with-existing-systems)
9. [Deployment Strategy](#deployment-strategy)
10. [Evaluation & Metrics](#evaluation-metrics)
11. [Implementation Phases](#implementation-phases)

---

## Executive Summary

This document outlines the architecture for a screenplay analysis AI system that combines:

1. **Fine-tuned Llama 3.3 70B** - Specialized for screenplay coverage, character analysis, and narrative structure
2. **RAG (Retrieval-Augmented Generation)** - Vector database storing screenplays, coverage examples, and industry knowledge
3. **Integration with ANIME CLI/Desktop** - Seamless workflow with existing coverage and analysis tools

### Key Capabilities

| Capability | Description |
|------------|-------------|
| **Script Coverage** | Generate professional studio-quality coverage reports |
| **Character Analysis** | Deep analysis of character arcs, voice, and development |
| **Structural Analysis** | Act structure, pacing, turning points identification |
| **Dialogue Assessment** | Voice consistency, subtext, commercial appeal |
| **Comparative Analysis** | Compare scripts against successful films in genre |
| **Industry Intelligence** | Budget estimation, casting suggestions, market positioning |

### Performance Targets

| Metric | Target |
|--------|--------|
| Coverage Quality (vs human) | 85%+ parity |
| Analysis Latency | < 30 seconds for full script |
| Context Window Utilization | Full 128K for complete screenplays |
| Retrieval Accuracy | 90%+ relevant context |

---

## System Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         SCREENPLAY ANALYSIS SYSTEM                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌───────────────┐    ┌───────────────┐    ┌───────────────────────────┐   │
│  │   ANIME CLI   │    │ ANIME Desktop │    │      External APIs        │   │
│  │   (Go/Cobra)  │    │ (Tauri/React) │    │   (REST/WebSocket)        │   │
│  └───────┬───────┘    └───────┬───────┘    └─────────────┬─────────────┘   │
│          │                    │                          │                  │
│          └────────────────────┴──────────────────────────┘                  │
│                               │                                             │
│                    ┌──────────▼──────────┐                                  │
│                    │   ANALYSIS GATEWAY  │                                  │
│                    │   (Go HTTP Server)  │                                  │
│                    └──────────┬──────────┘                                  │
│                               │                                             │
│          ┌────────────────────┼────────────────────┐                        │
│          │                    │                    │                        │
│  ┌───────▼───────┐   ┌───────▼───────┐   ┌───────▼───────┐                 │
│  │  SCRIPT PARSER │   │  RAG RETRIEVER │   │ ANALYSIS CACHE │                │
│  │   (Fountain)   │   │   (Vector DB)  │   │    (Redis)     │                │
│  └───────┬───────┘   └───────┬───────┘   └───────────────┘                 │
│          │                   │                                              │
│          └───────────┬───────┘                                              │
│                      │                                                      │
│           ┌──────────▼──────────┐                                           │
│           │   CONTEXT BUILDER   │                                           │
│           │  (Prompt Assembly)  │                                           │
│           └──────────┬──────────┘                                           │
│                      │                                                      │
│           ┌──────────▼──────────┐                                           │
│           │  LLAMA 3.3 70B-FT   │                                           │
│           │  (vLLM / TensorRT)  │                                           │
│           └──────────┬──────────┘                                           │
│                      │                                                      │
│           ┌──────────▼──────────┐                                           │
│           │  OUTPUT PROCESSOR   │                                           │
│           │ (JSON Structuring)  │                                           │
│           └──────────┬──────────┘                                           │
│                      │                                                      │
│           ┌──────────▼──────────┐                                           │
│           │  COVERAGE GENERATOR │                                           │
│           │ (Report Formatting) │                                           │
│           └─────────────────────┘                                           │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                           DATA LAYER                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────────┐  │
│  │   SCREENPLAY    │  │   COVERAGE      │  │      KNOWLEDGE BASE         │  │
│  │   VECTOR DB     │  │   EXAMPLES DB   │  │    (Industry Intelligence)  │  │
│  │   (Qdrant)      │  │   (Qdrant)      │  │         (Qdrant)            │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────────────────┘  │
│                                                                             │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────────┐  │
│  │   FINE-TUNING   │  │   MODEL         │  │      TRAINING               │  │
│  │   DATASET       │  │   CHECKPOINTS   │  │      METRICS                │  │
│  │   (HuggingFace) │  │   (S3/Local)    │  │      (W&B)                  │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Fine-Tuning Pipeline

### 3.1 Base Model Selection

| Model | Parameters | Context | Rationale |
|-------|-----------|---------|-----------|
| **Llama 3.3 70B Instruct** | 70B | 128K | Best balance of capability and trainability |

**Why Llama 3.3 70B:**
- 128K context fits complete screenplays (avg 40-60K tokens)
- Strong baseline creative writing (7.5-8.0 on benchmarks)
- Extensive LoRA/QLoRA tooling support
- Open weights allow unrestricted fine-tuning

### 3.2 Training Data Architecture

```
training_data/
├── screenplay_coverage/
│   ├── studio_coverage/           # Professional studio coverage examples
│   │   ├── format_a/             # Different coverage formats
│   │   └── format_b/
│   ├── professional_analyses/     # Industry analyst breakdowns
│   └── academic_analyses/         # Film school analyses
│
├── screenplay_corpus/
│   ├── produced_scripts/          # MovieSum dataset (2,200 scripts)
│   ├── award_winners/             # Oscar-nominated screenplays
│   ├── genre_exemplars/           # Best-in-class by genre
│   └── development_drafts/        # Multiple draft versions
│
├── structural_annotations/
│   ├── beat_sheets/              # Save the Cat, Hero's Journey
│   ├── turning_points/           # Annotated story turning points
│   └── act_breakdowns/           # Professional act analysis
│
├── character_analyses/
│   ├── arc_annotations/          # Character arc descriptions
│   ├── voice_samples/            # Dialogue by character type
│   └── casting_comps/            # Historical casting decisions
│
└── industry_knowledge/
    ├── box_office_data/          # Revenue and budget data
    ├── festival_results/         # Award history
    └── market_trends/            # Genre performance trends
```

### 3.3 Dataset Preparation

#### Source Datasets

| Dataset | Size | Use Case |
|---------|------|----------|
| [MovieSum](https://huggingface.co/datasets/rohitsaxena/MovieSum) | 2,200 scripts | Base screenplay corpus |
| [IMSDB Scripts](https://huggingface.co/datasets/aneeshas/imsdb-genre-movie-scripts) | ~1,000 scripts | Genre diversity |
| Custom Coverage | 500+ reports | Coverage format training |
| RealScripts (internal) | 10+ analyzed | High-quality exemplars |

#### Data Format for Fine-Tuning

```json
{
  "instruction": "Analyze the following screenplay and provide professional studio coverage.",
  "input": "<screenplay>\n[FULL SCREENPLAY TEXT]\n</screenplay>\n\n<analysis_requirements>\n- Provide logline (1 sentence)\n- Synopsis (500 words)\n- Rate: premise, character, dialogue, structure, pacing (1-10)\n- Identify 3 key strengths\n- Identify 3 areas for development\n- Recommendation: PASS/CONSIDER/RECOMMEND\n</analysis_requirements>",
  "output": "<coverage_report>\n<logline>A brilliant AI engineer must confront her past...</logline>\n<synopsis>...</synopsis>\n<ratings>\n<premise>9</premise>\n<character>8</character>\n...\n</ratings>\n...\n</coverage_report>",
  "metadata": {
    "genre": ["sci-fi", "thriller"],
    "budget_range": "medium",
    "source": "professional_coverage",
    "quality_score": 9.2
  }
}
```

#### Training Data Synthesis Pipeline

```python
# data_synthesis.py

from dataclasses import dataclass
from typing import List, Dict, Optional
import json

@dataclass
class TrainingExample:
    instruction: str
    input: str
    output: str
    metadata: Dict

class ScreenplayDataSynthesizer:
    """Generate training examples from screenplay corpus."""

    TASK_TEMPLATES = {
        "full_coverage": {
            "instruction": "Provide comprehensive professional coverage for this screenplay.",
            "output_schema": "ProfessionalCoverage"
        },
        "character_analysis": {
            "instruction": "Analyze the characters in this screenplay, focusing on arcs and development.",
            "output_schema": "CharacterAnalysis[]"
        },
        "scene_breakdown": {
            "instruction": "Break down this scene, analyzing function, dialogue quality, and production considerations.",
            "output_schema": "SceneAnalysis"
        },
        "structural_analysis": {
            "instruction": "Analyze the structural elements of this screenplay.",
            "output_schema": "StructuralAnalysis"
        },
        "dialogue_assessment": {
            "instruction": "Assess the dialogue quality and character voices in this screenplay.",
            "output_schema": "DialogueAnalysis"
        },
        "market_positioning": {
            "instruction": "Provide market analysis and positioning for this screenplay.",
            "output_schema": "IndustryIntelligence"
        }
    }

    def __init__(self, screenplay_corpus: List[str], coverage_examples: List[Dict]):
        self.screenplays = screenplay_corpus
        self.coverage = coverage_examples

    def generate_training_set(self, num_examples: int = 10000) -> List[TrainingExample]:
        """Generate diverse training examples."""
        examples = []

        # Full coverage examples (40%)
        examples.extend(self._generate_full_coverage(int(num_examples * 0.4)))

        # Character analysis (20%)
        examples.extend(self._generate_character_tasks(int(num_examples * 0.2)))

        # Scene breakdowns (15%)
        examples.extend(self._generate_scene_tasks(int(num_examples * 0.15)))

        # Structural analysis (15%)
        examples.extend(self._generate_structure_tasks(int(num_examples * 0.15)))

        # Market/industry (10%)
        examples.extend(self._generate_market_tasks(int(num_examples * 0.1)))

        return examples

    def _generate_full_coverage(self, count: int) -> List[TrainingExample]:
        """Generate full coverage training examples."""
        examples = []
        for i in range(count):
            screenplay = self._sample_screenplay()
            coverage = self._get_or_synthesize_coverage(screenplay)

            examples.append(TrainingExample(
                instruction=self.TASK_TEMPLATES["full_coverage"]["instruction"],
                input=f"<screenplay>\n{screenplay['text']}\n</screenplay>",
                output=self._format_coverage_output(coverage),
                metadata={
                    "task": "full_coverage",
                    "genre": screenplay.get("genre", []),
                    "quality_tier": coverage.get("quality_tier", "standard")
                }
            ))
        return examples
```

### 3.4 Fine-Tuning Configuration

#### QLoRA Configuration

```yaml
# finetune_config.yaml

model:
  base_model: "meta-llama/Llama-3.3-70B-Instruct"
  model_type: "llama"

quantization:
  load_in_4bit: true
  bnb_4bit_compute_dtype: "bfloat16"
  bnb_4bit_quant_type: "nf4"
  bnb_4bit_use_double_quant: true

lora:
  r: 64                          # Rank (higher for complex domain)
  lora_alpha: 128                # Alpha = 2 * r
  lora_dropout: 0.05
  target_modules:
    - "q_proj"
    - "k_proj"
    - "v_proj"
    - "o_proj"
    - "gate_proj"
    - "up_proj"
    - "down_proj"
  bias: "none"
  task_type: "CAUSAL_LM"

training:
  num_train_epochs: 3
  per_device_train_batch_size: 1
  gradient_accumulation_steps: 16
  learning_rate: 2e-4
  lr_scheduler_type: "cosine"
  warmup_ratio: 0.03
  weight_decay: 0.01
  max_grad_norm: 1.0

  # Long context training
  max_seq_length: 65536          # Half of 128K for efficiency

  # Memory optimization
  gradient_checkpointing: true
  optim: "paged_adamw_8bit"

  # Logging
  logging_steps: 10
  save_steps: 500
  eval_steps: 500
  save_total_limit: 3

data:
  train_dataset: "./training_data/train.jsonl"
  eval_dataset: "./training_data/eval.jsonl"
  dataset_format: "alpaca"

  # Data processing
  packing: true                   # Pack short examples
  max_packed_length: 32768

evaluation:
  metrics:
    - "coverage_quality"          # Custom metric
    - "character_consistency"
    - "structural_accuracy"
    - "dialogue_assessment"

output:
  output_dir: "./checkpoints/llama-3.3-70b-screenplay"
  hub_model_id: "anime/llama-screenplay-70b"
```

#### Training Script

```python
# train.py

import torch
from transformers import (
    AutoModelForCausalLM,
    AutoTokenizer,
    BitsAndBytesConfig,
    TrainingArguments,
)
from peft import LoraConfig, get_peft_model, prepare_model_for_kbit_training
from trl import SFTTrainer
from datasets import load_dataset
import yaml

def load_config(path: str) -> dict:
    with open(path) as f:
        return yaml.safe_load(f)

def main():
    config = load_config("finetune_config.yaml")

    # Quantization config
    bnb_config = BitsAndBytesConfig(
        load_in_4bit=config["quantization"]["load_in_4bit"],
        bnb_4bit_compute_dtype=getattr(torch, config["quantization"]["bnb_4bit_compute_dtype"]),
        bnb_4bit_quant_type=config["quantization"]["bnb_4bit_quant_type"],
        bnb_4bit_use_double_quant=config["quantization"]["bnb_4bit_use_double_quant"],
    )

    # Load model
    model = AutoModelForCausalLM.from_pretrained(
        config["model"]["base_model"],
        quantization_config=bnb_config,
        device_map="auto",
        trust_remote_code=True,
        attn_implementation="flash_attention_2",
    )

    tokenizer = AutoTokenizer.from_pretrained(
        config["model"]["base_model"],
        trust_remote_code=True,
    )
    tokenizer.pad_token = tokenizer.eos_token
    tokenizer.padding_side = "right"

    # Prepare for k-bit training
    model = prepare_model_for_kbit_training(model)

    # LoRA config
    lora_config = LoraConfig(
        r=config["lora"]["r"],
        lora_alpha=config["lora"]["lora_alpha"],
        lora_dropout=config["lora"]["lora_dropout"],
        target_modules=config["lora"]["target_modules"],
        bias=config["lora"]["bias"],
        task_type=config["lora"]["task_type"],
    )

    model = get_peft_model(model, lora_config)
    model.print_trainable_parameters()

    # Load dataset
    dataset = load_dataset(
        "json",
        data_files={
            "train": config["data"]["train_dataset"],
            "eval": config["data"]["eval_dataset"],
        },
    )

    # Format function
    def format_example(example):
        return f"""### Instruction:
{example['instruction']}

### Input:
{example['input']}

### Response:
{example['output']}"""

    # Training arguments
    training_args = TrainingArguments(
        output_dir=config["output"]["output_dir"],
        num_train_epochs=config["training"]["num_train_epochs"],
        per_device_train_batch_size=config["training"]["per_device_train_batch_size"],
        gradient_accumulation_steps=config["training"]["gradient_accumulation_steps"],
        learning_rate=config["training"]["learning_rate"],
        lr_scheduler_type=config["training"]["lr_scheduler_type"],
        warmup_ratio=config["training"]["warmup_ratio"],
        weight_decay=config["training"]["weight_decay"],
        max_grad_norm=config["training"]["max_grad_norm"],
        gradient_checkpointing=config["training"]["gradient_checkpointing"],
        optim=config["training"]["optim"],
        logging_steps=config["training"]["logging_steps"],
        save_steps=config["training"]["save_steps"],
        eval_steps=config["training"]["eval_steps"],
        save_total_limit=config["training"]["save_total_limit"],
        bf16=True,
        report_to="wandb",
    )

    # Trainer
    trainer = SFTTrainer(
        model=model,
        train_dataset=dataset["train"],
        eval_dataset=dataset["eval"],
        formatting_func=format_example,
        max_seq_length=config["training"]["max_seq_length"],
        tokenizer=tokenizer,
        args=training_args,
        packing=config["data"]["packing"],
    )

    # Train
    trainer.train()

    # Save
    trainer.save_model()
    trainer.push_to_hub(config["output"]["hub_model_id"])

if __name__ == "__main__":
    main()
```

### 3.5 Expected Performance After Fine-Tuning

| Metric | Base Llama 3.3 70B | After Fine-Tuning |
|--------|-------------------|-------------------|
| Creative Writing (lechmazur) | ~5.8 (Maverick) | 7.5-8.0 |
| Coverage Format Adherence | 60% | 95%+ |
| Character Consistency (CML-Bench CC) | Good | Excellent |
| Dialogue Coherence (CML-Bench DC) | Good | Excellent |
| Plot Reasonableness (CML-Bench PR) | Good | Excellent |
| Industry Terminology | Moderate | Expert |
| Screenplay Structure Recognition | Basic | Professional |

---

## RAG System Design

### 4.1 Vector Database Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           QDRANT CLUSTER                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    COLLECTION: screenplays                           │   │
│  ├─────────────────────────────────────────────────────────────────────┤   │
│  │  Vector Size: 1024 (Voyage-3 embeddings)                            │   │
│  │  Distance: Cosine                                                    │   │
│  │  Shards: 4                                                           │   │
│  │                                                                       │   │
│  │  Payload Schema:                                                      │   │
│  │  ├── title: string                                                   │   │
│  │  ├── author: string                                                  │   │
│  │  ├── genre: string[]                                                 │   │
│  │  ├── year: int                                                       │   │
│  │  ├── budget_tier: enum                                               │   │
│  │  ├── box_office: float                                               │   │
│  │  ├── awards: string[]                                                │   │
│  │  ├── chunk_type: enum (scene|act|full)                               │   │
│  │  ├── chunk_index: int                                                │   │
│  │  ├── page_range: string                                              │   │
│  │  └── content: string (raw text)                                      │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    COLLECTION: coverage_examples                      │   │
│  ├─────────────────────────────────────────────────────────────────────┤   │
│  │  Vector Size: 1024                                                   │   │
│  │  Distance: Cosine                                                    │   │
│  │  Shards: 2                                                           │   │
│  │                                                                       │   │
│  │  Payload Schema:                                                      │   │
│  │  ├── script_title: string                                            │   │
│  │  ├── coverage_type: enum (studio|contest|development)                │   │
│  │  ├── analyst: string                                                 │   │
│  │  ├── recommendation: enum (pass|consider|recommend)                  │   │
│  │  ├── overall_rating: float                                           │   │
│  │  ├── section: enum (logline|synopsis|character|dialogue|structure)   │   │
│  │  ├── quality_score: float                                            │   │
│  │  └── content: string                                                 │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    COLLECTION: industry_knowledge                     │   │
│  ├─────────────────────────────────────────────────────────────────────┤   │
│  │  Vector Size: 1024                                                   │   │
│  │  Distance: Cosine                                                    │   │
│  │  Shards: 2                                                           │   │
│  │                                                                       │   │
│  │  Payload Schema:                                                      │   │
│  │  ├── category: enum (market|casting|budget|festival|technique)       │   │
│  │  ├── source: string                                                  │   │
│  │  ├── date: datetime                                                  │   │
│  │  ├── relevance_genres: string[]                                      │   │
│  │  ├── budget_applicability: string[]                                  │   │
│  │  └── content: string                                                 │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    COLLECTION: character_archetypes                   │   │
│  ├─────────────────────────────────────────────────────────────────────┤   │
│  │  Vector Size: 1024                                                   │   │
│  │  Distance: Cosine                                                    │   │
│  │  Shards: 1                                                           │   │
│  │                                                                       │   │
│  │  Payload Schema:                                                      │   │
│  │  ├── archetype: string                                               │   │
│  │  ├── film_examples: string[]                                         │   │
│  │  ├── arc_patterns: string[]                                          │   │
│  │  ├── casting_comps: string[]                                         │   │
│  │  ├── genre_prevalence: map[string]float                              │   │
│  │  └── description: string                                             │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 4.2 Embedding Strategy

```python
# embeddings.py

from dataclasses import dataclass
from typing import List, Dict, Optional
import voyageai
from qdrant_client import QdrantClient
from qdrant_client.http import models

@dataclass
class EmbeddingConfig:
    model: str = "voyage-3"              # Best for long documents
    input_type: str = "document"          # or "query" for search
    truncation: bool = True

class ScreenplayEmbedder:
    """Handle embedding generation for screenplay content."""

    def __init__(self, voyage_api_key: str, qdrant_url: str):
        self.voyage = voyageai.Client(api_key=voyage_api_key)
        self.qdrant = QdrantClient(url=qdrant_url)

    def embed_screenplay(self, screenplay: Dict) -> List[Dict]:
        """Embed a screenplay in multiple chunk strategies."""
        chunks = []

        # Strategy 1: Full script (for overall similarity)
        full_text = screenplay["text"][:120000]  # Voyage limit
        chunks.append({
            "text": full_text,
            "chunk_type": "full",
            "chunk_index": 0,
            "metadata": screenplay["metadata"]
        })

        # Strategy 2: Scene-level chunks
        scenes = self._parse_scenes(screenplay["text"])
        for i, scene in enumerate(scenes):
            chunks.append({
                "text": scene["text"],
                "chunk_type": "scene",
                "chunk_index": i,
                "metadata": {
                    **screenplay["metadata"],
                    "scene_heading": scene["heading"],
                    "page_range": scene["pages"]
                }
            })

        # Strategy 3: Act-level chunks
        acts = self._parse_acts(screenplay["text"])
        for i, act in enumerate(acts):
            chunks.append({
                "text": act["text"],
                "chunk_type": "act",
                "chunk_index": i,
                "metadata": {
                    **screenplay["metadata"],
                    "act_number": i + 1
                }
            })

        # Generate embeddings
        texts = [c["text"] for c in chunks]
        embeddings = self.voyage.embed(
            texts,
            model="voyage-3",
            input_type="document"
        ).embeddings

        # Attach embeddings
        for chunk, embedding in zip(chunks, embeddings):
            chunk["embedding"] = embedding

        return chunks

    def _parse_scenes(self, text: str) -> List[Dict]:
        """Parse screenplay into scenes."""
        import re

        # Match scene headings: INT./EXT. LOCATION - TIME
        scene_pattern = r'^(INT\.|EXT\.|INT\./EXT\.|I/E\.)[\s\S]*?(?=^(?:INT\.|EXT\.|INT\./EXT\.|I/E\.)|$)'

        scenes = []
        for i, match in enumerate(re.finditer(scene_pattern, text, re.MULTILINE)):
            scene_text = match.group(0).strip()
            heading = scene_text.split('\n')[0]
            scenes.append({
                "text": scene_text,
                "heading": heading,
                "pages": f"{i*1.5:.0f}-{(i+1)*1.5:.0f}"  # Estimate
            })

        return scenes

    def _parse_acts(self, text: str) -> List[Dict]:
        """Parse screenplay into acts (approximate)."""
        # Use page-based heuristic: Act 1 = 1-30, Act 2 = 30-90, Act 3 = 90-120
        lines = text.split('\n')
        total_lines = len(lines)

        act_breaks = [
            (0, int(total_lines * 0.25)),           # Act 1: 25%
            (int(total_lines * 0.25), int(total_lines * 0.75)),  # Act 2: 50%
            (int(total_lines * 0.75), total_lines)  # Act 3: 25%
        ]

        acts = []
        for i, (start, end) in enumerate(act_breaks):
            acts.append({
                "text": '\n'.join(lines[start:end]),
                "act_number": i + 1
            })

        return acts

    def upsert_to_qdrant(self, collection: str, chunks: List[Dict]):
        """Upsert embedded chunks to Qdrant."""
        points = [
            models.PointStruct(
                id=f"{chunk['metadata']['title']}_{chunk['chunk_type']}_{chunk['chunk_index']}",
                vector=chunk["embedding"],
                payload={
                    **chunk["metadata"],
                    "chunk_type": chunk["chunk_type"],
                    "chunk_index": chunk["chunk_index"],
                    "content": chunk["text"][:10000]  # Truncate for payload
                }
            )
            for chunk in chunks
        ]

        self.qdrant.upsert(
            collection_name=collection,
            points=points
        )
```

### 4.3 Retrieval Strategy

```python
# retrieval.py

from dataclasses import dataclass
from typing import List, Dict, Optional
from qdrant_client import QdrantClient
from qdrant_client.http import models
import voyageai

@dataclass
class RetrievalConfig:
    screenplay_top_k: int = 5
    coverage_top_k: int = 10
    industry_top_k: int = 5
    character_top_k: int = 3

    # Reranking
    use_reranker: bool = True
    rerank_model: str = "rerank-2"

    # Filtering
    genre_weight: float = 0.3
    budget_weight: float = 0.2

class ScreenplayRetriever:
    """Multi-collection retrieval for screenplay analysis."""

    def __init__(
        self,
        qdrant_url: str,
        voyage_api_key: str,
        config: RetrievalConfig = RetrievalConfig()
    ):
        self.qdrant = QdrantClient(url=qdrant_url)
        self.voyage = voyageai.Client(api_key=voyage_api_key)
        self.config = config

    def retrieve_context(
        self,
        query_screenplay: str,
        genre: List[str],
        budget_tier: str,
        analysis_type: str = "full_coverage"
    ) -> Dict:
        """Retrieve relevant context for screenplay analysis."""

        # Generate query embedding
        query_embedding = self.voyage.embed(
            [query_screenplay[:10000]],  # Use beginning for similarity
            model="voyage-3",
            input_type="query"
        ).embeddings[0]

        context = {}

        # 1. Similar screenplays
        context["similar_scripts"] = self._search_screenplays(
            query_embedding, genre, budget_tier
        )

        # 2. Coverage examples
        context["coverage_examples"] = self._search_coverage(
            query_embedding, analysis_type
        )

        # 3. Industry knowledge
        context["industry_context"] = self._search_industry(
            query_embedding, genre, budget_tier
        )

        # 4. Character archetypes (for character analysis)
        if analysis_type in ["full_coverage", "character_analysis"]:
            context["character_archetypes"] = self._search_archetypes(
                query_embedding
            )

        # Rerank all context
        if self.config.use_reranker:
            context = self._rerank_context(query_screenplay, context)

        return context

    def _search_screenplays(
        self,
        embedding: List[float],
        genre: List[str],
        budget_tier: str
    ) -> List[Dict]:
        """Search for similar screenplays."""

        # Build filter
        filter_conditions = []
        if genre:
            filter_conditions.append(
                models.FieldCondition(
                    key="genre",
                    match=models.MatchAny(any=genre)
                )
            )
        if budget_tier:
            filter_conditions.append(
                models.FieldCondition(
                    key="budget_tier",
                    match=models.MatchValue(value=budget_tier)
                )
            )

        query_filter = models.Filter(
            should=filter_conditions
        ) if filter_conditions else None

        results = self.qdrant.search(
            collection_name="screenplays",
            query_vector=embedding,
            query_filter=query_filter,
            limit=self.config.screenplay_top_k,
            with_payload=True
        )

        return [
            {
                "title": r.payload.get("title"),
                "genre": r.payload.get("genre"),
                "content": r.payload.get("content"),
                "score": r.score
            }
            for r in results
        ]

    def _search_coverage(
        self,
        embedding: List[float],
        analysis_type: str
    ) -> List[Dict]:
        """Search for relevant coverage examples."""

        # Map analysis type to coverage sections
        section_map = {
            "full_coverage": None,  # All sections
            "character_analysis": "character",
            "structural_analysis": "structure",
            "dialogue_assessment": "dialogue",
            "market_positioning": "market"
        }

        section = section_map.get(analysis_type)

        query_filter = None
        if section:
            query_filter = models.Filter(
                must=[
                    models.FieldCondition(
                        key="section",
                        match=models.MatchValue(value=section)
                    )
                ]
            )

        results = self.qdrant.search(
            collection_name="coverage_examples",
            query_vector=embedding,
            query_filter=query_filter,
            limit=self.config.coverage_top_k,
            with_payload=True
        )

        return [
            {
                "script_title": r.payload.get("script_title"),
                "section": r.payload.get("section"),
                "content": r.payload.get("content"),
                "quality_score": r.payload.get("quality_score"),
                "score": r.score
            }
            for r in results
        ]

    def _search_industry(
        self,
        embedding: List[float],
        genre: List[str],
        budget_tier: str
    ) -> List[Dict]:
        """Search for relevant industry knowledge."""

        filter_conditions = []
        if genre:
            filter_conditions.append(
                models.FieldCondition(
                    key="relevance_genres",
                    match=models.MatchAny(any=genre)
                )
            )
        if budget_tier:
            filter_conditions.append(
                models.FieldCondition(
                    key="budget_applicability",
                    match=models.MatchAny(any=[budget_tier])
                )
            )

        query_filter = models.Filter(
            should=filter_conditions
        ) if filter_conditions else None

        results = self.qdrant.search(
            collection_name="industry_knowledge",
            query_vector=embedding,
            query_filter=query_filter,
            limit=self.config.industry_top_k,
            with_payload=True
        )

        return [
            {
                "category": r.payload.get("category"),
                "content": r.payload.get("content"),
                "source": r.payload.get("source"),
                "score": r.score
            }
            for r in results
        ]

    def _search_archetypes(self, embedding: List[float]) -> List[Dict]:
        """Search for relevant character archetypes."""

        results = self.qdrant.search(
            collection_name="character_archetypes",
            query_vector=embedding,
            limit=self.config.character_top_k,
            with_payload=True
        )

        return [
            {
                "archetype": r.payload.get("archetype"),
                "description": r.payload.get("description"),
                "film_examples": r.payload.get("film_examples"),
                "casting_comps": r.payload.get("casting_comps"),
                "score": r.score
            }
            for r in results
        ]

    def _rerank_context(
        self,
        query: str,
        context: Dict
    ) -> Dict:
        """Rerank retrieved context using Voyage reranker."""

        reranked = {}

        for key, items in context.items():
            if not items:
                reranked[key] = items
                continue

            documents = [item.get("content", str(item)) for item in items]

            rerank_result = self.voyage.rerank(
                query=query[:5000],
                documents=documents,
                model=self.config.rerank_model,
                top_k=len(documents)
            )

            # Reorder items based on rerank scores
            reranked_items = []
            for result in rerank_result.results:
                item = items[result.index].copy()
                item["rerank_score"] = result.relevance_score
                reranked_items.append(item)

            reranked[key] = reranked_items

        return reranked
```

### 4.4 Context Assembly

```python
# context_builder.py

from typing import Dict, List
from dataclasses import dataclass

@dataclass
class ContextConfig:
    max_similar_scripts: int = 3
    max_coverage_examples: int = 5
    max_industry_items: int = 3
    max_archetypes: int = 2

    # Token budgets (approximate)
    screenplay_budget: int = 50000      # Main screenplay
    similar_budget: int = 10000         # Similar scripts
    coverage_budget: int = 15000        # Coverage examples
    industry_budget: int = 5000         # Industry context
    archetype_budget: int = 3000        # Character archetypes
    instruction_budget: int = 2000      # System prompt

class ContextBuilder:
    """Assemble context for LLM prompt."""

    def __init__(self, config: ContextConfig = ContextConfig()):
        self.config = config

    def build_prompt(
        self,
        screenplay: str,
        retrieved_context: Dict,
        analysis_type: str,
        custom_instructions: str = ""
    ) -> str:
        """Build complete prompt with RAG context."""

        sections = []

        # 1. System instruction
        sections.append(self._build_system_section(analysis_type, custom_instructions))

        # 2. Similar screenplay examples
        sections.append(self._build_similar_section(
            retrieved_context.get("similar_scripts", [])
        ))

        # 3. Coverage format examples
        sections.append(self._build_coverage_section(
            retrieved_context.get("coverage_examples", [])
        ))

        # 4. Industry context
        sections.append(self._build_industry_section(
            retrieved_context.get("industry_context", [])
        ))

        # 5. Character archetypes
        if "character_archetypes" in retrieved_context:
            sections.append(self._build_archetype_section(
                retrieved_context["character_archetypes"]
            ))

        # 6. Main screenplay
        sections.append(self._build_screenplay_section(screenplay))

        # 7. Task instruction
        sections.append(self._build_task_section(analysis_type))

        return "\n\n".join(filter(None, sections))

    def _build_system_section(self, analysis_type: str, custom: str) -> str:
        """Build system instruction."""

        base_instruction = """You are an expert screenplay analyst with decades of experience in Hollywood studios.
You provide insightful, industry-standard coverage that helps development executives make informed decisions.
Your analysis is thorough yet concise, professional yet accessible.

Key principles:
- Be specific with examples from the script
- Consider commercial viability alongside artistic merit
- Identify both strengths and areas for development constructively
- Provide actionable next steps for the writer
- Use industry-standard terminology and formats"""

        type_instructions = {
            "full_coverage": """
For this full coverage analysis:
- Provide a compelling logline (1 sentence)
- Write a clear synopsis (500-700 words covering three acts)
- Rate each category 1-10 with brief justification
- Identify 3-5 key strengths with specific examples
- Identify 3-5 areas for development with constructive suggestions
- Provide clear recommendation (PASS/CONSIDER/RECOMMEND) with rationale""",

            "character_analysis": """
For this character analysis:
- Analyze each major character's arc, motivation, and development
- Assess voice consistency and dialogue quality per character
- Identify casting considerations and comparable roles
- Evaluate character relationships and dynamics
- Suggest specific improvements for character development""",

            "structural_analysis": """
For this structural analysis:
- Identify act breaks and key turning points
- Assess pacing through each act
- Evaluate inciting incident, midpoint, climax, and resolution
- Identify any structural issues or imbalances
- Compare to relevant structural paradigms (Save the Cat, Hero's Journey, etc.)"""
        }

        return f"""<system>
{base_instruction}
{type_instructions.get(analysis_type, "")}
{custom}
</system>"""

    def _build_similar_section(self, similar_scripts: List[Dict]) -> str:
        """Build section with similar screenplay examples."""
        if not similar_scripts:
            return ""

        items = similar_scripts[:self.config.max_similar_scripts]

        content = ["<similar_screenplays>"]
        content.append("Reference these similar produced screenplays for context:")

        for script in items:
            content.append(f"""
<script title="{script['title']}" genre="{','.join(script.get('genre', []))}">
{script.get('content', '')[:3000]}...
</script>""")

        content.append("</similar_screenplays>")
        return "\n".join(content)

    def _build_coverage_section(self, coverage_examples: List[Dict]) -> str:
        """Build section with coverage format examples."""
        if not coverage_examples:
            return ""

        items = coverage_examples[:self.config.max_coverage_examples]

        content = ["<coverage_examples>"]
        content.append("Use these as format and quality references:")

        for example in items:
            content.append(f"""
<example script="{example.get('script_title', 'Unknown')}" section="{example.get('section', 'general')}" quality="{example.get('quality_score', 'N/A')}">
{example.get('content', '')[:2500]}
</example>""")

        content.append("</coverage_examples>")
        return "\n".join(content)

    def _build_industry_section(self, industry_items: List[Dict]) -> str:
        """Build section with industry context."""
        if not industry_items:
            return ""

        items = industry_items[:self.config.max_industry_items]

        content = ["<industry_context>"]
        content.append("Consider this market and industry context:")

        for item in items:
            content.append(f"""
<context category="{item.get('category', 'general')}">
{item.get('content', '')[:1500]}
</context>""")

        content.append("</industry_context>")
        return "\n".join(content)

    def _build_archetype_section(self, archetypes: List[Dict]) -> str:
        """Build section with character archetype references."""
        if not archetypes:
            return ""

        items = archetypes[:self.config.max_archetypes]

        content = ["<character_archetypes>"]
        content.append("Reference character archetypes for analysis:")

        for archetype in items:
            content.append(f"""
<archetype name="{archetype.get('archetype', 'Unknown')}">
Description: {archetype.get('description', '')}
Film Examples: {', '.join(archetype.get('film_examples', [])[:3])}
Casting Comparables: {', '.join(archetype.get('casting_comps', [])[:3])}
</archetype>""")

        content.append("</character_archetypes>")
        return "\n".join(content)

    def _build_screenplay_section(self, screenplay: str) -> str:
        """Build section with the main screenplay."""

        # Truncate if needed (reserve space for other sections)
        max_chars = self.config.screenplay_budget * 4  # ~4 chars per token
        truncated = screenplay[:max_chars]

        return f"""<screenplay_to_analyze>
{truncated}
</screenplay_to_analyze>"""

    def _build_task_section(self, analysis_type: str) -> str:
        """Build the task instruction section."""

        task_prompts = {
            "full_coverage": """<task>
Provide complete professional coverage for the screenplay above.

Output your analysis in the following XML format:
<coverage>
  <logline>One sentence hook</logline>
  <synopsis>
    <act_one>Setup and inciting incident...</act_one>
    <act_two>Confrontation and complications...</act_two>
    <act_three>Climax and resolution...</act_three>
  </synopsis>
  <ratings>
    <premise score="X">Brief justification</premise>
    <character score="X">Brief justification</character>
    <dialogue score="X">Brief justification</dialogue>
    <structure score="X">Brief justification</structure>
    <pacing score="X">Brief justification</pacing>
    <marketability score="X">Brief justification</marketability>
    <originality score="X">Brief justification</originality>
    <execution score="X">Brief justification</execution>
  </ratings>
  <strengths>
    <strength>Specific strength with example</strength>
    ...
  </strengths>
  <areas_for_development>
    <area>Specific issue with constructive suggestion</area>
    ...
  </areas_for_development>
  <recommendation verdict="PASS|CONSIDER|RECOMMEND" confidence="X">
    Recommendation summary and next steps...
  </recommendation>
</coverage>
</task>""",

            "character_analysis": """<task>
Analyze all significant characters in this screenplay.

Output your analysis in the following XML format:
<character_analysis>
  <character name="CHARACTER NAME" role="lead|supporting|minor">
    <arc>Character arc description</arc>
    <motivation>Core motivation</motivation>
    <voice>Dialogue voice characteristics</voice>
    <development>How character develops through story</development>
    <strengths>What works well</strengths>
    <opportunities>Areas for improvement</opportunities>
    <casting_notes>Casting considerations and comparables</casting_notes>
  </character>
  ...
</character_analysis>
</task>""",

            "structural_analysis": """<task>
Analyze the structural elements of this screenplay.

Output your analysis in the following XML format:
<structural_analysis>
  <act_structure type="3-act|4-act|5-act|non-traditional">
    <act number="1" pages="X-Y">
      <key_events>Major events in this act</key_events>
      <turning_point>How act ends/transitions</turning_point>
      <pacing_notes>Pacing observations</pacing_notes>
    </act>
    ...
  </act_structure>
  <key_moments>
    <inciting_incident page="X">Description</inciting_incident>
    <midpoint page="X">Description</midpoint>
    <climax page="X">Description</climax>
  </key_moments>
  <strengths>Structural strengths</strengths>
  <issues>Structural issues with suggestions</issues>
</structural_analysis>
</task>"""
        }

        return task_prompts.get(analysis_type, task_prompts["full_coverage"])
```

---

## Data Models

### 5.1 Go Data Structures (CLI Integration)

```go
// internal/screenplay/types.go

package screenplay

import "time"

// ScreenplayAnalysisRequest represents a request to analyze a screenplay
type ScreenplayAnalysisRequest struct {
    ID               string            `json:"id" yaml:"id"`
    ScreenplayPath   string            `json:"screenplay_path" yaml:"screenplay_path"`
    ScreenplayText   string            `json:"screenplay_text,omitempty" yaml:"screenplay_text,omitempty"`
    AnalysisType     AnalysisType      `json:"analysis_type" yaml:"analysis_type"`
    Genre            []string          `json:"genre,omitempty" yaml:"genre,omitempty"`
    BudgetTier       string            `json:"budget_tier,omitempty" yaml:"budget_tier,omitempty"`
    CustomPrompt     string            `json:"custom_prompt,omitempty" yaml:"custom_prompt,omitempty"`
    IncludeRAG       bool              `json:"include_rag" yaml:"include_rag"`
    OutputFormat     OutputFormat      `json:"output_format" yaml:"output_format"`
    Metadata         map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type AnalysisType string

const (
    AnalysisTypeFull       AnalysisType = "full_coverage"
    AnalysisTypeCharacter  AnalysisType = "character_analysis"
    AnalysisTypeStructure  AnalysisType = "structural_analysis"
    AnalysisTypeDialogue   AnalysisType = "dialogue_assessment"
    AnalysisTypeMarket     AnalysisType = "market_positioning"
    AnalysisTypeScene      AnalysisType = "scene_breakdown"
)

type OutputFormat string

const (
    OutputFormatJSON     OutputFormat = "json"
    OutputFormatMarkdown OutputFormat = "markdown"
    OutputFormatXML      OutputFormat = "xml"
    OutputFormatPDF      OutputFormat = "pdf"
)

// CoverageReport represents the full analysis output
type CoverageReport struct {
    ID                  string              `json:"id" yaml:"id"`
    Title               string              `json:"title" yaml:"title"`
    Author              string              `json:"author,omitempty" yaml:"author,omitempty"`
    AnalyzedAt          time.Time           `json:"analyzed_at" yaml:"analyzed_at"`
    AnalysisType        AnalysisType        `json:"analysis_type" yaml:"analysis_type"`
    ModelVersion        string              `json:"model_version" yaml:"model_version"`

    // Core coverage
    Logline             string              `json:"logline" yaml:"logline"`
    Synopsis            Synopsis            `json:"synopsis" yaml:"synopsis"`
    Ratings             Ratings             `json:"ratings" yaml:"ratings"`
    Strengths           []string            `json:"strengths" yaml:"strengths"`
    AreasForDevelopment []string            `json:"areas_for_development" yaml:"areas_for_development"`
    Recommendation      Recommendation      `json:"recommendation" yaml:"recommendation"`

    // Extended analysis
    CharacterAnalyses   []CharacterAnalysis `json:"character_analyses,omitempty" yaml:"character_analyses,omitempty"`
    StructuralAnalysis  *StructuralAnalysis `json:"structural_analysis,omitempty" yaml:"structural_analysis,omitempty"`
    DialogueAnalysis    *DialogueAnalysis   `json:"dialogue_analysis,omitempty" yaml:"dialogue_analysis,omitempty"`
    SceneAnalyses       []SceneAnalysis     `json:"scene_analyses,omitempty" yaml:"scene_analyses,omitempty"`
    IndustryIntel       *IndustryIntel      `json:"industry_intelligence,omitempty" yaml:"industry_intelligence,omitempty"`

    // Metadata
    Genre               []string            `json:"genre" yaml:"genre"`
    BudgetEstimate      string              `json:"budget_estimate,omitempty" yaml:"budget_estimate,omitempty"`
    RAGContextUsed      bool                `json:"rag_context_used" yaml:"rag_context_used"`
    ProcessingTimeMs    int64               `json:"processing_time_ms" yaml:"processing_time_ms"`
}

type Synopsis struct {
    ActOne   string `json:"act_one" yaml:"act_one"`
    ActTwo   string `json:"act_two" yaml:"act_two"`
    ActThree string `json:"act_three" yaml:"act_three"`
    Full     string `json:"full,omitempty" yaml:"full,omitempty"`
}

type Ratings struct {
    Overall       float64 `json:"overall" yaml:"overall"`
    Premise       Rating  `json:"premise" yaml:"premise"`
    Character     Rating  `json:"character" yaml:"character"`
    Dialogue      Rating  `json:"dialogue" yaml:"dialogue"`
    Structure     Rating  `json:"structure" yaml:"structure"`
    Pacing        Rating  `json:"pacing" yaml:"pacing"`
    Marketability Rating  `json:"marketability" yaml:"marketability"`
    Originality   Rating  `json:"originality" yaml:"originality"`
    Execution     Rating  `json:"execution" yaml:"execution"`
}

type Rating struct {
    Score         float64 `json:"score" yaml:"score"`
    Justification string  `json:"justification" yaml:"justification"`
}

type Recommendation struct {
    Verdict    string   `json:"verdict" yaml:"verdict"` // PASS, CONSIDER, RECOMMEND
    Confidence float64  `json:"confidence" yaml:"confidence"`
    Summary    string   `json:"summary" yaml:"summary"`
    NextSteps  []string `json:"next_steps" yaml:"next_steps"`
}

type CharacterAnalysis struct {
    Name                 string   `json:"name" yaml:"name"`
    Role                 string   `json:"role" yaml:"role"` // lead, supporting, minor
    ScreenTimePercent    float64  `json:"screen_time_percent,omitempty" yaml:"screen_time_percent,omitempty"`
    ArcDescription       string   `json:"arc_description" yaml:"arc_description"`
    Motivation           string   `json:"motivation" yaml:"motivation"`
    VoiceDescription     string   `json:"voice_description" yaml:"voice_description"`
    VoiceEvolution       VoiceEvo `json:"voice_evolution,omitempty" yaml:"voice_evolution,omitempty"`
    Strengths            []string `json:"strengths" yaml:"strengths"`
    Opportunities        []string `json:"opportunities" yaml:"opportunities"`
    CastingRecommendations []string `json:"casting_recommendations,omitempty" yaml:"casting_recommendations,omitempty"`
}

type VoiceEvo struct {
    ActOne   []string `json:"act_one" yaml:"act_one"`
    ActTwo   []string `json:"act_two" yaml:"act_two"`
    ActThree []string `json:"act_three" yaml:"act_three"`
}

type StructuralAnalysis struct {
    PageCount       int           `json:"page_count" yaml:"page_count"`
    ActStructure    string        `json:"act_structure" yaml:"act_structure"` // 3-act, 4-act, etc.
    ActBreakdowns   []ActAnalysis `json:"act_breakdowns" yaml:"act_breakdowns"`
    KeyMoments      KeyMoments    `json:"key_moments" yaml:"key_moments"`
    PacingRhythm    string        `json:"pacing_rhythm" yaml:"pacing_rhythm"`
    Strengths       []string      `json:"strengths" yaml:"strengths"`
    Issues          []string      `json:"issues" yaml:"issues"`
}

type ActAnalysis struct {
    ActNumber     int      `json:"act_number" yaml:"act_number"`
    PageRange     string   `json:"page_range" yaml:"page_range"`
    KeyEvents     []string `json:"key_events" yaml:"key_events"`
    TurningPoint  string   `json:"turning_point" yaml:"turning_point"`
    PacingNotes   string   `json:"pacing_notes" yaml:"pacing_notes"`
}

type KeyMoments struct {
    IncitingIncident Moment `json:"inciting_incident" yaml:"inciting_incident"`
    Midpoint         Moment `json:"midpoint" yaml:"midpoint"`
    Climax           Moment `json:"climax" yaml:"climax"`
    Resolution       Moment `json:"resolution,omitempty" yaml:"resolution,omitempty"`
}

type Moment struct {
    Page        int    `json:"page" yaml:"page"`
    Description string `json:"description" yaml:"description"`
}

type DialogueAnalysis struct {
    OverallQuality   string          `json:"overall_quality" yaml:"overall_quality"`
    CharacterVoices  []CharacterVoice `json:"character_voices" yaml:"character_voices"`
    SubtextExamples  []SubtextExample `json:"subtext_examples,omitempty" yaml:"subtext_examples,omitempty"`
    Strengths        []string        `json:"strengths" yaml:"strengths"`
    Improvements     []string        `json:"improvements" yaml:"improvements"`
}

type CharacterVoice struct {
    Character   string   `json:"character" yaml:"character"`
    Description string   `json:"description" yaml:"description"`
    Examples    []string `json:"examples" yaml:"examples"`
    Rating      float64  `json:"rating" yaml:"rating"`
}

type SubtextExample struct {
    Scene    int    `json:"scene" yaml:"scene"`
    Exchange string `json:"exchange" yaml:"exchange"`
    Surface  string `json:"surface" yaml:"surface"`
    Subtext  string `json:"subtext" yaml:"subtext"`
}

type SceneAnalysis struct {
    SceneNumber           int                   `json:"scene_number" yaml:"scene_number"`
    SceneHeading          string                `json:"scene_heading" yaml:"scene_heading"`
    PageStart             int                   `json:"page_start" yaml:"page_start"`
    PageEnd               int                   `json:"page_end" yaml:"page_end"`
    SceneFunction         string                `json:"scene_function" yaml:"scene_function"`
    DialogueQualityRating float64               `json:"dialogue_quality_rating" yaml:"dialogue_quality_rating"`
    CommercialAppeal      CommercialAppeal      `json:"commercial_appeal" yaml:"commercial_appeal"`
    ProductionNotes       ProductionNotes       `json:"production_notes" yaml:"production_notes"`
    Strengths             []string              `json:"strengths" yaml:"strengths"`
    Concerns              []string              `json:"concerns" yaml:"concerns"`
}

type CommercialAppeal struct {
    FestivalCircuit string `json:"festival_circuit" yaml:"festival_circuit"`
    ActorShowcase   string `json:"actor_showcase" yaml:"actor_showcase"`
    International   string `json:"international" yaml:"international"`
}

type ProductionNotes struct {
    BudgetImpact         string `json:"budget_impact" yaml:"budget_impact"` // LOW, MEDIUM, HIGH
    ShootingDays         string `json:"shooting_days" yaml:"shooting_days"`
    LocationRequirements string `json:"location_requirements" yaml:"location_requirements"`
    VFXRequirements      string `json:"vfx_requirements" yaml:"vfx_requirements"`
}

type IndustryIntel struct {
    CompTitles        []CompTitle `json:"comp_titles" yaml:"comp_titles"`
    MarketPosition    string      `json:"market_position" yaml:"market_position"`
    BudgetRange       string      `json:"budget_range" yaml:"budget_range"`
    RevenueProjection string      `json:"revenue_projection,omitempty" yaml:"revenue_projection,omitempty"`
    AwardsPotential   string      `json:"awards_potential" yaml:"awards_potential"` // VERY HIGH, HIGH, MODERATE, LOW
    FestivalStrategy  string      `json:"festival_strategy" yaml:"festival_strategy"`
    TargetDistributors []string   `json:"target_distributors,omitempty" yaml:"target_distributors,omitempty"`
    CastingTier       string      `json:"casting_tier" yaml:"casting_tier"`
}

type CompTitle struct {
    Title      string  `json:"title" yaml:"title"`
    Year       int     `json:"year" yaml:"year"`
    Similarity float64 `json:"similarity" yaml:"similarity"`
    Notes      string  `json:"notes" yaml:"notes"`
}
```

### 5.2 Vector Database Schema (Qdrant)

```python
# schemas/qdrant_schemas.py

from qdrant_client.http import models

COLLECTIONS = {
    "screenplays": {
        "vectors_config": models.VectorParams(
            size=1024,  # Voyage-3 dimension
            distance=models.Distance.COSINE,
        ),
        "optimizers_config": models.OptimizersConfigDiff(
            indexing_threshold=20000,
        ),
        "hnsw_config": models.HnswConfigDiff(
            m=16,
            ef_construct=100,
        ),
    },

    "coverage_examples": {
        "vectors_config": models.VectorParams(
            size=1024,
            distance=models.Distance.COSINE,
        ),
    },

    "industry_knowledge": {
        "vectors_config": models.VectorParams(
            size=1024,
            distance=models.Distance.COSINE,
        ),
    },

    "character_archetypes": {
        "vectors_config": models.VectorParams(
            size=1024,
            distance=models.Distance.COSINE,
        ),
    },
}

PAYLOAD_INDICES = {
    "screenplays": [
        models.PayloadSchemaType.KEYWORD,  # genre
        models.PayloadSchemaType.KEYWORD,  # budget_tier
        models.PayloadSchemaType.INTEGER,  # year
        models.PayloadSchemaType.KEYWORD,  # chunk_type
    ],

    "coverage_examples": [
        models.PayloadSchemaType.KEYWORD,  # coverage_type
        models.PayloadSchemaType.KEYWORD,  # section
        models.PayloadSchemaType.KEYWORD,  # recommendation
        models.PayloadSchemaType.FLOAT,    # quality_score
    ],

    "industry_knowledge": [
        models.PayloadSchemaType.KEYWORD,  # category
        models.PayloadSchemaType.KEYWORD,  # relevance_genres
        models.PayloadSchemaType.KEYWORD,  # budget_applicability
    ],
}
```

---

## Infrastructure Requirements

### 6.1 Training Infrastructure

| Component | Specification | Purpose |
|-----------|--------------|---------|
| **GPU** | 8x NVIDIA H100 80GB | QLoRA training of 70B model |
| **CPU** | 128 cores | Data preprocessing |
| **RAM** | 512GB | Dataset loading |
| **Storage** | 4TB NVMe SSD | Checkpoints + datasets |
| **Network** | 100Gbps InfiniBand | Multi-GPU communication |

**Estimated Training Time:**
- 10,000 examples, 3 epochs: ~24-48 hours
- Cost estimate: ~$2,000-4,000 on Lambda Labs

### 6.2 Inference Infrastructure

| Component | Specification | Purpose |
|-----------|--------------|---------|
| **GPU** | 2x NVIDIA A100 80GB or 4x A6000 | vLLM inference |
| **CPU** | 32 cores | Request handling |
| **RAM** | 256GB | Context caching |
| **Storage** | 1TB NVMe | Model weights + cache |

**Alternative: Cloud Inference**
- Lambda Labs: $1.10/hr per A100
- RunPod: $0.79/hr per A100
- Together.ai: API pricing for fine-tuned models

### 6.3 RAG Infrastructure

| Component | Specification | Purpose |
|-----------|--------------|---------|
| **Qdrant** | 3-node cluster, 64GB RAM each | Vector storage |
| **Redis** | 32GB RAM | Analysis caching |
| **Storage** | 500GB SSD | Qdrant persistence |

### 6.4 Complete System Architecture

```yaml
# docker-compose.yml

version: '3.8'

services:
  # Inference Server
  llama-inference:
    image: vllm/vllm-openai:latest
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 2
              capabilities: [gpu]
    volumes:
      - ./models:/models
    environment:
      - MODEL_NAME=/models/llama-3.3-70b-screenplay
      - MAX_MODEL_LEN=65536
      - GPU_MEMORY_UTILIZATION=0.9
      - TENSOR_PARALLEL_SIZE=2
    ports:
      - "8000:8000"
    command: >
      --model /models/llama-3.3-70b-screenplay
      --tensor-parallel-size 2
      --max-model-len 65536
      --trust-remote-code

  # Vector Database
  qdrant:
    image: qdrant/qdrant:latest
    volumes:
      - ./qdrant_storage:/qdrant/storage
    ports:
      - "6333:6333"
      - "6334:6334"
    environment:
      - QDRANT__SERVICE__GRPC_PORT=6334

  # Analysis Gateway
  gateway:
    build:
      context: ./gateway
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - LLAMA_ENDPOINT=http://llama-inference:8000
      - QDRANT_URL=http://qdrant:6333
      - VOYAGE_API_KEY=${VOYAGE_API_KEY}
      - REDIS_URL=redis://redis:6379
    depends_on:
      - llama-inference
      - qdrant
      - redis

  # Cache
  redis:
    image: redis:7-alpine
    volumes:
      - ./redis_data:/data
    ports:
      - "6379:6379"

  # Embedding Service (for batch processing)
  embedder:
    build:
      context: ./embedder
      dockerfile: Dockerfile
    environment:
      - VOYAGE_API_KEY=${VOYAGE_API_KEY}
      - QDRANT_URL=http://qdrant:6333
    depends_on:
      - qdrant
```

---

## API Specifications

### 7.1 Analysis Gateway API (Go)

```go
// internal/screenplay/api/handlers.go

package api

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/joshkornreich/anime/internal/screenplay"
)

// AnalyzeHandler handles screenplay analysis requests
func (s *Server) AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
    var req screenplay.ScreenplayAnalysisRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate request
    if req.ScreenplayPath == "" && req.ScreenplayText == "" {
        http.Error(w, "Either screenplay_path or screenplay_text required", http.StatusBadRequest)
        return
    }

    // Start analysis pipeline
    ctx := r.Context()
    startTime := time.Now()

    // 1. Load/parse screenplay
    text, err := s.loadScreenplay(ctx, req)
    if err != nil {
        http.Error(w, "Failed to load screenplay: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // 2. Retrieve RAG context
    var ragContext map[string]interface{}
    if req.IncludeRAG {
        ragContext, err = s.retriever.RetrieveContext(ctx, text, req.Genre, req.BudgetTier, string(req.AnalysisType))
        if err != nil {
            // Log but continue without RAG
            s.logger.Warn("RAG retrieval failed", "error", err)
        }
    }

    // 3. Build prompt
    prompt := s.contextBuilder.BuildPrompt(text, ragContext, string(req.AnalysisType), req.CustomPrompt)

    // 4. Call LLM
    response, err := s.llmClient.Complete(ctx, prompt)
    if err != nil {
        http.Error(w, "LLM inference failed: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // 5. Parse response into structured format
    report, err := s.parseResponse(response, req.AnalysisType)
    if err != nil {
        http.Error(w, "Failed to parse response: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // 6. Add metadata
    report.ProcessingTimeMs = time.Since(startTime).Milliseconds()
    report.RAGContextUsed = req.IncludeRAG && ragContext != nil
    report.AnalysisType = req.AnalysisType
    report.ModelVersion = s.modelVersion

    // 7. Cache result
    if err := s.cache.Set(ctx, req.ID, report, 24*time.Hour); err != nil {
        s.logger.Warn("Failed to cache result", "error", err)
    }

    // 8. Return response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(report)
}

// StreamAnalyzeHandler handles streaming analysis
func (s *Server) StreamAnalyzeHandler(w http.ResponseWriter, r *http.Request) {
    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming not supported", http.StatusInternalServerError)
        return
    }

    var req screenplay.ScreenplayAnalysisRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    ctx := r.Context()

    // Load screenplay
    text, err := s.loadScreenplay(ctx, req)
    if err != nil {
        sendSSE(w, flusher, "error", map[string]string{"message": err.Error()})
        return
    }

    sendSSE(w, flusher, "status", map[string]string{"stage": "loaded_screenplay"})

    // Retrieve RAG context
    var ragContext map[string]interface{}
    if req.IncludeRAG {
        ragContext, _ = s.retriever.RetrieveContext(ctx, text, req.Genre, req.BudgetTier, string(req.AnalysisType))
        sendSSE(w, flusher, "status", map[string]string{"stage": "retrieved_context"})
    }

    // Build prompt
    prompt := s.contextBuilder.BuildPrompt(text, ragContext, string(req.AnalysisType), req.CustomPrompt)
    sendSSE(w, flusher, "status", map[string]string{"stage": "built_prompt"})

    // Stream LLM response
    stream, err := s.llmClient.StreamComplete(ctx, prompt)
    if err != nil {
        sendSSE(w, flusher, "error", map[string]string{"message": err.Error()})
        return
    }

    var fullResponse strings.Builder
    for chunk := range stream {
        fullResponse.WriteString(chunk)
        sendSSE(w, flusher, "chunk", map[string]string{"text": chunk})
    }

    // Parse and send final report
    report, err := s.parseResponse(fullResponse.String(), req.AnalysisType)
    if err != nil {
        sendSSE(w, flusher, "error", map[string]string{"message": err.Error()})
        return
    }

    sendSSE(w, flusher, "complete", report)
}

func sendSSE(w http.ResponseWriter, flusher http.Flusher, event string, data interface{}) {
    jsonData, _ := json.Marshal(data)
    fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, jsonData)
    flusher.Flush()
}
```

### 7.2 CLI Commands

```go
// cmd/analyze.go

package cmd

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "time"

    "github.com/spf13/cobra"
    "github.com/joshkornreich/anime/internal/screenplay"
    "github.com/joshkornreich/anime/internal/screenplay/api"
)

var analyzeCmd = &cobra.Command{
    Use:   "analyze [screenplay.fountain]",
    Short: "Analyze a screenplay using AI",
    Long: `Analyze a screenplay using the fine-tuned Llama 3.3 70B model.

Provides professional studio-quality coverage including:
- Logline and synopsis
- Ratings across 8 dimensions
- Character analysis
- Structural analysis
- Industry positioning
- Recommendation (PASS/CONSIDER/RECOMMEND)

Examples:
  anime analyze script.fountain
  anime analyze script.fountain --type character_analysis
  anime analyze script.fountain --output coverage.json --format json
  anime analyze script.fountain --no-rag --stream`,
    Args: cobra.ExactArgs(1),
    RunE: runAnalyze,
}

var (
    analysisType   string
    outputPath     string
    outputFormat   string
    includeRAG     bool
    stream         bool
    genre          []string
    budgetTier     string
    customPrompt   string
)

func init() {
    rootCmd.AddCommand(analyzeCmd)

    analyzeCmd.Flags().StringVarP(&analysisType, "type", "t", "full_coverage",
        "Analysis type: full_coverage, character_analysis, structural_analysis, dialogue_assessment, market_positioning")
    analyzeCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path")
    analyzeCmd.Flags().StringVarP(&outputFormat, "format", "f", "markdown", "Output format: json, markdown, xml, pdf")
    analyzeCmd.Flags().BoolVar(&includeRAG, "rag", true, "Include RAG context from similar screenplays")
    analyzeCmd.Flags().BoolVar(&stream, "stream", false, "Stream output in real-time")
    analyzeCmd.Flags().StringSliceVar(&genre, "genre", nil, "Genre tags for context retrieval")
    analyzeCmd.Flags().StringVar(&budgetTier, "budget", "", "Budget tier: micro, low, medium, high, studio")
    analyzeCmd.Flags().StringVar(&customPrompt, "prompt", "", "Additional custom instructions")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
    scriptPath := args[0]

    // Validate file exists
    if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
        return fmt.Errorf("screenplay file not found: %s", scriptPath)
    }

    // Create request
    req := screenplay.ScreenplayAnalysisRequest{
        ID:             generateID(),
        ScreenplayPath: scriptPath,
        AnalysisType:   screenplay.AnalysisType(analysisType),
        Genre:          genre,
        BudgetTier:     budgetTier,
        CustomPrompt:   customPrompt,
        IncludeRAG:     includeRAG,
        OutputFormat:   screenplay.OutputFormat(outputFormat),
    }

    // Load config
    cfg := loadConfig()

    // Create client
    client := api.NewClient(cfg.AnalysisGatewayURL)

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()

    if stream {
        return streamAnalysis(ctx, client, req)
    }

    return runAnalysisSync(ctx, client, req)
}

func streamAnalysis(ctx context.Context, client *api.Client, req screenplay.ScreenplayAnalysisRequest) error {
    events, err := client.StreamAnalyze(ctx, req)
    if err != nil {
        return err
    }

    for event := range events {
        switch event.Type {
        case "status":
            fmt.Printf("📊 %s\n", event.Data["stage"])
        case "chunk":
            fmt.Print(event.Data["text"])
        case "complete":
            fmt.Println("\n\n✅ Analysis complete")
            return saveOutput(event.Report)
        case "error":
            return fmt.Errorf("analysis failed: %s", event.Data["message"])
        }
    }

    return nil
}

func runAnalysisSync(ctx context.Context, client *api.Client, req screenplay.ScreenplayAnalysisRequest) error {
    fmt.Println("🎬 Analyzing screenplay...")

    report, err := client.Analyze(ctx, req)
    if err != nil {
        return err
    }

    fmt.Printf("✅ Analysis complete in %dms\n\n", report.ProcessingTimeMs)

    return saveOutput(report)
}

func saveOutput(report *screenplay.CoverageReport) error {
    var output []byte
    var err error

    switch outputFormat {
    case "json":
        output, err = json.MarshalIndent(report, "", "  ")
    case "markdown":
        output = formatMarkdown(report)
    case "xml":
        output = formatXML(report)
    default:
        output, err = json.MarshalIndent(report, "", "  ")
    }

    if err != nil {
        return err
    }

    if outputPath != "" {
        return os.WriteFile(outputPath, output, 0644)
    }

    fmt.Println(string(output))
    return nil
}

func formatMarkdown(report *screenplay.CoverageReport) []byte {
    // Generate markdown coverage report
    var md strings.Builder

    md.WriteString(fmt.Sprintf("# Coverage Report: %s\n\n", report.Title))
    md.WriteString(fmt.Sprintf("**Analyzed:** %s\n\n", report.AnalyzedAt.Format("January 2, 2006")))

    md.WriteString("## Logline\n\n")
    md.WriteString(report.Logline + "\n\n")

    md.WriteString("## Synopsis\n\n")
    md.WriteString("### Act One\n" + report.Synopsis.ActOne + "\n\n")
    md.WriteString("### Act Two\n" + report.Synopsis.ActTwo + "\n\n")
    md.WriteString("### Act Three\n" + report.Synopsis.ActThree + "\n\n")

    md.WriteString("## Ratings\n\n")
    md.WriteString(fmt.Sprintf("| Category | Score | Notes |\n"))
    md.WriteString("|----------|-------|-------|\n")
    md.WriteString(fmt.Sprintf("| Overall | %.1f | |\n", report.Ratings.Overall))
    md.WriteString(fmt.Sprintf("| Premise | %.1f | %s |\n", report.Ratings.Premise.Score, report.Ratings.Premise.Justification))
    // ... continue for all ratings

    md.WriteString("\n## Strengths\n\n")
    for _, s := range report.Strengths {
        md.WriteString(fmt.Sprintf("- %s\n", s))
    }

    md.WriteString("\n## Areas for Development\n\n")
    for _, a := range report.AreasForDevelopment {
        md.WriteString(fmt.Sprintf("- %s\n", a))
    }

    md.WriteString("\n## Recommendation\n\n")
    md.WriteString(fmt.Sprintf("**%s** (Confidence: %.0f%%)\n\n", report.Recommendation.Verdict, report.Recommendation.Confidence*10))
    md.WriteString(report.Recommendation.Summary + "\n\n")

    if len(report.Recommendation.NextSteps) > 0 {
        md.WriteString("### Next Steps\n\n")
        for i, step := range report.Recommendation.NextSteps {
            md.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
        }
    }

    return []byte(md.String())
}
```

---

## Integration with Existing Systems

### 8.1 ANIME CLI Integration

```go
// cmd/embedded/sky/internal/analysis/analysis.go

package analysis

import (
    "context"

    "github.com/joshkornreich/anime/internal/screenplay"
    "github.com/joshkornreich/anime/internal/screenplay/api"
)

// ScreenplayAnalyzer integrates with the existing ANIME protocol system
type ScreenplayAnalyzer struct {
    client  *api.Client
    cache   *Cache
    config  *Config
}

// NewScreenplayAnalyzer creates analyzer with config from ~/.config/anime/
func NewScreenplayAnalyzer(cfg *Config) *ScreenplayAnalyzer {
    return &ScreenplayAnalyzer{
        client: api.NewClient(cfg.AnalysisGatewayURL),
        cache:  NewCache(cfg.CachePath),
        config: cfg,
    }
}

// AnalyzeWithProtocol runs analysis as part of a protocol phase
func (a *ScreenplayAnalyzer) AnalyzeWithProtocol(ctx context.Context, phase *protocol.Phase) error {
    // Extract screenplay path from phase config
    scriptPath := phase.Config["screenplay_path"]
    analysisType := phase.Config.Get("analysis_type", "full_coverage")

    req := screenplay.ScreenplayAnalysisRequest{
        ScreenplayPath: scriptPath,
        AnalysisType:   screenplay.AnalysisType(analysisType),
        IncludeRAG:     phase.Config.GetBool("include_rag", true),
        Genre:          phase.Config.GetStringSlice("genre"),
        BudgetTier:     phase.Config.Get("budget_tier", ""),
    }

    // Check cache first
    if cached, ok := a.cache.Get(req.ID); ok {
        phase.SetOutput("coverage_report", cached)
        return nil
    }

    // Run analysis
    report, err := a.client.Analyze(ctx, req)
    if err != nil {
        return err
    }

    // Store in phase output
    phase.SetOutput("coverage_report", report)

    // Cache for future use
    a.cache.Set(req.ID, report)

    return nil
}
```

### 8.2 Desktop App Integration (Tauri)

```rust
// src-tauri/src/screenplay.rs

use serde::{Deserialize, Serialize};
use tauri::State;
use reqwest::Client;

#[derive(Debug, Serialize, Deserialize)]
pub struct AnalysisRequest {
    pub screenplay_path: Option<String>,
    pub screenplay_text: Option<String>,
    pub analysis_type: String,
    pub include_rag: bool,
    pub genre: Vec<String>,
    pub budget_tier: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct CoverageReport {
    pub id: String,
    pub title: String,
    pub logline: String,
    pub synopsis: Synopsis,
    pub ratings: Ratings,
    pub strengths: Vec<String>,
    pub areas_for_development: Vec<String>,
    pub recommendation: Recommendation,
    pub processing_time_ms: i64,
    pub rag_context_used: bool,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Synopsis {
    pub act_one: String,
    pub act_two: String,
    pub act_three: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Ratings {
    pub overall: f64,
    pub premise: Rating,
    pub character: Rating,
    pub dialogue: Rating,
    pub structure: Rating,
    pub pacing: Rating,
    pub marketability: Rating,
    pub originality: Rating,
    pub execution: Rating,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Rating {
    pub score: f64,
    pub justification: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Recommendation {
    pub verdict: String,
    pub confidence: f64,
    pub summary: String,
    pub next_steps: Vec<String>,
}

pub struct AnalysisClient {
    http: Client,
    gateway_url: String,
}

impl AnalysisClient {
    pub fn new(gateway_url: String) -> Self {
        Self {
            http: Client::new(),
            gateway_url,
        }
    }

    pub async fn analyze(&self, req: AnalysisRequest) -> Result<CoverageReport, String> {
        let response = self.http
            .post(format!("{}/api/v1/analyze", self.gateway_url))
            .json(&req)
            .send()
            .await
            .map_err(|e| format!("Request failed: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("Analysis failed: {}", response.status()));
        }

        response.json()
            .await
            .map_err(|e| format!("Failed to parse response: {}", e))
    }
}

#[tauri::command]
pub async fn analyze_screenplay(
    request: AnalysisRequest,
    client: State<'_, AnalysisClient>,
) -> Result<CoverageReport, String> {
    client.analyze(request).await
}

#[tauri::command]
pub async fn analyze_screenplay_stream(
    request: AnalysisRequest,
    client: State<'_, AnalysisClient>,
    window: tauri::Window,
) -> Result<(), String> {
    // Stream events to frontend via Tauri events
    let events = client.stream_analyze(request).await?;

    for event in events {
        window.emit("analysis-event", event)
            .map_err(|e| format!("Failed to emit event: {}", e))?;
    }

    Ok(())
}
```

### 8.3 React Frontend Integration

```typescript
// src/hooks/useScreenplayAnalysis.ts

import { useState, useCallback } from 'react';
import { invoke } from '@tauri-apps/api/core';
import { listen } from '@tauri-apps/api/event';

interface AnalysisRequest {
  screenplay_path?: string;
  screenplay_text?: string;
  analysis_type: string;
  include_rag: boolean;
  genre: string[];
  budget_tier?: string;
}

interface CoverageReport {
  id: string;
  title: string;
  logline: string;
  synopsis: {
    act_one: string;
    act_two: string;
    act_three: string;
  };
  ratings: {
    overall: number;
    premise: { score: number; justification: string };
    character: { score: number; justification: string };
    dialogue: { score: number; justification: string };
    structure: { score: number; justification: string };
    pacing: { score: number; justification: string };
    marketability: { score: number; justification: string };
    originality: { score: number; justification: string };
    execution: { score: number; justification: string };
  };
  strengths: string[];
  areas_for_development: string[];
  recommendation: {
    verdict: 'PASS' | 'CONSIDER' | 'RECOMMEND';
    confidence: number;
    summary: string;
    next_steps: string[];
  };
  processing_time_ms: number;
  rag_context_used: boolean;
}

interface AnalysisState {
  status: 'idle' | 'loading' | 'streaming' | 'complete' | 'error';
  progress: string;
  streamedText: string;
  report: CoverageReport | null;
  error: string | null;
}

export function useScreenplayAnalysis() {
  const [state, setState] = useState<AnalysisState>({
    status: 'idle',
    progress: '',
    streamedText: '',
    report: null,
    error: null,
  });

  const analyze = useCallback(async (request: AnalysisRequest) => {
    setState({
      status: 'loading',
      progress: 'Starting analysis...',
      streamedText: '',
      report: null,
      error: null,
    });

    try {
      const report = await invoke<CoverageReport>('analyze_screenplay', {
        request,
      });

      setState((prev) => ({
        ...prev,
        status: 'complete',
        report,
      }));

      return report;
    } catch (error) {
      setState((prev) => ({
        ...prev,
        status: 'error',
        error: String(error),
      }));
      throw error;
    }
  }, []);

  const analyzeStream = useCallback(async (request: AnalysisRequest) => {
    setState({
      status: 'streaming',
      progress: 'Starting analysis...',
      streamedText: '',
      report: null,
      error: null,
    });

    // Listen for streaming events
    const unlisten = await listen<{
      type: string;
      data: any;
    }>('analysis-event', (event) => {
      const { type, data } = event.payload;

      switch (type) {
        case 'status':
          setState((prev) => ({
            ...prev,
            progress: data.stage,
          }));
          break;
        case 'chunk':
          setState((prev) => ({
            ...prev,
            streamedText: prev.streamedText + data.text,
          }));
          break;
        case 'complete':
          setState((prev) => ({
            ...prev,
            status: 'complete',
            report: data,
          }));
          break;
        case 'error':
          setState((prev) => ({
            ...prev,
            status: 'error',
            error: data.message,
          }));
          break;
      }
    });

    try {
      await invoke('analyze_screenplay_stream', { request });
    } catch (error) {
      setState((prev) => ({
        ...prev,
        status: 'error',
        error: String(error),
      }));
    } finally {
      unlisten();
    }
  }, []);

  const reset = useCallback(() => {
    setState({
      status: 'idle',
      progress: '',
      streamedText: '',
      report: null,
      error: null,
    });
  }, []);

  return {
    ...state,
    analyze,
    analyzeStream,
    reset,
  };
}
```

---

## Deployment Strategy

### 9.1 Phase 1: Local Development

```bash
# Development setup

# 1. Clone and setup
git clone https://github.com/joshkornreich/anime
cd anime

# 2. Setup Python environment for training
python -m venv venv
source venv/bin/activate
pip install -r requirements-training.txt

# 3. Start local services
docker-compose -f docker-compose.dev.yml up -d

# 4. Run training (on GPU machine)
python train.py --config finetune_config.yaml

# 5. Test inference locally
python test_inference.py --model ./checkpoints/llama-3.3-70b-screenplay
```

### 9.2 Phase 2: Cloud Deployment

```yaml
# kubernetes/deployment.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: screenplay-inference
spec:
  replicas: 2
  selector:
    matchLabels:
      app: screenplay-inference
  template:
    metadata:
      labels:
        app: screenplay-inference
    spec:
      containers:
      - name: vllm
        image: vllm/vllm-openai:latest
        resources:
          limits:
            nvidia.com/gpu: 2
        env:
        - name: MODEL_NAME
          value: "/models/llama-3.3-70b-screenplay"
        - name: MAX_MODEL_LEN
          value: "65536"
        volumeMounts:
        - name: model-storage
          mountPath: /models
      volumes:
      - name: model-storage
        persistentVolumeClaim:
          claimName: model-pvc
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: screenplay-gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: screenplay-gateway
  template:
    spec:
      containers:
      - name: gateway
        image: anime/screenplay-gateway:latest
        env:
        - name: LLAMA_ENDPOINT
          value: "http://screenplay-inference:8000"
        - name: QDRANT_URL
          value: "http://qdrant:6333"
        - name: VOYAGE_API_KEY
          valueFrom:
            secretKeyRef:
              name: api-keys
              key: voyage
```

### 9.3 Phase 3: Production

| Component | Provider | Specs | Monthly Cost |
|-----------|----------|-------|--------------|
| Inference | Lambda Labs | 2x H100 | ~$6,000 |
| Qdrant | Qdrant Cloud | 3-node cluster | ~$500 |
| Gateway | AWS EKS | 3x c6i.xlarge | ~$400 |
| Redis | AWS ElastiCache | r6g.large | ~$200 |
| **Total** | | | **~$7,100** |

---

## Evaluation & Metrics

### 10.1 Quality Metrics

```python
# evaluation/metrics.py

from dataclasses import dataclass
from typing import List, Dict
import numpy as np

@dataclass
class EvaluationResult:
    coverage_quality: float      # CML-Bench style composite
    format_adherence: float      # Correct output structure
    character_consistency: float # CC dimension
    dialogue_coherence: float    # DC dimension
    plot_reasonableness: float   # PR dimension
    industry_accuracy: float     # Correct terminology, budget estimates
    human_parity: float          # Blind comparison vs human coverage

class ScreenplayEvaluator:
    """Evaluate fine-tuned model performance."""

    def evaluate_coverage(
        self,
        generated: Dict,
        reference: Dict,
        screenplay: str
    ) -> EvaluationResult:
        """Evaluate generated coverage against reference."""

        return EvaluationResult(
            coverage_quality=self._score_coverage_quality(generated, reference),
            format_adherence=self._score_format(generated),
            character_consistency=self._score_character_consistency(generated, screenplay),
            dialogue_coherence=self._score_dialogue_coherence(generated, screenplay),
            plot_reasonableness=self._score_plot_reasonableness(generated, screenplay),
            industry_accuracy=self._score_industry_accuracy(generated),
            human_parity=0.0  # Requires human eval
        )

    def _score_coverage_quality(self, generated: Dict, reference: Dict) -> float:
        """Score overall coverage quality using weighted metrics."""
        weights = {
            "logline": 0.1,
            "synopsis": 0.2,
            "ratings_accuracy": 0.2,
            "strengths_recall": 0.15,
            "weaknesses_recall": 0.15,
            "recommendation_match": 0.2
        }

        scores = {}

        # Logline similarity
        scores["logline"] = self._semantic_similarity(
            generated.get("logline", ""),
            reference.get("logline", "")
        )

        # Synopsis similarity
        scores["synopsis"] = self._semantic_similarity(
            generated.get("synopsis", {}).get("full", ""),
            reference.get("synopsis", {}).get("full", "")
        )

        # Ratings accuracy (within 1 point)
        gen_ratings = generated.get("ratings", {})
        ref_ratings = reference.get("ratings", {})
        rating_diffs = []
        for key in ["premise", "character", "dialogue", "structure", "pacing"]:
            gen_score = gen_ratings.get(key, {}).get("score", 5)
            ref_score = ref_ratings.get(key, {}).get("score", 5)
            rating_diffs.append(abs(gen_score - ref_score))
        scores["ratings_accuracy"] = 1.0 - (np.mean(rating_diffs) / 10.0)

        # Strengths recall
        gen_strengths = set(generated.get("strengths", []))
        ref_strengths = set(reference.get("strengths", []))
        if ref_strengths:
            scores["strengths_recall"] = len(gen_strengths & ref_strengths) / len(ref_strengths)
        else:
            scores["strengths_recall"] = 1.0 if not gen_strengths else 0.0

        # Weaknesses recall
        gen_weaknesses = set(generated.get("areas_for_development", []))
        ref_weaknesses = set(reference.get("areas_for_development", []))
        if ref_weaknesses:
            scores["weaknesses_recall"] = len(gen_weaknesses & ref_weaknesses) / len(ref_weaknesses)
        else:
            scores["weaknesses_recall"] = 1.0 if not gen_weaknesses else 0.0

        # Recommendation match
        gen_rec = generated.get("recommendation", {}).get("verdict", "")
        ref_rec = reference.get("recommendation", {}).get("verdict", "")
        scores["recommendation_match"] = 1.0 if gen_rec == ref_rec else 0.0

        # Weighted average
        return sum(scores[k] * weights[k] for k in weights)

    def _semantic_similarity(self, text1: str, text2: str) -> float:
        """Compute semantic similarity using embeddings."""
        # Use Voyage or sentence-transformers
        import voyageai
        client = voyageai.Client()

        embeddings = client.embed([text1, text2], model="voyage-3").embeddings

        # Cosine similarity
        return np.dot(embeddings[0], embeddings[1]) / (
            np.linalg.norm(embeddings[0]) * np.linalg.norm(embeddings[1])
        )
```

### 10.2 Benchmark Suite

```python
# evaluation/benchmark.py

BENCHMARK_SCRIPTS = [
    {
        "name": "LeatherApron",
        "path": "./test_scripts/leather_apron.fountain",
        "genre": ["thriller", "period"],
        "reference_coverage": "./references/leather_apron_coverage.json"
    },
    {
        "name": "TheArrangement",
        "path": "./test_scripts/the_arrangement.fountain",
        "genre": ["drama"],
        "reference_coverage": "./references/the_arrangement_coverage.json"
    },
    # ... 50+ test scripts
]

async def run_benchmark(model_path: str) -> Dict:
    """Run full benchmark suite."""

    evaluator = ScreenplayEvaluator()
    client = InferenceClient(model_path)

    results = []

    for script in BENCHMARK_SCRIPTS:
        # Load screenplay
        with open(script["path"]) as f:
            screenplay = f.read()

        # Load reference
        with open(script["reference_coverage"]) as f:
            reference = json.load(f)

        # Generate coverage
        generated = await client.analyze(screenplay, script["genre"])

        # Evaluate
        result = evaluator.evaluate_coverage(generated, reference, screenplay)
        result.script_name = script["name"]
        results.append(result)

    # Aggregate
    return {
        "mean_coverage_quality": np.mean([r.coverage_quality for r in results]),
        "mean_format_adherence": np.mean([r.format_adherence for r in results]),
        "mean_character_consistency": np.mean([r.character_consistency for r in results]),
        "mean_dialogue_coherence": np.mean([r.dialogue_coherence for r in results]),
        "mean_plot_reasonableness": np.mean([r.plot_reasonableness for r in results]),
        "mean_industry_accuracy": np.mean([r.industry_accuracy for r in results]),
        "per_script_results": results
    }
```

---

## Implementation Phases

### Phase 1: Data Collection & Preparation (2 weeks)

| Task | Duration | Deliverable |
|------|----------|-------------|
| Collect screenplay corpus | 3 days | 2,000+ scripts in Fountain format |
| Annotate coverage examples | 5 days | 500+ coverage reports |
| Build data pipeline | 3 days | Training data generator |
| Create evaluation set | 3 days | 50 held-out test scripts |

### Phase 2: Model Training (2 weeks)

| Task | Duration | Deliverable |
|------|----------|-------------|
| Setup training infrastructure | 2 days | Lambda Labs H100 cluster |
| Initial training run | 3 days | Base fine-tuned model |
| Hyperparameter tuning | 4 days | Optimized model |
| Evaluation & iteration | 5 days | Production model |

### Phase 3: RAG System (2 weeks)

| Task | Duration | Deliverable |
|------|----------|-------------|
| Deploy Qdrant | 1 day | Vector database cluster |
| Implement embedding pipeline | 3 days | Screenplay embedder |
| Build retrieval system | 3 days | Multi-collection retriever |
| Context assembly | 3 days | Prompt builder |
| Integration testing | 4 days | End-to-end RAG pipeline |

### Phase 4: API & Integration (2 weeks)

| Task | Duration | Deliverable |
|------|----------|-------------|
| Build Go gateway | 4 days | REST API server |
| CLI integration | 3 days | `anime analyze` command |
| Desktop integration | 4 days | Tauri commands + React hooks |
| Streaming support | 3 days | SSE endpoints |

### Phase 5: Production & Polish (2 weeks)

| Task | Duration | Deliverable |
|------|----------|-------------|
| Kubernetes deployment | 3 days | Production cluster |
| Monitoring & logging | 2 days | Observability stack |
| Load testing | 2 days | Performance validation |
| Documentation | 3 days | User guide + API docs |
| Launch | 4 days | Production release |

---

## Appendix: Estimated Costs

### One-Time Costs

| Item | Cost |
|------|------|
| Training compute (100 GPU-hours) | $4,000 |
| Dataset preparation (labor) | $2,000 |
| Initial RAG population | $500 |
| **Total One-Time** | **$6,500** |

### Monthly Operating Costs

| Item | Cost |
|------|------|
| Inference (2x H100) | $6,000 |
| Qdrant Cloud | $500 |
| Gateway infrastructure | $400 |
| Redis cache | $200 |
| Voyage API (embeddings) | $300 |
| **Total Monthly** | **$7,400** |

### Cost Per Analysis

- Average tokens per analysis: ~80,000 (input) + 5,000 (output)
- Estimated cost: ~$0.50-1.00 per full coverage
- With caching: ~$0.10-0.25 per analysis (70% cache hit rate)

---

## Conclusion

This architecture provides a production-ready system for AI-powered screenplay analysis that:

1. **Leverages fine-tuned Llama 3.3 70B** for domain-specific expertise
2. **Uses RAG** to ground analysis in industry knowledge and examples
3. **Integrates seamlessly** with existing ANIME CLI and Desktop apps
4. **Scales efficiently** with caching and cloud infrastructure
5. **Maintains quality** through comprehensive evaluation metrics

The system will enable rapid, consistent, professional-quality screenplay coverage at a fraction of the cost and time of human analysis, while preserving the nuance and insight that makes coverage valuable.
