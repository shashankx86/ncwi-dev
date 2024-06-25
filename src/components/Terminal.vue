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
      commandHistory: [],
      historyIndex: 0,
      currentCommand: '',
      cursorPosition: 0,
      prompt: '$ '
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

    this.socket.onopen = () => {
      this.terminal.writeln('Connected to reverse shell');
      this.printPrompt();
    };

    this.socket.onmessage = (event) => {
      this.terminal.write(event.data);
    };

    this.terminal.onData(this.handleInput);
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
          this.executeCommand();
          break;
        case '\u007F': // Backspace key
          if (this.cursorPosition > 0) {
            this.currentCommand = this.currentCommand.slice(0, this.cursorPosition - 1) + this.currentCommand.slice(this.cursorPosition);
            this.cursorPosition--;
            this.updateTerminalDisplay();
          }
          break;
        case '\u001B[A': // Up arrow key
          this.previousCommand();
          break;
        case '\u001B[B': // Down arrow key
          this.nextCommand();
          break;
        case '\u001B[C': // Right arrow key
          if (this.cursorPosition < this.currentCommand.length) {
            this.cursorPosition++;
            this.terminal.write('\x1b[C');
          }
          break;
        case '\u001B[D': // Left arrow key
          if (this.cursorPosition > 0) {
            this.cursorPosition--;
            this.terminal.write('\x1b[D');
          }
          break;
        default:
          if (data >= ' ' && data <= '~') { // Only process printable characters
            this.currentCommand = this.currentCommand.slice(0, this.cursorPosition) + data + this.currentCommand.slice(this.cursorPosition);
            this.cursorPosition++;
            this.terminal.write(data);
          }
          break;
      }
    },
    updateTerminalDisplay() {
      this.terminal.write('\x1b[2K\r' + this.prompt + this.currentCommand.slice(0, this.cursorPosition) + '\x1b[7m' + (this.currentCommand[this.cursorPosition] || ' ') + '\x1b[27m' + this.currentCommand.slice(this.cursorPosition + 1) + '\x1b[' + (this.prompt.length + this.cursorPosition + 1) + 'G');
    },
    executeCommand() {
      this.terminal.write('\r\n');
      if (this.currentCommand.trim()) {
        this.commandHistory.push(this.currentCommand);
        this.historyIndex = this.commandHistory.length;
        this.socket.send(this.currentCommand + '\n');
      }
      this.currentCommand = '';
      this.cursorPosition = 0;
      this.printPrompt();
    },
    previousCommand() {
      if (this.historyIndex > 0) {
        this.historyIndex--;
        this.currentCommand = this.commandHistory[this.historyIndex];
        this.cursorPosition = this.currentCommand.length;
        this.updateTerminalDisplay();
      }
    },
    nextCommand() {
      if (this.historyIndex < this.commandHistory.length - 1) {
        this.historyIndex++;
        this.currentCommand = this.commandHistory[this.historyIndex];
        this.cursorPosition = this.currentCommand.length;
        this.updateTerminalDisplay();
      } else {
        this.historyIndex = this.commandHistory.length;
        this.currentCommand = '';
        this.cursorPosition = 0;
        this.updateTerminalDisplay();
      }
    },
    printPrompt() {
      this.terminal.write(this.prompt);
    }
  }
};
</script>

<style scoped>
.terminal-container {
  width: 100%;
  height: 100%;
}

.terminal {
  width: 100%;
  height: 100%;
}
</style>
