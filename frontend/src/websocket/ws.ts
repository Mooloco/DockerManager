import type { WSMessage } from '../api/types'

/** WebSocket 封装:自动重连、统一消息回调 */

interface WSHandlers {
  onMessage: (msg: WSMessage) => void
  onOpen?: () => void
  onClose?: () => void
  onError?: (msg: string) => void
}

export class WSClient {
  private ws: WebSocket | null = null
  private url: string
  private handlers: WSHandlers
  private closed = false

  constructor(path: string, handlers: WSHandlers) {
    this.url = `${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}${path}`
    this.handlers = handlers
  }

  connect() {
    this.closed = false
    this.open()
  }

  private open() {
    try {
      this.ws = new WebSocket(this.url)
    } catch (e) {
      this.handlers.onError?.(`WebSocket 连接失败: ${e}`)
      return
    }

    this.ws.onopen = () => this.handlers.onOpen?.()
    this.ws.onmessage = (ev) => {
      try {
        const msg = JSON.parse(ev.data) as WSMessage
        this.handlers.onMessage(msg)
      } catch {
        // 忽略无法解析的消息
      }
    }
    this.ws.onerror = () => this.handlers.onError?.('WebSocket 连接错误')
    this.ws.onclose = () => {
      this.ws = null
      this.handlers.onClose?.()
      // 非主动关闭时 3 秒后重连
      if (!this.closed) {
        setTimeout(() => {
          if (!this.closed) this.open()
        }, 3000)
      }
    }
  }

  close() {
    this.closed = true
    this.ws?.close()
    this.ws = null
  }
}
