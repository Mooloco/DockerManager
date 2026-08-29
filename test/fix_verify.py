"""验证三个 bug 修复:1) pull 完成后不重复提示 2) 完成按钮 3) 表格列不重叠"""
import time
from playwright.sync_api import sync_playwright

BASE = "http://192.168.1.24:18080"

with sync_playwright() as p:
    browser = p.chromium.launch(channel="msedge", headless=True)
    page = browser.new_page(viewport={"width": 1440, "height": 900})

    # 登录
    page.goto(f"{BASE}/login", wait_until="networkidle")
    page.fill('input[placeholder="用户名"]', "admin")
    page.fill('input[placeholder="密码"]', "test1234")
    page.click("button:has-text('登 录')")
    page.wait_for_url("**/dashboard**", timeout=10000)
    print("登录 OK")

    # === Bug 1+2: 镜像拉取 ===
    page.goto(f"{BASE}/images", wait_until="networkidle")

    # 注入 MutationObserver 统计"镜像拉取完成"提示出现次数(ElMessage 3 秒后自动消失,不能靠 DOM 存量)
    page.evaluate("""() => {
        window.__pullDoneCount = 0;
        new MutationObserver(() => {
            document.querySelectorAll('.el-message--success').forEach(m => {
                if (m.textContent.includes('镜像拉取完成') && !m.dataset.counted) {
                    m.dataset.counted = '1';
                    window.__pullDoneCount++;
                }
            });
        }).observe(document.body, { childList: true, subtree: true });
    }""")

    page.click("button:has-text('拉取镜像')")
    page.wait_for_selector(".el-dialog:visible", timeout=5000)
    page.fill('.el-dialog input[placeholder*="nginx"]', "busybox:1.36")
    page.click(".el-dialog button:has-text('拉取')")

    # 等待完成提示出现
    page.wait_for_selector(".el-message--success:has-text('镜像拉取完成')", timeout=60000)
    print(f"首次完成提示出现,等待 8 秒观察是否重复...")

    # 等待 8 秒(超过 3 秒重连间隔),统计提示出现总次数
    time.sleep(8)
    count = page.evaluate("window.__pullDoneCount")
    print(f"8 秒内'镜像拉取完成'提示出现次数: {count}")
    assert count == 1, f"BUG: 完成提示出现 {count} 次(自动重连未修复)"

    # 完成按钮
    done_btn = page.locator(".el-dialog button:has-text('完成')")
    assert done_btn.is_visible(), "BUG: 拉取完成后没有'完成'按钮"
    print("完成按钮 OK")
    page.screenshot(path="build/verify-pull.png")

    # 关闭对话框
    done_btn.click()
    page.wait_for_timeout(500)

    # === Bug 3: 容器表格列重叠 ===
    page.goto(f"{BASE}/containers", wait_until="networkidle")
    page.wait_for_selector("table", timeout=10000)
    page.wait_for_timeout(1000)

    # 取"创建时间"列头与"操作"列头的边界,检查是否有重叠
    bounds = page.evaluate("""() => {
        const headers = document.querySelectorAll('.el-table__header th');
        const info = {};
        headers.forEach(th => {
            const label = th.innerText.trim();
            if (['创建时间', '操作', 'ID', 'CPU', '内存'].includes(label)) {
                const r = th.getBoundingClientRect();
                info[label] = { left: r.left, right: r.right, width: r.width };
            }
        });
        return info;
    }""")
    print("列边界:", bounds)

    # 检查重叠:相邻列 right > 下一列 left 视为重叠
    cols = bounds
    if '创建时间' in cols and '操作' in cols:
        assert cols['创建时间']['right'] <= cols['操作']['left'], f"BUG: 创建时间列与操作列重叠 {cols}"
    # 检查操作列是否覆盖行内容(操作列 fixed 移除后不应遮住任何行)
    overlap = page.evaluate("""() => {
        const rows = document.querySelectorAll('.el-table__body tr');
        let worst = 0;
        rows.forEach(tr => {
            const cells = tr.querySelectorAll('td');
            for (let i = 0; i < cells.length - 1; i++) {
                const a = cells[i].getBoundingClientRect();
                const b = cells[i+1].getBoundingClientRect();
                if (a.right > b.left) worst = Math.max(worst, a.right - b.left);
            }
        });
        return worst;
    }""")
    print(f"行内最大重叠像素: {overlap}px")
    assert overlap < 2, f"BUG: 行内容重叠 {overlap}px"
    print("表格列布局 OK")
    page.screenshot(path="build/verify-containers.png")

    browser.close()
    print("\n=== 三个 bug 修复验证全部通过 ===")
