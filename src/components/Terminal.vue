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
      currentCommand: '', // Store the current command being typed
    };
  },
  mounted() {
    this.terminal = new Terminal();
    this.fitAddon = new FitAddon();
    this.terminal.loadAddon(this.fitAddon);
    this.terminal.open(this.$refs.terminal);
    this.fitAddon.fit();

    // Connect to the WebSocket server
    this.socket = new WebSocket('ws://localhost:5490');

    this.socket.onmessage = (event) => {
      this.terminal.write(event.data);
    };

    this.terminal.onData((data) => {
      this.handleInput(data);
    });

    this.terminal.writeln('Connected to reverse shell');
  },
  beforeDestroy() {
    if (this.socket) {
      this.socket.close();
    }
  },
  methods: {
    handleInput(data) {
      switch (data) {
        case '\r': // Enter key
          this.socket.send(this.currentCommand);
          this.currentCommand = '';
          this.terminal.write('\r\n');
          break;
        case '\u007F': // Backspace key
          if (this.currentCommand.length > 0) {
            this.currentCommand = this.currentCommand.slice(0, -1);
            this.terminal.write('\b \b');
          }
          break;
        default:
          this.currentCommand += data;
          this.terminal.write(data);
          break;
      }
    },
  },
};
</script>