use chrono::{Duration, Utc};
use serde_json::json;
use uuid::Uuid;
use super::types::{Todo, TodoStatus, TodoPriority, Category};

/// Get all predefined categories
pub fn get_categories() -> Vec<Category> {
    let now = Utc::now();
    vec![
        Category {
            id: "bug-fixes".to_string(),
            name: "Bug Fixes".to_string(),
            icon: Some("🐛".to_string()),
            color: "#ef4444".to_string(),
            created_at: now,
        },
        Category {
            id: "features".to_string(),
            name: "Features".to_string(),
            icon: Some("✨".to_string()),
            color: "#8b5cf6".to_string(),
            created_at: now,
        },
        Category {
            id: "refactoring".to_string(),
            name: "Refactoring".to_string(),
            icon: Some("🔧".to_string()),
            color: "#f59e0b".to_string(),
            created_at: now,
        },
        Category {
            id: "documentation".to_string(),
            name: "Documentation".to_string(),
            icon: Some("📚".to_string()),
            color: "#3b82f6".to_string(),
            created_at: now,
        },
        Category {
            id: "testing".to_string(),
            name: "Testing".to_string(),
            icon: Some("🧪".to_string()),
            color: "#10b981".to_string(),
            created_at: now,
        },
        Category {
            id: "ui-ux".to_string(),
            name: "UI/UX".to_string(),
            icon: Some("🎨".to_string()),
            color: "#ec4899".to_string(),
            created_at: now,
        },
        Category {
            id: "performance".to_string(),
            name: "Performance".to_string(),
            icon: Some("⚡".to_string()),
            color: "#eab308".to_string(),
            created_at: now,
        },
        Category {
            id: "security".to_string(),
            name: "Security".to_string(),
            icon: Some("🔒".to_string()),
            color: "#dc2626".to_string(),
            created_at: now,
        },
    ]
}

/// Helper function to create a Todo
fn create_todo(
    title: &str,
    description: &str,
    category: &str,
    priority: TodoPriority,
    assignee: &str,
    tags: Vec<&str>,
    due_date_offset: Option<Duration>,
    estimated_hours: Option<f64>,
    file_refs: Vec<&str>,
) -> Todo {
    let now = Utc::now();
    let due_date = due_date_offset.map(|offset| now + offset);

    let metadata = if !file_refs.is_empty() {
        Some(json!({
            "file_references": file_refs
        }))
    } else {
        None
    };

    Todo {
        id: Uuid::new_v4().to_string(),
        title: title.to_string(),
        description: Some(description.to_string()),
        category: Some(category.to_string()),
        priority,
        status: TodoStatus::Pending,
        assignee: Some(assignee.to_string()),
        tags: tags.iter().map(|t| t.to_string()).collect(),
        due_date,
        created_at: now,
        updated_at: now,
        completed_at: None,
        parent_id: None,
        estimated_hours,
        actual_hours: None,
        metadata,
    }
}

/// Get all predefined todos with comprehensive initial dataset
pub fn get_initial_todos() -> Vec<Todo> {
    let week = Some(Duration::days(7));
    let two_weeks = Some(Duration::days(14));
    let month = Some(Duration::days(30));

    vec![
        // ========== CRITICAL PRIORITY ==========
        create_todo(
            "Complete SSH functionality implementation",
            "Multiple SSH-related TODOs need to be completed including connection handling, authentication, and session management.\n\nSubtasks:\n- Implement connection pooling\n- Add authentication methods\n- Implement session management\n- Add error recovery",
            "bug-fixes",
            TodoPriority::Critical,
            "Agent",
            vec!["ssh", "networking", "backend"],
            week,
            Some(8.0),
            vec!["/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/ssh.rs"],
        ),
        create_todo(
            "Implement real AI text generation",
            "Replace mock implementation with real AI text generation in creative.rs line 296. Connect to Lambda GPU instances for actual AI inference.\n\nSubtasks:\n- Select appropriate AI model\n- Implement API integration\n- Add streaming support\n- Handle rate limiting",
            "features",
            TodoPriority::Critical,
            "Human",
            vec!["ai", "creative", "backend", "mock"],
            week,
            Some(12.0),
            vec!["/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/creative.rs"],
        ),
        create_todo(
            "Implement real script analysis",
            "Replace mock implementation with real script analysis functionality in creative.rs line 339. Should analyze screenplay structure, pacing, and provide meaningful insights.\n\nSubtasks:\n- Define analysis criteria\n- Implement screenplay parser\n- Add AI-powered insights\n- Generate visualization data",
            "features",
            TodoPriority::Critical,
            "Human",
            vec!["ai", "creative", "nlp", "mock"],
            week,
            Some(16.0),
            vec!["/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/creative.rs"],
        ),
        create_todo(
            "Implement real screenplay parser",
            "Replace mock implementation with real screenplay parsing in creative.rs line 499. Should support standard screenplay formats (FDX, Fountain, etc).\n\nSubtasks:\n- Research screenplay format specs\n- Implement FDX parser\n- Implement Fountain parser\n- Add validation and error handling",
            "features",
            TodoPriority::Critical,
            "Human",
            vec!["parser", "creative", "backend", "mock"],
            week,
            Some(20.0),
            vec!["/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/creative.rs"],
        ),
        create_todo(
            "Implement real shot suggestions",
            "Replace mock implementation with real shot suggestions in creative.rs line 541. Should provide cinematography suggestions based on screenplay analysis.\n\nSubtasks:\n- Build cinematography knowledge base\n- Implement AI-powered shot analysis\n- Generate shot list templates",
            "features",
            TodoPriority::Critical,
            "Human",
            vec!["ai", "creative", "cinematography", "mock"],
            two_weeks,
            Some(14.0),
            vec!["/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/creative.rs"],
        ),
        create_todo(
            "Implement real storyboard image generation",
            "Replace mock implementation with real storyboard generation in creative.rs line 573. Should use Stable Diffusion or similar models on Lambda GPUs.\n\nSubtasks:\n- Set up Stable Diffusion on Lambda GPU\n- Implement prompt generation from shots\n- Add image streaming/download\n- Implement style consistency",
            "features",
            TodoPriority::Critical,
            "Human",
            vec!["ai", "creative", "image-generation", "mock"],
            two_weeks,
            Some(18.0),
            vec!["/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/creative.rs"],
        ),
        create_todo(
            "Implement real export functionality",
            "Replace mock implementation with real export functionality in creative.rs line 583. Should support multiple formats (PDF, FDX, CSV, etc).\n\nSubtasks:\n- Implement PDF export\n- Implement FDX export\n- Implement CSV export\n- Add export templates",
            "features",
            TodoPriority::Critical,
            "Human",
            vec!["export", "creative", "backend", "mock"],
            two_weeks,
            Some(10.0),
            vec!["/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/creative.rs"],
        ),

        // ========== HIGH PRIORITY ==========
        create_todo(
            "Add version tracking to installer",
            "Implement version tracking functionality in installer.rs:701 to track installed package versions.\n\nSubtasks:\n- Design version tracking schema\n- Implement version storage\n- Add version comparison logic",
            "bug-fixes",
            TodoPriority::High,
            "Agent",
            vec!["installer", "versioning", "backend"],
            two_weeks,
            Some(4.0),
            vec!["/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/installer.rs"],
        ),
        create_todo(
            "Complete dependency resolver implementation",
            "Finish implementing the dependency resolver in main.rs:21 for package management.\n\nSubtasks:\n- Implement dependency graph\n- Add circular dependency detection\n- Implement resolution algorithm",
            "features",
            TodoPriority::High,
            "Agent",
            vec!["packages", "dependencies", "backend"],
            two_weeks,
            Some(8.0),
            vec!["/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/main.rs"],
        ),
        create_todo(
            "Complete installation logic",
            "Finish implementing the installation logic in main.rs:43 for package installation.\n\nSubtasks:\n- Implement installation workflow\n- Add rollback functionality\n- Implement progress tracking",
            "features",
            TodoPriority::High,
            "Agent",
            vec!["packages", "installer", "backend"],
            two_weeks,
            Some(6.0),
            vec!["/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/main.rs"],
        ),
        create_todo(
            "Implement server loading from config",
            "Implement loading servers from configuration in main.rs:86.\n\nSubtasks:\n- Define config schema\n- Implement config parser\n- Add config validation",
            "features",
            TodoPriority::High,
            "Agent",
            vec!["config", "server", "backend"],
            two_weeks,
            Some(3.0),
            vec!["/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/main.rs"],
        ),
        create_todo(
            "Implement add server functionality",
            "Implement the add server functionality in main.rs:93 for managing server connections.\n\nSubtasks:\n- Implement server validation\n- Add server to config\n- Test connection on add",
            "features",
            TodoPriority::High,
            "Agent",
            vec!["server", "backend"],
            two_weeks,
            Some(4.0),
            vec!["/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/main.rs"],
        ),
        create_todo(
            "Fix unused variable warnings in installer.rs",
            "Fix unused variable warnings on lines 581 and 625 in installer.rs. Either use the variables or mark them with underscore prefix.",
            "refactoring",
            TodoPriority::High,
            "Agent",
            vec!["warnings", "code-quality", "quick-win"],
            week,
            Some(0.5),
            vec!["/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/installer.rs"],
        ),
        create_todo(
            "Fix unused variable warnings in models.rs",
            "Fix unused variable warnings on lines 702 and 712 in models.rs.",
            "refactoring",
            TodoPriority::High,
            "Agent",
            vec!["warnings", "code-quality", "quick-win"],
            week,
            Some(0.5),
            vec!["/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/models.rs"],
        ),
        create_todo(
            "Fix unused variable warnings in animation.rs",
            "Fix unused variable warnings on lines 358 and 459 in animation.rs.",
            "refactoring",
            TodoPriority::High,
            "Agent",
            vec!["warnings", "code-quality", "quick-win"],
            week,
            Some(0.5),
            vec!["/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/animation.rs"],
        ),
        create_todo(
            "Fix unused variable warnings in creative.rs",
            "Fix unused variable warnings on lines 298 and 626 in creative.rs.",
            "refactoring",
            TodoPriority::High,
            "Agent",
            vec!["warnings", "code-quality", "quick-win"],
            week,
            Some(0.5),
            vec!["/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/creative.rs"],
        ),
        create_todo(
            "Fix unused field api_key in LambdaClient",
            "Fix unused field warning for api_key in LambdaClient struct in client.rs:10. Either use it or remove it.",
            "refactoring",
            TodoPriority::High,
            "Agent",
            vec!["warnings", "code-quality", "lambda"],
            week,
            Some(0.5),
            vec!["/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/lambda/client.rs"],
        ),
        create_todo(
            "Remove unused import json in comfyui/commands.rs",
            "Remove unused import on line 4 in comfyui/commands.rs.",
            "refactoring",
            TodoPriority::High,
            "Agent",
            vec!["warnings", "code-quality", "quick-win"],
            week,
            Some(0.25),
            vec!["/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/comfyui/commands.rs"],
        ),

        // ========== MEDIUM PRIORITY - Code Quality ==========
        create_todo(
            "Standardize error handling patterns across components",
            "Create consistent error handling patterns across all Rust modules. Currently using mix of Result, anyhow, and thiserror.\n\nSubtasks:\n- Audit current error handling\n- Define error handling standards\n- Implement custom error types\n- Refactor existing code",
            "refactoring",
            TodoPriority::Medium,
            "Agent",
            vec!["error-handling", "architecture", "code-quality"],
            month,
            Some(12.0),
            vec![],
        ),
        create_todo(
            "Consolidate async/await vs promise patterns",
            "Standardize asynchronous code patterns in frontend. Mix of async/await and .then() promises causes inconsistency.\n\nSubtasks:\n- Audit async patterns\n- Choose standard approach\n- Refactor to standard",
            "refactoring",
            TodoPriority::Medium,
            "Agent",
            vec!["frontend", "async", "code-quality"],
            month,
            Some(6.0),
            vec![],
        ),
        create_todo(
            "Create clear state management guidelines",
            "Document state management patterns and create guidelines for when to use local state vs global state.\n\nSubtasks:\n- Analyze current state usage\n- Define state management rules\n- Document with examples",
            "documentation",
            TodoPriority::Medium,
            "Human",
            vec!["documentation", "frontend", "architecture"],
            month,
            Some(4.0),
            vec![],
        ),
        create_todo(
            "Standardize API call patterns",
            "Create consistent patterns for making API calls, including error handling, loading states, and data transformation.\n\nSubtasks:\n- Create API utility hooks\n- Implement loading state wrapper\n- Refactor existing API calls",
            "refactoring",
            TodoPriority::Medium,
            "Agent",
            vec!["api", "frontend", "code-quality"],
            month,
            Some(8.0),
            vec![],
        ),
        create_todo(
            "Create consistent className formatting standard",
            "Standardize Tailwind CSS className formatting. Mix of single line vs multi-line, inconsistent ordering.\n\nSubtasks:\n- Define className formatting rules\n- Set up Prettier/ESLint plugin\n- Refactor existing classes",
            "refactoring",
            TodoPriority::Medium,
            "Agent",
            vec!["frontend", "styling", "code-quality"],
            month,
            Some(4.0),
            vec![],
        ),
        create_todo(
            "Consolidate duplicate type definitions",
            "Many types are duplicated across frontend and backend. Create shared type definitions.\n\nSubtasks:\n- Identify duplicate types\n- Create shared types package\n- Update imports across codebase",
            "refactoring",
            TodoPriority::Medium,
            "Agent",
            vec!["types", "code-quality", "architecture"],
            month,
            Some(6.0),
            vec![],
        ),
        create_todo(
            "Add comprehensive E2E tests",
            "Set up end-to-end testing framework (Playwright/Cypress) and add tests for critical user flows.\n\nSubtasks:\n- Choose E2E testing framework\n- Set up test infrastructure\n- Write tests for critical flows\n- Set up CI/CD integration",
            "testing",
            TodoPriority::Medium,
            "Human",
            vec!["testing", "e2e", "quality"],
            month,
            Some(16.0),
            vec![],
        ),
        create_todo(
            "Add unit tests for critical functions",
            "Add unit tests for critical business logic, parsers, and utilities. Currently minimal test coverage.\n\nSubtasks:\n- Identify critical functions\n- Write tests for installers\n- Write tests for parsers\n- Write tests for utilities",
            "testing",
            TodoPriority::Medium,
            "Agent",
            vec!["testing", "unit-tests", "quality"],
            month,
            Some(12.0),
            vec![],
        ),
        create_todo(
            "Add integration tests for Lambda API",
            "Create integration tests for Lambda Cloud API interactions, including instance management and monitoring.\n\nSubtasks:\n- Set up test Lambda instance\n- Write instance management tests\n- Write SSH connection tests",
            "testing",
            TodoPriority::Medium,
            "Agent",
            vec!["testing", "integration", "lambda"],
            month,
            Some(8.0),
            vec![],
        ),
        create_todo(
            "Create architecture documentation",
            "Document overall architecture, component interactions, and data flow. Include diagrams and examples.\n\nSubtasks:\n- Create system architecture diagram\n- Document component relationships\n- Document data flow\n- Add deployment guide",
            "documentation",
            TodoPriority::Medium,
            "Human",
            vec!["documentation", "architecture"],
            month,
            Some(10.0),
            vec![],
        ),
        create_todo(
            "Document mock feature limitations",
            "Create clear documentation of which features are currently mocked and what real implementation will require.\n\nSubtasks:\n- List all mocked features\n- Document implementation requirements\n- Create implementation roadmap",
            "documentation",
            TodoPriority::Medium,
            "Agent",
            vec!["documentation", "mock"],
            month,
            Some(3.0),
            vec![],
        ),
        create_todo(
            "Create developer onboarding guide",
            "Create comprehensive guide for new developers to get started with the project, including setup, architecture overview, and contribution guidelines.\n\nSubtasks:\n- Write setup instructions\n- Document development workflow\n- Create contribution guidelines\n- Add troubleshooting section",
            "documentation",
            TodoPriority::Medium,
            "Human",
            vec!["documentation", "onboarding"],
            month,
            Some(8.0),
            vec![],
        ),

        // ========== MEDIUM PRIORITY - Features ==========
        create_todo(
            "Add dark mode theme support",
            "Implement comprehensive dark mode theme with user preference detection and toggle.\n\nSubtasks:\n- Define dark mode color palette\n- Implement theme context\n- Update all components\n- Add theme toggle UI",
            "ui-ux",
            TodoPriority::Medium,
            "Human",
            vec!["ui", "theme", "frontend"],
            month,
            Some(12.0),
            vec![],
        ),
        create_todo(
            "Implement proper logging utility for production",
            "Replace console.log with proper logging utility that supports log levels, filtering, and production mode.\n\nSubtasks:\n- Choose logging library\n- Implement log levels\n- Add log filtering\n- Replace console.log calls",
            "features",
            TodoPriority::Medium,
            "Agent",
            vec!["logging", "backend", "production"],
            month,
            Some(6.0),
            vec![],
        ),
        create_todo(
            "Add error tracking service integration",
            "Integrate error tracking service (Sentry, Rollbar, etc.) for production error monitoring.\n\nSubtasks:\n- Choose error tracking service\n- Set up service account\n- Integrate SDK\n- Configure error boundaries",
            "features",
            TodoPriority::Medium,
            "Human",
            vec!["monitoring", "production", "errors"],
            month,
            Some(4.0),
            vec![],
        ),
        create_todo(
            "Implement proper Toast notification system improvements",
            "Enhance toast notification system with better positioning, animations, and action buttons.\n\nSubtasks:\n- Add positioning options\n- Improve animations\n- Add action buttons\n- Add stacking management",
            "ui-ux",
            TodoPriority::Medium,
            "Agent",
            vec!["ui", "notifications", "frontend"],
            month,
            Some(6.0),
            vec![],
        ),
        create_todo(
            "Add keyboard shortcuts for common actions",
            "Implement keyboard shortcuts for navigation and common actions to improve power user experience.\n\nSubtasks:\n- Define shortcut schema\n- Implement shortcut handler\n- Add visual indicators\n- Create shortcuts help modal",
            "ui-ux",
            TodoPriority::Medium,
            "Agent",
            vec!["ui", "keyboard", "accessibility"],
            month,
            Some(8.0),
            vec![],
        ),
        create_todo(
            "Implement search across all views",
            "Add global search functionality to quickly find servers, packages, models, and workflows.\n\nSubtasks:\n- Implement search index\n- Create search UI component\n- Add fuzzy matching\n- Implement result navigation",
            "features",
            TodoPriority::Medium,
            "Agent",
            vec!["search", "ui", "frontend"],
            month,
            Some(10.0),
            vec![],
        ),
        create_todo(
            "Add export data functionality",
            "Implement data export functionality for servers, packages, and project configurations.\n\nSubtasks:\n- Define export formats\n- Implement JSON export\n- Implement CSV export\n- Add export UI",
            "features",
            TodoPriority::Medium,
            "Agent",
            vec!["export", "data", "backend"],
            month,
            Some(6.0),
            vec![],
        ),
        create_todo(
            "Implement settings persistence",
            "Persist user settings and preferences across sessions using Tauri store.\n\nSubtasks:\n- Define settings schema\n- Implement settings store\n- Add settings sync",
            "features",
            TodoPriority::Medium,
            "Agent",
            vec!["settings", "persistence", "backend"],
            month,
            Some(4.0),
            vec![],
        ),
        create_todo(
            "Add multi-instance management",
            "Support managing multiple Lambda instances simultaneously with bulk operations.\n\nSubtasks:\n- Design multi-instance UI\n- Implement instance grouping\n- Add bulk operations\n- Implement parallel monitoring",
            "features",
            TodoPriority::Medium,
            "Human",
            vec!["lambda", "instances", "backend"],
            month,
            Some(16.0),
            vec![],
        ),
        create_todo(
            "Implement connection retry logic",
            "Add automatic retry logic for SSH connections and API calls with exponential backoff.\n\nSubtasks:\n- Implement retry utility\n- Add exponential backoff\n- Integrate with SSH\n- Integrate with API calls",
            "features",
            TodoPriority::Medium,
            "Agent",
            vec!["networking", "reliability", "backend"],
            month,
            Some(6.0),
            vec![],
        ),

        // ========== LOW PRIORITY - Nice to Have ==========
        create_todo(
            "Add animation transitions between views",
            "Add smooth animations when transitioning between different views and components.",
            "ui-ux",
            TodoPriority::Low,
            "Agent",
            vec!["ui", "animations", "polish"],
            None,
            Some(4.0),
            vec![],
        ),
        create_todo(
            "Implement drag-and-drop file uploads",
            "Add drag-and-drop support for uploading screenplay files and assets.",
            "ui-ux",
            TodoPriority::Low,
            "Agent",
            vec!["ui", "upload", "frontend"],
            None,
            Some(4.0),
            vec![],
        ),
        create_todo(
            "Add command palette (Cmd+K)",
            "Implement command palette for quick access to all actions and navigation.\n\nSubtasks:\n- Design command palette UI\n- Implement command registry\n- Add fuzzy search",
            "ui-ux",
            TodoPriority::Low,
            "Agent",
            vec!["ui", "navigation", "productivity"],
            None,
            Some(8.0),
            vec![],
        ),
        create_todo(
            "Create custom icon set",
            "Design and implement custom icon set matching the application's visual style.",
            "ui-ux",
            TodoPriority::Low,
            "Human",
            vec!["design", "ui", "assets"],
            None,
            Some(12.0),
            vec![],
        ),
        create_todo(
            "Add tutorial/onboarding flow",
            "Create interactive tutorial for first-time users explaining key features.",
            "ui-ux",
            TodoPriority::Low,
            "Human",
            vec!["onboarding", "ui", "ux"],
            None,
            Some(8.0),
            vec![],
        ),
        create_todo(
            "Implement undo/redo functionality",
            "Add undo/redo support for creative workflow actions.",
            "features",
            TodoPriority::Low,
            "Agent",
            vec!["ui", "workflow", "frontend"],
            None,
            Some(10.0),
            vec![],
        ),
        create_todo(
            "Add bulk operations for packages",
            "Implement bulk install, update, and remove operations for packages.",
            "features",
            TodoPriority::Low,
            "Agent",
            vec!["packages", "backend"],
            None,
            Some(6.0),
            vec![],
        ),
        create_todo(
            "Create dashboard analytics view",
            "Add dashboard with analytics showing instance usage, cost tracking, and workflow statistics.",
            "features",
            TodoPriority::Low,
            "Human",
            vec!["analytics", "dashboard", "ui"],
            None,
            Some(16.0),
            vec![],
        ),
        create_todo(
            "Add notification system",
            "Implement system notifications for long-running tasks and important events.",
            "features",
            TodoPriority::Low,
            "Agent",
            vec!["notifications", "ui"],
            None,
            Some(4.0),
            vec![],
        ),
        create_todo(
            "Implement backup/restore functionality",
            "Add ability to backup and restore all application data and configurations.",
            "features",
            TodoPriority::Low,
            "Agent",
            vec!["backup", "data", "backend"],
            None,
            Some(8.0),
            vec![],
        ),
    ]
}

/// Initialize the todo dataset and return it for storage
pub fn initialize_todos() -> (Vec<Category>, Vec<Todo>) {
    (get_categories(), get_initial_todos())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_categories_have_unique_ids() {
        let categories = get_categories();
        let mut ids = std::collections::HashSet::new();
        for category in categories {
            assert!(ids.insert(category.id.clone()), "Duplicate category ID found");
        }
    }

    #[test]
    fn test_todos_have_valid_category_refs() {
        let categories = get_categories();
        let category_ids: std::collections::HashSet<_> =
            categories.iter().map(|c| c.id.clone()).collect();

        let todos = get_initial_todos();
        for todo in todos {
            if let Some(cat) = &todo.category {
                assert!(
                    category_ids.contains(cat),
                    "Todo references invalid category: {}",
                    cat
                );
            }
        }
    }

    #[test]
    fn test_all_priorities_covered() {
        let todos = get_initial_todos();
        let priorities: std::collections::HashSet<_> =
            todos.iter().map(|t| t.priority).collect();

        assert!(priorities.contains(&TodoPriority::Critical));
        assert!(priorities.contains(&TodoPriority::High));
        assert!(priorities.contains(&TodoPriority::Medium));
        assert!(priorities.contains(&TodoPriority::Low));
    }

    #[test]
    fn test_todo_count() {
        let todos = get_initial_todos();
        assert_eq!(todos.len(), 50, "Expected exactly 50 todos");
    }
}
