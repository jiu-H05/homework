// 桌面壳：负责生命周期与 Go 后端 sidecar 的拉起、就绪探测与退出回收。
// 业务逻辑全部位于 Go 后端与 Web UI，Rust 侧仅做进程编排，保持高内聚低耦合。
use std::sync::Mutex;
use std::time::{Duration, Instant};
use tauri::{Manager, RunEvent};
use tauri_plugin_shell::process::CommandChild;
use tauri_plugin_shell::ShellExt;

// SERVER_ADDR 用于就绪探测；BIND_ADDR 让后端监听所有网卡，
// 前端通过本机私网 IP 访问，避开 WebView2 对环回地址(loopback)的限制。
const SERVER_ADDR: &str = "127.0.0.1:8765";
const BIND_ADDR: &str = "0.0.0.0:8765";

// ChildHolder 持有 sidecar 子进程句柄，便于退出时回收。
struct ChildHolder(Mutex<Option<CommandChild>>);

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_http::init())
        .setup(|app| {
            // 数据目录：生产环境下 sidecar 位于临时资源目录，数据库需放到持久化的 AppData。
            let data_dir = app.path().app_data_dir().expect("resolve app data dir");
            std::fs::create_dir_all(&data_dir).ok();
            let db_path = data_dir.join("library_data.db");

            // 拉起 Go 后端 sidecar，并通过环境变量注入监听地址、数据库路径与签名密钥。
            let sidecar = app
                .shell()
                .sidecar("lms-server")
                .expect("resolve sidecar binary")
                .env("LMS_ADDR", BIND_ADDR)
                .env("LMS_DB", db_path.to_string_lossy().to_string())
                .env("LMS_SECRET", "lms-local-secret-2026");

            let (mut rx, child) = sidecar.spawn().expect("spawn sidecar");

            // 后台排空 sidecar 输出事件，避免管道阻塞。
            tauri::async_runtime::spawn(async move {
                while let Some(_event) = rx.recv().await {}
            });

            app.manage(ChildHolder(Mutex::new(Some(child))));

            // 后台线程轮询端口就绪，仅用于日志，不阻塞窗口显示；UI 侧会自行重试。
            let addr = SERVER_ADDR.to_string();
            std::thread::spawn(move || {
                let start = Instant::now();
                while start.elapsed() < Duration::from_secs(20) {
                    if std::net::TcpStream::connect(&addr).is_ok() {
                        println!("backend ready at {}", addr);
                        break;
                    }
                    std::thread::sleep(Duration::from_millis(300));
                }
            });

            // E2E 自动化通道（仅当设置 LMS_E2E 环境变量时启用）：
            // 轮询临时命令文件，把其中 JS eval 进当前 webview，随后删除文件。
            // 用于在无法注入物理按键的环境下驱动真实界面逻辑。
            if std::env::var("LMS_E2E").is_ok() {
                if let Some(win) = app.get_webview_window("main") {
                    std::thread::spawn(move || loop {
                        let cmd = std::env::temp_dir().join("lms_e2e_cmd.js");
                        if let Ok(js) = std::fs::read_to_string(&cmd) {
                            let _ = win.eval(&js);
                            let _ = std::fs::remove_file(&cmd);
                        }
                        std::thread::sleep(Duration::from_millis(250));
                    });
                }
            }

            Ok(())
        })
        .build(tauri::generate_context!())
        .expect("error while building tauri application")
        .run(|app, event| {
            // 应用退出时终止 Go 后端，避免残留进程。
            if let RunEvent::ExitRequested { .. } = event {
                if let Some(holder) = app.try_state::<ChildHolder>() {
                    if let Ok(mut guard) = holder.0.lock() {
                        if let Some(child) = guard.take() {
                            let _ = child.kill();
                        }
                    }
                }
            }
        });
}
