#!/usr/bin/env python3
"""
parity_random_100.py

生成 100 个随机经纬度点，对比：
 1. Python reverse_geocoder 库本地查询结果
 2. 已运行的 Go HTTP 服务 /reverse 接口返回

要求：name, admin1, admin2, cc 四个字段全部一致，否则视为失败。

使用前提：
 - 已安装: pip install reverse_geocoder
 - Go 服务已在 localhost:8080 运行 (可通过 test.bat 启动或单独运行 rgeocoder)
 - 使用的 cities 数据应保持一致。

运行：
  python parity_random_100.py

退出码:
  0 -> 全部匹配
  1 -> 有不匹配或错误
"""
import json
import math
import os
import random
import sys
import time
from typing import Dict, Tuple

import urllib.request
import urllib.error

try:
    import reverse_geocoder as rg
except ImportError:
    print("ERROR: missing reverse_geocoder. Install with: pip install reverse_geocoder", file=sys.stderr)
    sys.exit(1)

GO_SERVER = os.environ.get("GO_SERVER", "http://localhost:8080")
SAMPLES = 100
FIELDS = ["name", "admin1", "admin2", "cc"]
RANDOM_SEED = int(os.environ.get("PARITY_SEED", "20250928"))
random.seed(RANDOM_SEED)

def rand_coord() -> Tuple[float,float]:
    # Uniform latitude distribution (not area-corrected) is acceptable for parity test
    lat = random.uniform(-89.5, 89.5)  # avoid exact poles (where some kd-tree may have edge behavior)
    lon = random.uniform(-179.5, 179.5)
    return round(lat, 6), round(lon, 6)

def query_python(lat: float, lon: float) -> Dict:
    # reverse_geocoder expects list of tuples
    res = rg.search([(lat, lon)], mode=1)  # mode=1 single-thread for determinism
    if not res:
        return {}
    return res[0]

def query_go(lat: float, lon: float) -> Dict:
    url = f"{GO_SERVER}/reverse?lat={lat}&lon={lon}"
    with urllib.request.urlopen(url, timeout=5) as resp:
        data = json.loads(resp.read().decode('utf-8'))
        if not isinstance(data, dict):
            raise ValueError('invalid json root')
        if data.get('code') != 0:
            raise ValueError(f"go error code={data.get('code')} message={data.get('message')}")
        return data.get('data') or {}

def compare(go_rec: Dict, py_rec: Dict):
    diffs = []
    for f in FIELDS:
        if go_rec.get(f) != py_rec.get(f):
            diffs.append(f"{f}: go='{go_rec.get(f)}' python='{py_rec.get(f)}'")
    return diffs

def main():
    start = time.time()
    mismatches = 0
    tested = 0
    for i in range(SAMPLES):
        lat, lon = rand_coord()
        try:
            py_rec = query_python(lat, lon)
        except Exception as e:
            print(f"[ERROR] python query failed idx={i} lat={lat} lon={lon}: {e}")
            mismatches += 1
            continue
        try:
            go_rec = query_go(lat, lon)
        except Exception as e:
            print(f"[ERROR] go query failed idx={i} lat={lat} lon={lon}: {e}")
            mismatches += 1
            continue
        tested += 1
        diffs = compare(go_rec, py_rec)
        if diffs:
            mismatches += 1
            print(f"[DIFF] idx={i} lat={lat} lon={lon} -> {'; '.join(diffs)}")
    dur = time.time() - start
    print(f"Completed tested={tested} mismatches={mismatches} duration={dur:.2f}s")
    if mismatches == 0:
        print("ALL MATCH ✔")
        return 0
    else:
        print("HAS MISMATCH ✖")
        return 1

if __name__ == '__main__':
    sys.exit(main())
