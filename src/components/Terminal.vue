<template>
  <div class="terminal-container" ref="container" @dblclick="captureCursor" @contextmenu.prevent="showContextMenu">
    <div ref="terminal" class="terminal"></div>
    <ul v-if="contextMenuVisible" :style="{ top: `${contextMenuY}px`, left: `${contextMenuX}px` }" class="context-menu">
      <li @click="copyText">Copy</li>
      <li @click="pasteText">Paste</li>
      <li @click="cutText">Cut</li>
      <li @click="restartShell">Restart</li>
      <li @click="reconnectShell">Reconnect</li>
      <li @click="saveOutput">Save</li>
    </ul>
  </div>
</template>

<script>
import { Terminal } from 'xterm';
import { FitAddon } from '@xterm/addon-fit';
import 'xterm/css/xterm.css';

export default {
  name: 'Terminal',
  data() {
    return {
      terminal: null,
      fitAddon: null,
      socket: null,
      cursorCaptured: false,
      contextMenuVisible: false,
      contextMenuX: 0,
      contextMenuY: 0,
      terminalOutput: ''
    };
  },
  mounted() {
    this.initializeTerminal();
    this.setupEventListeners();
  },
  beforeDestroy() {
    if (this.socket) {
      this.socket.close();
    }
    this.removeEventListeners();
  },
  methods: {
    initializeTerminal() {
      this.terminal = new Terminal();
      this.fitAddon = new FitAddon();
      this.terminal.loadAddon(this.fitAddon);
      this.terminal.open(this.$refs.terminal);
      this.fitAddon.fit();

      // Connect to the WebSocket server
      this.socket = new WebSocket('ws://localhost:5492');

      this.socket.onopen = () => {
        this.terminal.writeln('Connected to reverse shell');
      };

      this.socket.onmessage = (event) => {
        if (event.data === '__RESTART__') {
          this.terminal.writeln('\r\nShell restarted');
        } else {
          this.terminal.write(event.data);
          this.terminalOutput += event.data;
        }
      };

      this.socket.onclose = () => {
        this.terminal.writeln('\r\nConnection closed');
      };

      this.terminal.onData(this.handleInput);
    },
    handleInput(data) {
      this.socket.send(data);
      this.terminalOutput += data;
    },
    captureCursor() {
      if (!this.cursorCaptured) {
        this.terminal.focus();
        this.cursorCaptured = true;
      }
    },
    releaseCursor() {
      if (this.cursorCaptured) {
        this.terminal.blur();
        this.cursorCaptured = false;
      }
    },
    showContextMenu(event) {
      this.contextMenuVisible = true;
      this.contextMenuX = event.clientX;
      this.contextMenuY = event.clientY;
    },
    hideContextMenu() {
      this.contextMenuVisible = false;
    },
    copyText() {
      navigator.clipboard.writeText(this.terminal.getSelection());
      this.hideContextMenu();
    },
    pasteText() {
      navigator.clipboard.readText().then(text => {
        this.terminal.write(text);
        this.socket.send(text);
      });
      this.hideContextMenu();
    },
    cutText() {
      const selection = this.terminal.getSelection();
      navigator.clipboard.writeText(selection);
      this.terminal.write('\b \b'.repeat(selection.length));
      this.hideContextMenu();
    },
    restartShell() {
      this.socket.send('exit\n');
      setTimeout(() => {
        this.socket.close();
        this.initializeTerminal();
      }, 1000);
      this.hideContextMenu();
    },
    reconnectShell() {
      this.socket.close();
      this.initializeTerminal();
      this.hideContextMenu();
    },
    saveOutput() {
      const blob = new Blob([this.terminalOutput], { type: 'text/plain' });
      const a = document.createElement('a');
      a.href = URL.createObjectURL(blob);
      a.download = 'shell.txt';
      a.click();
      this.hideContextMenu();
    },
    setupEventListeners() {
      document.addEventListener('keydown', this.handleDocumentKeyDown);
      document.addEventListener('mousedown', this.handleDocumentMouseDown);
      window.addEventListener('click', this.hideContextMenu);
    },
    removeEventListeners() {
      document.removeEventListener('keydown', this.handleDocumentKeyDown);
      document.removeEventListener('mousedown', this.handleDocumentMouseDown);
      window.removeEventListener('click', this.hideContextMenu);
    },
    handleDocumentKeyDown(event) {
      // Prevent default browser actions when terminal is focused
      if (this.cursorCaptured) {
        event.preventDefault();
        this.terminal.write(event.key);
      }
    },
    handleDocumentMouseDown(event) {
      // Release cursor capture if user clicks outside the terminal
      if (this.cursorCaptured && !this.$refs.container.contains(event.target)) {
        this.releaseCursor();
      }
    },
  },
};
</script>
