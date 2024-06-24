const express = require('express');
const bodyParser = require('body-parser');
const cors = require('cors');
const fs = require('fs');
const { exec } = require('child_process');
require('dotenv').config();
const os = require('os');
const disk = require('diskusage');

const app = express();
const PORT = process.env.PORT || 5490;
const VERSION = '0.0.1';

app.use(cors());
app.use(bodyParser.json());

// const credentials = JSON.parse(fs.readFileSync('credentials.json', 'utf8'));

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

// Helper function to parse services or sockets using regex
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

// Endpoint to write a file
systemRouter.post('/write', (req, res) => {
  const { filename, filepath, filecontent } = req.query;
  if (!filename || !filepath || !filecontent) {
    return res.status(400).json({ message: 'Filename, filepath, and filecontent are required' });
  }

  const fullPath = `${filepath}/${filename}`;
  
  try {
    fs.writeFileSync(fullPath, filecontent, 'utf8');
    res.status(200).json({ message: `File ${filename} saved successfully at ${filepath}` });
  } catch (error) {
    res.status(500).json({ message: `Error saving file ${filename} at ${filepath}`, error });
  }
});

// Endpoint to read a file
systemRouter.get('/read', (req, res) => {
  const { filename, filepath } = req.query;
  if (!filename || !filepath) {
    return res.status(400).json({ message: 'Filename and filepath are required' });
  }

  const fullPath = `${filepath}/${filename}`;

  try {
    const fileContent = fs.readFileSync(fullPath, 'utf8');
    res.status(200).json({ content: fileContent });
  } catch (error) {
    res.status(500).json({ message: `Error reading file ${filename} at ${filepath}`, error });
  }
});

// Endpoint to schedule a task
systemRouter.post('/at', (req, res) => {
  const { time, command } = req.query;
  if (!time || !command) {
    return res.status(400).json({ message: 'Both time and command are required' });
  }

  const atCommand = `echo "${command}" | at ${time}`;
  executeCommand(atCommand, (error, stdout, stderr) => {
    if (error) {
      res.status(500).json({ message: `Error scheduling task at ${time}`, error: stderr || error.message });
    } else {
      res.status(200).json({ message: `Task scheduled at ${time}`, output: stdout });
    }
  });
});

// Mount the systemd related routes under /system
app.use('/system', systemRouter);

// Create a router for Docker related routes
const dockerRouter = express.Router();

// Endpoint to list running Docker containers
dockerRouter.get('/running', (req, res) => {
  executeCommand('docker ps', (error, stdout) => {
    if (error) {
      return res.status(500).json({ message: 'Error fetching running containers', error });
    }

    const lines = stdout.trim().split('\n');
    const headers = lines[0].split(/\s{2,}/);
    const containers = lines.slice(1).map(line => {
      const columns = line.split(/\s{2,}/);
      return headers.reduce((container, header, index) => {
        container[header.replace(/ /g, '_')] = columns[index];
        return container;
      }, {});
    });

    res.status(200).json({ containers });
  });
});

// Endpoint to start a Docker container
dockerRouter.post('/start', (req, res) => {
  const { targetid, targetname } = req.query;
  if (!targetid && !targetname) {
    return res.status(400).json({ message: 'Either targetid or targetname is required' });
  }

  const target = targetid || targetname;
  executeCommand(`docker start ${target}`, (error) => {
    if (error) {
      res.status(500).json({ message: `Error starting container ${target}`, error });
    } else {
      res.status(200).json({ message: `Container ${target} started successfully` });
    }
  });
});

// Endpoint to stop a Docker container
dockerRouter.post('/stop', (req, res) => {
  const { targetid, targetname } = req.query;
  if (!targetid && !targetname) {
    return res.status(400).json({ message: 'Either targetid or targetname is required' });
  }

  const target = targetid || targetname;
  executeCommand(`docker stop ${target}`, (error) => {
    if (error) {
      res.status(500).json({ message: `Error stopping container ${target}`, error });
    } else {
      res.status(200).json({ message: `Container ${target} stopped successfully` });
    }
  });
});

// Endpoint to restart a Docker container
dockerRouter.post('/restart', (req, res) => {
  const { targetid, targetname } = req.query;
  if (!targetid && !targetname) {
    return res.status(400).json({ message: 'Either targetid or targetname is required' });
  }

  const target = targetid || targetname;
  executeCommand(`docker restart ${target}`, (error) => {
    if (error) {
      res.status(500).json({ message: `Error restarting container ${target}`, error });
    } else {
      res.status(200).json({ message: `Container ${target} restarted successfully` });
    }
  });
});

// Endpoint to list Docker images
dockerRouter.get('/image/ls', (req, res) => {
  executeCommand('docker image ls', (error, stdout) => {
    if (error) {
      return res.status(500).json({ message: 'Error fetching Docker images', error });
    }

    const lines = stdout.trim().split('\n');
    const headers = lines[0].split(/\s{2,}/);
    const images = lines.slice(1).map(line => {
      const columns = line.split(/\s{2,}/);
      return headers.reduce((image, header, index) => {
        image[header.replace(/ /g, '_')] = columns[index];
        return image;
      }, {});
    });

    res.status(200).json({ images });
  });
});

// Endpoint to remove a Docker image
dockerRouter.delete('/image/rm', (req, res) => {
  const { targetid, tokill } = req.query;
  if (!targetid) {
    return res.status(400).json({ message: 'targetid is required' });
  }

  const forceFlag = tokill === 'true' ? '--force' : '';
  executeCommand(`docker image rm ${targetid} ${forceFlag}`, (error, stdout, stderr) => {
    if (error) {
      return res.status(500).json({ message: `Error removing image ${targetid}`, error: stderr || error.message });
    }

    res.status(200).json({ message: `Image ${targetid} removed successfully` });
  });
});

// Mount the Docker related routes under /docker
app.use('/docker', dockerRouter);

// Function to get disk usage
const getDiskUsage = (callback) => {
  disk.check('/', (err, info) => {
    if (err) {
      return callback(err, null);
    }
    const used = (info.total - info.free) / (1024 ** 3); // Convert bytes to GB
    const total = info.total / (1024 ** 3); // Convert bytes to GB
    callback(null, { used: used.toFixed(2), total: total.toFixed(2) });
  });
};

// Function to get memory usage
const getMemoryUsage = () => {
  const total = os.totalmem() / (1024 ** 3); // Convert bytes to GB
  const free = os.freemem() / (1024 ** 3); // Convert bytes to GB
  const used = total - free;
  return { used: used.toFixed(2), total: total.toFixed(2) };
};

app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});
