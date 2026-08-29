"""Docker Manager WebSocket 测试:日志流 + stats 流"""
import json
import sys
import time
import urllib.request

HOST = "192.168.1.24"
PORT = 18080

def login():
    req = urllib.request.Request(
        f"http://{HOST}:{PORT}/api/v1/auth/login",
        data=json.dumps({"username": "admin", "password": "test1234"}).encode(),
        headers={"Content-Type": "application/json"},
    )
    resp = urllib.request.urlopen(req)
    return resp.headers.get("Set-Cookie", "").split(";")[0]

def get_container_id(cookie):
    req = urllib.request.Request(
        f"http://{HOST}:{PORT}/api/v1/containers",
        headers={"Cookie": cookie},
    )
    data = json.loads(urllib.request.urlopen(req).read())
    for c in data["data"]:
        if c["name"] == "dm-logtest":
            return c["id"]
    raise SystemExit("dm-logtest 不存在")

import websocket

def test_logs(cookie, cid):
    print("== 测试日志流 ==")
    ws = websocket.create_connection(
        f"ws://{HOST}:{PORT}/api/v1/ws/containers/{cid}/logs?follow=true&tail=50",
        header=[f"Cookie: {cookie}"],
        timeout=8,
    )
    msgs = []
    try:
        while len(msgs) < 3:
            msg = json.loads(ws.recv())
            msgs.append(msg)
            if msg["type"] == "log":
                print(f"  log[{msg['stream']}]: {msg['data'][:80]!r}")
    except Exception as e:
        print(f"  (接收结束: {e})")
    ws.close()
    types = {m["type"] for m in msgs}
    assert "log" in types, f"日志流未收到 log 消息: {msgs}"
    assert any(m["stream"] in ("stdout", "stderr") for m in msgs), "日志流缺少 stream 标记"
    print("  日志流 OK")

def test_stats(cookie, cid):
    print("== 测试 stats 流 ==")
    ws = websocket.create_connection(
        f"ws://{HOST}:{PORT}/api/v1/ws/containers/{cid}/stats?interval=2",
        header=[f"Cookie: {cookie}"],
        timeout=10,
    )
    msgs = []
    try:
        while len(msgs) < 2:
            msg = json.loads(ws.recv())
            msgs.append(msg)
            if msg["type"] == "stats":
                d = msg["data"]
                print(f"  stats: cpu={d['cpu_percent']}% mem={d['memory_bytes']/1048576:.1f}MB pids={d['pids']}")
    except Exception as e:
        print(f"  (接收结束: {e})")
    ws.close()
    assert any(m["type"] == "stats" for m in msgs), f"stats 流未收到 stats 消息: {msgs}"
    print("  stats 流 OK")

def test_pull(cookie):
    print("== 测试镜像拉取进度流 ==")
    ws = websocket.create_connection(
        f"ws://{HOST}:{PORT}/api/v1/ws/images/pull?ref={urllib.parse.quote('alpine:3.20')}",
        header=[f"Cookie: {cookie}"],
        timeout=60,
    )
    events = []
    try:
        while True:
            msg = json.loads(ws.recv())
            events.append(msg)
            if msg["type"] in ("end", "error"):
                break
            if len(events) <= 5:
                d = msg.get("data", {})
                print(f"  pull: status={d.get('status')!r} progress={d.get('progress', '')[:40]!r}")
    except Exception as e:
        print(f"  (接收结束: {e})")
    ws.close()
    types = {m["type"] for m in events}
    assert "pull" in types and "end" in types, f"pull 流异常: {types}"
    print(f"  pull 流 OK(共 {len(events)} 个事件)")

if __name__ == "__main__":
    import urllib.parse
    cookie = login()
    print(f"登录成功 cookie={cookie[:30]}...")
    cid = get_container_id(cookie)
    print(f"容器 ID={cid[:12]}")
    test_logs(cookie, cid)
    test_stats(cookie, cid)
    test_pull(cookie)
    print("\n=== 全部 WebSocket 测试通过 ===")
