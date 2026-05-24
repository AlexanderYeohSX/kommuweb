const crypto = require('crypto');
const Razorpay = require('razorpay');

let client;

function getClient() {
  if (!client) {
    const key_id = process.env.RAZORPAY_KEY_ID;
    const key_secret = process.env.RAZORPAY_KEY_SECRET;
    if (!key_id || !key_secret) {
      throw new Error('RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET must be set');
    }
    client = new Razorpay({ key_id, key_secret });
  }
  return client;
}

function verifyPaymentSignature(orderId, paymentId, signature) {
  const secret = process.env.RAZORPAY_KEY_SECRET;
  const body = `${orderId}|${paymentId}`;
  const expected = crypto.createHmac('sha256', secret).update(body).digest('hex');
  return expected === signature;
}

function verifyWebhookSignature(rawBody, signature) {
  const secret = process.env.RAZORPAY_WEBHOOK_SECRET;
  if (!secret) return false;
  const expected = crypto
    .createHmac('sha256', secret)
    .update(rawBody)
    .digest('hex');
  return expected === signature;
}

function trimNotes(notes, max = 15) {
  if (!notes || typeof notes !== 'object') return {};
  const entries = Object.entries(notes)
    .slice(0, max)
    .map(([k, v]) => [String(k).slice(0, 256), String(v).slice(0, 256)]);
  return Object.fromEntries(entries);
}

module.exports = {
  getClient,
  verifyPaymentSignature,
  verifyWebhookSignature,
  trimNotes
};
