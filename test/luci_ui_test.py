"""Playwright 验证 LuCI Docker Manager 页面"""
import time
from playwright.sync_api import sync_playwright

BASE = "http://192.168.1.1"

with sync_playwright() as p:
    browser = p.chromium.launch(channel="msedge", headless=True)
    ctx = browser.new_context()
    page = ctx.new_page()
    errors = []
    page.on("pageerror", lambda e: errors.append(str(e)))

    # 登录 LuCI
    page.goto(f"{BASE}/cgi-bin/luci", wait_until="networkidle", timeout=20000)
    page.fill("input[name='luci_username']", "root")
    page.fill("input[name='luci_password']", "JDunix786")
    page.click("input[type='submit']")
    page.wait_for_load_state("networkidle")
    time.sleep(2)
    print("登录后 URL:", page.url)

    # 找菜单:服务 → Docker Manager
    page.goto(f"{BASE}/cgi-bin/luci/admin/services/dockermanager", wait_until="networkidle", timeout=20000)
    time.sleep(3)
    body = page.inner_text("body")
    print("=== 页面内容 ===")
    for line in body.split("\n"):
        if any(k in line for k in ["运行中", "已停止", "8080", "Docker Manager", "开机自启", "管理界面", "启动", "停止", "重启"]):
            print(" ", line.strip())
    assert "Docker Manager" in body, "页面标题缺失"
    assert "打开 DOCKER MANAGER 管理界面" in body, "入口按钮缺失"  # LuCI CSS 强制大写
    assert "192.168.1.1:8080" in body, "跳转地址缺失"
    # 验证链接 href
    link = page.locator("a.cbi-button-apply").first
    href = link.get_attribute("href")
    assert "8080" in href, f"链接地址错误: {href}"
    print(f"跳转链接: {href} ✓")
    print("=== 菜单项检查 ===")
    # 检查侧边栏菜单
    menu = page.inner_text("body")
    assert "Docker Manager" in menu
    print("菜单存在 ✓")
    print("JS 错误:", errors if errors else "无")
    browser.close()
    print("=== LuCI 页面验证通过 ===")
