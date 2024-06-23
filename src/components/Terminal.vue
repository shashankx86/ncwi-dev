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
    };
  },
  mounted() {
    this.terminal = new Terminal();
    this.fitAddon = new FitAddon();
    this.terminal.loadAddon(this.fitAddon);
    this.terminal.open(this.$refs.terminal);
    this.fitAddon.fit();

    // Simulating terminal interaction
    this.terminal.writeln('Welcome to the xterm.js terminal');
  },
  methods: {
    openTerminal() {
      this.terminal.open(this.$refs.terminal);
      this.fitAddon.fit();
    },
  },
};
</script>
