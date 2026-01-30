specialize - Domain adaptation engine for industry-specific variants

Usage: Transform general-purpose artifacts (libraries, templates, frameworks, processes, documentation) into domain-specific variants optimized for particular industries, use cases, or contexts.

**Sequential Specialization Protocol:**

🔍 **Phase 1: Source Analysis**
- Parse source artifact (code, docs, template, process)
- Identify generalizable components vs fixed elements
- Map extension points and customization surfaces
- Extract implicit assumptions and constraints
- Assess adaptation complexity and scope

🎯 **Phase 2: Domain Context Loading**
- Load domain expertise for target specialization:
  - **Industry Knowledge**: Vertical-specific patterns and practices
  - **Regulatory Framework**: Compliance requirements and constraints
  - **Terminology Map**: Domain vocabulary and conventions
  - **Best Practices**: Industry-standard approaches
  - **Anti-Patterns**: Domain-specific pitfalls to avoid
- Synthesize domain lens from three perspectives:
  - **Domain Expert**: What does this field actually need?
  - **Translator**: How do general concepts map to domain terms?
  - **Compliance Officer**: What constraints must be honored?

🔧 **Phase 3: Adaptation Execution**
- Apply domain specialization transformations:
  1. **Terminology**: Replace generic terms with domain vocabulary
  2. **Constraints**: Add domain-specific validations and limits
  3. **Compliance**: Inject required controls and documentation
  4. **Patterns**: Apply industry best practices
  5. **Examples**: Replace generic examples with domain-relevant ones
  6. **Warnings**: Add domain-specific cautions and considerations
- Preserve core functionality while adapting surface

✅ **Phase 4: Compliance Validation**
- Verify all regulatory requirements addressed
- Confirm terminology consistency throughout
- Validate domain patterns correctly applied
- Check that no generic assumptions leak through
- Assess fitness for stated domain purpose

📊 **Phase 5: Specialization Documentation**
- Generate SPECIALIZATION.md detailing:
  - Source artifact and target domain
  - Transformations applied
  - Compliance mappings
  - Terminology translations
  - Domain-specific considerations
- Create domain-specific usage documentation
- Note any limitations or caveats

**Specialization Domains:**
| Domain | Key Concerns | Compliance |
|--------|--------------|------------|
| Healthcare | PHI protection, patient safety | HIPAA, HITECH |
| Finance | Transaction integrity, fraud prevention | SOX, PCI-DSS |
| Education | Student privacy, accessibility | FERPA, WCAG |
| Government | Security clearance, audit trails | FedRAMP, FISMA |
| Legal | Privilege, retention, discovery | Jurisdiction-specific |
| Manufacturing | Safety, quality, traceability | ISO, OSHA |

**Integration Patterns:**

```bash
# Specialize a generic HTTP client for healthcare
specialize ./http-client --domain healthcare --compliance HIPAA

# Adapt project template for finance industry
specialize ./project-template --domain finance --compliance "SOX,PCI-DSS"

# Specialize deployment process for government
specialize ./deploy-process.md --domain government --compliance FedRAMP

# Adapt user manual for education sector
specialize ./user-manual --domain education --compliance "FERPA,WCAG"

# Healthcare specialization with strict compliance
specialize ./patient-portal --domain healthcare --compliance HIPAA --strict

# Custom domain with terminology glossary
specialize ./crm-template --domain "real-estate" --terminology ./realestate-glossary.yaml
```

**Quality Standards:**
- ✅ **Domain Fit**: Artifact feels native to target domain
- ✅ **Compliance Met**: All regulatory requirements addressed
- ✅ **Terminology Aligned**: Domain vocabulary used throughout
- ✅ **Patterns Applied**: Industry best practices integrated
- ✅ **No Leakage**: Generic assumptions don't leak through
- ✅ **Documentation Complete**: Specialization fully documented

**Domain Fit Test:**
The specialization passes if a domain expert would:
- Recognize it as domain-appropriate immediately
- Find no generic terminology that feels out of place
- Trust it meets compliance requirements
- Consider it following industry best practices

Target: $ARGUMENTS

The specialize command transforms general-purpose artifacts into domain-specific solutions, applying industry expertise, compliance requirements, and best practices to create purpose-fit variants that feel native to their target domain.
