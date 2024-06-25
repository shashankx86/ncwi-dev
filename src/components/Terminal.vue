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
      socket: null
    };
  },
  mounted() {
    this.initializeTerminal();
  },
  beforeDestroy() {
    if (this.socket) {
      this.socket.close();
    }
  },
  methods: {
    initializeTerminal() {
      this.terminal = new Terminal();
      this.fitAddon = new FitAddon();
      this.terminal.loadAddon(this.fitAddon);
      this.terminal.open(this.$refs.terminal);
      this.fitAddon.fit();

      // Connect to the WebSocket server
      this.socket = new WebSocket('ws://localhost:5490');

      this.socket.onopen = () => {
        this.terminal.writeln('Connected to reverse shell');
      };

      this.socket.onmessage = (event) => {
        if (event.data === '__RESTART__') {
          this.terminal.writeln('\r\nShell restarted');
        } else {
          this.terminal.write(event.data);
        }
      };

      this.socket.onclose = () => {
        this.terminal.writeln('\r\nConnection closed');
      };

      this.terminal.onData(this.handleInput);
    },
    handleInput(data) {
      this.socket.send(data);
    }
  }
};
</script>

<style scoped>
.terminal-container {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: flex-start;
  align-items: flex-start;
}

.terminal {
  width: 100%;
  height: 100%;
  text-align: left;
}
</style>
