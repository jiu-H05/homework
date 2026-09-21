// 验证前端真实排序模块 sort.mjs（即 frontend/src/js/sort.js 的副本）在 127 本书上的行为。
import { sortBooks } from "./sort.mjs";

const base = "http://127.0.0.1:8765";
const login = await fetch(`${base}/api/auth/login`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ username: "admin", password: "admin123" }),
}).then((r) => r.json());
const books = await fetch(`${base}/api/books?page=1&page_size=200`, {
  headers: { Authorization: `Bearer ${login.token}` },
}).then((r) => r.json());

console.log("fetched books:", books.length);

const collator = new Intl.Collator(["zh-Hans-CN", "zh-Hans", "en"], {
  sensitivity: "base",
  numeric: true,
});

// 1) title-asc 必须单调非降
const asc = sortBooks(books, "title-asc").map((b) => b.title);
let violations = 0;
for (let i = 1; i < asc.length; i++) {
  if (collator.compare(asc[i - 1], asc[i]) > 0) violations++;
}
console.log("title-asc monotonic violations:", violations);
console.log("A->Z first 20:", asc.slice(0, 20).join(" | "));
console.log("A->Z last 8:", asc.slice(-8).join(" | "));

// 2) title-desc 必须是 asc 的逆序
const desc = sortBooks(books, "title-desc").map((b) => b.title);
const descOk = desc.every((t, i) => t === asc[asc.length - 1 - i]);
console.log("title-desc is reverse of asc:", descOk);

// 3) available-desc：可借数量非增
const byAvail = sortBooks(books, "available-desc");
let avViol = 0;
for (let i = 1; i < byAvail.length; i++) {
  if (byAvail[i - 1].availableQty < byAvail[i].availableQty) avViol++;
}
console.log("available-desc violations:", avViol, "| top3:", byAvail.slice(0, 3).map((b) => `${b.title}(${b.availableQty})`).join(", "));

// 4) default 不改变顺序
const def = sortBooks(books, "default");
console.log("default preserves order:", def.every((b, i) => b === books[i]));
