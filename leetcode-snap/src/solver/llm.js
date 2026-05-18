const fetch = require('node-fetch');

const SYSTEM_PROMPT = `你是 LeetCode 解题专家。用户会给你一段 OCR 自截图识别出来的题面（可能含噪声）。

请你：
1. 先识别题目编号、标题与核心要求；
2. 给出 **多种** 主流解法（通常 2~4 种，从最朴素到最优解递进），每种都必须是可直接提交 AC 的完整代码；
3. 只用一种语言：{LANG}；
4. 每个方法包含：精炼的中文标题（≤12 字，如"暴力枚举"、"哈希表 O(n)"、"双指针"、"动态规划"、"单调栈"、"分治+剪枝"）、时间/空间复杂度、4-6 行思路说明。

针对 Go 语言的特别要求（仅当 {LANG} == go 时生效）：
- 不要写 package 声明，不要写 main 函数；
- 函数名和签名必须严格匹配 LeetCode Go 模板（如 twoSum、maxSubArray、reverseList，参数类型 []int / *ListNode / string / int 等）；
- 链表/二叉树题目假定 ListNode/TreeNode 已由系统提供，不要自己定义；
- 标准库直接 import "math" / "sort" 等；
- 优先使用切片、map、闭包，避免泛型；
- 函数首字母小写。

严格输出 JSON（**不要** markdown 代码块、**不要** 任何额外文字）：
{
  "title": "题号. 题目名称",
  "methods": [
    {
      "name": "方法标题（≤12字）",
      "complexity": "时间 O(...)，空间 O(...)",
      "code": "完整可AC代码",
      "explanation": "4-6 行思路说明"
    }
  ]
}

至少返回 1 种方法；如果题目本身只有唯一最优解就只给 1 种，否则尽量给 2-4 种。`;

async function solve(text, opts) {
  const { provider, endpoint, model, apiKey, language } = opts;
  const sys = SYSTEM_PROMPT.replace(/\{LANG\}/g, language);

  let url, headers, body;

  if (provider === 'anthropic') {
    url = endpoint || 'https://api.anthropic.com/v1/messages';
    headers = {
      'x-api-key': apiKey,
      'anthropic-version': '2023-06-01',
      'content-type': 'application/json'
    };
    body = {
      model: model || 'claude-3-5-sonnet-latest',
      max_tokens: 4096,
      system: sys,
      messages: [{ role: 'user', content: text.slice(0, 6000) }]
    };
  } else {
    url = endpoint || 'https://api.openai.com/v1/chat/completions';
    headers = {
      'Authorization': `Bearer ${apiKey}`,
      'Content-Type': 'application/json'
    };
    body = {
      model: model || 'gpt-4o-mini',
      temperature: 0.1,
      max_tokens: 4096,
      response_format: { type: 'json_object' },
      messages: [
        { role: 'system', content: sys },
        { role: 'user', content: text.slice(0, 6000) }
      ]
    };
  }

  const resp = await fetch(url, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
    timeout: 90000
  });
  if (!resp.ok) {
    const errText = await resp.text();
    throw new Error(`LLM 调用失败 ${resp.status}: ${errText.slice(0, 200)}`);
  }
  const data = await resp.json();

  let raw;
  if (provider === 'anthropic') {
    raw = data.content?.[0]?.text || '';
  } else {
    raw = data.choices?.[0]?.message?.content || '';
  }

  let parsed;
  try {
    parsed = JSON.parse(raw);
  } catch (e) {
    const m = raw.match(/\{[\s\S]*\}/);
    parsed = m ? JSON.parse(m[0]) : null;
  }

  if (!parsed) {
    return { title: '解析失败', methods: [{ name: '原始输出', complexity: '-', code: raw, explanation: '模型未按 JSON 格式返回，已显示原文' }], language };
  }

  // 兼容老格式（只返回 code/explanation 字段、没有 methods）
  if (!parsed.methods && parsed.code) {
    parsed.methods = [{
      name: '解法',
      complexity: '-',
      code: parsed.code,
      explanation: parsed.explanation || ''
    }];
  }

  if (!Array.isArray(parsed.methods) || parsed.methods.length === 0) {
    parsed.methods = [{ name: '解法', complexity: '-', code: '', explanation: '模型未返回有效方法' }];
  }

  return {
    title: parsed.title || 'LeetCode 题解',
    methods: parsed.methods.map(m => ({
      name: m.name || '解法',
      complexity: m.complexity || '-',
      code: m.code || '',
      explanation: m.explanation || ''
    })),
    language
  };
}

module.exports = { solve };
