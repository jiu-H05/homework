# -*- coding: utf-8 -*-
"""用系统 LoadLibrary 逐个解析 PE 导入（含 API 集），报告真正无法加载的模块。"""
import os
import sys
import ctypes
from ctypes import wintypes
import pefile

kernel32 = ctypes.WinDLL("kernel32", use_last_error=True)
LoadLibraryEx = kernel32.LoadLibraryExW
LoadLibraryEx.restype = wintypes.HMODULE
LoadLibraryEx.argtypes = [wintypes.LPCWSTR, wintypes.HANDLE, wintypes.DWORD]
FreeLibrary = kernel32.FreeLibrary
FreeLibrary.argtypes = [wintypes.HMODULE]

LOAD_WITH_ALTERED_SEARCH_PATH = 0x8


def try_load(name, hint_dir=None):
    h = LoadLibraryEx(name, None, LOAD_WITH_ALTERED_SEARCH_PATH)
    if not h:
        err = ctypes.get_last_error()
        return False, err
    FreeLibrary(h)
    return True, 0


def imports_of(path):
    pe = pefile.PE(path, fast_load=True)
    pe.parse_data_directories(directories=[pefile.DIRECTORY_ENTRY['IMAGE_DIRECTORY_ENTRY_IMPORT']])
    out = []
    if hasattr(pe, "DIRECTORY_ENTRY_IMPORT"):
        for e in pe.DIRECTORY_ENTRY_IMPORT:
            out.append(e.dll.decode(errors="ignore"))
    return out


start = sys.argv[1]
bin_dir = os.path.dirname(start)

print("== 直接加载目标 ==")
ok, err = try_load(start)
print("Load rustc.exe:", ok, "err", err)

print("\n== 逐个加载 rustc.exe 的导入 ==")
for dll in imports_of(start):
    full = os.path.join(bin_dir, dll)
    candidate = full if os.path.exists(full) else dll
    ok, err = try_load(candidate)
    print(("OK  " if ok else "FAIL") + " %-45s %s" % (dll, "" if ok else "err=%d" % err))

# 深入 rustc_driver
driver = [f for f in os.listdir(bin_dir) if f.startswith("rustc_driver")][0]
print("\n== 逐个加载 rustc_driver 的导入 ==")
for dll in imports_of(os.path.join(bin_dir, driver)):
    full = os.path.join(bin_dir, dll)
    candidate = full if os.path.exists(full) else dll
    ok, err = try_load(candidate)
    print(("OK  " if ok else "FAIL") + " %-45s %s" % (dll, "" if ok else "err=%d" % err))
