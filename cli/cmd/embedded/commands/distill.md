# /distill

Compress a slash command to minimum token cost. Same info, fewer tokens. Target: 60%+ reduction.

Scope: $ARGUMENTS (command name or file path to compress)

## Method
1. Read target command. Count lines, words.
2. Apply transforms in order:

**Prose→terse**: "Verify build + tests after each elephant. All must pass before proceeding." → "Build+test between each."
**Tables→lists**: `| Pattern | Solution |` rows → `- pattern → solution`
**Decorative rm**: `---` dividers, `*"quotes"*`, empty lines between every section
**Phase names→codes**: "Phase 0: Audit" → "P0: Audit". "Phase 1: Kill the Elephants" → "P1: Elephants"
**Common abbrevs**: function→fn, lines→ln, remove→rm, with→w/, without→w/o, parameters→params, check→chk
**Report collapse**: multi-line tables in template → single-line `Key: [vars] | Col | Col |`
**Redundant explanation rm**: "Can't do X before Y because Z" only if Z isn't obvious from context
**Principle sections**: fold into 1-2 lines or rm if protocol already embodies them

3. Verify: diff before/after. Every rule, phase, report field preserved. No info loss.
4. Measure: lines, words before/after. Report reduction %.

## Abbreviation dictionary
fn=function, ln=lines, rm=remove, w/=with, w/o=without, params=parameters, chk=check, vars=variables, desc=descending, dep=dependency, ref=reference, impl=implementation, HOF=higher-order function, arg=argument, fns=functions

## Anti-patterns
- Don't compress command names or code literals (`cargo check`, `#[allow(dead_code)]`)
- Don't merge phases — phase boundaries are load-bearing
- Don't rm report templates — output format must be preserved
- Don't abbreviate domain terms the agent might not know
