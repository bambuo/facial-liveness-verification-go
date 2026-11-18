#!/usr/bin/env node
/**
 * 从 idswyft 路由源码直接抽取响应契约：res.json(...) 的状态码与顶层字段，
 * 以及 buildVerificationResponse 的返回结构。
 *
 * 用法:
 *   node contracts/tools/extract-responses.mjs [--repo <idswyft-repo>] [--out <contracts-dir>]
 *
 * 输出:
 *   contracts/responses.json  按接口聚合的响应状态码与字段
 *   contracts/responses.md    人工评审表
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

function propName(node) {
  if (!node) return null;
  if (ts.isIdentifier(node) || ts.isStringLiteral(node) || ts.isNumericLiteral(node)) return node.text;
  return null;
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

function isResJsonCall(node) {
  if (!ts.isCallExpression(node)) return false;
  const e = node.expression;
  if (!ts.isPropertyAccessExpression(e) || e.name.text !== 'json') return false;
  const base = e.expression;
  if (ts.isIdentifier(base) && base.text === 'res') return true;
  if (ts.isCallExpression(base)) {
    const inner = base.expression;
    if (ts.isPropertyAccessExpression(inner)
      && ts.isIdentifier(inner.expression)
      && inner.expression.text === 'res'
      && inner.name.text === 'status') return true;
  }
  return false;
}

function statusOf(node) {
  const e = node.expression;
  if (ts.isPropertyAccessExpression(e) && ts.isCallExpression(e.expression)) {
    const inner = e.expression;
    if (ts.isPropertyAccessExpression(inner.expression)
      && ts.isIdentifier(inner.expression.expression)
      && inner.expression.expression.text === 'res'
      && inner.expression.name.text === 'status') {
      const arg = inner.arguments[0];
      if (arg && ts.isNumericLiteral(arg)) return arg.text;
      return arg ? arg.getText() : 'dynamic';
    }
  }
  return '200';
}

function shapeOf(entry, arg) {
  if (!arg) return { keys: [], kind: 'empty' };
  if (ts.isObjectLiteralExpression(arg)) {
    const keys = [];
    const spreads = [];
    for (const p of arg.properties) {
      if (ts.isPropertyAssignment(p)) {
        keys.push(propName(p.name) || p.name.getText(entry.sf));
      } else if (ts.isShorthandPropertyAssignment(p)) {
        keys.push(p.name.text);
      } else if (ts.isSpreadAssignment(p)) {
        spreads.push(p.expression.getText(entry.sf).replace(/\s+/g, ' ').slice(0, 120));
      } else if (ts.isMethodDeclaration(p)) {
        keys.push(propName(p.name) || '<method>');
      }
    }
    return { keys, spreads, kind: 'object' };
  }
  if (ts.isCallExpression(arg)) {
    const name = callName(arg);
    return { keys: [], ref: name || arg.expression.getText(entry.sf), kind: 'call' };
  }
  if (ts.isIdentifier(arg)) return { keys: [], ref: arg.text, kind: 'identifier' };
  return { keys: [], ref: arg.getText(entry.sf).replace(/\s+/g, ' ').slice(0, 120), kind: 'expr' };
}

function extractBuilderShape(file) {
  if (!fs.existsSync(file)) return null;
  const entry = load(file);
  let result = null;
  const visit = (node) => {
    if (ts.isFunctionDeclaration(node) && node.name && node.name.text === 'buildVerificationResponse' && node.body) {
      const findReturn = (n) => {
        if (ts.isReturnStatement(n) && n.expression && ts.isObjectLiteralExpression(n.expression)) {
          const keys = [];
          const spreads = [];
          for (const p of n.expression.properties) {
            if (ts.isPropertyAssignment(p)) keys.push(propName(p.name) || p.name.getText(entry.sf));
            else if (ts.isShorthandPropertyAssignment(p)) keys.push(p.name.text);
            else if (ts.isSpreadAssignment(p)) spreads.push(p.expression.getText(entry.sf).replace(/\s+/g, ' ').slice(0, 140));
          }
          return { keys, spreads };
        }
        let found = null;
        n.forEachChild((c) => {
          if (!found) found = findReturn(c);
        });
        return found;
      };
      result = { name: 'buildVerificationResponse', file: path.relative(REPO, file), line: lineOf(entry, node), ...findReturn(node.body) };
      return;
    }
    node.forEachChild(visit);
  };
  visit(entry.sf);
  return result;
}

function extractFileResponses(file) {
  const entry = load(file);
  const routes = [];
  const responses = [];
  const walk = (node) => {
    if (ts.isCallExpression(node)) {
      const e = node.expression;
      if (ts.isPropertyAccessExpression(e) && ts.isIdentifier(e.expression) && e.expression.text === 'router') {
        const routePath = node.arguments[0] && ts.isStringLiteral(node.arguments[0]) ? node.arguments[0].text : null;
        if (routePath !== null) {
          routes.push({ method: e.name.text.toUpperCase(), path: routePath, line: lineOf(entry, node), pos: node.getStart(entry.sf) });
        }
      }
      if (isResJsonCall(node)) {
        responses.push({
          status: statusOf(node),
          shape: shapeOf(entry, node.arguments[0]),
          line: lineOf(entry, node),
          pos: node.getStart(entry.sf),
        });
      }
    }
    node.forEachChild(walk);
  };
  walk(entry.sf);
  routes.sort((a, b) => a.pos - b.pos);
  return { routes, responses };
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
  lines.push('# idswyft 响应契约（M0 契约冻结输入）');
  lines.push('');
  lines.push(`- 来源：\`${result.source.repo}\` @ \`${result.source.commit}\``);
  lines.push(`- 生成时间：${result.generatedAt}`);
  lines.push(`- 口径：按源码位置就近归属到路由；同一接口聚合状态码与顶层字段。`);
  lines.push('');
  lines.push('## 汇总');
  lines.push('');
  lines.push('| 指标 | 数值 |');
  lines.push('| --- | --- |');
  lines.push(`| 接口 | ${result.statistics.endpoints} |`);
  lines.push(`| 有响应证据的接口 | ${result.statistics.endpointsWithResponses} |`);
  lines.push(`| res.json 调用 | ${result.statistics.responses} |`);
  lines.push(`| 状态码种类 | ${result.statistics.statusCodes} |`);
  lines.push('');
  lines.push('## 响应构建器');
  lines.push('');
  if (result.builder) {
    lines.push(`\`${result.builder.name}\`（\`${result.builder.file}:${result.builder.line}\`）`);
    lines.push('');
    lines.push(`- 顶层字段：${result.builder.keys.join(', ')}`);
    lines.push(`- 展开：${result.builder.spreads.join('；') || '-'}`);
  } else {
    lines.push('未找到 `buildVerificationResponse`。');
  }
  lines.push('');
  lines.push('## 接口响应');
  lines.push('');
  for (const ep of result.endpoints) {
    if (ep.responses.length === 0) continue;
    lines.push(`### ${ep.method} ${ep.path}`);
    lines.push('');
    lines.push(`来源：\`${ep.source.file}\``);
    lines.push('');
    lines.push('| 状态码 | 顶层字段 / 引用 | 源码位置 |');
    lines.push('| --- | --- | --- |');
    for (const r of ep.responses) {
      const shape = r.keys.length ? r.keys.join(', ') : r.ref ? `→ ${r.ref}` : '-';
      lines.push(`| ${r.status} | ${shape} | \`${ep.source.file}:${r.line}\` |`);
    }
    lines.push('');
  }
  lines.push('## 未归属的 res.json');
  lines.push('');
  if (result.unattributed.length === 0) lines.push('无。');
  for (const u of result.unattributed) lines.push(`- \`${u.file}:${u.line}\` status=${u.status} keys=${u.keys.join(', ') || u.ref || '-'}`);
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
  for (const ep of endpointsDoc.endpoints) {
    if (!byFile.has(ep.source.file)) byFile.set(ep.source.file, extractFileResponses(path.join(REPO, ep.source.file)));
  }

  const endpoints = [];
  const attributed = new Map();
  for (const ep of endpointsDoc.endpoints) {
    const data = byFile.get(ep.source.file);
    const route = data?.routes.find((r) => r.line === ep.source.line);
    if (!route) {
      endpoints.push({ method: ep.method, path: ep.path, module: ep.module, source: ep.source, responses: [] });
      continue;
    }
    const list = data.responses.filter((r) => r.pos >= route.pos && r.pos < (data.routes[data.routes.indexOf(route) + 1]?.pos ?? Infinity));
    for (const r of list) attributed.set(`${ep.source.file}:${r.line}`, true);
    endpoints.push({
      method: ep.method,
      path: ep.path,
      module: ep.module,
      source: ep.source,
      responses: list.map((r) => ({
        status: r.status,
        keys: r.shape.keys,
        spreads: r.shape.spreads || [],
        ref: r.shape.ref || null,
        kind: r.shape.kind,
        line: r.line,
      })),
    });
  }

  const builder = extractBuilderShape(path.join(REPO, 'backend/src/verification/statusReader.ts'));
  const unattributed = [];
  for (const [file, data] of byFile) {
    for (const r of data.responses) {
      if (!attributed.has(`${file}:${r.line}`)) {
        unattributed.push({ file, line: r.line, status: r.status, keys: r.shape.keys, ref: r.shape.ref || null });
      }
    }
  }

  const allResponses = endpoints.reduce((n, e) => n + e.responses.length, 0);
  const statusCodes = new Set();
  for (const e of endpoints) for (const r of e.responses) statusCodes.add(r.status);

  const result = {
    generatedAt: new Date().toISOString(),
    source: {
      repo: REPO,
      commit: execGit(REPO, ['rev-parse', 'HEAD']),
      commitDate: execGit(REPO, ['log', '-1', '--format=%cI']),
    },
    statistics: {
      endpoints: endpoints.length,
      endpointsWithResponses: endpoints.filter((e) => e.responses.length > 0).length,
      responses: allResponses,
      statusCodes: statusCodes.size,
    },
    builder,
    endpoints,
    unattributed,
    caveats: [
      '按源码位置就近归属；helper 函数内产生的 res.json 可能归属到相邻路由，需人工复核',
      '对象字面量只取顶层字段；嵌套结构与类型约束需结合 TypeScript 类型继续展开',
      'buildVerificationResponse 的展开表达式（条件对象）未逐字段展开',
      'C8 已决定 status 与 final_result 统一 12 值；本文件反映 idswyft 当前实现，落地时按 C8 改造',
    ],
  };

  fs.mkdirSync(OUT_DIR, { recursive: true });
  fs.writeFileSync(path.join(OUT_DIR, 'responses.json'), JSON.stringify(result, null, 2) + '\n');
  fs.writeFileSync(path.join(OUT_DIR, 'responses.md'), renderMarkdown(result));
  console.log(JSON.stringify(result.statistics, null, 2));
}

main();
