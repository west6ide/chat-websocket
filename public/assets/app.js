var app = new Vue({
  el: '#app',
  data: {
    ws: null,
    status: 'disconnected',
    lastMessage: '',
    serverUrl: (location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/ws',
    messages: [],           // [{ message: '...', mine: bool, ts: number }]
    newMessage: ''
  },
  mounted() {
    this.connectToWebsocket();
  },
  methods: {
    connectToWebsocket() {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) return;

      this.ws = new WebSocket(this.serverUrl);

      this.ws.addEventListener('open', () => {
        this.status = 'connected';
        console.log('connected to WS!');
      });

      this.ws.addEventListener('message', (event) => this.handleNewMessage(event));

      this.ws.addEventListener('close', (e) => {
        this.status = `closed (${e.code})`;
      });

      this.ws.addEventListener('error', (e) => {
        this.status = 'error';
        console.error('ws error', e);
      });
    },

    handleNewMessage(event) {
      const lines = String(event.data).split(/\r?\n/);
      for (const line of lines) {
        if (!line) continue;

        let text;
        try {
          const parsed = JSON.parse(line);
          text = (parsed && typeof parsed === 'object' && 'message' in parsed)
            ? parsed.message
            : String(line);
        } catch {
          text = String(line);
        }

        const item = { message: text, mine: false, ts: Date.now() };
        this.messages.push(item);
        this.lastMessage = text;
      }
      this.$nextTick(this.scrollToBottom);
    },

    sendMessage() {
      const text = this.newMessage.trim();
      if (!text) return;

      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        console.warn('WS is not open');
        this.status = 'disconnected';
        return;
      }

      // 1) Показать сразу как «моё» сообщение
      this.messages.push({ message: text, mine: true, ts: Date.now() });
      this.$nextTick(this.scrollToBottom);

      // 2) Отправить на сервер
      this.ws.send(JSON.stringify({ message: text }));

      // 3) Очистить поле
      this.newMessage = '';
    },

    scrollToBottom() {
      const el = this.$refs.msgBody;
      if (!el) return;
      el.scrollTop = el.scrollHeight;
    },

    formatTime(ts) {
      const d = new Date(ts);
      return d.toLocaleTimeString();
    }
  }
});
