const WebSocket = require('ws');
const { spawn } = require('child_process');

const wss = new WebSocket.Server({ port: 5490 });

wss.on('connection', (ws) => {
  const shell = spawn('sh');

  shell.stdout.on('data', (data) => {
    ws.send(data.toString());
  });

  shell.stderr.on('data', (data) => {
    ws.send(data.toString());
  });

  ws.on('message', (message) => {
    shell.stdin.write(message + '\n');
  });

  ws.on('close', () => {
    shell.kill();
  });
});

console.log('WebSocket server is running on ws://localhost:8080');
