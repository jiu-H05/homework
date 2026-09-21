// logs.js — 管理员操作日志查看。
import { api } from "../api.js";
import { loading, renderTable } from "../components.js";

export async function renderLogs(container) {
  loading(container);
  const logs = await api.logs();
  container.innerHTML = `<div class="card"><h3 class="card-title">最近操作日志</h3><div id="table"></div></div>`;
  renderTable(
    container.querySelector("#table"),
    [
      { key: "createdAt", title: "时间" },
      { key: "username", title: "操作用户" },
      { key: "action", title: "操作类型" },
      { key: "detail", title: "详情" },
    ],
    logs
  );
}
