const { test } = require('node:test');
const assert = require('node:assert/strict');
const { rmToSen } = require('../src/lib/amount');

test('rmToSen converts ringgit to sen', () => {
  assert.equal(rmToSen(2222.5), 222250);
  assert.equal(rmToSen(1999), 199900);
});

test('rmToSen handles invalid input', () => {
  assert.equal(rmToSen('bad'), 0);
  assert.equal(rmToSen(-1), 0);
});
