const WebSocket = require('ws');
const pty = require('node-pty');
const { execSync } = require('child_process');

const wss = new WebSocket.Server({ port: 5492 });

wss.on('connection', (ws) => {
  let shell;

  try {
    // Check if the tmux session named "ncwi-shell" exists
    const existingSessions = execSync('tmux ls').toString();
    if (existingSessions.includes('ncwi-shell')) {
      shell = pty.spawn('tmux', ['attach-session', '-t', 'ncwi-shell'], {
        name: 'xterm-color',
        cols: 80,
        rows: 24,
        cwd: process.env.HOME,
        env: process.env
      });
    } else {
      shell = pty.spawn('tmux', ['new-session', '-s', 'ncwi-shell'], {
        name: 'xterm-color',
        cols: 80,
        rows: 24,
        cwd: process.env.HOME,
        env: process.env
      });
    }
  } catch (error) {
    // If the "tmux ls" command fails, create a new session
    shell = pty.spawn('tmux', ['new-session', '-s', 'ncwi-shell'], {
      name: 'xterm-color',
      cols: 80,
      rows: 24,
      cwd: process.env.HOME,
      env: process.env
    });
  }

  shell.on('data', (data) => {
    ws.send(data);
  });

  ws.on('message', (message) => {
    shell.write(message);
  });

  ws.on('close', () => {
    shell.kill();
  });
});

console.log('WebSocket server is running on ws://localhost:5490');
