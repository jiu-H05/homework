# -*- coding: utf-8 -*-
"""
认证模块：口令 PBKDF2-SHA256 加盐哈希、校验。
"""
import hashlib
import os

ITERATIONS = 100_000


def hash_password(password: str, salt: str = None) -> str:
    """返回 "salt$hash" 形式的口令摘要。"""
    if salt is None:
        salt = os.urandom(16).hex()
    digest = hashlib.pbkdf2_hmac("sha256", password.encode("utf-8"), salt.encode("utf-8"), ITERATIONS).hex()
    return f"{salt}${digest}"


def verify_password(password: str, stored: str) -> bool:
    """校验明文口令与存储的摘要是否一致。"""
    try:
        salt, _ = stored.split("$", 1)
    except ValueError:
        return False
    return hash_password(password, salt) == stored
