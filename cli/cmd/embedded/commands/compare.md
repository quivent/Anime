# Document Comparison Command

Comprehensive document analysis and comparison with 10+ metrics, natural language summary, and utility assessment.

**Features:**
📊 **Comprehensive Analysis** - 10 detailed metrics with customizable selection
🔍 **Content Similarity** - TF-IDF cosine similarity with sequence matching  
📋 **Natural Language** - Human-readable comparison summary
🎯 **Target Profiling** - Automatic audience identification and recommendations
⚡ **Fast Processing** - Sub-second analysis for documents up to 50KB
📈 **Quality Scoring** - Readability, complexity, and utility assessment

**Usage Examples:**
- Compare two files: `/compare file1.md file2.md`
- Content focus: `/compare --focus content doc1.txt doc2.txt`
- Custom metrics: `/compare --metrics content_similarity,quality_assessment file1 file2`
- JSON output: `/compare --format json document1.md document2.md`

**Available Formats:** markdown (default), json, html, text
**Available Metrics:** content_similarity, structural_similarity, conceptual_overlap, intent_alignment, readability_score, complexity_index, information_density, quality_assessment, target_audience_match, utility_value

Target: $ARGUMENTS

Analyzes the specified documents using the Document Comparison Suite with comprehensive metrics, natural language comparison, similarity index, target user analysis, and utility value assessment based on document quality and understanding.