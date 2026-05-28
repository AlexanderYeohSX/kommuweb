/** Convert major currency units (e.g. RM) to sen (smallest unit). */
function rmToSen(amount) {
  const n = parseFloat(amount);
  if (Number.isNaN(n) || n < 0) return 0;
  return Math.round(n * 100);
}

module.exports = { rmToSen };
