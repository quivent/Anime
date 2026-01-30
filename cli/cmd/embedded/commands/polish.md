# polish - Iterative UI/UX Refinement

Discover and apply user-facing design improvements through an iterative, one-change-at-a-time workflow.

Usage: `/polish [target] [focus]` - Analyze UI/UX and propose refinements iteratively.

**Philosophy:** Great interfaces emerge through careful iteration. Each refinement is applied, verified, and approved before moving to the next. Small, visible improvements compound into exceptional user experiences.

---

## Core Focus: User-Facing Design

This command focuses exclusively on what users see and interact with:

- **Visual Design** — Colors, typography, spacing, shadows, gradients
- **Layout & Composition** — Element arrangement, visual hierarchy, balance
- **Interactions** — Hover states, transitions, animations, feedback
- **Usability** — Clarity, discoverability, accessibility, affordances
- **Consistency** — Design system coherence, pattern harmony
- **Delight** — Micro-interactions, polish details, finishing touches

**Not in scope:** Code architecture, naming conventions, internal refactoring (use `/refactor` for those).

---

## Iterative Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│                    ITERATIVE POLISH LOOP                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   1. ASSESS    →  Analyze current UI/UX state                   │
│       │            Identify improvement opportunities           │
│       ▼                                                         │
│   2. PROPOSE   →  Present ONE refinement with before/after      │
│       │            Explain the UX benefit                       │
│       ▼                                                         │
│   3. APPLY     →  Make the single change                        │
│       │            Keep change isolated and reversible          │
│       ▼                                                         │
│   4. VERIFY    →  Confirm the improvement visually              │
│       │            Check for regressions                        │
│       ▼                                                         │
│   5. CONFIRM   →  User approves or requests adjustment          │
│       │                                                         │
│       └──────→  Loop back to step 2 for next refinement         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Key Principle:** One change at a time. Apply, verify, confirm, then proceed.

---

## UI/UX Refinement Categories

### Visual Polish
```
┌─────────────────────┬────────────────────────────────────────────────────────────┐
│ Category            │ Refinements                                                │
├─────────────────────┼────────────────────────────────────────────────────────────┤
│ Color & Contrast    │ • Improve color harmony                                    │
│                     │ • Enhance contrast for readability                         │
│                     │ • Add subtle gradients or depth                            │
│                     │ • Refine accent color usage                                │
├─────────────────────┼────────────────────────────────────────────────────────────┤
│ Typography          │ • Optimize font sizes and hierarchy                        │
│                     │ • Improve line heights and spacing                         │
│                     │ • Enhance text contrast                                    │
│                     │ • Refine font weights for emphasis                         │
├─────────────────────┼────────────────────────────────────────────────────────────┤
│ Spacing & Layout    │ • Balance whitespace                                       │
│                     │ • Align elements consistently                              │
│                     │ • Improve visual grouping                                  │
│                     │ • Optimize density for scannability                        │
├─────────────────────┼────────────────────────────────────────────────────────────┤
│ Depth & Dimension   │ • Add appropriate shadows                                  │
│                     │ • Layer elements with z-index                              │
│                     │ • Create visual hierarchy through elevation                │
│                     │ • Refine border treatments                                 │
└─────────────────────┴────────────────────────────────────────────────────────────┘
```

### Interaction Polish
```
┌─────────────────────┬────────────────────────────────────────────────────────────┐
│ Category            │ Refinements                                                │
├─────────────────────┼────────────────────────────────────────────────────────────┤
│ Hover & Focus       │ • Add hover state feedback                                 │
│                     │ • Improve focus indicators for accessibility               │
│                     │ • Create smooth state transitions                          │
│                     │ • Enhance interactive affordances                          │
├─────────────────────┼────────────────────────────────────────────────────────────┤
│ Animations          │ • Add entrance/exit transitions                            │
│                     │ • Smooth layout shifts                                     │
│                     │ • Micro-interactions for feedback                          │
│                     │ • Loading state animations                                 │
├─────────────────────┼────────────────────────────────────────────────────────────┤
│ Feedback            │ • Visual confirmation of actions                           │
│                     │ • Progress indicators                                      │
│                     │ • Error state styling                                      │
│                     │ • Success/completion states                                │
└─────────────────────┴────────────────────────────────────────────────────────────┘
```

### Usability Polish
```
┌─────────────────────┬────────────────────────────────────────────────────────────┐
│ Category            │ Refinements                                                │
├─────────────────────┼────────────────────────────────────────────────────────────┤
│ Clarity             │ • Improve label clarity                                    │
│                     │ • Add helpful placeholders                                 │
│                     │ • Clarify button actions                                   │
│                     │ • Simplify complex interfaces                              │
├─────────────────────┼────────────────────────────────────────────────────────────┤
│ Accessibility       │ • Improve color contrast (WCAG)                            │
│                     │ • Add ARIA labels where needed                             │
│                     │ • Ensure keyboard navigation                               │
│                     │ • Screen reader improvements                               │
├─────────────────────┼────────────────────────────────────────────────────────────┤
│ Responsiveness      │ • Mobile layout improvements                               │
│                     │ • Touch target sizing                                      │
│                     │ • Breakpoint refinements                                   │
│                     │ • Fluid typography and spacing                             │
└─────────────────────┴────────────────────────────────────────────────────────────┘
```

---

## Execution Protocol

### Step 1: Initial Assessment

Analyze the target UI and identify opportunities:

```
╔════════════════════════════════════════════════════════════════╗
║                    UI/UX ASSESSMENT                            ║
╠════════════════════════════════════════════════════════════════╣
║                                                                ║
║  Target: [component/page name]                                 ║
║                                                                ║
║  Current State:                                                ║
║  • [What's working well]                                       ║
║  • [Areas that could be improved]                              ║
║                                                                ║
║  Refinement Opportunities (prioritized):                       ║
║  1. [Highest impact improvement]                               ║
║  2. [Second improvement]                                       ║
║  3. [Third improvement]                                        ║
║  ...                                                           ║
║                                                                ║
║  Ready to start with refinement #1?                            ║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝
```

### Step 2: Propose Single Refinement

Present one specific change:

```
╔════════════════════════════════════════════════════════════════╗
║                 REFINEMENT PROPOSAL                            ║
╠════════════════════════════════════════════════════════════════╣
║                                                                ║
║  Refinement: [Name of improvement]                             ║
║  Category: [Visual/Interaction/Usability]                      ║
║                                                                ║
║  ┌─────────────────────┬─────────────────────┐                 ║
║  │      BEFORE         │       AFTER         │                 ║
║  ├─────────────────────┼─────────────────────┤                 ║
║  │ [Current state]     │ [Improved state]    │                 ║
║  │                     │                     │                 ║
║  └─────────────────────┴─────────────────────┘                 ║
║                                                                ║
║  UX Benefit: [Why this improves the experience]                ║
║                                                                ║
║  File: [path/to/file.tsx]                                      ║
║  Lines: [affected lines]                                       ║
║                                                                ║
║  Apply this refinement? [yes/skip/modify]                      ║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝
```

### Step 3-4: Apply & Verify

Make the change and confirm:

```
╔════════════════════════════════════════════════════════════════╗
║                 REFINEMENT APPLIED                             ║
╠════════════════════════════════════════════════════════════════╣
║                                                                ║
║  ✓ Change applied to [file]                                    ║
║                                                                ║
║  Verification:                                                 ║
║  • [Specific improvement visible]                              ║
║  • [No regressions detected]                                   ║
║                                                                ║
║  Continue to next refinement? [yes/undo/done]                  ║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝
```

---

## Command Invocation

```bash
# Basic usage
/polish                          # Polish current component/page
/polish ./src/components/Card    # Polish specific component
/polish dashboard                # Polish by name

# With focus area
/polish --focus visual           # Focus on colors, typography, spacing
/polish --focus interactions     # Focus on hover, animations, transitions
/polish --focus usability        # Focus on clarity, accessibility
/polish --focus consistency      # Focus on design system harmony

# Control options
/polish --assess-only            # Just analyze, don't apply changes
/polish --quick                  # Top 3 high-impact refinements only
/polish --thorough               # Comprehensive analysis, all opportunities
```

---

## Parameters

| Parameter      | Type   | Default  | Description                              |
|----------------|--------|----------|------------------------------------------|
| target         | path   | .        | Component, page, or directory to polish  |
| --focus        | enum   | auto     | visual, interactions, usability, consistency |
| --assess-only  | flag   | false    | Analysis only, no changes                |
| --quick        | flag   | false    | Top 3 improvements only                  |
| --thorough     | flag   | false    | All improvement opportunities            |

---

## Example Session

```
User: /polish

Claude: Let me assess the current UI...

╔════════════════════════════════════════════════════════════════╗
║                    UI/UX ASSESSMENT                            ║
╠════════════════════════════════════════════════════════════════╣
║  Target: Dashboard                                             ║
║                                                                ║
║  Current State:                                                ║
║  ✓ Clean layout with good visual hierarchy                     ║
║  ✓ Consistent color palette                                    ║
║  • Cards could use subtle hover feedback                       ║
║  • Progress bars lack smooth transitions                       ║
║  • Some text contrast could be improved                        ║
║                                                                ║
║  Refinement Opportunities:                                     ║
║  1. Add hover elevation to metric cards                        ║
║  2. Animate progress bar value changes                         ║
║  3. Improve secondary text contrast                            ║
║  4. Add loading skeleton for async data                        ║
║                                                                ║
║  Starting with #1: Card hover states                           ║
╚════════════════════════════════════════════════════════════════╝

╔════════════════════════════════════════════════════════════════╗
║                 REFINEMENT PROPOSAL                            ║
╠════════════════════════════════════════════════════════════════╣
║                                                                ║
║  Refinement: Add hover elevation to metric cards               ║
║  Category: Interaction                                         ║
║                                                                ║
║  Before: Cards are static, no hover feedback                   ║
║  After:  Cards lift with shadow on hover, cursor pointer       ║
║                                                                ║
║  UX Benefit: Clearer affordance that cards are interactive,    ║
║              subtle delight on interaction                     ║
║                                                                ║
║  Apply? [yes/skip/modify]                                      ║
╚════════════════════════════════════════════════════════════════╝

User: yes

Claude: ✓ Applied hover elevation to MetricCard

[Makes the single change, shows the code diff]

Continue to refinement #2 (progress bar animation)?
```

---

## Philosophy

### Iterative Over Batch
- One visible change at a time
- User confirms each improvement
- Easy to undo or adjust
- Builds confidence through visible progress

### User Experience First
- Every change must benefit the end user
- Aesthetic improvements serve usability
- Polish should feel natural, not flashy
- Respect the existing design language

### Compound Excellence
- Small improvements stack
- Each refinement builds on the last
- The sum exceeds the parts
- Quality emerges through iteration

---

The polish command delivers user-facing design improvements through careful, iterative refinement—one change at a time, verified and approved, building toward exceptional user experience.
