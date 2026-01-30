use serde::{Deserialize, Serialize};
use std::fs;
use std::path::PathBuf;
use tauri::{AppHandle, Manager};
use uuid::Uuid;

// ============================================================================
// Writing Tab Structures
// ============================================================================

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Document {
    pub id: String,
    pub title: String,
    pub content: String,
    pub created_at: String,
    pub updated_at: String,
    pub word_count: usize,
    pub tags: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WritingRequest {
    pub mode: String,
    pub context: String,
    pub prompt: String,
    pub max_tokens: Option<usize>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WritingResponse {
    pub generated_text: String,
    pub usage: TokenUsage,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TokenUsage {
    pub prompt_tokens: usize,
    pub completion_tokens: usize,
    pub total_tokens: usize,
}

// ============================================================================
// Analysis Tab Structures
// ============================================================================

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CharacterAnalysis {
    pub characters: Vec<Character>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Character {
    pub name: String,
    pub role: String,
    pub traits: Vec<String>,
    pub arc: String,
    pub dialogue_count: usize,
    pub importance_score: f32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PlotAnalysis {
    pub structure: PlotStructure,
    pub pacing: PacingInfo,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PlotStructure {
    pub act_breakdown: Vec<String>,
    pub turning_points: Vec<String>,
    pub conflicts: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PacingInfo {
    pub slow_sections: Vec<usize>,
    pub fast_sections: Vec<usize>,
    pub overall_score: f32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DialogueAnalysis {
    pub total_lines: usize,
    pub avg_length: usize,
    pub unique_voices: usize,
    pub readability_score: f32,
    pub suggestions: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PacingAnalysis {
    pub beats_per_scene: Vec<usize>,
    pub scene_lengths: Vec<usize>,
    pub tension_curve: Vec<f32>,
    pub recommendations: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ThemeAnalysis {
    pub primary_themes: Vec<String>,
    pub recurring_motifs: Vec<String>,
    pub symbolism: Vec<Symbol>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Symbol {
    pub symbol: String,
    pub meaning: String,
    pub occurrences: usize,
}

// ============================================================================
// Storyboard Tab Structures
// ============================================================================

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StoryboardProject {
    pub id: String,
    pub name: String,
    pub created_at: String,
    pub updated_at: String,
    pub script: String,
    pub scenes: Vec<StoryboardScene>,
    pub panels: Vec<StoryboardPanel>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StoryboardScene {
    pub id: String,
    pub scene_number: usize,
    pub title: String,
    pub description: String,
    pub duration: Option<String>,
    pub location: Option<String>,
    pub time: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StoryboardPanel {
    pub id: String,
    pub scene_id: String,
    pub panel_number: usize,
    pub shot_type: String,
    pub composition: String,
    pub description: String,
    pub dialogue: Option<String>,
    pub image_url: Option<String>,
    pub notes: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ScriptParseResult {
    pub scenes: Vec<StoryboardScene>,
    pub total_scenes: usize,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ShotSuggestion {
    #[serde(rename = "type")]
    pub shot_type: String,
    pub description: String,
    pub composition: String,
    pub camera_angle: String,
    pub lighting: String,
}

// ============================================================================
// Helper Functions
// ============================================================================

fn get_documents_dir(app: &AppHandle) -> Result<PathBuf, String> {
    let app_dir = app
        .path()
        .app_data_dir()
        .map_err(|e| format!("Failed to get app data dir: {}", e))?;
    let docs_dir = app_dir.join("documents");
    fs::create_dir_all(&docs_dir).map_err(|e| format!("Failed to create documents dir: {}", e))?;
    Ok(docs_dir)
}

fn get_storyboards_dir(app: &AppHandle) -> Result<PathBuf, String> {
    let app_dir = app
        .path()
        .app_data_dir()
        .map_err(|e| format!("Failed to get app data dir: {}", e))?;
    let sb_dir = app_dir.join("storyboards");
    fs::create_dir_all(&sb_dir).map_err(|e| format!("Failed to create storyboards dir: {}", e))?;
    Ok(sb_dir)
}

// ============================================================================
// Writing Commands
// ============================================================================

#[tauri::command]
pub async fn list_documents(app: AppHandle) -> Result<Vec<Document>, String> {
    let docs_dir = get_documents_dir(&app)?;
    let mut documents = Vec::new();

    if let Ok(entries) = fs::read_dir(docs_dir) {
        for entry in entries.flatten() {
            if let Ok(content) = fs::read_to_string(entry.path()) {
                if let Ok(doc) = serde_json::from_str::<Document>(&content) {
                    documents.push(doc);
                }
            }
        }
    }

    documents.sort_by(|a, b| b.updated_at.cmp(&a.updated_at));
    Ok(documents)
}

#[tauri::command]
pub async fn create_document(app: AppHandle, title: String) -> Result<Document, String> {
    let doc = Document {
        id: Uuid::new_v4().to_string(),
        title,
        content: String::new(),
        created_at: chrono::Utc::now().to_rfc3339(),
        updated_at: chrono::Utc::now().to_rfc3339(),
        word_count: 0,
        tags: Vec::new(),
    };

    let docs_dir = get_documents_dir(&app)?;
    let doc_path = docs_dir.join(format!("{}.json", doc.id));
    let content = serde_json::to_string_pretty(&doc)
        .map_err(|e| format!("Failed to serialize document: {}", e))?;
    fs::write(doc_path, content).map_err(|e| format!("Failed to write document: {}", e))?;

    Ok(doc)
}

#[tauri::command]
pub async fn save_document(app: AppHandle, id: String, content: String) -> Result<Document, String> {
    let docs_dir = get_documents_dir(&app)?;
    let doc_path = docs_dir.join(format!("{}.json", id));

    let mut doc: Document = if doc_path.exists() {
        let file_content = fs::read_to_string(&doc_path)
            .map_err(|e| format!("Failed to read document: {}", e))?;
        serde_json::from_str(&file_content)
            .map_err(|e| format!("Failed to parse document: {}", e))?
    } else {
        return Err("Document not found".to_string());
    };

    doc.content = content.clone();
    doc.word_count = content.split_whitespace().count();
    doc.updated_at = chrono::Utc::now().to_rfc3339();

    let serialized = serde_json::to_string_pretty(&doc)
        .map_err(|e| format!("Failed to serialize document: {}", e))?;
    fs::write(doc_path, serialized).map_err(|e| format!("Failed to write document: {}", e))?;

    Ok(doc)
}

#[tauri::command]
pub async fn delete_document(app: AppHandle, id: String) -> Result<(), String> {
    let docs_dir = get_documents_dir(&app)?;
    let doc_path = docs_dir.join(format!("{}.json", id));
    fs::remove_file(doc_path).map_err(|e| format!("Failed to delete document: {}", e))?;
    Ok(())
}

#[tauri::command]
pub async fn export_document(app: AppHandle, id: String) -> Result<(), String> {
    let docs_dir = get_documents_dir(&app)?;
    let doc_path = docs_dir.join(format!("{}.json", id));

    let content = fs::read_to_string(&doc_path)
        .map_err(|e| format!("Failed to read document: {}", e))?;
    let doc: Document = serde_json::from_str(&content)
        .map_err(|e| format!("Failed to parse document: {}", e))?;

    let export_path = dirs::download_dir()
        .ok_or("Failed to get downloads directory")?
        .join(format!("{}.txt", doc.title));

    fs::write(export_path, doc.content)
        .map_err(|e| format!("Failed to export document: {}", e))?;

    Ok(())
}

/// **WARNING: MOCK IMPLEMENTATION - NOT PRODUCTION READY**
///
/// This function returns pre-written mock text and does not use real AI generation.
/// In production, this should integrate with an actual LLM API (e.g., OpenAI, Anthropic, etc.)
#[tauri::command]
pub async fn generate_text(
    mode: String,
    context: String,
    prompt: String,
    max_tokens: Option<usize>,
) -> Result<WritingResponse, String> {
    // ⚠️ WARNING: This is a MOCK implementation


    // Mock AI generation - in production, this would call an LLM API
    let generated = match mode.as_str() {
        "continuation" => {
            format!("The story continues from where it left off. The protagonist discovers that the mysterious artifact holds a power beyond imagination. As the sun sets over the horizon, a new chapter begins...")
        }
        "dialogue" => {
            format!("\"I never expected to find you here,\" Sarah said, her voice trembling slightly.\n\n\"Life has a way of bringing people together when they least expect it,\" Marcus replied with a knowing smile.\n\n\"But after all these years... why now?\"\n\n\"Because some stories aren't meant to end.\"")
        }
        "scene" => {
            format!("The abandoned warehouse loomed before them, its broken windows reflecting the pale moonlight. Graffiti covered the weathered brick walls, and the air hung thick with the scent of rust and decay. A faint sound echoed from within - footsteps on metal, deliberate and approaching.")
        }
        "outline" => {
            format!("Act 1: Setup and Introduction\n- Introduce protagonist in their ordinary world\n- Inciting incident disrupts their life\n- Initial resistance to change\n\nAct 2: Rising Action\n- Protagonist accepts the challenge\n- Series of obstacles and setbacks\n- Midpoint twist changes everything\n\nAct 3: Climax and Resolution\n- Final confrontation\n- Character transformation\n- New equilibrium established")
        }
        _ => "Generated text...".to_string(),
    };

    let completion_tokens = generated.split_whitespace().count();
    let prompt_tokens = context.split_whitespace().count() + prompt.split_whitespace().count();

    // Add warning prefix to the generated text
    let warning_prefix = "⚠️ MOCK DATA - This is placeholder text, not real AI generation.\n\n";
    let generated_with_warning = format!("{}{}", warning_prefix, generated);

    Ok(WritingResponse {
        generated_text: generated_with_warning,
        usage: TokenUsage {
            prompt_tokens,
            completion_tokens,
            total_tokens: prompt_tokens + completion_tokens,
        },
    })
}

// ============================================================================
// Analysis Commands
// ============================================================================

#[tauri::command]
pub async fn read_file(path: String) -> Result<String, String> {
    fs::read_to_string(&path).map_err(|e| format!("Failed to read file: {}", e))
}

/// **WARNING: MOCK IMPLEMENTATION - NOT PRODUCTION READY**
///
/// This function returns hardcoded analysis data and does not perform real NLP/AI analysis.
/// In production, this should integrate with actual analysis services or ML models.
#[tauri::command]
pub async fn analyze_content(
    r#type: String,
    _content: String,
) -> Result<serde_json::Value, String> {
    // ⚠️ WARNING: This is a MOCK implementation


    // Mock analysis - in production, this would use AI/NLP
    let result = match r#type.as_str() {
        "character" => {
            let analysis = CharacterAnalysis {
                characters: vec![
                    Character {
                        name: "Alex Morgan".to_string(),
                        role: "Protagonist".to_string(),
                        traits: vec!["Determined".to_string(), "Curious".to_string(), "Resourceful".to_string()],
                        arc: "Transforms from hesitant observer to confident leader".to_string(),
                        dialogue_count: 42,
                        importance_score: 0.95,
                    },
                    Character {
                        name: "Dr. Sarah Chen".to_string(),
                        role: "Mentor".to_string(),
                        traits: vec!["Wise".to_string(), "Experienced".to_string(), "Supportive".to_string()],
                        arc: "Guides the protagonist while facing her own challenges".to_string(),
                        dialogue_count: 28,
                        importance_score: 0.78,
                    },
                ],
            };
            serde_json::json!({"type": "character", "data": analysis})
        }
        "plot" => {
            let analysis = PlotAnalysis {
                structure: PlotStructure {
                    act_breakdown: vec![
                        "Act 1: Introduction and setup (pages 1-25)".to_string(),
                        "Act 2: Rising action and complications (pages 26-75)".to_string(),
                        "Act 3: Climax and resolution (pages 76-100)".to_string(),
                    ],
                    turning_points: vec![
                        "Discovery of the hidden message (page 15)".to_string(),
                        "Betrayal by trusted ally (page 45)".to_string(),
                        "Revelation of true antagonist (page 70)".to_string(),
                    ],
                    conflicts: vec![
                        "Internal: Self-doubt vs confidence".to_string(),
                        "External: Protagonist vs antagonist forces".to_string(),
                        "Relational: Trust issues with team".to_string(),
                    ],
                },
                pacing: PacingInfo {
                    slow_sections: vec![1, 2, 5],
                    fast_sections: vec![3, 7, 9],
                    overall_score: 75.0,
                },
            };
            serde_json::json!({"type": "plot", "data": analysis})
        }
        "dialogue" => {
            let analysis = DialogueAnalysis {
                total_lines: 156,
                avg_length: 12,
                unique_voices: 5,
                readability_score: 82.0,
                suggestions: vec![
                    "Consider varying sentence length for more natural flow".to_string(),
                    "Character voices could be more distinct in emotional scenes".to_string(),
                    "Strong use of subtext in confrontation scenes".to_string(),
                ],
            };
            serde_json::json!({"type": "dialogue", "data": analysis})
        }
        "pacing" => {
            let analysis = PacingAnalysis {
                beats_per_scene: vec![3, 5, 2, 4, 6, 3, 5, 4, 7],
                scene_lengths: vec![450, 680, 320, 550, 720, 390, 610, 480, 850],
                tension_curve: vec![0.2, 0.4, 0.3, 0.6, 0.8, 0.5, 0.9, 0.95, 1.0],
                recommendations: vec![
                    "Scene 3 could benefit from additional tension building".to_string(),
                    "Consider shortening scene 5 to maintain momentum".to_string(),
                    "Excellent escalation in final act".to_string(),
                ],
            };
            serde_json::json!({"type": "pacing", "data": analysis})
        }
        "theme" => {
            let analysis = ThemeAnalysis {
                primary_themes: vec![
                    "Redemption".to_string(),
                    "Identity and self-discovery".to_string(),
                    "The cost of ambition".to_string(),
                ],
                recurring_motifs: vec![
                    "Mirrors and reflections".to_string(),
                    "Broken clocks".to_string(),
                    "Water imagery".to_string(),
                ],
                symbolism: vec![
                    Symbol {
                        symbol: "The red door".to_string(),
                        meaning: "Represents choice and point of no return".to_string(),
                        occurrences: 7,
                    },
                    Symbol {
                        symbol: "Lighthouse".to_string(),
                        meaning: "Guidance and hope in darkness".to_string(),
                        occurrences: 4,
                    },
                ],
            };
            serde_json::json!({"type": "theme", "data": analysis})
        }
        _ => return Err("Unknown analysis type".to_string()),
    };

    Ok(result)
}

#[tauri::command]
pub async fn export_analysis(
    result: serde_json::Value,
    file_name: String,
) -> Result<(), String> {
    let export_path = dirs::download_dir()
        .ok_or("Failed to get downloads directory")?
        .join(format!("{}_analysis.json", file_name));

    let content = serde_json::to_string_pretty(&result)
        .map_err(|e| format!("Failed to serialize analysis: {}", e))?;

    fs::write(export_path, content)
        .map_err(|e| format!("Failed to export analysis: {}", e))?;

    Ok(())
}

// ============================================================================
// Storyboard Commands
// ============================================================================

#[tauri::command]
pub async fn create_storyboard_project(
    app: AppHandle,
    name: String,
) -> Result<StoryboardProject, String> {
    let project = StoryboardProject {
        id: Uuid::new_v4().to_string(),
        name,
        created_at: chrono::Utc::now().to_rfc3339(),
        updated_at: chrono::Utc::now().to_rfc3339(),
        script: String::new(),
        scenes: Vec::new(),
        panels: Vec::new(),
    };

    let sb_dir = get_storyboards_dir(&app)?;
    let project_path = sb_dir.join(format!("{}.json", project.id));
    let content = serde_json::to_string_pretty(&project)
        .map_err(|e| format!("Failed to serialize project: {}", e))?;
    fs::write(project_path, content).map_err(|e| format!("Failed to write project: {}", e))?;

    Ok(project)
}

/// **WARNING: MOCK IMPLEMENTATION - NOT PRODUCTION READY**
///
/// This function returns hardcoded scene data and does not parse actual scripts.
/// In production, this should implement proper screenplay parsing (Fountain format, etc.)
#[tauri::command]
pub async fn parse_script(_script: String) -> Result<ScriptParseResult, String> {
    // ⚠️ WARNING: This is a MOCK implementation


    // Mock script parser - in production, this would use proper screenplay parsing
    let scenes: Vec<StoryboardScene> = vec![
        StoryboardScene {
            id: Uuid::new_v4().to_string(),
            scene_number: 1,
            title: "Opening - City Streets".to_string(),
            description: "Establishing shot of bustling city streets at dawn".to_string(),
            duration: Some("30 seconds".to_string()),
            location: Some("Downtown Manhattan".to_string()),
            time: Some("Early Morning".to_string()),
        },
        StoryboardScene {
            id: Uuid::new_v4().to_string(),
            scene_number: 2,
            title: "Apartment - Morning Routine".to_string(),
            description: "Protagonist wakes up in small apartment, preparing for the day".to_string(),
            duration: Some("45 seconds".to_string()),
            location: Some("Alex's Apartment".to_string()),
            time: Some("Morning".to_string()),
        },
        StoryboardScene {
            id: Uuid::new_v4().to_string(),
            scene_number: 3,
            title: "Discovery - The Message".to_string(),
            description: "Protagonist finds mysterious envelope under the door".to_string(),
            duration: Some("1 minute".to_string()),
            location: Some("Apartment Hallway".to_string()),
            time: Some("Morning".to_string()),
        },
    ];

    Ok(ScriptParseResult {
        total_scenes: scenes.len(),
        scenes,
    })
}

/// **WARNING: MOCK IMPLEMENTATION - NOT PRODUCTION READY**
///
/// This function returns hardcoded shot suggestions and does not use AI analysis.
/// In production, this should integrate with AI services to analyze scene context.
#[tauri::command]
pub async fn generate_shot_suggestions(
    _project_id: String,
    _scene_id: String,
) -> Result<Vec<ShotSuggestion>, String> {
    // ⚠️ WARNING: This is a MOCK implementation


    // Mock shot suggestions - in production, this would use AI
    Ok(vec![
        ShotSuggestion {
            shot_type: "Wide Shot".to_string(),
            description: "Establish the location and setting".to_string(),
            composition: "Rule of thirds, deep focus".to_string(),
            camera_angle: "Eye level".to_string(),
            lighting: "Natural daylight, soft shadows".to_string(),
        },
        ShotSuggestion {
            shot_type: "Medium Shot".to_string(),
            description: "Character enters frame, showing emotion".to_string(),
            composition: "Center frame, shallow depth of field".to_string(),
            camera_angle: "Slight low angle".to_string(),
            lighting: "Three-point lighting".to_string(),
        },
        ShotSuggestion {
            shot_type: "Close-up".to_string(),
            description: "Focus on character's reaction to discovery".to_string(),
            composition: "Tight frame on face".to_string(),
            camera_angle: "Eye level, slightly off-center".to_string(),
            lighting: "Dramatic side lighting".to_string(),
        },
    ])
}

/// **WARNING: MOCK IMPLEMENTATION - NOT PRODUCTION READY**
///
/// This function returns placeholder.com URLs and does not generate real images.
/// In production, this should integrate with ComfyUI, Stable Diffusion, or other image generation.
#[tauri::command]
pub async fn generate_storyboard_image(
    _description: String,
    shot_type: String,
    _composition: String,
) -> Result<String, String> {
    // ⚠️ WARNING: This is a MOCK implementation


    // Mock image generation - in production, this would call ComfyUI or Stable Diffusion
    // For now, return a placeholder URL
    Ok(format!(
        "https://via.placeholder.com/800x450/1a1a1a/ffffff?text={}+Shot",
        shot_type.replace(' ', "+")
    ))
}

/// **WARNING: MOCK IMPLEMENTATION - NOT PRODUCTION READY**
///
/// This function does not actually export anything - it's a placeholder.
/// In production, this should generate PDFs or export images using a proper rendering library.
#[tauri::command]
pub async fn export_storyboard(project_id: String, format: String) -> Result<(), String> {
    // ⚠️ WARNING: This is a MOCK implementation


    // Mock export - in production, this would generate PDF/images

    Ok(())
}
