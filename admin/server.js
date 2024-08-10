const express = require('express');
const cors = require('cors');
const path = require('path');

const app = express();
const port = 8000;

// Enable CORS for all routes
app.use(cors());

// Serve static files from the 'public' directory
app.use(express.static(path.join(__dirname, 'public')));

// Sample API route
app.get('/api/hello', (req, res) => {
  res.json({ message: 'Hello from the server!' });
});

// Catch-all route to serve your single-page application
app.get('*', (req, res) => {
  res.sendFile(path.join(__dirname, 'public', 'index.html'));
});

app.listen(port, () => {
  console.log(`Server running at http://localhost:${port}`);
});
