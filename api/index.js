const express = require('express');
const bodyParser = require('body-parser');
const cors = require('cors');
const fs = require('fs');
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

app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});
