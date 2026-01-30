use super::types::*;
use super::db::TodoDatabase;
use std::sync::{Arc, Mutex};
use std::path::PathBuf;
use anyhow::Result;

#[derive(Clone)]
pub struct TodoStorage {
    db: Arc<Mutex<TodoDatabase>>,
}

impl TodoStorage {
    pub fn new() -> Self {
        let db = Self::initialize_database().expect("Failed to initialize database");
        Self {
            db: Arc::new(Mutex::new(db)),
        }
    }

    fn initialize_database() -> Result<TodoDatabase> {
        let home = std::env::var("HOME")
            .or_else(|_| std::env::var("USERPROFILE"))
            .expect("Could not determine home directory");

        let data_dir = PathBuf::from(home).join(".anime-desktop");
        std::fs::create_dir_all(&data_dir)?;

        let db_path = data_dir.join("todos.db");
        TodoDatabase::new(db_path)
    }

    pub fn insert_todo(&self, todo: Todo) -> Result<Todo, String> {
        let db = self.db.lock().unwrap();
        db.create_todo(&todo).map_err(|e| e.to_string())
    }

    pub fn get_todo(&self, id: &str) -> Result<Option<Todo>, String> {
        let db = self.db.lock().unwrap();
        match db.get_todo_by_id(id) {
            Ok(todo) => Ok(Some(todo)),
            Err(_) => Ok(None),
        }
    }

    pub fn get_all_todos(&self) -> Result<Vec<Todo>, String> {
        let db = self.db.lock().unwrap();
        let filters = TodoFilters {
            status: None,
            priority: None,
            category: None,
            assignee: None,
            search_query: None,
            tags: None,
            include_subtasks: Some(true),
        };
        db.get_todos(&filters).map_err(|e| e.to_string())
    }

    pub fn update_todo(&self, id: &str, updates: TodoUpdate) -> Result<Option<Todo>, String> {
        let db = self.db.lock().unwrap();
        db.update_todo(id, &updates).map_err(|e| e.to_string())
    }

    pub fn delete_todo(&self, id: &str) -> Result<bool, String> {
        let db = self.db.lock().unwrap();
        db.delete_todo(id).map_err(|e| e.to_string())
    }

    pub fn bulk_update_status(&self, ids: Vec<String>, status: TodoStatus) -> Result<usize, String> {
        let db = self.db.lock().unwrap();
        db.bulk_update_status(&ids, &status).map_err(|e| e.to_string())
    }

    pub fn insert_category(&self, category: Category) -> Result<Category, String> {
        let db = self.db.lock().unwrap();
        db.create_category(&category).map_err(|e| e.to_string())
    }

    pub fn get_all_categories(&self) -> Result<Vec<Category>, String> {
        let db = self.db.lock().unwrap();
        db.get_all_categories().map_err(|e| e.to_string())
    }

    pub fn insert_tag(&self, tag: Tag) -> Result<Tag, String> {
        let db = self.db.lock().unwrap();
        db.create_or_get_tag(&tag.name, tag.color).map_err(|e| e.to_string())
    }

    pub fn get_all_tags(&self) -> Result<Vec<Tag>, String> {
        let db = self.db.lock().unwrap();
        db.get_all_tags().map_err(|e| e.to_string())
    }

    pub fn increment_tag_usage(&self, _tag_name: &str) -> Result<(), String> {
        // Tag usage is now automatically tracked by the database via todo_tags junction table
        Ok(())
    }

    pub fn get_todos_filtered(&self, filters: &TodoFilters) -> Result<Vec<Todo>, String> {
        let db = self.db.lock().unwrap();
        db.get_todos(filters).map_err(|e| e.to_string())
    }

    pub fn get_stats(&self) -> Result<TodoStats, String> {
        let db = self.db.lock().unwrap();
        db.get_stats().map_err(|e| e.to_string())
    }
}

impl Default for TodoStorage {
    fn default() -> Self {
        Self::new()
    }
}
