require('dotenv').config();
const express = require('express');
const cors = require('cors');
const curlecRoutes = require('./routes/curlec');

const app = express();
const port = process.env.PORT || 3000;

const allowedOrigins = (process.env.STOREFRONT_ORIGIN || 'https://kommu.ai')
  .split(',')
  .map((s) => s.trim())
  .filter(Boolean);

app.use(
  cors({
    origin(origin, cb) {
      if (!origin) return cb(null, true);
      if (allowedOrigins.some((o) => origin === o || origin.startsWith(o.replace(/\/$/, '')))) {
        return cb(null, true);
      }
      if (process.env.NODE_ENV !== 'production') return cb(null, true);
      cb(new Error('Not allowed by CORS'));
    }
  })
);

app.post(
  '/curlec/webhook',
  express.raw({ type: 'application/json' }),
  require('./routes/webhook')
);
app.use(express.json());

app.get('/health', (_req, res) => res.json({ ok: true }));

app.use('/curlec', curlecRoutes);

app.use((err, _req, res, _next) => {
  console.error(err);
  res.status(500).json({ error: err.message || 'Internal server error' });
});

app.listen(port, () => {
  console.info(`Kommu Curlec API listening on port ${port}`);
});
