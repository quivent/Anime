use tauri::State;
use chrono::{DateTime, Utc};
use uuid::Uuid;
use super::types::*;
use super::storage::TodoStorage;
use super::seed;

pub struct TodoState {
    pub storage: TodoStorage,
}

impl TodoState {
    pub fn new() -> Self {
        Self {
            storage: TodoStorage::new(),
        }
    }
}

impl Default for TodoState {
    fn default() -> Self {
        Self::new()
    }
}

/// Create a new todo item
#[tauri::command]
pub async fn create_todo(
    title: String,
    description: Option<String>,
    priority: TodoPriority,
    category: Option<String>,
    assignee: Option<String>,
    due_date: Option<DateTime<Utc>>,
    parent_id: Option<String>,
    tags: Vec<String>,
    state: State<'_, TodoState>,
) -> Result<Todo, String> {
    log::info!("Creating todo: {}", title);

    // Validate title is not empty
    if title.trim().is_empty() {
        return Err("Todo title cannot be empty".to_string());
    }

    // Validate parent exists if parent_id is provided
    if let Some(ref parent_id) = parent_id {
        if state.storage.get_todo(parent_id)?.is_none() {
            return Err(format!("Parent todo with id '{}' not found", parent_id));
        }
    }

    let now = Utc::now();
    let todo = Todo {
        id: Uuid::new_v4().to_string(),
        title: title.trim().to_string(),
        description,
        status: TodoStatus::Pending,
        priority,
        category,
        assignee,
        due_date,
        parent_id,
        tags: tags.clone(),
        created_at: now,
        updated_at: now,
        completed_at: None,
        estimated_hours: None,
        actual_hours: None,
        metadata: None,
    };

    // Increment tag usage counts
    for tag in &tags {
        if let Err(e) = state.storage.increment_tag_usage(tag) {
            log::warn!("Failed to increment tag usage for '{}': {}", tag, e);
        }
    }

    let result = state.storage.insert_todo(todo)?;
    log::info!("Todo created successfully: {}", result.id);

    Ok(result)
}

/// Get all todos with optional filtering
#[tauri::command]
pub async fn get_todos(
    filters: Option<TodoFilters>,
    state: State<'_, TodoState>,
) -> Result<Vec<Todo>, String> {
    log::info!("Fetching todos with filters: {:?}", filters);

    let mut todos = state.storage.get_all_todos()?;

    // Apply filters if provided
    if let Some(filters) = filters {
        // Filter by status
        if let Some(status) = filters.status {
            todos.retain(|t| matches!((&t.status, &status),
                (TodoStatus::Pending, TodoStatus::Pending) |
                (TodoStatus::InProgress, TodoStatus::InProgress) |
                (TodoStatus::Completed, TodoStatus::Completed) |
                (TodoStatus::Blocked, TodoStatus::Blocked) |
                (TodoStatus::Cancelled, TodoStatus::Cancelled)
            ));
        }

        // Filter by priority
        if let Some(priority) = filters.priority {
            todos.retain(|t| matches!((&t.priority, &priority),
                (TodoPriority::Low, TodoPriority::Low) |
                (TodoPriority::Medium, TodoPriority::Medium) |
                (TodoPriority::High, TodoPriority::High) |
                (TodoPriority::Critical, TodoPriority::Critical)
            ));
        }

        // Filter by category
        if let Some(category) = filters.category {
            todos.retain(|t| t.category.as_ref().map_or(false, |c| c == &category));
        }

        // Filter by assignee
        if let Some(assignee) = filters.assignee {
            todos.retain(|t| t.assignee.as_ref().map_or(false, |a| a == &assignee));
        }

        // Filter by tags
        if let Some(tags) = filters.tags {
            todos.retain(|t| {
                tags.iter().any(|filter_tag| {
                    t.tags.iter().any(|todo_tag| todo_tag.to_lowercase() == filter_tag.to_lowercase())
                })
            });
        }

        // Filter by search query (searches in title and description)
        if let Some(query) = filters.search_query {
            let query_lower = query.to_lowercase();
            todos.retain(|t| {
                t.title.to_lowercase().contains(&query_lower) ||
                t.description.as_ref().map_or(false, |d| d.to_lowercase().contains(&query_lower))
            });
        }

        // Filter out subtasks if requested
        if let Some(false) = filters.include_subtasks {
            todos.retain(|t| t.parent_id.is_none());
        }
    }

    log::info!("Returning {} todos", todos.len());
    Ok(todos)
}

/// Get a single todo by ID
#[tauri::command]
pub async fn get_todo_by_id(
    id: String,
    state: State<'_, TodoState>,
) -> Result<Option<Todo>, String> {
    log::info!("Fetching todo by id: {}", id);

    let todo = state.storage.get_todo(&id)?;

    if todo.is_some() {
        log::info!("Todo found: {}", id);
    } else {
        log::warn!("Todo not found: {}", id);
    }

    Ok(todo)
}

/// Update an existing todo
#[tauri::command]
pub async fn update_todo(
    id: String,
    updates: TodoUpdate,
    state: State<'_, TodoState>,
) -> Result<Todo, String> {
    log::info!("Updating todo: {}", id);

    // Validate title if provided
    if let Some(ref title) = updates.title {
        if title.trim().is_empty() {
            return Err("Todo title cannot be empty".to_string());
        }
    }

    // Update tag usage counts if tags are being updated
    if let Some(ref new_tags) = updates.tags {
        // Get the old tags
        if let Some(old_todo) = state.storage.get_todo(&id)? {
            // Increment usage for new tags
            for tag in new_tags {
                if !old_todo.tags.contains(tag) {
                    if let Err(e) = state.storage.increment_tag_usage(tag) {
                        log::warn!("Failed to increment tag usage for '{}': {}", tag, e);
                    }
                }
            }
        }
    }

    let updated_todo = state.storage.update_todo(&id, updates)?
        .ok_or_else(|| format!("Todo with id '{}' not found", id))?;

    log::info!("Todo updated successfully: {}", id);
    Ok(updated_todo)
}

/// Delete a todo by ID
#[tauri::command]
pub async fn delete_todo(
    id: String,
    state: State<'_, TodoState>,
) -> Result<bool, String> {
    log::info!("Deleting todo: {}", id);

    let deleted = state.storage.delete_todo(&id)?;

    if deleted {
        log::info!("Todo deleted successfully: {}", id);
    } else {
        log::warn!("Todo not found for deletion: {}", id);
    }

    Ok(deleted)
}

/// Bulk update todo status
#[tauri::command]
pub async fn bulk_update_todos(
    ids: Vec<String>,
    status: TodoStatus,
    state: State<'_, TodoState>,
) -> Result<usize, String> {
    log::info!("Bulk updating {} todos to status: {:?}", ids.len(), status);

    if ids.is_empty() {
        return Err("No todo IDs provided for bulk update".to_string());
    }

    let updated_count = state.storage.bulk_update_status(ids, status)?;

    log::info!("Successfully updated {} todos", updated_count);
    Ok(updated_count)
}

/// Get statistics about todos
#[tauri::command]
pub async fn get_todo_stats(
    state: State<'_, TodoState>,
) -> Result<TodoStats, String> {
    log::info!("Fetching todo statistics");

    let todos = state.storage.get_all_todos()?;
    let now = Utc::now();

    let mut stats = TodoStats {
        total: todos.len(),
        pending: 0,
        in_progress: 0,
        completed: 0,
        blocked: 0,
        cancelled: 0,
        by_priority: PriorityStats {
            low: 0,
            medium: 0,
            high: 0,
            critical: 0,
        },
        overdue: 0,
    };

    for todo in todos {
        // Count by status
        match todo.status {
            TodoStatus::Pending => stats.pending += 1,
            TodoStatus::InProgress => stats.in_progress += 1,
            TodoStatus::Completed => stats.completed += 1,
            TodoStatus::Blocked => stats.blocked += 1,
            TodoStatus::Cancelled => stats.cancelled += 1,
        }

        // Count by priority
        match todo.priority {
            TodoPriority::Low => stats.by_priority.low += 1,
            TodoPriority::Medium => stats.by_priority.medium += 1,
            TodoPriority::High => stats.by_priority.high += 1,
            TodoPriority::Critical => stats.by_priority.critical += 1,
        }

        // Count overdue todos (not completed and past due date)
        if !matches!(todo.status, TodoStatus::Completed | TodoStatus::Cancelled) {
            if let Some(due_date) = todo.due_date {
                if due_date < now {
                    stats.overdue += 1;
                }
            }
        }
    }

    log::info!("Statistics calculated: {} total, {} overdue", stats.total, stats.overdue);
    Ok(stats)
}

/// Search todos by query
#[tauri::command]
pub async fn search_todos(
    query: String,
    state: State<'_, TodoState>,
) -> Result<Vec<Todo>, String> {
    log::info!("Searching todos with query: {}", query);

    if query.trim().is_empty() {
        return Ok(Vec::new());
    }

    let filters = TodoFilters {
        status: None,
        priority: None,
        category: None,
        assignee: None,
        search_query: Some(query),
        tags: None,
        include_subtasks: Some(true),
    };

    get_todos(Some(filters), state).await
}

/// Create a new category
#[tauri::command]
pub async fn create_category(
    name: String,
    color: String,
    icon: Option<String>,
    state: State<'_, TodoState>,
) -> Result<Category, String> {
    log::info!("Creating category: {}", name);

    // Validate name is not empty
    if name.trim().is_empty() {
        return Err("Category name cannot be empty".to_string());
    }

    // Validate color format (basic hex color validation)
    if !color.starts_with('#') || (color.len() != 4 && color.len() != 7) {
        return Err("Color must be in hex format (#RGB or #RRGGBB)".to_string());
    }

    let category = Category {
        id: Uuid::new_v4().to_string(),
        name: name.trim().to_string(),
        color,
        icon,
        created_at: Utc::now(),
    };

    let result = state.storage.insert_category(category)?;
    log::info!("Category created successfully: {}", result.id);

    Ok(result)
}

/// Get all categories
#[tauri::command]
pub async fn get_categories(
    state: State<'_, TodoState>,
) -> Result<Vec<Category>, String> {
    log::info!("Fetching all categories");

    let categories = state.storage.get_all_categories()?;

    log::info!("Returning {} categories", categories.len());
    Ok(categories)
}

/// Add tags to a todo
#[tauri::command]
pub async fn add_tags_to_todo(
    todo_id: String,
    tags: Vec<String>,
    state: State<'_, TodoState>,
) -> Result<bool, String> {
    log::info!("Adding {} tags to todo: {}", tags.len(), todo_id);

    // Get the current todo
    let todo = state.storage.get_todo(&todo_id)?
        .ok_or_else(|| format!("Todo with id '{}' not found", todo_id))?;

    // Merge tags (avoiding duplicates)
    let mut new_tags = todo.tags.clone();
    for tag in tags {
        if !new_tags.contains(&tag) {
            new_tags.push(tag.clone());

            // Increment usage count
            if let Err(e) = state.storage.increment_tag_usage(&tag) {
                log::warn!("Failed to increment tag usage for '{}': {}", tag, e);
            }
        }
    }

    // Update the todo with new tags
    let updates = TodoUpdate {
        title: None,
        description: None,
        status: None,
        priority: None,
        category: None,
        assignee: None,
        due_date: None,
        tags: Some(new_tags),
        estimated_hours: None,
        actual_hours: None,
        metadata: None,
    };

    state.storage.update_todo(&todo_id, updates)?;

    log::info!("Tags added successfully to todo: {}", todo_id);
    Ok(true)
}

/// Get all tags
#[tauri::command]
pub async fn get_tags(
    state: State<'_, TodoState>,
) -> Result<Vec<Tag>, String> {
    log::info!("Fetching all tags");

    let tags = state.storage.get_all_tags()?;

    log::info!("Returning {} tags", tags.len());
    Ok(tags)
}

/// Create a new tag (helper for pre-defining tags)
#[tauri::command]
pub async fn create_tag(
    name: String,
    color: Option<String>,
    state: State<'_, TodoState>,
) -> Result<Tag, String> {
    log::info!("Creating tag: {}", name);

    // Validate name is not empty
    if name.trim().is_empty() {
        return Err("Tag name cannot be empty".to_string());
    }

    // Validate color format if provided
    if let Some(ref color) = color {
        if !color.starts_with('#') || (color.len() != 4 && color.len() != 7) {
            return Err("Color must be in hex format (#RGB or #RRGGBB)".to_string());
        }
    }

    let tag = Tag {
        id: Uuid::new_v4().to_string(),
        name: name.trim().to_string(),
        color,
        usage_count: 0,
    };

    let result = state.storage.insert_tag(tag)?;
    log::info!("Tag created successfully: {}", result.id);

    Ok(result)
}

/// Seed the database with initial todos and categories
#[tauri::command]
pub async fn seed_initial_todos(
    state: State<'_, TodoState>,
) -> Result<(usize, usize), String> {
    log::info!("Seeding initial todos and categories");

    // Get initial data
    let (categories, todos) = seed::initialize_todos();

    // Insert all categories
    let mut categories_count = 0;
    for category in categories {
        match state.storage.insert_category(category) {
            Ok(_) => categories_count += 1,
            Err(e) => log::warn!("Failed to insert category: {}", e),
        }
    }

    // Insert all todos
    let mut todos_count = 0;
    for todo in todos {
        match state.storage.insert_todo(todo) {
            Ok(_) => todos_count += 1,
            Err(e) => log::warn!("Failed to insert todo: {}", e),
        }
    }

    log::info!(
        "Seeded {} categories and {} todos",
        categories_count,
        todos_count
    );

    Ok((categories_count, todos_count))
}

/// Check if the database has been seeded
#[tauri::command]
pub async fn is_seeded(
    state: State<'_, TodoState>,
) -> Result<bool, String> {
    let todos = state.storage.get_all_todos()?;
    let categories = state.storage.get_all_categories()?;

    // Consider seeded if there are any todos or categories
    Ok(!todos.is_empty() || !categories.is_empty())
}
