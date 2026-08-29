"""调试 LuCI 前端 RPC 调用失败原因"""
import time
from playwright.sync_api import sync_playwright

BASE = "http://192.168.1.1"

with sync_playwright() as p:
    browser = p.chromium.launch(channel="msedge", headless=True)
    ctx = browser.new_context()
    page = ctx.new_page()
    console_msgs = []
    page.on("console", lambda m: console_msgs.append(f"[{m.type}] {m.text[:200]}"))
    page.on("pageerror", lambda e: console_msgs.append(f"[PAGEERROR] {str(e)[:200]}"))

    page.goto(f"{BASE}/cgi-bin/luci", wait_until="networkidle", timeout=20000)
    page.fill("input[name='luci_username']", "root")
    page.fill("input[name='luci_password']", "JDunix786")
    page.click("input[type='submit']")
    page.wait_for_load_state("networkidle")
    time.sleep(2)

    page.goto(f"{BASE}/cgi-bin/luci/admin/services/dockermanager", wait_until="networkidle", timeout=20000)
    time.sleep(4)

    print("=== console 消息 ===")
    for m in console_msgs[-15:]:
        print(" ", m)
    print("=== 页面状态文本 ===")
    body = page.inner_text("body")
    import re
    for line in body.split("\n"):
        if "运行" in line or "停止" in line or "8080" in line:
            print(" ", line.strip()[:80])
    browser.close()
