#!/usr/bin/env node
/**
 * 从 idswyft 源码抽取字段级契约：
 *   1. Zod schema（shared/src/verification/models）
 *   2. 路由请求字段（express-validator body/param/query、multer 文件字段）
 *   3. 状态/拒绝原因词表（含冲突检测）
 *
 * 用法:
 *   node contracts/tools/extract-fields.mjs [--repo <idswyft-repo>] [--out <contracts-dir>]
 *
 * 输出:
 *   contracts/fields.json  机器可读字段契约
 *   contracts/fields.md    人工评审表
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

function unwrap(node) {
  let cur = node;
  while (cur && (ts.isAsExpression(cur) || ts.isParenthesizedExpression(cur) || ts.isSatisfiesExpression(cur))) {
    cur = cur.expression;
  }
  return cur;
}

function findZodObject(call) {
  let cur = call;
  while (cur && ts.isCallExpression(cur)) {
    const callee = cur.expression;
    if (ts.isPropertyAccessExpression(callee)
      && callee.name.text === 'object'
      && ts.isIdentifier(callee.expression)
      && callee.expression.text === 'z') {
      return cur;
    }
    if (ts.isPropertyAccessExpression(callee)) {
      cur = callee.expression;
      continue;
    }
    break;
  }
  return null;
}

function callName(node) {
  if (!ts.isCallExpression(node)) return null;
  if (ts.isIdentifier(node.expression)) return node.expression.text;
  if (ts.isPropertyAccessExpression(node.expression)) return node.expression.name.text;
  return null;
}

function chainInfo(node) {
  // 从最外层链式调用向下拆：body('x').isEmail().optional()
  const methods = [];
  let cur = node;
  let root = null;
  let field = null;
  while (ts.isCallExpression(cur)) {
    methods.push(callName(cur) || '<call>');
    const callee = cur.expression;
    if (ts.isPropertyAccessExpression(callee)) {
      cur = callee.expression;
      continue;
    }
    if (ts.isIdentifier(callee)) {
      root = callee.text;
      const first = cur.arguments[0];
      if (first && ts.isStringLiteral(first)) field = first.text;
      break;
    }
    break;
  }
  return { root, field, methods: methods.reverse() };
}

const VALIDATOR_ROOTS = new Set(['body', 'param', 'query', 'cookie', 'header', 'check']);

function collectRequestArgs(node, out) {
  if (!node) return;
  if (ts.isArrayLiteralExpression(node)) {
    for (const el of node.elements) collectRequestArgs(el, out);
    return;
  }
  if (ts.isCallExpression(node)) {
    const info = chainInfo(node);
    if (info.root && VALIDATOR_ROOTS.has(info.root) && info.field) {
      out.push({ location: info.root === 'check' ? 'body' : info.root, name: info.field, validators: info.methods });
      return;
    }
    const name = callName(node);
    if (name === 'checkSchema' && node.arguments[0] && ts.isObjectLiteralExpression(node.arguments[0])) {
      for (const p of node.arguments[0].properties) {
        const field = propName(p.name);
        if (!field) continue;
        out.push({ location: 'body', name: field, validators: ['checkSchema'] });
      }
      return;
    }
    if (name === 'single' || name === 'array' || name === 'fields' || name === 'none') {
      const obj = node.expression;
      if (ts.isPropertyAccessExpression(obj) && ts.isIdentifier(obj.expression) && obj.expression.text === 'upload') {
        const first = node.arguments[0];
        if (first && ts.isStringLiteral(first)) out.push({ location: 'files', name: first.text, validators: [`upload.${name}`] });
      }
      return;
    }
  }
}

function extractEndpointFields(file) {
  const entry = load(file);
  const callsByLine = new Map();
  const walk = (node) => {
    if (ts.isCallExpression(node)) {
      const callee = node.expression;
      if (ts.isPropertyAccessExpression(callee) && ts.isIdentifier(callee.expression) && callee.expression.text === 'router') {
        const line = entry.sf.getLineAndCharacterOfPosition(node.getStart(entry.sf)).line + 1;
        callsByLine.set(line, node);
      }
    }
    node.forEachChild(walk);
  };
  walk(entry.sf);
  return callsByLine;
}

function extractZod() {
  const schemas = [];
  const vocabularies = [];
  const dir = path.join(REPO, 'shared/src');
  const files = [];
  const walkDir = (d) => {
    for (const e of fs.readdirSync(d, { withFileTypes: true })) {
      const p = path.join(d, e.name);
      if (e.isDirectory()) walkDir(p);
      else if (e.isFile() && p.endsWith('.ts')) files.push(p);
    }
  };
  walkDir(dir);

  for (const file of files) {
    const entry = load(file);
    const text = entry.text;
    if (!text.includes("from 'zod'") && !text.includes('from "zod"')) continue;
    const visit = (node) => {
      if (ts.isVariableDeclaration(node) && node.initializer && ts.isCallExpression(unwrap(node.initializer))) {
        const name = propName(node.name);
        const zodObject = findZodObject(unwrap(node.initializer));
        if (name && zodObject && zodObject.arguments[0] && ts.isObjectLiteralExpression(zodObject.arguments[0])) {
          const fields = [];
          for (const p of zodObject.arguments[0].properties) {
            const fname = propName(p.name);
            if (!fname || !p.initializer) continue;
            const initText = p.initializer.getText(entry.sf);
            fields.push({
              name: fname,
              type: initText.replace(/\s+/g, ' ').slice(0, 160),
              optional: /\boptional\(\)/.test(initText) || Boolean(p.questionToken),
              nullable: /\bnullable\(\)/.test(initText),
            });
          }
          const slice = text.slice(node.getStart(entry.sf), node.getEnd());
          const passthrough = /\.passthrough\(\)/.test(slice);
          schemas.push({
            name,
            file: path.relative(REPO, file),
            line: entry.sf.getLineAndCharacterOfPosition(node.getStart(entry.sf)).line + 1,
            passthrough,
            fields,
          });
        }
      }
      if (ts.isVariableStatement(node)) {
        const exported = node.modifiers?.some((m) => m.kind === ts.SyntaxKind.ExportKeyword);
        if (exported) {
          for (const decl of node.declarationList.declarations) {
            const name = propName(decl.name);
            const init = decl.initializer ? unwrap(decl.initializer) : null;
            if (!name || !init || !ts.isObjectLiteralExpression(init)) continue;
            const values = [];
            for (const p of init.properties) {
              if (ts.isPropertyAssignment(p) && ts.isStringLiteral(unwrap(p.initializer))) values.push(unwrap(p.initializer).text);
            }
            if (values.length > 0) vocabularies.push({ name, file: path.relative(REPO, file), values });
          }
        }
      }
      node.forEachChild(visit);
    };
    visit(entry.sf);
  }
  return { schemas, vocabularies };
}

function extractTypeUnions(file, wanted) {
  const entry = load(file);
  const result = [];
  const visit = (node) => {
    if (ts.isTypeAliasDeclaration(node) && wanted.has(node.name.text) && ts.isUnionTypeNode(node.type)) {
      const values = node.type.types
        .filter((t) => ts.isLiteralTypeNode(t) && ts.isStringLiteral(t.literal))
        .map((t) => t.literal.text);
      result.push({
        name: node.name.text,
        file: path.relative(REPO, file),
        line: entry.sf.getLineAndCharacterOfPosition(node.getStart(entry.sf)).line + 1,
        values,
      });
    }
    node.forEachChild(visit);
  };
  visit(entry.sf);
  return result;
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
  lines.push('# idswyft 字段契约（M0 契约冻结输入）');
  lines.push('');
  lines.push(`- 来源：\`${result.source.repo}\` @ \`${result.source.commit}\``);
  lines.push(`- 生成时间：${result.generatedAt}`);
  lines.push('');
  lines.push('## 汇总');
  lines.push('');
  lines.push('| 指标 | 数值 |');
  lines.push('| --- | --- |');
  lines.push(`| Zod schema | ${result.statistics.zodSchemas} |`);
  lines.push(`| Zod 字段 | ${result.statistics.zodFields} |`);
  lines.push(`| 有请求字段的接口 | ${result.statistics.endpointsWithFields} / ${result.statistics.endpoints} |`);
  lines.push(`| 请求字段总数 | ${result.statistics.requestFields} |`);
  lines.push(`| 状态词表 | ${result.statistics.vocabularies} |`);
  lines.push('');
  lines.push('## 状态词表与冲突');
  lines.push('');
  lines.push('| 名称 | 值数量 | 值 | 来源 |');
  lines.push('| --- | --- | --- | --- |');
  for (const v of result.vocabularies) {
    lines.push(`| ${v.name} | ${v.values.length} | ${v.values.join(', ')} | ${v.file} |`);
  }
  lines.push('');
  for (const c of result.vocabularyConflicts) lines.push(`- **冲突**：${c}`);
  lines.push('');
  lines.push('## Zod schema');
  lines.push('');
  for (const s of result.zodSchemas) {
    lines.push(`### ${s.name}`);
    lines.push('');
    lines.push(`来源：\`${s.file}:${s.line}\`${s.passthrough ? '（passthrough）' : ''}`);
    lines.push('');
    lines.push('| 字段 | 类型 | 可选 | 可空 |');
    lines.push('| --- | --- | --- | --- |');
    for (const f of s.fields) lines.push(`| ${f.name} | \`${f.type}\` | ${f.optional ? '是' : ''} | ${f.nullable ? '是' : ''} |`);
    lines.push('');
  }
  lines.push('## 接口请求字段');
  lines.push('');
  lines.push('| 方法 | 路径 | body | query | param | files |');
  lines.push('| --- | --- | --- | --- | --- | --- |');
  for (const ep of result.endpoints) {
    const fmt = (arr) => (arr.length ? arr.map((f) => `${f.name}(${f.validators.join('.')})`).join('<br>') : '-');
    lines.push(`| ${ep.method} | \`${ep.path}\` | ${fmt(ep.body)} | ${fmt(ep.query)} | ${fmt(ep.param)} | ${fmt(ep.files)} |`);
  }
  lines.push('');
  lines.push('## 说明');
  lines.push('');
  lines.push('- 请求字段来自 express-validator 链与 multer 字段；未覆盖 handler 内直接读 `req.body.x` 的分支。');
  lines.push('- 响应字段尚未逐接口展开；M0 下一步从 shared/backend 类型与 `res.json` 结构生成。');
  lines.push('- 状态词表存在多套口径，必须在 C8 决策后统一。');
  lines.push('');
  return lines.join('\n');
}

function main() {
  const endpointsDoc = JSON.parse(fs.readFileSync(ENDPOINTS, 'utf8'));
  const { schemas, vocabularies } = extractZod();
  const unions = [
    ...extractTypeUnions(path.join(REPO, 'backend/src/services/verificationStateMachine.ts'), new Set(['VerificationState', 'VerificationEvent'])),
    ...extractTypeUnions(path.join(REPO, 'backend/src/types/index.ts'), new Set(['VerificationStatus'])),
  ];
  const allVocabularies = [...vocabularies, ...unions];

  const byFile = new Map();
  for (const ep of endpointsDoc.endpoints) {
    if (!byFile.has(ep.source.file)) byFile.set(ep.source.file, extractEndpointFields(path.join(REPO, ep.source.file)));
  }

  const endpoints = [];
  for (const ep of endpointsDoc.endpoints) {
    const calls = byFile.get(ep.source.file);
    const call = calls?.get(ep.source.line);
    const fields = [];
    if (call) {
      for (const arg of call.arguments.slice(1, -1)) collectRequestArgs(arg, fields);
    }
    const group = (loc) => fields.filter((f) => f.location === loc).map(({ name, validators }) => ({ name, validators }));
    endpoints.push({
      method: ep.method,
      path: ep.path,
      module: ep.module,
      source: ep.source,
      body: group('body'),
      query: group('query'),
      param: group('param'),
      files: group('files'),
    });
  }

  const statusNames = new Map();
  for (const v of allVocabularies) {
    if (!statusNames.has(v.name)) statusNames.set(v.name, []);
    statusNames.get(v.name).push(v);
  }
  const vocabularyConflicts = [];
  for (const [name, list] of statusNames) {
    if (list.length > 1) {
      const unique = new Set(list.map((x) => x.values.join('|')));
      if (unique.size > 1) vocabularyConflicts.push(`${name} 有 ${list.length} 套不同取值：${list.map((x) => `${x.file}(${x.values.length})`).join('；')}`);
    }
  }
  if (statusNames.has('VerificationStatus') && statusNames.has('VerificationState')) {
    vocabularyConflicts.push('对外 VerificationStatus 与内部 VerificationState 是两套状态；API 响应 status 使用前者，内部状态机使用后者');
  }

  const requestFieldCount = endpoints.reduce((n, e) => n + e.body.length + e.query.length + e.param.length + e.files.length, 0);
  const result = {
    generatedAt: new Date().toISOString(),
    source: {
      repo: REPO,
      commit: execGit(REPO, ['rev-parse', 'HEAD']),
      commitDate: execGit(REPO, ['log', '-1', '--format=%cI']),
    },
    statistics: {
      endpoints: endpoints.length,
      endpointsWithFields: endpoints.filter((e) => e.body.length + e.query.length + e.param.length + e.files.length > 0).length,
      requestFields: requestFieldCount,
      zodSchemas: schemas.length,
      zodFields: schemas.reduce((n, s) => n + s.fields.length, 0),
      vocabularies: allVocabularies.length,
    },
    zodSchemas: schemas,
    vocabularies: allVocabularies,
    vocabularyConflicts,
    endpoints,
    caveats: [
      '请求字段来自 express-validator 链与 multer；handler 内直接读 req.body 的分支未覆盖',
      '响应字段未逐接口展开，需从类型与 res.json 结构继续抽取',
      '状态词表多套口径，C8 决策后统一',
    ],
  };

  fs.mkdirSync(OUT_DIR, { recursive: true });
  fs.writeFileSync(path.join(OUT_DIR, 'fields.json'), JSON.stringify(result, null, 2) + '\n');
  fs.writeFileSync(path.join(OUT_DIR, 'fields.md'), renderMarkdown(result));
  console.log(JSON.stringify(result.statistics, null, 2));
  console.log('conflicts:', result.vocabularyConflicts.length);
}

main();
