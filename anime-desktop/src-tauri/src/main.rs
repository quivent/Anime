// Prevents additional console window on Windows in release
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use anime_desktop::{get_packages, resolve_dependencies, Package, Server};
use anime_desktop::lambda::commands::LambdaState;
use anime_desktop::server::ServerState;
use std::sync::Mutex;
use std::collections::HashMap;

// Tauri command to get all packages
#[tauri::command]
fn get_packages_command() -> Vec<Package> {
    get_packages()
}

// Tauri command to resolve dependencies
#[tauri::command]
fn resolve_dependencies_command(package_ids: Vec<String>) -> Result<Vec<Package>, String> {
    let all_packages = get_packages();
    resolve_dependencies(package_ids, &all_packages)
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
    println!("Installing packages: {:?}", package_ids);
    Ok(format!("Installing {} packages", package_ids.len()))
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
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .manage(LambdaState {
            client: Mutex::new(None),
        })
        .manage(ServerState {
            connections: Mutex::new(HashMap::new()),
        })
        .invoke_handler(tauri::generate_handler![
            get_packages_command,
            resolve_dependencies_command,
            get_package,
            install_packages,
            get_servers,
            add_server,
            test_server_connection,
            // Lambda commands
            anime_desktop::set_lambda_api_key,
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
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
