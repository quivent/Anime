# /abstract - Find the Right Level

*"The right mathematical abstraction reveals the essential structure of any problem."* — Claude Shannon

---

## PURPOSE

Every problem can be described at multiple levels of abstraction. Too concrete: you're lost in details. Too abstract: you've lost the problem. This command finds the level where structure becomes clear.

---

## THE INSIGHT

When I looked at relay circuits, I could have seen:
- Copper and magnets (too concrete)
- General computation (too abstract)

Instead I saw Boolean algebra - exactly the right level where the structure of switching circuits became clear and manipulable.

The goal is not maximum abstraction. The goal is **appropriate** abstraction.

---

## PROTOCOL

**Arguments**: $ARGUMENTS (system, code, or problem to abstract)

### Phase 1: Describe at Multiple Levels

Write descriptions at increasing abstraction:

1. **Physical/Literal**: What literally happens, step by step
2. **Operational**: What operations are performed
3. **Structural**: What patterns and relationships exist
4. **Functional**: What transformation is achieved
5. **Essential**: What fundamental problem is being solved

### Phase 2: Find the Inflection Point

At which level does:
- The structure become clear?
- The solution become obvious?
- Similar problems become recognizable?
- Unnecessary details fall away?

This is the right abstraction level.

### Phase 3: Verify the Level

Test the abstraction:
- Can you solve the problem at this level?
- Can you translate solutions back to implementation?
- Does it connect to known theory or patterns?
- Is it simpler than lower levels without losing essentials?

### Phase 4: Name the Abstraction

A good abstraction deserves a name. The name should:
- Capture the essential concept
- Connect to existing knowledge
- Be memorable and usable

---

## OUTPUT FORMAT

```
ABSTRACTION ANALYSIS: [target]

LEVEL 1 - Physical/Literal:
[Concrete description]

LEVEL 2 - Operational:
[What operations occur]

LEVEL 3 - Structural:
[What patterns exist]

LEVEL 4 - Functional:
[What transformation happens]

LEVEL 5 - Essential:
[What fundamental problem is solved]

RIGHT LEVEL: [N]
Justification: [Why this level reveals structure]

THE ABSTRACTION:
Name: [What to call it]
Definition: [Precise statement]
Connections: [What known concepts it relates to]

SOLUTIONS VISIBLE AT THIS LEVEL:
[What becomes clear once properly abstracted]
```

---

## EXAMPLES

### A REST API

- **Physical**: HTTP requests over TCP/IP sockets
- **Operational**: GET, POST, PUT, DELETE on resource paths
- **Structural**: Resources with relationships, CRUD operations
- **Functional**: State transfer between client and server
- **Essential**: Distributed hypermedia interaction

**Right level**: Structural (resources and operations)
**Why**: CRUD patterns become clear, implementation details hidden

### A Sorting Algorithm

- **Physical**: Memory swaps, comparisons
- **Operational**: Compare pairs, swap if needed, repeat
- **Structural**: Divide, conquer, merge
- **Functional**: Transform unordered → ordered sequence
- **Essential**: Establish total ordering

**Right level**: Structural (for algorithm choice) or Essential (for complexity analysis)
**Why**: Structure reveals algorithm class; essence reveals bounds

### A Neural Network

- **Physical**: Matrix multiplications, memory access
- **Operational**: Forward pass, loss, backward pass, update
- **Structural**: Layers, connections, activations
- **Functional**: Function approximation through optimization
- **Essential**: Learning a mapping from data

**Right level**: Functional (for understanding) or Structural (for design)
**Why**: Functional level connects to theory; structural level enables architecture choices

---

## SIGNS YOU'RE AT THE WRONG LEVEL

### Too Concrete
- Drowning in implementation details
- Can't see the pattern across instances
- Solutions are brittle and specific
- Hard to explain to others

### Too Abstract
- Lost connection to the actual problem
- Can't translate back to implementation
- Everything looks the same (over-generalization)
- No actionable insights

---

## THE WISDOM

I didn't invent Boolean algebra. Boole did, decades earlier, for philosophy.

I didn't invent relay circuits. Engineers built them for telephone switching.

What I did was see that Boolean algebra was the right abstraction for relay circuits. The abstraction existed. The problem existed. I found where they met.

*The right abstraction is not created. It is discovered.*
