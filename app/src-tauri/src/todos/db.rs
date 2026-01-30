//! Database operations for the todo system using SQLite
//!
//! This module provides all CRUD operations and queries for the todo system.

use super::types::*;
use anyhow::{Context, Result};
use chrono::{DateTime, Utc};
use rusqlite::{params, Connection, Row};
use std::path::PathBuf;
use std::sync::{Arc, Mutex};

/// Database manager for todo system
pub struct TodoDatabase {
    conn: Arc<Mutex<Connection>>,
}

impl TodoDatabase {
    /// Initialize the database with schema
    pub fn new(db_path: PathBuf) -> Result<Self> {
        log::info!("Initializing todo database at: {:?}", db_path);

        let conn = Connection::open(&db_path)
            .context("Failed to open database connection")?;

        // Read and execute schema
        let schema = include_str!("../../schema/todos.sql");
        conn.execute_batch(schema)
            .context("Failed to execute schema")?;

        log::info!("Todo database initialized successfully");

        Ok(Self {
            conn: Arc::new(Mutex::new(conn)),
        })
    }

    // ===== TODO OPERATIONS =====

    /// Create a new todo
    pub fn create_todo(&self, todo: &Todo) -> Result<Todo> {
        let conn = self.conn.lock().unwrap();

        let metadata_json = todo.metadata.as_ref()
            .map(|m| serde_json::to_string(m).unwrap_or_default());

        conn.execute(
            "INSERT INTO todos (
                title, description, status, priority, assignee, category_id,
                parent_id, due_date, estimated_hours, actual_hours, metadata
            ) VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, ?10, ?11)",
            params![
                todo.title,
                todo.description,
                status_to_string(&todo.status),
                priority_to_string(&todo.priority),
                todo.assignee,
                self.get_category_id_by_name(&conn, &todo.category)?,
                todo.parent_id,
                todo.due_date.map(|d| d.to_rfc3339()),
                todo.estimated_hours,
                todo.actual_hours,
                metadata_json,
            ],
        )?;

        let id = conn.last_insert_rowid();
        self.get_todo_by_id(&format!("{}", id))
    }

    /// Get a todo by ID
    pub fn get_todo_by_id(&self, id: &str) -> Result<Todo> {
        let conn = self.conn.lock().unwrap();

        let mut stmt = conn.prepare(
            "SELECT
                t.id, t.title, t.description, t.status, t.priority, t.assignee,
                c.name as category_name, t.parent_id, t.created_at, t.updated_at,
                t.completed_at, t.due_date, t.estimated_hours, t.actual_hours, t.metadata
            FROM todos t
            LEFT JOIN categories c ON t.category_id = c.id
            WHERE t.id = ?1"
        )?;

        let todo = stmt.query_row(params![id], |row| {
            Ok(self.row_to_todo(row)?)
        })?;

        // Get tags for this todo
        let tags = self.get_tags_for_todo(&conn, id)?;

        Ok(Todo { tags, ..todo })
    }

    /// Get all todos with optional filters
    pub fn get_todos(&self, filters: &TodoFilters) -> Result<Vec<Todo>> {
        let conn = self.conn.lock().unwrap();

        let mut query = String::from(
            "SELECT
                t.id, t.title, t.description, t.status, t.priority, t.assignee,
                c.name as category_name, t.parent_id, t.created_at, t.updated_at,
                t.completed_at, t.due_date, t.estimated_hours, t.actual_hours, t.metadata
            FROM todos t
            LEFT JOIN categories c ON t.category_id = c.id
            WHERE 1=1"
        );

        let mut params: Vec<Box<dyn rusqlite::ToSql>> = Vec::new();

        // Apply filters
        if let Some(status) = &filters.status {
            query.push_str(" AND t.status = ?");
            params.push(Box::new(status_to_string(status)));
        }

        if let Some(priority) = &filters.priority {
            query.push_str(" AND t.priority = ?");
            params.push(Box::new(priority_to_string(priority)));
        }

        if let Some(category) = &filters.category {
            query.push_str(" AND c.name = ?");
            params.push(Box::new(category.clone()));
        }

        if let Some(assignee) = &filters.assignee {
            query.push_str(" AND t.assignee = ?");
            params.push(Box::new(assignee.clone()));
        }

        if let Some(search_query) = &filters.search_query {
            query.push_str(" AND (t.title LIKE ? OR t.description LIKE ?)");
            let search_pattern = format!("%{}%", search_query);
            params.push(Box::new(search_pattern.clone()));
            params.push(Box::new(search_pattern));
        }

        if let Some(false) = filters.include_subtasks {
            query.push_str(" AND t.parent_id IS NULL");
        }

        query.push_str(" ORDER BY t.created_at DESC");

        let mut stmt = conn.prepare(&query)?;
        let param_refs: Vec<&dyn rusqlite::ToSql> = params.iter().map(|p| p.as_ref()).collect();

        let todos = stmt.query_map(param_refs.as_slice(), |row| {
            Ok(self.row_to_todo(row)?)
        })?
        .collect::<Result<Vec<_>, _>>()?;

        // Get tags for each todo
        let mut todos_with_tags = Vec::new();
        for mut todo in todos {
            let tags = self.get_tags_for_todo(&conn, &todo.id)?;
            todo.tags = tags;
            todos_with_tags.push(todo);
        }

        Ok(todos_with_tags)
    }

    /// Update a todo
    pub fn update_todo(&self, id: &str, updates: &TodoUpdate) -> Result<Option<Todo>> {
        let conn = self.conn.lock().unwrap();

        // Check if todo exists
        if !self.todo_exists(&conn, id)? {
            return Ok(None);
        }

        let mut set_clauses = Vec::new();
        let mut params: Vec<Box<dyn rusqlite::ToSql>> = Vec::new();

        if let Some(title) = &updates.title {
            set_clauses.push("title = ?");
            params.push(Box::new(title.clone()));
        }

        if let Some(description) = &updates.description {
            set_clauses.push("description = ?");
            params.push(Box::new(description.clone()));
        }

        if let Some(status) = &updates.status {
            set_clauses.push("status = ?");
            params.push(Box::new(status_to_string(status)));
        }

        if let Some(priority) = &updates.priority {
            set_clauses.push("priority = ?");
            params.push(Box::new(priority_to_string(priority)));
        }

        if let Some(category) = &updates.category {
            let category_id = self.get_category_id_by_name(&conn, &Some(category.clone()))?;
            set_clauses.push("category_id = ?");
            params.push(Box::new(category_id));
        }

        if let Some(assignee) = &updates.assignee {
            set_clauses.push("assignee = ?");
            params.push(Box::new(assignee.clone()));
        }

        if let Some(due_date) = &updates.due_date {
            set_clauses.push("due_date = ?");
            params.push(Box::new(due_date.to_rfc3339()));
        }

        if let Some(estimated_hours) = updates.estimated_hours {
            set_clauses.push("estimated_hours = ?");
            params.push(Box::new(estimated_hours));
        }

        if let Some(actual_hours) = updates.actual_hours {
            set_clauses.push("actual_hours = ?");
            params.push(Box::new(actual_hours));
        }

        if let Some(metadata) = &updates.metadata {
            set_clauses.push("metadata = ?");
            let metadata_json = serde_json::to_string(metadata)?;
            params.push(Box::new(metadata_json));
        }

        if !set_clauses.is_empty() {
            let query = format!(
                "UPDATE todos SET {} WHERE id = ?",
                set_clauses.join(", ")
            );
            params.push(Box::new(id.to_string()));

            let param_refs: Vec<&dyn rusqlite::ToSql> = params.iter().map(|p| p.as_ref()).collect();
            conn.execute(&query, param_refs.as_slice())?;
        }

        // Update tags if provided
        if let Some(tags) = &updates.tags {
            self.update_todo_tags(&conn, id, tags)?;
        }

        drop(conn);
        Ok(self.get_todo_by_id(id).ok())
    }

    /// Delete a todo
    pub fn delete_todo(&self, id: &str) -> Result<bool> {
        let conn = self.conn.lock().unwrap();

        let rows_affected = conn.execute("DELETE FROM todos WHERE id = ?", params![id])?;

        Ok(rows_affected > 0)
    }

    /// Bulk update todo status
    pub fn bulk_update_status(&self, ids: &[String], status: &TodoStatus) -> Result<usize> {
        let conn = self.conn.lock().unwrap();
        let status_str = status_to_string(status);

        let placeholders = ids.iter().map(|_| "?").collect::<Vec<_>>().join(",");
        let query = format!("UPDATE todos SET status = ? WHERE id IN ({})", placeholders);

        let mut params: Vec<Box<dyn rusqlite::ToSql>> = vec![Box::new(status_str)];
        for id in ids {
            params.push(Box::new(id.clone()));
        }

        let param_refs: Vec<&dyn rusqlite::ToSql> = params.iter().map(|p| p.as_ref()).collect();
        let rows_affected = conn.execute(&query, param_refs.as_slice())?;

        Ok(rows_affected)
    }

    // ===== CATEGORY OPERATIONS =====

    /// Create a new category
    pub fn create_category(&self, category: &Category) -> Result<Category> {
        let conn = self.conn.lock().unwrap();

        conn.execute(
            "INSERT INTO categories (name, color, icon, description) VALUES (?1, ?2, ?3, ?4)",
            params![category.name, category.color, category.icon, None::<String>],
        )?;

        let id = conn.last_insert_rowid();

        let mut stmt = conn.prepare(
            "SELECT id, name, color, icon, created_at FROM categories WHERE id = ?"
        )?;

        stmt.query_row(params![id], |row| {
            Ok(Category {
                id: row.get::<_, String>(0)?,
                name: row.get(1)?,
                color: row.get(2)?,
                icon: row.get(3)?,
                created_at: parse_datetime(&row.get::<_, String>(4)?).unwrap_or_else(|| Utc::now()),
            })
        }).map_err(Into::into)
    }

    /// Get all categories
    pub fn get_all_categories(&self) -> Result<Vec<Category>> {
        let conn = self.conn.lock().unwrap();

        let mut stmt = conn.prepare(
            "SELECT id, name, color, icon, created_at FROM categories ORDER BY name"
        )?;

        let categories = stmt.query_map([], |row| {
            Ok(Category {
                id: row.get::<_, String>(0)?,
                name: row.get(1)?,
                color: row.get(2)?,
                icon: row.get(3)?,
                created_at: parse_datetime(&row.get::<_, String>(4)?).unwrap_or_else(|| Utc::now()),
            })
        })?
        .collect::<Result<Vec<_>, _>>()?;

        Ok(categories)
    }

    // ===== TAG OPERATIONS =====

    /// Create or get a tag
    pub fn create_or_get_tag(&self, name: &str, color: Option<String>) -> Result<Tag> {
        let conn = self.conn.lock().unwrap();

        // Try to get existing tag
        let mut stmt = conn.prepare("SELECT id, name, color, created_at FROM tags WHERE name = ?")?;

        if let Ok(tag) = stmt.query_row(params![name], |row| {
            Ok(Tag {
                id: row.get::<_, String>(0)?,
                name: row.get(1)?,
                color: row.get(2)?,
                usage_count: 0, // Will be counted separately
            })
        }) {
            return Ok(tag);
        }

        // Create new tag
        conn.execute(
            "INSERT INTO tags (name, color) VALUES (?1, ?2)",
            params![name, color],
        )?;

        let id = conn.last_insert_rowid();

        Ok(Tag {
            id: format!("{}", id),
            name: name.to_string(),
            color,
            usage_count: 0,
        })
    }

    /// Get all tags with usage counts
    pub fn get_all_tags(&self) -> Result<Vec<Tag>> {
        let conn = self.conn.lock().unwrap();

        let mut stmt = conn.prepare(
            "SELECT
                t.id, t.name, t.color,
                COUNT(tt.todo_id) as usage_count
            FROM tags t
            LEFT JOIN todo_tags tt ON t.id = tt.tag_id
            GROUP BY t.id
            ORDER BY usage_count DESC, t.name"
        )?;

        let tags = stmt.query_map([], |row| {
            Ok(Tag {
                id: row.get::<_, i64>(0)?.to_string(),
                name: row.get(1)?,
                color: row.get(2)?,
                usage_count: row.get::<_, i64>(3)? as usize,
            })
        })?
        .collect::<Result<Vec<_>, _>>()?;

        Ok(tags)
    }

    // ===== STATISTICS =====

    /// Get todo statistics
    pub fn get_stats(&self) -> Result<TodoStats> {
        let conn = self.conn.lock().unwrap();

        let mut stmt = conn.prepare(
            "SELECT
                COUNT(*) as total,
                SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
                SUM(CASE WHEN status = 'in_progress' THEN 1 ELSE 0 END) as in_progress,
                SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed,
                SUM(CASE WHEN status = 'blocked' THEN 1 ELSE 0 END) as blocked,
                SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END) as cancelled,
                SUM(CASE WHEN priority = 'low' THEN 1 ELSE 0 END) as low_priority,
                SUM(CASE WHEN priority = 'medium' THEN 1 ELSE 0 END) as medium_priority,
                SUM(CASE WHEN priority = 'high' THEN 1 ELSE 0 END) as high_priority,
                SUM(CASE WHEN priority = 'critical' THEN 1 ELSE 0 END) as critical_priority,
                SUM(CASE WHEN status NOT IN ('completed', 'cancelled') AND due_date IS NOT NULL AND due_date < datetime('now') THEN 1 ELSE 0 END) as overdue
            FROM todos"
        )?;

        let stats = stmt.query_row([], |row| {
            Ok(TodoStats {
                total: row.get::<_, i64>(0)? as usize,
                pending: row.get::<_, i64>(1)? as usize,
                in_progress: row.get::<_, i64>(2)? as usize,
                completed: row.get::<_, i64>(3)? as usize,
                blocked: row.get::<_, i64>(4)? as usize,
                cancelled: row.get::<_, i64>(5)? as usize,
                by_priority: PriorityStats {
                    low: row.get::<_, i64>(6)? as usize,
                    medium: row.get::<_, i64>(7)? as usize,
                    high: row.get::<_, i64>(8)? as usize,
                    critical: row.get::<_, i64>(9)? as usize,
                },
                overdue: row.get::<_, i64>(10)? as usize,
            })
        })?;

        Ok(stats)
    }

    // ===== HELPER METHODS =====

    fn row_to_todo(&self, row: &Row) -> rusqlite::Result<Todo> {
        let metadata_str: Option<String> = row.get(14)?;
        let metadata = metadata_str.and_then(|s| serde_json::from_str(&s).ok());

        Ok(Todo {
            id: row.get::<_, i64>(0)?.to_string(),
            title: row.get(1)?,
            description: row.get(2)?,
            status: string_to_status(&row.get::<_, String>(3)?).unwrap_or(TodoStatus::Pending),
            priority: string_to_priority(&row.get::<_, String>(4)?).unwrap_or(TodoPriority::Medium),
            assignee: row.get(5)?,
            category: row.get(6)?,
            parent_id: row.get::<_, Option<i64>>(7)?.map(|id| id.to_string()),
            created_at: parse_datetime(&row.get::<_, String>(8)?).unwrap_or_else(|| Utc::now()),
            updated_at: parse_datetime(&row.get::<_, String>(9)?).unwrap_or_else(|| Utc::now()),
            completed_at: row.get::<_, Option<String>>(10)?.and_then(|s| parse_datetime(&s)),
            due_date: row.get::<_, Option<String>>(11)?.and_then(|s| parse_datetime(&s)),
            estimated_hours: row.get(12)?,
            actual_hours: row.get(13)?,
            metadata,
            tags: Vec::new(), // Tags are loaded separately
        })
    }

    fn get_tags_for_todo(&self, conn: &Connection, todo_id: &str) -> Result<Vec<String>> {
        let mut stmt = conn.prepare(
            "SELECT t.name
             FROM tags t
             JOIN todo_tags tt ON t.id = tt.tag_id
             WHERE tt.todo_id = ?"
        )?;

        let tags = stmt.query_map(params![todo_id], |row| {
            row.get::<_, String>(0)
        })?
        .collect::<Result<Vec<_>, _>>()?;

        Ok(tags)
    }

    fn update_todo_tags(&self, conn: &Connection, todo_id: &str, tags: &[String]) -> Result<()> {
        // Remove existing tags
        conn.execute("DELETE FROM todo_tags WHERE todo_id = ?", params![todo_id])?;

        // Add new tags
        for tag_name in tags {
            let tag = self.create_or_get_tag(tag_name, None)?;
            conn.execute(
                "INSERT INTO todo_tags (todo_id, tag_id) VALUES (?1, ?2)",
                params![todo_id, tag.id],
            )?;
        }

        Ok(())
    }

    fn get_category_id_by_name(&self, conn: &Connection, category: &Option<String>) -> Result<Option<i64>> {
        if let Some(name) = category {
            let mut stmt = conn.prepare("SELECT id FROM categories WHERE name = ?")?;
            if let Ok(id) = stmt.query_row(params![name], |row| row.get::<_, i64>(0)) {
                return Ok(Some(id));
            }
        }
        Ok(None)
    }

    fn todo_exists(&self, conn: &Connection, id: &str) -> Result<bool> {
        let mut stmt = conn.prepare("SELECT 1 FROM todos WHERE id = ?")?;
        Ok(stmt.exists(params![id])?)
    }
}

// ===== CONVERSION HELPERS =====

fn status_to_string(status: &TodoStatus) -> String {
    match status {
        TodoStatus::Pending => "pending".to_string(),
        TodoStatus::InProgress => "in_progress".to_string(),
        TodoStatus::Completed => "completed".to_string(),
        TodoStatus::Blocked => "blocked".to_string(),
        TodoStatus::Cancelled => "cancelled".to_string(),
    }
}

fn string_to_status(s: &str) -> Option<TodoStatus> {
    match s {
        "pending" => Some(TodoStatus::Pending),
        "in_progress" => Some(TodoStatus::InProgress),
        "completed" => Some(TodoStatus::Completed),
        "blocked" => Some(TodoStatus::Blocked),
        "cancelled" => Some(TodoStatus::Cancelled),
        _ => None,
    }
}

fn priority_to_string(priority: &TodoPriority) -> String {
    match priority {
        TodoPriority::Low => "low".to_string(),
        TodoPriority::Medium => "medium".to_string(),
        TodoPriority::High => "high".to_string(),
        TodoPriority::Critical => "critical".to_string(),
    }
}

fn string_to_priority(s: &str) -> Option<TodoPriority> {
    match s {
        "low" => Some(TodoPriority::Low),
        "medium" => Some(TodoPriority::Medium),
        "high" => Some(TodoPriority::High),
        "critical" => Some(TodoPriority::Critical),
        _ => None,
    }
}

fn parse_datetime(s: &str) -> Option<DateTime<Utc>> {
    DateTime::parse_from_rfc3339(s)
        .ok()
        .map(|dt| dt.with_timezone(&Utc))
}
