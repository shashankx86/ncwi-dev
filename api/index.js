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

// const credentials = JSON.parse(fs.readFileSync('./credentials.json', 'utf8'));

// // Login endpoint
// app.post('/login', (req, res) => {
//   const { username, password } = req.body;
//   if (
//     username === credentials.username &&
//     password === credentials.password
//   ) {
//     res.status(200).json({ message: 'Login successful' });
//   } else {
//     res.status(401).json({ message: 'Invalid username or password' });
//   }
// });

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

// Helper function to parse docker ps output using regex
const parseDockerPs = (data) => {
  const lines = data.split('\n').filter(line => line.trim() !== '');
  const headers = lines[0].split(/\s{2,}/);
  const containers = [];

  const regex = /(\S+)\s+(\S+)\s+(\"[^\"]*\"|\S+)\s+(.*?)(?=\s{2,})\s{2,}(.*?)\s{2,}(.*?)\s{2,}(.*?)$/;

  for (let i = 1; i < lines.length; i++) {
    const match = lines[i].match(regex);
    if (match) {
      containers.push({
        CONTAINER_ID: match[1],
        IMAGE: match[2],
        COMMAND: match[3],
        CREATED: match[4],
        STATUS: match[5],
        PORTS: match[6],
        NAMES: match[7],
      });
    }
  }

  return containers;
};

// Create a router for Docker-related routes
const dockerRouter = express.Router();

// Endpoint to list all running Docker containers
dockerRouter.get('/running', (req, res) => {
  executeCommand('docker ps', (error, stdout) => {
    if (error) {
      return res.status(500).json({ message: 'Error fetching Docker containers', error });
    }

    const containers = parseDockerPs(stdout);
    res.status(200).json({ containers });
  });
});

// Mount the Docker-related routes under /docker
app.use('/docker', dockerRouter);

// Systemd routes
const parseUnits = (data) => {
  const unitRegex = /^\s*(\S+\.service|\S+\.socket)\s+(\S+)\s+(\S+)\s+(\S+)\s+(.+)$/gm;
  let match;
  const units = [];

  while ((match = unitRegex.exec(data)) !== null) {
    const [ , UNIT, LOAD, ACTIVE, SUB, DESCRIPTION ] = match;
    units.push({ UNIT, LOAD, ACTIVE, SUB, DESCRIPTION });
  }

  return units;
};

// Create a router for systemd related routes
const systemRouter = express.Router();

// Endpoint to list all running services and sockets owned by the user
systemRouter.get('/services', (req, res) => {
  // Execute both commands and wait for both to complete
  executeCommand('systemctl --user list-units --type=service --state=running', (serviceError, serviceStdout) => {
    if (serviceError) {
      return res.status(500).json({ message: 'Error fetching services', error: serviceError });
    }

    const services = parseUnits(serviceStdout);

    executeCommand('systemctl --user list-units --type=socket', (socketError, socketStdout) => {
      if (socketError) {
        return res.status(500).json({ message: 'Error fetching sockets', error: socketError });
      }

      const sockets = parseUnits(socketStdout);

      res.status(200).json({ services, sockets });
    });
  });
});

// Endpoint to start a service
systemRouter.post('/services/start', (req, res) => {
  const service = req.query.target;
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
systemRouter.post('/services/stop', (req, res) => {
  const service = req.query.target;
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

// Endpoint to restart a service
systemRouter.post('/services/restart', (req, res) => {
  const service = req.query.target;
  if (!service) {
    return res.status(400).json({ message: 'Service name is required' });
  }
  executeCommand(`systemctl --user restart ${service}`, (error) => {
    if (error) {
      res.status(500).json({ message: `Error restarting service ${service}`, error });
    } else {
      res.status(200).json({ message: `Service ${service} restarted successfully` });
    }
  });
});

// Mount the systemd related routes under /system
app.use('/system', systemRouter);

app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});
