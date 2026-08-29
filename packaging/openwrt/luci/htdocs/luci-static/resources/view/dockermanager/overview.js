'use strict';
// Docker Manager - LuCI 入口页(纯静态,无 RPC)
// 管理界面是独立 Web 应用,默认端口 8080(可在 /etc/config/dockermanager 的 port 修改)

return L.view.extend({
	render: function() {
		var url = window.location.protocol + '//' + window.location.hostname + ':8080';

		return E('div', { 'class': 'cbi-map' }, [
			E('h2', {}, 'Docker Manager'),
			E('p', { 'class': 'cbi-map-descr' },
				'Docker Manager 是一个独立的 Web 管理界面,提供容器 / Compose 项目 / 镜像 / 网络 / 卷管理。' +
				'点击下方按钮在新窗口打开。'),
			E('div', { 'class': 'cbi-page-actions' }, [
				E('a', { 'class': 'cbi-button cbi-button-apply', 'href': url, 'target': '_blank' },
					'打开 Docker Manager 管理界面 (' + window.location.hostname + ':8080)')
			])
		]);
	}
});
