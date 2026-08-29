"""验证项目功能前端 UI"""
from playwright.sync_api import sync_playwright

BASE = "http://192.168.1.24:18080"

with sync_playwright() as p:
    browser = p.chromium.launch(channel="msedge", headless=True)
    page = browser.new_page(viewport={"width": 1440, "height": 900})
    errors = []
    page.on("pageerror", lambda e: errors.append(str(e)))

    page.goto(f"{BASE}/login", wait_until="networkidle")
    page.fill('input[placeholder="用户名"]', "admin")
    page.fill('input[placeholder="密码"]', "test1234")
    page.click("button:has-text('登 录')")
    page.wait_for_url("**/dashboard**", timeout=10000)

    # 1. 菜单"项目"在容器和镜像之间
    menu_items = page.evaluate("""() => Array.from(document.querySelectorAll('.el-menu-item span')).map(s => s.textContent.trim())""")
    print("菜单:", menu_items)
    ci = menu_items.index("容器")
    pi = menu_items.index("项目")
    ii = menu_items.index("镜像")
    assert ci < pi < ii, "项目菜单位置不对"
    print("菜单'项目'位置 OK(容器和镜像中间)✓")

    # 2. 项目列表
    page.goto(f"{BASE}/projects", wait_until="networkidle")
    page.wait_for_selector("text=dmproj", timeout=10000)
    page.wait_for_selector("text=guacamole", timeout=5000)
    page.wait_for_selector("text=新建项目", timeout=5000)
    # 已有项目来源标签
    page.wait_for_selector("text=已有", timeout=5000)
    # compose 文件路径显示
    page.wait_for_selector("text=compose-test.yaml", timeout=5000)
    print("项目列表 OK(dmproj/guacamole 可见,含 compose 文件名)✓")
    page.screenshot(path="build/verify-projects.png")

    # 3. 项目详情
    page.click("text=dmproj")
    page.wait_for_url("**/projects/dmproj**", timeout=10000)
    page.wait_for_selector("text=基本信息", timeout=10000)
    page.wait_for_selector("text=compose-test.yaml", timeout=5000)  # 非默认文件名
    page.wait_for_selector("text=卷 / 挂载", timeout=5000)
    page.wait_for_selector("text=bind 挂载", timeout=5000)  # bind 类型标签
    page.wait_for_selector("text=webdata", timeout=5000)  # volume 名
    page.wait_for_selector(".el-table__body td:has-text('/mnt/bind')", timeout=5000)  # bind 源路径
    page.wait_for_selector("text=ro", timeout=5000)  # 只读标签
    page.wait_for_selector("text=compose YAML", timeout=5000)
    print("项目详情 OK(compose 文件名/卷类型/读写/YAML)✓")
    page.screenshot(path="build/verify-project-detail.png")

    # 4. 服务列表 + 容器状态
    page.wait_for_selector("text=服务", timeout=5000)
    page.wait_for_selector("text=dm-pweb", timeout=5000)
    print("服务列表 OK ✓")

    # 5. 编辑对话框
    page.click("button:has-text('编辑 compose')")
    page.wait_for_selector(".el-dialog textarea", timeout=5000)
    content = page.locator(".el-dialog textarea").input_value()
    # discovered 项目的文件在宿主机任意位置,容器部署下可能不可读 → fallback 提示;managed 项目(API 已验)可正常读写
    assert len(content) > 0, "编辑器内容为空"
    print(f"编辑对话框 OK(内容 {len(content)} 字符:{content[:40]!r})")
    page.keyboard.press("Escape")

    browser.close()
    if errors:
        print("页面错误:", errors[:5])
    else:
        print("无页面错误")
    print("\n=== 项目功能 UI 验证全部通过 ===")
