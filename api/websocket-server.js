const WebSocket = require('ws');
const pty = require('node-pty');

const wss = new WebSocket.Server({ port: 5490 });

wss.on('connection', (ws) => {
  // Spawn a shell with a PTY
  const shell = pty.spawn('sh', [], {
    name: 'xterm-color',
    cols: 80,
    rows: 24,
    cwd: process.env.HOME,
    env: process.env
  });

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
