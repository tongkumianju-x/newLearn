const fetch = require('node-fetch');

const SYSTEM_PROMPT = `你是 LeetCode 解题专家。用户会给你一段 OCR 自截图识别出来的题面（可能含噪声）。

请你：
1. 先识别题目编号、标题与核心要求；
2. 给出主流解法（**通常 1~4 种**，由你判断：如果题目只有一种最优解就只给 1 种，无需为了凑数硬造劣解；如有明显不同思路则递进给出多种）；
3. 只用一种语言：{LANG}；
4. 每个方法包含：精炼的中文标题（≤12 字，如"暴力枚举"、"哈希表 O(n)"、"双指针"、"动态规划"、"单调栈"、"分治+剪枝"）、时间/空间复杂度、4-6 行思路说明。

针对 Go 语言的特别要求（仅当 {LANG} == go 时生效）：
- 不要写 package 声明，不要写 main 函数；
- 函数名和签名必须严格匹配 LeetCode Go 模板（如 twoSum、maxSubArray、reverseList，参数类型 []int / *ListNode / string / int 等）；
- 链表/二叉树题目假定 ListNode/TreeNode 已由系统提供，不要自己定义；
- 标准库直接 import "math" / "sort" 等；
- 优先使用切片、map、闭包，避免泛型；
- 函数首字母小写。

**输出格式（极其重要，违反将无法解析）**：
- 必须是单个合法 JSON 对象，不要 markdown 代码块（不要 \`\`\`json），不要解释文字；
- 字符串里的换行用 \\n，双引号用 \\"，反斜杠用 \\\\；
- 代码不要超过 2000 字符/段。

JSON 结构：
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

至少返回 1 种方法（单解题也完全 OK）；多种思路则递进给出，最多 4 种。`;

// 强力 JSON 提取：剥 markdown 包装，修复尾部截断，try 多种模式
function extractJson(raw) {
  if (!raw) return null;
  let s = String(raw).trim();

  // 1. 剥 markdown 代码块包装 ```json ... ``` / ``` ... ```
  const codeBlock = s.match(/```(?:json)?\s*([\s\S]*?)\s*```/i);
  if (codeBlock) s = codeBlock[1].trim();

  // 2. 直接 parse
  try { return JSON.parse(s); } catch (_) {}

  // 3. 找第一个 { 到最后一个 } 的子串
  const first = s.indexOf('{');
  const last = s.lastIndexOf('}');
  if (first >= 0 && last > first) {
    const sub = s.slice(first, last + 1);
    try { return JSON.parse(sub); } catch (_) {}
  }

  // 4. 从第一个 { 开始，统计未闭合的 { [，自动补 } ]，并先关闭未闭合的字符串
  if (first >= 0) {
    let candidate = s.slice(first);
    // 4a. 关闭未闭合的字符串：如果有奇数个未转义的 "，末尾补一个 "
    const quoteMatches = candidate.match(/(?<!\\)"/g);
    const quotes = quoteMatches ? quoteMatches.length : 0;
    if (quotes % 2 === 1) candidate += '"';

    // 4b. 统计未闭合的 { [，按栈顺序补齐
    const stack = [];
    let inStr = false;
    let esc = false;
    for (let i = 0; i < candidate.length; i++) {
      const c = candidate[i];
      if (esc) { esc = false; continue; }
      if (c === '\\') { esc = true; continue; }
      if (c === '"') { inStr = !inStr; continue; }
      if (inStr) continue;
      if (c === '{' || c === '[') stack.push(c);
      else if (c === '}' && stack[stack.length - 1] === '{') stack.pop();
      else if (c === ']' && stack[stack.length - 1] === '[') stack.pop();
    }
    while (stack.length) {
      candidate += stack.pop() === '{' ? '}' : ']';
    }
    try { return JSON.parse(candidate); } catch (_) {}
  }

  return null;
}

async function callOnce(url, headers, body, controller, TIMEOUT) {
  const resp = await fetch(url, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
    signal: controller.signal,
    timeout: TIMEOUT
  });
  if (!resp.ok) {
    const errText = await resp.text();
    throw new Error(`LLM 调用失败 ${resp.status}: ${errText.slice(0, 200)}`);
  }
  return resp.json();
}

async function solve(text, opts) {
  const { provider, endpoint, model, apiKey, language, timeoutMs, onProgress } = opts;
  const sys = SYSTEM_PROMPT.replace(/\{LANG\}/g, language);

  const TIMEOUT = Math.max(30 * 1000, Math.min(30 * 60 * 1000, Number(timeoutMs) || 300000));

  let url, headers, body;
  let supportJsonMode = false;

  if (provider === 'anthropic') {
    url = endpoint || 'https://api.anthropic.com/v1/messages';
    headers = {
      'x-api-key': apiKey,
      'anthropic-version': '2023-06-01',
      'content-type': 'application/json'
    };
    body = {
      model: model || 'claude-3-5-sonnet-latest',
      max_tokens: 24576,
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
      max_tokens: 24576,
      response_format: { type: 'json_object' },
      messages: [
        { role: 'system', content: sys },
        { role: 'user', content: text.slice(0, 6000) }
      ]
    };
    supportJsonMode = true;
  }

  const controller = new AbortController();
  const startedAt = Date.now();
  let progressTimer = null;
  let aborted = false;

  if (typeof onProgress === 'function') {
    progressTimer = setInterval(() => {
      const elapsed = Math.floor((Date.now() - startedAt) / 1000);
      const remain = Math.max(0, Math.floor(TIMEOUT / 1000) - elapsed);
      onProgress({ elapsed, remain, timeoutSec: Math.floor(TIMEOUT / 1000) });
    }, 1000);
  }

  const abortTimer = setTimeout(() => {
    aborted = true;
    controller.abort();
  }, TIMEOUT);

  try {
    let data;
    try {
      data = await callOnce(url, headers, body, controller, TIMEOUT);
    } catch (err) {
      // JSON mode 部分服务商不支持，自动 retry 不带 response_format
      if (supportJsonMode && /response_format|json_object|unsupported|invalid/i.test(err.message || '')) {
        console.log('[llm] JSON mode 不支持，回退到普通模式重试');
        delete body.response_format;
        data = await callOnce(url, headers, body, controller, TIMEOUT);
      } else {
        throw err;
      }
    }

    clearTimeout(abortTimer);
    if (progressTimer) clearInterval(progressTimer);

    let raw;
    if (provider === 'anthropic') {
      raw = data.content?.[0]?.text || '';
    } else {
      raw = data.choices?.[0]?.message?.content || '';
    }

    if (!raw) {
      throw new Error('LLM 返回空内容，可能模型生成被过滤或 max_tokens 太小');
    }

    let parsed = extractJson(raw);

    if (!parsed) {
      // 真的解析不出来：把原始内容当 code 字段返回，至少让用户看到模型输出了什么
      console.log('[llm] JSON 解析失败，原文长度', raw.length);
      console.log('[llm] 原文前 500 字符:', raw.slice(0, 500));
      return {
        title: 'JSON 解析失败（已显示模型原文）',
        methods: [{
          name: '原始输出',
          complexity: '-',
          code: raw,
          explanation: `模型未按 JSON 格式返回。可能原因：1) 模型 (${model}) 不严格支持 JSON mode；2) max_tokens 被打满导致截断；3) 模型主动加了 markdown 包装但内容残缺。\n\n建议：到设置切换为 deepseek-chat 或 glm-4-flash 等较稳模型，或减小题面长度重试。`
        }],
        language,
        elapsedMs: Date.now() - startedAt,
        rawOutput: raw
      };
    }

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
      language,
      elapsedMs: Date.now() - startedAt
    };
  } catch (err) {
    clearTimeout(abortTimer);
    if (progressTimer) clearInterval(progressTimer);
    if (aborted || err.name === 'AbortError' || (err.type && err.type === 'aborted')) {
      const sec = Math.floor(TIMEOUT / 1000);
      throw new Error(`LLM 调用超时（已等待 ${sec}s）。请到设置面板把"LLM 超时时间"调大，或换个更快的模型/服务商。`);
    }
    throw err;
  }
}

module.exports = { solve };
