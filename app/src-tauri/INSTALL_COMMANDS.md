# Package Installation Commands to Add

Add these to main.rs after the existing commands:

```rust
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
```

Add to `tauri::Builder` invoke_handler list:
- check_package_installed
- install_package_remote
- get_packages_status_remote
