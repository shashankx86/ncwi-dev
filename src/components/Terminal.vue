<template>
  <div class="terminal-container">
    <div ref="terminal" class="terminal"></div>
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
    };
  },
  mounted() {
    this.terminal = new Terminal({
      cursorBlink: true,
      rows: 24,
      cols: 80,
      fontFamily: 'monospace',
      fontSize: 14,
      theme: {
        background: '#1a1a1a',
        foreground: '#ffffff',
      },
    });
    this.fitAddon = new FitAddon();
    this.terminal.loadAddon(this.fitAddon);
    this.terminal.open(this.$refs.terminal);
    this.fitAddon.fit();

    // Connect to the WebSocket server
    this.socket = new WebSocket('ws://localhost:5490');

    this.socket.onmessage = (event) => {
      this.terminal.write(event.data.replace(/\n/g, '\r\n'));
    };

    this.terminal.onData((data) => {
      this.socket.send(data);
    });

    this.terminal.onResize(({ cols, rows }) => {
      this.socket.send(JSON.stringify({ event: 'resize', cols, rows }));
    });

    this.terminal.writeln('Connected to reverse shell');
  },
  beforeDestroy() {
    if (this.socket) {
      this.socket.close();
    }
  },
};
</script>

<style scoped>
.terminal-container {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #1a1a1a;
}
.terminal {
  width: 100%;
  height: 100%;
}
</style>
