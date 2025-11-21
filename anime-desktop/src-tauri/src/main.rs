// Prevents additional console window on Windows in release
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use anime_desktop::{get_packages, Package, Server};
use anime_desktop::lambda::commands::LambdaState;
use anime_desktop::server::ServerState;
use anime_desktop::terminal::TerminalState;
use anime_desktop::models::{ModelManagerState, ModelDownloadManager};
use anime_desktop::animation::AnimationState;
use anime_desktop::comfyui::commands::ComfyUIState;
use anime_desktop::todos::commands::TodoState;
use std::sync::{Arc, Mutex};
use std::collections::HashMap;
use tokio::sync::Mutex as TokioMutex;

// Tauri command to get all packages
#[tauri::command]
fn get_packages_command() -> Vec<Package> {
    get_packages()
}

// Tauri command to resolve dependencies (stub - not implemented yet)
#[tauri::command]
fn resolve_dependencies_command(package_ids: Vec<String>) -> Result<Vec<Package>, String> {
    let all_packages = get_packages();
    // For now, just return the requested packages without dependency resolution
    let requested_packages: Vec<Package> = all_packages
        .into_iter()
        .filter(|p| package_ids.contains(&p.id))
        .collect();
    Ok(requested_packages)
}

// Tauri command to get package by ID
#[tauri::command]
fn get_package(package_id: String) -> Result<Package, String> {
    let packages = get_packages();
    packages
        .into_iter()
        .find(|p| p.id == package_id)
        .ok_or_else(|| format!("Package '{}' not found", package_id))
}

// Tauri command to start installation (stub)
#[tauri::command]
async fn install_packages(package_ids: Vec<String>) -> Result<String, String> {
    // TODO: Implement actual installation logic

    Ok(format!("Installing {} packages", package_ids.len()))
}

// Package installation commands
#[tauri::command]
async fn check_package_installed(
    host: String,
    username: String,
    package_id: String,
) -> Result<bool, String> {
    anime_desktop::check_package_installed(&host, &username, &package_id)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn install_package_remote(
    app: tauri::AppHandle,
    host: String,
    username: String,
    package_id: String,
) -> Result<(), String> {
    anime_desktop::install_package(app, &host, &username, &package_id)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn get_packages_status_remote(
    host: String,
    username: String,
    package_ids: Vec<String>,
) -> Result<Vec<anime_desktop::PackageStatus>, String> {
    anime_desktop::get_packages_status(&host, &username, &package_ids)
        .await
        .map_err(|e| e.to_string())
}

// Tauri command to get servers (stub)
#[tauri::command]
fn get_servers() -> Vec<Server> {
    // TODO: Load from config file
    vec![]
}

// Tauri command to add server (stub)
#[tauri::command]
fn add_server(server: Server) -> Result<Server, String> {
    // TODO: Save to config and test connection
    Ok(server)
}

// Tauri command to test server connection (stub)
#[tauri::command]
async fn test_server_connection(server_id: String) -> Result<String, String> {
    // TODO: Implement SSH connection test
    Ok(format!("Testing connection to {}", server_id))
}

fn main() {
    // Initialize model manager
    let model_manager = ModelDownloadManager::new()
        .expect("Failed to initialize model download manager");

    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_store::Builder::new().build())
        .plugin(tauri_plugin_dialog::init())
        .manage(LambdaState {
            client: Mutex::new(None),
        })
        .manage(ServerState {
            connections: Mutex::new(HashMap::new()),
        })
        .manage(TerminalState {
            sessions: Mutex::new(HashMap::new()),
        })
        .manage(ModelManagerState {
            manager: Arc::new(TokioMutex::new(model_manager)),
        })
        .manage(AnimationState::new())
        .manage(ComfyUIState::default())
        .manage(TodoState::new())
        .invoke_handler(tauri::generate_handler![
            get_packages_command,
            resolve_dependencies_command,
            get_package,
            install_packages,
            check_package_installed,
            install_package_remote,
            get_packages_status_remote,
            get_servers,
            add_server,
            test_server_connection,
            // Lambda commands
            anime_desktop::set_lambda_api_key,
            anime_desktop::load_lambda_api_key,
            anime_desktop::check_lambda_connection,
            anime_desktop::lambda_list_instances,
            anime_desktop::lambda_list_instance_types,
            anime_desktop::lambda_launch_instance,
            anime_desktop::lambda_terminate_instances,
            anime_desktop::lambda_restart_instances,
            anime_desktop::lambda_list_ssh_keys,
            anime_desktop::lambda_add_ssh_key,
            anime_desktop::lambda_list_file_systems,
            // Server monitoring commands
            anime_desktop::connect_to_server,
            anime_desktop::disconnect_from_server,
            anime_desktop::get_server_status,
            anime_desktop::is_server_connected,
            anime_desktop::list_connected_servers,
            anime_desktop::find_ssh_keys,
            anime_desktop::validate_ssh_key,
            // Terminal commands
            anime_desktop::terminal_connect,
            anime_desktop::terminal_connect_local,
            anime_desktop::terminal_input,
            anime_desktop::terminal_disconnect,
            anime_desktop::terminal_resize,
            // Model management commands
            anime_desktop::download_model,
            anime_desktop::list_installed_models,
            anime_desktop::delete_model,
            anime_desktop::load_model,
            anime_desktop::cancel_model_download,
            anime_desktop::get_model_info,
            // Animation commands
            anime_desktop::get_animation_models,
            anime_desktop::get_animation_style_presets,
            anime_desktop::submit_generation_job,
            anime_desktop::get_generation_jobs,
            anime_desktop::get_generation_job,
            anime_desktop::cancel_generation_job,
            anime_desktop::retry_generation_job,
            anime_desktop::delete_generation_job,
            // Creative tools - Writing commands
            anime_desktop::list_documents,
            anime_desktop::create_document,
            anime_desktop::save_document,
            anime_desktop::delete_document,
            anime_desktop::export_document,
            anime_desktop::generate_text,
            // Creative tools - Analysis commands
            anime_desktop::read_file,
            anime_desktop::analyze_content,
            anime_desktop::export_analysis,
            // Creative tools - Storyboard commands
            anime_desktop::create_storyboard_project,
            anime_desktop::parse_script,
            // ComfyUI workflow commands
            anime_desktop::comfyui_set_connection,
            anime_desktop::comfyui_check_connection,
            anime_desktop::comfyui_get_status,
            anime_desktop::comfyui_list_workflows,
            anime_desktop::comfyui_get_workflow,
            anime_desktop::comfyui_execute_workflow,
            anime_desktop::comfyui_get_execution,
            anime_desktop::comfyui_list_executions,
            anime_desktop::comfyui_cancel_execution,
            anime_desktop::comfyui_interrupt,
            anime_desktop::comfyui_clear_queue,
            anime_desktop::comfyui_upload_image,
            anime_desktop::generate_shot_suggestions,
            anime_desktop::generate_storyboard_image,
            anime_desktop::export_storyboard,
            // Todo system commands
            anime_desktop::create_todo,
            anime_desktop::get_todos,
            anime_desktop::get_todo_by_id,
            anime_desktop::update_todo,
            anime_desktop::delete_todo,
            anime_desktop::bulk_update_todos,
            anime_desktop::get_todo_stats,
            anime_desktop::search_todos,
            anime_desktop::create_category,
            anime_desktop::get_categories,
            anime_desktop::add_tags_to_todo,
            anime_desktop::get_tags,
            anime_desktop::create_tag,
            anime_desktop::seed_initial_todos,
            anime_desktop::is_seeded,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
