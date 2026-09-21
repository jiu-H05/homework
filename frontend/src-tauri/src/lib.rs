// 桌面壳：前后端分离后，它是纯客户端——只负责窗口与 WebView，
// 后端跑在另一台（或本机）服务器上，地址由用户在"服务器设置"里填写。
use tauri::{Builder};

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    Builder::default()
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_http::init())
        .build(tauri::generate_context!())
        .expect("error while building tauri application")
        .run(|_app, _event| {});
}
