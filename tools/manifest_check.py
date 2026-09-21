# -*- coding: utf-8 -*-
"""提取 PE 嵌入的 RT_MANIFEST 与导入的 SxS 程序集。"""
import sys
import pefile

path = sys.argv[1]
pe = pefile.PE(path)
RT_MANIFEST = 24
found = False
if hasattr(pe, "DIRECTORY_ENTRY_RESOURCE"):
    for entry in pe.DIRECTORY_ENTRY_RESOURCE.entries:
        rid = entry.id
        name = pefile.RESOURCE_TYPE.get(rid, rid)
        if rid == RT_MANIFEST:
            found = True
            for sub in entry.directory.entries:
                for lang in sub.directory.entries:
                    data = pe.get_data(lang.data.struct.OffsetToData, lang.data.struct.Size)
                    print("--- MANIFEST ---")
                    try:
                        print(data.decode("utf-8"))
                    except UnicodeDecodeError:
                        print(data.decode("utf-16"))
if not found:
    print("无 RT_MANIFEST")
