var app = new Vue({
  el: '#app',
  data: {
    ws: null,
    serverUrlBase: (location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/ws',
    status: 'disconnected',

    // текущий юзер
    user: { name: '' },

    // онлайн-юзеры (для v-for="user in users")
    users: [],

    // комнаты
    rooms: [],
    roomInput: ''
  },
  methods: {
    connect() {
      const name = (this.user.name || '').trim();
      if (!name) return;
      this.connectToWebsocket(this.serverUrlBase + '?name=' + encodeURIComponent(name));
    },

    connectToWebsocket(url) {
      if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) return;
      this.ws = new WebSocket(url);
      this.ws.addEventListener('open', () => this.status = 'connected');
      this.ws.addEventListener('message', (e) => this.handleNewMessage(e));
      this.ws.addEventListener('close', (e) => this.status = `closed (${e.code})`);
      this.ws.addEventListener('error', () => this.status = 'error');
    },

    handleNewMessage(event) {
      const lines = String(event.data).split(/\r?\n/);
      for (const line of lines) {
        if (!line) continue;
        let msg;
        try { msg = JSON.parse(line); } catch { continue; }

        switch (msg.action) {
          case 'send-message': this.handleChatMessage(msg); break;
          case 'user-join':   this.handleUserJoined(msg);   break;
          case 'user-left':   this.handleUserLeft(msg);     break;
          case "room-joined": this.handleRoomJoined(msg);   break;
          default: break;
        }
      }
    },

    handleChatMessage(msg) {
      const room = this.findRoom(msg.target.id);
      if (typeof room !== "undefined") {
        room.messages.push(msg);
      }
    },

    handleRoomJoined(msg) {
      room = msg.target;
      room.name = room.private ? msg.sender.name : room.name;
      room["messages"] = [];
      this.rooms.push(room);
    },

    sendMessage(room) {
      if (room.newMessage !== "") {
        this.ws.send(JSON.stringify({
          action: 'send-message',
          message: room.newMessage,
          target: {
            id: room.id,
            name: room.name
          }
        }));
        room.newMessage = "";
      }
    },

    findRoom(roomId) {
      for (let i = 0; i < this.rooms.length; i++) {
        if (this.rooms[i].id === roomId) {
          return this.rooms[i];
        }
      }
    },

    joinRoom() {
      this.ws.send(JSON.stringify({ action: 'join-room', message: this.roomInput }));
      this.roomInput = "";
    },

    joinPrivateRoom(room) {
      this.ws.send(JSON.stringify({ action: 'join-room-private', message: room.id }));
    },

    leaveRoom(room) {
      this.ws.send(JSON.stringify({ action: 'leave-room', message: room.name }));
      const i = this.rooms.findIndex(r => r.name === room.name);
      if (i > -1) this.rooms.splice(i, 1);
    },

    handleUserJoined(msg) {
      // добавляем только если ещё нет
      if (!this.users.some(u => u.id === msg.sender.id)) {
        this.users.push(msg.sender);
      }
    },

    handleUserLeft(msg) {
      const i = this.users.findIndex(u => u.id === msg.sender.id);
      if (i > -1) this.users.splice(i, 1);
    }
  }
});
