const { verifyWebhookSignature } = require('../lib/razorpay');

module.exports = function razorpayWebhook(req, res) {
  const signature = req.headers['x-razorpay-signature'];
  const rawBody = req.body;
  if (!Buffer.isBuffer(rawBody)) {
    return res.status(400).send('Invalid body');
  }
  if (!verifyWebhookSignature(rawBody, signature)) {
    return res.status(400).send('Invalid signature');
  }

  let event;
  try {
    event = JSON.parse(rawBody.toString('utf8'));
  } catch (_) {
    return res.status(400).send('Invalid JSON');
  }

  console.info('Razorpay webhook', event.event);

  res.json({ status: 'ok' });
};
