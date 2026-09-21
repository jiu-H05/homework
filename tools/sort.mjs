// sort.js — 图书排序工具：使用 ICU Collator，中文按拼音首字母、英文按字母，支持数字感知。
// 在 WebView2（Chromium ICU）下，new Intl.Collator('zh-Hans-CN') 会按拼音对汉字排序，
// 从而实现“书名首字母 A→Z”，便于读者按首字母检索借阅。

const collator = new Intl.Collator(["zh-Hans-CN", "zh-Hans", "en"], {
  sensitivity: "base",
  numeric: true,
});

export const BOOK_SORTS = [
  { value: "default", label: "默认排序" },
  { value: "title-asc", label: "书名首字母 A→Z" },
  { value: "title-desc", label: "书名 Z→A" },
  { value: "available-desc", label: "可借数量优先" },
];

// sortBooks 返回排序后的新数组，不修改原数组。
export function sortBooks(books, mode) {
  const rows = books.slice();
  switch (mode) {
    case "title-asc":
      rows.sort((a, b) => collator.compare(a.title, b.title));
      break;
    case "title-desc":
      rows.sort((a, b) => collator.compare(b.title, a.title));
      break;
    case "available-desc":
      rows.sort(
        (a, b) =>
          b.availableQty - a.availableQty || collator.compare(a.title, b.title)
      );
      break;
    default:
      break;
  }
  return rows;
}

// titleFirstLetter 取书名首字符并大写，用于首字母分组/展示。
export function titleFirstLetter(title) {
  const ch = (title || "").trim().charAt(0);
  return ch ? ch.toUpperCase() : "#";
}
