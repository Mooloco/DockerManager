"""验证新功能:1) 容器页复选框+批量操作 2) 镜像输入校验 3) 总览菜单+刷新频率"""
import time
from playwright.sync_api import sync_playwright

BASE = "http://192.168.1.24:18080"

with sync_playwright() as p:
    browser = p.chromium.launch(channel="msedge", headless=True)
    page = browser.new_page(viewport={"width": 1440, "height": 900})
    errors = []
    page.on("pageerror", lambda e: errors.append(str(e)))

    # 登录
    page.goto(f"{BASE}/login", wait_until="networkidle")
    page.fill('input[placeholder="用户名"]', "admin")
    page.fill('input[placeholder="密码"]', "test1234")
    page.click("button:has-text('登 录')")
    page.wait_for_url("**/dashboard**", timeout=10000)

    # === 3a. 菜单"总览" ===
    menu = page.locator(".el-menu-item span", has_text="总览")
    assert menu.is_visible(), "菜单没有'总览'"
    print("菜单'总览' OK")
    assert "总览" in page.title(), f"页面标题应为'总览': {page.title()}"
    print("页面标题'总览' OK")

    # === 3b. 刷新频率按钮 ===
    refresh_btn = page.locator(".refresh-btn")
    assert refresh_btn.is_visible(), "顶栏没有刷新频率按钮"
    print("刷新频率按钮:", refresh_btn.inner_text())
    # 打开下拉,点击 2 秒
    refresh_btn.click()
    page.wait_for_timeout(500)
    page.locator(".el-dropdown-menu__item", has_text="2 秒").click()
    page.wait_for_timeout(500)
    assert "2s" in refresh_btn.inner_text(), f"刷新频率未更新: {refresh_btn.inner_text()}"
    print("刷新频率切换 2s OK,持久化:", page.evaluate("localStorage.getItem('dm-refresh-interval')"))

    # === 1. 容器页复选框 + 批量操作 ===
    page.goto(f"{BASE}/containers", wait_until="networkidle")
    page.wait_for_selector("table", timeout=10000)
    page.wait_for_timeout(800)

    # 复选框列存在
    sel_col = page.locator(".el-table__header th .el-checkbox").count()
    assert sel_col >= 1, "表格没有复选框列"
    print("复选框列 OK")

    # 批量操作按钮带文字
    for label in ["启动", "停止", "重启", "暂停", "恢复", "强制终止", "删除"]:
        btn = page.locator(f".page-toolbar button:has-text('{label}')").first
        assert btn.is_visible(), f"缺少批量操作按钮: {label}"
    print("批量操作按钮(带文字) OK")

    # 未选中时禁用
    disabled = page.locator(".page-toolbar button:has-text('删除')").first.is_disabled()
    assert disabled, "未选中容器时删除按钮应禁用"
    print("未选中时按钮禁用 OK")

    # 选中一个容器 → 按钮可用
    page.locator(".el-table__body .el-checkbox").first.click()
    page.wait_for_timeout(300)
    enabled = not page.locator(".page-toolbar button:has-text('删除')").first.is_disabled()
    assert enabled, "选中容器后删除按钮应可用"
    count_hint = page.locator(".count-hint").inner_text()
    print(f"选中后提示: {count_hint}")
    assert "已选 1 个" in count_hint, "计数提示未更新"
    print("选中容器 → 操作可用 OK")

    # 行内操作列已移除(操作按钮不在行内)
    row_btns = page.locator(".el-table__body .el-button").count()
    assert row_btns == 0, f"行内不应有操作按钮,实际 {row_btns} 个"
    print("行内操作列已移除 OK")

    # 取消选择
    page.locator(".el-table__body .el-checkbox").first.click()
    page.screenshot(path="build/verify-containers2.png")

    # === 2. 镜像输入校验 ===
    page.goto(f"{BASE}/images", wait_until="networkidle")
    page.click("button:has-text('拉取镜像')")
    page.wait_for_selector(".el-dialog:visible", timeout=5000)
    page.fill('.el-dialog input[placeholder*="nginx"]', "NGINX:latest 非法输入!")
    page.click(".el-dialog button:has-text('拉取')")
    page.wait_for_selector(".el-message--warning:has-text('输入错误,请重新输入')", timeout=5000)
    print("非法输入提示 OK")
    # 输入框被清空
    val = page.locator('.el-dialog input[placeholder*="nginx"]').input_value()
    assert val == "", f"输入框应被清空,实际: {val}"
    print("输入框清空 OK")
    page.screenshot(path="build/verify-pull-invalid.png")

    # 合法输入正常拉取
    page.fill('.el-dialog input[placeholder*="nginx"]', "busybox:1.36")
    page.click(".el-dialog button:has-text('拉取')")
    page.wait_for_selector(".el-message--success:has-text('镜像拉取完成')", timeout=60000)
    print("合法输入拉取成功 OK")

    browser.close()
    if errors:
        print("页面错误:", errors[:5])
    else:
        print("无页面错误")
    print("\n=== 新功能验证全部通过 ===")
