require('express');

const router = express.Router();
const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

/**
 * Reference handler for POST /newsletter/subscribe.
 * Production: Go Lambda (CurlecGateway) writes to Google Sheet tab "Newsletter".
 */
router.post('/subscribe', async (req, res) => {
  const email = String(req.body?.email || '').trim().toLowerCase();
  const name = String(req.body?.name || '').trim();
  const source = String(req.body?.source || 'homepage').trim() || 'homepage';

  if (!EMAIL_RE.test(email)) {
    return res.status(400).json({ error: 'Invalid email address' });
  }

  const allowedSources = ['homepage', 'checkout', 'import'];
  if (!allowedSources.includes(source)) {
    return res.status(400).json({ error: 'Invalid source' });
  }

  // Production upserts: email, name, source, subscribed_at, sequence_step=0, status=active
  console.info('Newsletter subscribe (reference)', { email, name, source });

  return res.json({
    ok: true,
    email,
    message: 'Subscribed. Reference server does not write to Google Sheets — deploy Lambda handler.'
  });
});

module.exports = router;
