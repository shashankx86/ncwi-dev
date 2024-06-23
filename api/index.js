const express = require('express');
const bodyParser = require('body-parser');
const cors = require('cors');
const fs = require('fs');
const { exec } = require('child_process');
require('dotenv').config();

const app = express();
const PORT = process.env.PORT || 5490;
const VERSION = '0.0.1';

app.use(cors());
app.use(bodyParser.json());

const credentials = JSON.parse(fs.readFileSync('credentials.json', 'utf8'));

// Login endpoint
app.post('/login', (req, res) => {
  const { username, password } = req.body;
  if (
    username === credentials.username &&
    password === credentials.password
  ) {
    res.status(200).json({ message: 'Login successful' });
  } else {
    res.status(401).json({ message: 'Invalid username or password' });
  }
});

// Version endpoint
app.get('/version', (req, res) => {
  res.status(200).json({ version: VERSION });
});

// Helper function to execute shell commands
const executeCommand = (command, callback) => {
  exec(command, (error, stdout, stderr) => {
    if (error) {
      callback(stderr || error.message, null);
    } else {
      callback(null, stdout);
    }
  });
};

// Endpoint to list all running services owned by the user
app.get('/services', (req, res) => {
  executeCommand('systemctl --user list-units --type=service --state=running', (error, stdout) => {
    if (error) {
      res.status(500).json({ message: 'Error fetching services', error });
    } else {
      const services = stdout.split('\n').filter(line => line).map(line => line.split(/\s+/)[0]);
      res.status(200).json({ services });
    }
  });
});

// Endpoint to start a service
app.post('/services/start', (req, res) => {
  const { service } = req.body;
  if (!service) {
    return res.status(400).json({ message: 'Service name is required' });
  }
  executeCommand(`systemctl --user start ${service}`, (error) => {
    if (error) {
      res.status(500).json({ message: `Error starting service ${service}`, error });
    } else {
      res.status(200).json({ message: `Service ${service} started successfully` });
    }
  });
});

// Endpoint to stop a service
app.post('/services/stop', (req, res) => {
  const { service } = req.body;
  if (!service) {
    return res.status(400).json({ message: 'Service name is required' });
  }
  executeCommand(`systemctl --user stop ${service}`, (error) => {
    if (error) {
      res.status(500).json({ message: `Error stopping service ${service}`, error });
    } else {
      res.status(200).json({ message: `Service ${service} stopped successfully` });
    }
  });
});

app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});
