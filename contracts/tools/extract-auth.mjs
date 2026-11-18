#!/usr/bin/env node
/**
 * 直接抽取 idswyft 路由 handler 内的鉴权证据，解决 endpoints.json 中 55 条
 * none-or-inline 的“待复核”项。
 *
 * 用法:
 *   node contracts/tools/extract-auth.mjs [--repo <idswyft-repo>] [--out <contracts-dir>]
 *
 * 输出:
 *   contracts/auth.json      按接口的鉴权分类与源码证据
 *   contracts/auth-review.md 人工评审表
 */
import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import ts from 'typescript';

const args = process.argv.slice(2);
function argValue(name, fallback) {
  const i = args.indexOf(name);
  return i >= 0 && args[i + 1] ? args[i + 1] : fallback;
}

const REPO = path.resolve(argValue('--repo', '/Users/johana/Codes/github/idswyft-community-1.12.39'));
const OUT_DIR = path.resolve(argValue('--out', path.join(process.cwd(), 'contracts')));
const ENDPOINTS = path.join(OUT_DIR, 'endpoints.json');

const cache = new Map();
function load(file) {
  if (cache.has(file)) return cache.get(file);
  const text = fs.readFileSync(file, 'utf8');
  const sf = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS);
  const entry = { file, text, sf };
  cache.set(file, entry);
  return entry;
}

function lineOf(entry, node) {
  return entry.sf.getLineAndCharacterOfPosition(node.getStart(entry.sf)).line + 1;
}

function callName(node) {
  if (!ts.isCallExpression(node)) return null;
  if (ts.isIdentifier(node.expression)) return node.expression.text;
  if (ts.isPropertyAccessExpression(node.expression)) return node.expression.name.text;
  return null;
}

const MECHANISMS = [
  ['authenticateComplianceRequest', 'compliance-request'],
  ['authenticateAdminOrReviewer', 'admin-or-reviewer-jwt'],
  ['authenticateReviewPrincipal', 'review-principal'],
  ['authenticateReviewerJWT', 'reviewer-jwt'],
  ['authenticateDeveloperJWT', 'developer-jwt'],
  ['authenticateAPIKey', 'api-key'],
  ['authenticateServiceToken', 'service-token'],
  ['authenticateHandoffToken', 'handoff-token'],
  ['authenticateUser', 'user-jwt'],
  ['authenticateJWT', 'jwt'],
  ['verifyRegistrationToken', 'registration-token'],
  ['verifyToken', 'jwt'],
  ['hashHandoffToken', 'handoff-token'],
  ['scopeForRequest', 'developer-scope'],
  ['getComplianceDeveloperId', 'compliance-request'],
];

const MEMBERS = [
  ['req.developer', 'developer-context'],
  ['req.reviewer', 'reviewer-context'],
  ['req.admin', 'admin-context'],
  ['req.user', 'user-context'],
  ['req.apiKey', 'api-key-context'],
  ['req.session', 'session-context'],
  ['req.params.token', 'path-token'],
  ['req.params.shareToken', 'share-token'],
  ['req.params.jti', 'path-jti'],
  ['req.body.api_key', 'api-key-body'],
  ['req.query.api_key', 'api-key-query'],
  ['req.query.key', 'api-key-query'],
  ['req.cookies', 'cookie'],
  ['jwt.verify', 'jwt'],
];

const HEADERS = [
  ['x-api-key', 'api-key-header'],
  ['authorization', 'authorization-header'],
  ['x-session-token', 'session-token-header'],
];

function handlerBodies(node) {
  const out = [];
  const visit = (n) => {
    if (!n) return;
    if (ts.isArrowFunction(n) || ts.isFunctionExpression(n)) {
      out.push(n);
      return;
    }
    if (ts.isCallExpression(n)) {
      for (const a of n.arguments) visit(a);
      return;
    }
    if (ts.isAsExpression(n) || ts.isParenthesizedExpression(n) || ts.isSatisfiesExpression(n)) {
      visit(n.expression);
    }
  };
  visit(node);
  return out;
}

function extractRouteCalls(file) {
  const entry = load(file);
  const calls = new Map();
  const walk = (node) => {
    if (ts.isCallExpression(node)) {
      const e = node.expression;
      if (ts.isPropertyAccessExpression(e) && ts.isIdentifier(e.expression) && e.expression.text === 'router') {
        calls.set(lineOf(entry, node), node);
      }
    }
    node.forEachChild(walk);
  };
  walk(entry.sf);
  return { entry, calls };
}

function evidenceFromHandler(entry, handler) {
  const evidence = [];
  const push = (mechanism, node, detail) => {
    evidence.push({ mechanism, detail, line: lineOf(entry, node) });
  };
  const visit = (node) => {
    if (ts.isCallExpression(node)) {
      const name = callName(node);
      if (name) {
        for (const [fn, mech] of MECHANISMS) {
          if (name === fn) push(mech, node, `${fn}()`);
        }
        if (name === 'clearCookie') push('cookie', node, 'res.clearCookie()');
        if (name === 'generateToken') push('csrf', node, 'generateToken()');
        if (name === 'from' && node.arguments[0] && ts.isStringLiteral(node.arguments[0])) {
          const table = node.arguments[0].text;
          if (['api_keys', 'developers', 'verification_reviewers', 'users'].includes(table)) {
            push('table-lookup', node, `from('${table}')`);
          }
        }
      }
    }
    if (ts.isVariableDeclaration(node) && node.initializer
      && ts.isObjectBindingPattern(node.name)
      && node.initializer.getText(entry.sf).startsWith('req.params')) {
      for (const el of node.name.elements) {
        const key = el.name.getText(entry.sf);
        if (key === 'shareToken') push('share-token', node, 'const { shareToken } = req.params');
        else if (key === 'token') push('path-token', node, 'const { token } = req.params');
        else if (key === 'jti') push('path-jti', node, 'const { jti } = req.params');
      }
    }
    if (ts.isPropertyAccessExpression(node)) {
      const text = node.getText(entry.sf);
      for (const [member, mech] of MEMBERS) {
        if (text === member) push(mech, node, member);
      }
    }
    if (ts.isElementAccessExpression(node)) {
      const text = node.getText(entry.sf);
      for (const [header, mech] of HEADERS) {
        if (text === `req.headers['${header}']` || text === `req.headers["${header}"]`) {
          push(mech, node, `req.headers['${header}']`);
        }
      }
    }
    node.forEachChild(visit);
  };
  visit(handler);
  const dedup = new Map();
  for (const e of evidence) dedup.set(`${e.mechanism}:${e.line}:${e.detail}`, e);
  return [...dedup.values()];
}

function classify(middleware, evidence) {
  const mw = middleware.filter((m) => !['inline', 'validate', 'conditionalCsrf'].includes(m));
  if (mw.some((m) => m.startsWith('authenticate') || m.startsWith('require'))) {
    return { auth: 'middleware', mechanisms: [...new Set(mw.filter((m) => m.startsWith('authenticate') || m.startsWith('require')))] };
  }
  const strong = evidence.filter((e) => e.mechanism !== 'table-lookup');
  if (strong.length > 0) {
    return { auth: 'inline', mechanisms: [...new Set(strong.map((e) => e.mechanism))] };
  }
  return { auth: 'public-candidate', mechanisms: [] };
}

function execGit(repo, gitArgs) {
  try {
    return execFileSync('git', ['-C', repo, ...gitArgs], { encoding: 'utf8' }).trim();
  } catch {
    return null;
  }
}

function renderMarkdown(result) {
  const lines = [];
  lines.push('# idswyft 鉴权证据（M0 契约冻结输入）');
  lines.push('');
  lines.push(`- 来源：\`${result.source.repo}\` @ \`${result.source.commit}\``);
  lines.push(`- 生成时间：${result.generatedAt}`);
  lines.push('- 口径：中间件鉴权来自 endpoints.json；内联鉴权来自 handler 内调用/成员/请求头；无证据标记为 public-candidate，仍需人工签字。');
  lines.push('');
  lines.push('## 汇总');
  lines.push('');
  lines.push('| 分类 | 数量 |');
  lines.push('| --- | --- |');
  for (const [k, v] of Object.entries(result.statistics.byAuth)) lines.push(`| ${k} | ${v} |`);
  lines.push('');
  lines.push('## 内联鉴权证据');
  lines.push('');
  lines.push('| 方法 | 路径 | 机制 | 证据 |');
  lines.push('| --- | --- | --- | --- |');
  for (const ep of result.endpoints.filter((e) => e.auth === 'inline')) {
    const ev = ep.evidence.map((x) => `${x.detail}@${x.line}`).join('<br>');
    lines.push(`| ${ep.method} | \`${ep.path}\` | ${ep.mechanisms.join(', ')} | ${ev} |`);
  }
  lines.push('');
  lines.push('## public-candidate（需签字）');
  lines.push('');
  lines.push('| 方法 | 路径 | 模块 | 说明 |');
  lines.push('| --- | --- | --- | --- |');
  for (const ep of result.endpoints.filter((e) => e.auth === 'public-candidate')) {
    lines.push(`| ${ep.method} | \`${ep.path}\` | ${ep.module} | 未发现中间件或内联鉴权证据 |`);
  }
  lines.push('');
  lines.push('## 说明');
  lines.push('');
  for (const c of result.caveats) lines.push(`- ${c}`);
  lines.push('');
  return lines.join('\n');
}

function main() {
  const endpointsDoc = JSON.parse(fs.readFileSync(ENDPOINTS, 'utf8'));
  const byFile = new Map();
  const endpoints = [];

  for (const ep of endpointsDoc.endpoints) {
    if (!byFile.has(ep.source.file)) byFile.set(ep.source.file, extractRouteCalls(path.join(REPO, ep.source.file)));
    const { entry, calls } = byFile.get(ep.source.file);
    const call = calls.get(ep.source.line);
    let evidence = [];
    if (call && call.arguments.length >= 2) {
      const handler = call.arguments[call.arguments.length - 1];
      for (const body of handlerBodies(handler)) {
        evidence = evidence.concat(evidenceFromHandler(entry, body));
      }
    }
    const cls = classify(ep.middleware, evidence);
    endpoints.push({
      method: ep.method,
      path: ep.path,
      module: ep.module,
      source: ep.source,
      middlewareAuth: ep.auth,
      auth: cls.auth,
      mechanisms: cls.mechanisms,
      evidence,
    });
  }

  const byAuth = {};
  for (const e of endpoints) byAuth[e.auth] = (byAuth[e.auth] || 0) + 1;

  const result = {
    generatedAt: new Date().toISOString(),
    source: {
      repo: REPO,
      commit: execGit(REPO, ['rev-parse', 'HEAD']),
      commitDate: execGit(REPO, ['log', '-1', '--format=%cI']),
    },
    statistics: { endpoints: endpoints.length, byAuth },
    endpoints,
    caveats: [
      '内联鉴权按调用名/成员访问/请求头匹配，不判断完整分支逻辑；public-candidate 必须人工签字',
      'authenticateComplianceRequest 的内部实现可能是 API Key 或 Reviewer JWT，需在 M1 打开实现确认',
      'C3 已决定完全采用 idswyft 认证模型；本文件作为 M1 鉴权矩阵的实现依据',
    ],
  };

  fs.mkdirSync(OUT_DIR, { recursive: true });
  fs.writeFileSync(path.join(OUT_DIR, 'auth.json'), JSON.stringify(result, null, 2) + '\n');
  fs.writeFileSync(path.join(OUT_DIR, 'auth-review.md'), renderMarkdown(result));
  console.log(JSON.stringify(result.statistics, null, 2));
}

main();
