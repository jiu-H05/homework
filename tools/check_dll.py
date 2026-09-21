# -*- coding: utf-8 -*-
"""递归检查 PE 导入表，报告无法定位的 DLL。"""
import os
import sys
import pefile

start = sys.argv[1]
search_dirs = [
    os.path.dirname(start),
    r"C:\Windows\System32",
    r"C:\Windows\SysWOW64",
    r"C:\Windows",
]
search_dirs += os.environ.get("PATH", "").split(os.pathsep)

seen = set()
missing = set()


def locate(name):
    for d in search_dirs:
        if d and os.path.exists(os.path.join(d, name)):
            return os.path.join(d, name)
    return None


def walk(path, depth=0):
    if path.lower() in seen:
        return
    seen.add(path.lower())
    try:
        pe = pefile.PE(path, fast_load=True)
        pe.parse_data_directories(directories=[pefile.DIRECTORY_ENTRY['IMAGE_DIRECTORY_ENTRY_IMPORT']])
    except Exception as e:
        print("无法解析", path, e)
        return
    if not hasattr(pe, "DIRECTORY_ENTRY_IMPORT"):
        return
    for entry in pe.DIRECTORY_ENTRY_IMPORT:
        dll = entry.dll.decode(errors="ignore")
        if dll.lower().startswith("api-ms-win-") or dll.lower().startswith("ext-ms-"):
            continue  # API 虚拟集，由系统加载器解析，无实体文件
        p = locate(dll)
        if not p:
            missing.add((dll, os.path.basename(path)))
            print("缺失: %-40s （被 %s 需要）" % (dll, os.path.basename(path)))
        else:
            walk(p, depth + 1)


walk(start)
print("\n共检查 %d 个 DLL，缺失 %d 个。" % (len(seen), len(missing)))
