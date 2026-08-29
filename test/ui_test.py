"""Docker Manager 前端 UI 验证(Playwright + Edge 无头)"""
from playwright.sync_api import sync_playwright

BASE = "http://192.168.1.24:18080"

with sync_playwright() as p:
    browser = p.chromium.launch(channel="msedge", headless=True)
    page = browser.new_page(viewport={"width": 1440, "height": 900})
    errors = []

    page.on("console", lambda m: errors.append(f"[{m.type}] {m.text}") if m.type == "error" else None)
    page.on("pageerror", lambda e: errors.append(f"[pageerror] {e}"))

    # 1. 登录页
    page.goto(f"{BASE}/login", wait_until="networkidle")
    assert "Docker Manager" in page.title(), f"登录页标题异常: {page.title()}"
    print("登录页 OK:", page.title())

    # 2. 登录
    page.fill('input[placeholder="用户名"]', "admin")
    page.fill('input[placeholder="密码"]', "test1234")
    page.click("button:has-text('登 录')")
    page.wait_for_url("**/dashboard**", timeout=10000)
    page.wait_for_load_state("networkidle")
    print("登录跳转 OK ->", page.url)

    # 3. Dashboard 内容
    page.wait_for_selector("text=Docker Engine", timeout=10000)
    page.wait_for_selector("text=Ubuntu", timeout=10000)
    page.wait_for_selector("text=容器状态分布", timeout=5000)
    print("Dashboard 内容 OK(引擎信息/统计可见)")

    # 4. 容器列表
    page.goto(f"{BASE}/containers", wait_until="networkidle")
    page.wait_for_selector("text=dm-test", timeout=10000)
    page.wait_for_selector("text=dm-logtest", timeout=5000)
    page.wait_for_selector("text=运行中", timeout=5000)
    print("容器列表 OK(dm-test/dm-logtest 可见)")

    # 5. 容器详情(Overview)
    page.click("text=dm-logtest")
    page.wait_for_url("**/containers/**", timeout=10000)
    page.wait_for_selector("text=Overview", timeout=10000)
    page.wait_for_selector("text=nginx", timeout=5000)
    page.wait_for_selector("text=/docker-entrypoint.sh", timeout=5000)
    print("容器详情 Overview OK")

    # 6. 日志 tab
    page.click("text=日志", timeout=5000)
    page.wait_for_selector(".log-box", timeout=10000)
    page.wait_for_timeout(2500)  # 等 WS 日志到达
    log_lines = page.locator(".log-line").count()
    assert log_lines > 0, "日志区没有内容"
    print(f"日志 tab OK(收到 {log_lines} 行日志)")

    # 7. 实时监控 tab
    page.click("text=实时监控", timeout=5000)
    page.wait_for_selector("text=CPU 使用率", timeout=5000)
    page.wait_for_timeout(2500)  # 等 stats 推送
    cpu_text = page.locator(".stat-big").first.inner_text()
    print(f"实时监控 OK(CPU: {cpu_text})")

    # 8. Inspect tab
    page.click("text=Inspect", timeout=5000)
    page.wait_for_selector("text=Raw JSON", timeout=5000)
    page.click("text=Raw JSON", timeout=5000)
    page.wait_for_selector(".raw-json", timeout=5000)
    print("Inspect OK(Raw JSON 视图)")

    # 9. 终端占位
    page.click("text=终端", timeout=5000)
    page.wait_for_selector("text=终端功能未开放,后续版本提供", timeout=5000)
    print("终端占位 OK")

    # 10. 镜像页
    page.goto(f"{BASE}/images", wait_until="networkidle")
    page.wait_for_selector("text=nginx", timeout=10000)
    page.wait_for_selector("text=拉取镜像", timeout=5000)
    print("镜像页 OK")

    # 11. 网络页
    page.goto(f"{BASE}/networks", wait_until="networkidle")
    page.wait_for_selector("text=bridge", timeout=10000)
    print("网络页 OK")

    # 12. 卷页
    page.goto(f"{BASE}/volumes", wait_until="networkidle")
    page.wait_for_selector("text=local", timeout=10000)
    print("卷页 OK")

    # 13. 设置页
    page.goto(f"{BASE}/settings", wait_until="networkidle")
    page.wait_for_selector("text=修改密码", timeout=10000)
    print("设置页 OK")

    # 14. 暗色模式切换
    page.click(".header-right .el-button", timeout=5000)
    page.wait_for_timeout(500)
    is_dark = page.evaluate("document.documentElement.classList.contains('dark')")
    print(f"暗色模式切换 OK(dark={is_dark})")

    browser.close()

    real_errors = [e for e in errors if "favicon" not in e.lower()]
    if real_errors:
        print("\n控制台错误:")
        for e in real_errors[:10]:
            print(" ", e)
    else:
        print("\n无控制台错误")
    print("\n=== 前端 UI 验证完成 ===")
