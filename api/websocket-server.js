const WebSocket = require('ws');
const pty = require('node-pty');

const wss = new WebSocket.Server({ port: 5490 });

function startShell(ws) {
  // Spawn a bash shell with pty
  const shell = pty.spawn('bash', [], {
    name: 'xterm-color',
    cols: 80,
    rows: 24,
    cwd: process.env.HOME,
    env: process.env
  });

  shell.on('data', (data) => {
    ws.send(data);
  });

  shell.on('exit', () => {
    ws.send('__RESTART__');
    startShell(ws); // Restart the shell
  });

  ws.on('message', (message) => {
    shell.write(message);
  });

  ws.on('close', () => {
    shell.kill();
  });
}

wss.on('connection', (ws) => {
  startShell(ws);
});

console.log('WebSocket server is running on ws://localhost:5490');
