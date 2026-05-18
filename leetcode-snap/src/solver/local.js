const fs = require('fs');
const path = require('path');

const BANK_DIR = path.join(__dirname, 'bank');

function loadBank() {
  if (!fs.existsSync(BANK_DIR)) return [];
  return fs.readdirSync(BANK_DIR)
    .filter(f => f.endsWith('.json'))
    .map(f => JSON.parse(fs.readFileSync(path.join(BANK_DIR, f), 'utf8')));
}

const BANK = loadBank();

function normalize(s) {
  return (s || '').toLowerCase().replace(/\s+/g, ' ').trim();
}

function score(text, item) {
  const normText = normalize(text);
  let s = 0;
  for (const kw of item.keywords) {
    const k = kw.toLowerCase();
    if (normText.includes(k)) s += k.length >= 6 ? 3 : 2;
  }
  if (item.titleEn && normText.includes(item.titleEn.toLowerCase())) s += 5;
  if (item.titleCn && normText.includes(item.titleCn.toLowerCase())) s += 5;
  if (item.id && normText.includes(`leetcode ${item.id}`)) s += 4;
  return s;
}

function match(text, language = 'python') {
  if (!BANK.length) return null;
  let best = null;
  let bestScore = 0;
  for (const item of BANK) {
    const sc = score(text, item);
    if (sc > bestScore) {
      bestScore = sc;
      best = item;
    }
  }
  if (bestScore < 4) return null;

  const code = best.solutions[language] || best.solutions.python || Object.values(best.solutions)[0];
  return {
    title: `LC ${best.id}. ${best.titleEn} / ${best.titleCn}`,
    methods: [{
      name: '本地题库标准解法',
      complexity: best.complexity || '-',
      code,
      explanation: best.explanation
    }],
    language,
    matchScore: bestScore
  };
}

module.exports = { match, loadBank };
