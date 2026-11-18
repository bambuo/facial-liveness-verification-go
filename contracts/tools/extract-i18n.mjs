#!/usr/bin/env node
/**
 * 生成错误码 / i18n messageKey 对账表。
 *
 * 用法:
 *   node contracts/tools/extract-i18n.mjs [--repo <idswyft-repo>] [--out <contracts-dir>]
 */
import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';

const args = process.argv.slice(2);
function argValue(name, fallback) {
  const i = args.indexOf(name);
  return i >= 0 && args[i + 1] ? args[i + 1] : fallback;
}

const REPO = path.resolve(argValue('--repo', '/Users/johana/Codes/github/idswyft-community-1.12.39'));
const OUT_DIR = path.resolve(argValue('--out', path.join(process.cwd(), 'contracts')));

function flatten(obj, prefix = '', out = {}) {
  for (const [k, v] of Object.entries(obj)) {
    const key = prefix ? `${prefix}.${k}` : k;
    if (v && typeof v === 'object' && !Array.isArray(v)) flatten(v, key, out);
    else out[key] = String(v);
  }
  return out;
}

function execGit(repo, gitArgs) {
  try {
    return execFileSync('git', ['-C', repo, ...gitArgs], { encoding: 'utf8' }).trim();
  } catch {
    return null;
  }
}

function walk(dir, files = []) {
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) walk(p, files);
    else if (e.isFile() && (p.endsWith('.ts') || p.endsWith('.tsx'))) files.push(p);
  }
  return files;
}

function cell(text) {
  return String(text ?? '').replace(/\|/g, '\\|').replace(/\s+/g, ' ').trim();
}

const en = flatten(JSON.parse(fs.readFileSync(path.join(REPO, 'backend/src/i18n/locales/en.json'), 'utf8')));
const zh = flatten(JSON.parse(fs.readFileSync(path.join(REPO, 'backend/src/i18n/locales/zh-CN.json'), 'utf8')));

const known = new Set([...Object.keys(en), ...Object.keys(zh)]);
const usages = new Map();
const files = [...walk(path.join(REPO, 'backend/src')), ...walk(path.join(REPO, 'shared/src'))];
const quoted = /(['"])([A-Za-z][A-Za-z0-9_.]*\.[A-Za-z0-9_.]+)\1/g;
for (const file of files) {
  const text = fs.readFileSync(file, 'utf8');
  let m;
  while ((m = quoted.exec(text))) {
    const key = m[2];
    if (!known.has(key) && !key.startsWith('errors.')) continue;
    const rel = path.relative(REPO, file);
    if (!usages.has(key)) usages.set(key, { count: 0, files: new Set() });
    const item = usages.get(key);
    item.count++;
    item.files.add(rel);
  }
}

const allKeys = [...new Set([...known, ...usages.keys()])].sort();
const missingEn = allKeys.filter((k) => !(k in en));
const missingZh = allKeys.filter((k) => !(k in zh));
const usedUndefined = [...usages.keys()].filter((k) => !known.has(k)).sort();
const unused = [...known].filter((k) => !usages.has(k)).sort();

const lines = [];
lines.push('# 错误码与 i18n 对账表（P0）');
lines.push('');
lines.push(`- 来源：\`${REPO}\` @ \`${execGit(REPO, ['rev-parse', 'HEAD'])}\``);
lines.push('- 口径：en / zh-CN locale 全量键 + 源码中的 `messageKey` 字面量引用');
lines.push('');
lines.push('## 汇总');
lines.push('');
lines.push('| 指标 | 数值 |');
lines.push('| --- | --- |');
lines.push(`| en 键 | ${Object.keys(en).length} |`);
lines.push(`| zh-CN 键 | ${Object.keys(zh).length} |`);
lines.push(`| 合并键 | ${allKeys.length} |`);
lines.push(`| 缺少 en | ${missingEn.length} |`);
lines.push(`| 缺少 zh-CN | ${missingZh.length} |`);
lines.push(`| 源码引用但未定义 | ${usedUndefined.length} |`);
lines.push(`| 已定义但未引用 | ${unused.length} |`);
lines.push('');
lines.push('## 对账明细');
lines.push('');
lines.push('| Key | en | zh-CN | 引用次数 | 引用文件 | 状态 |');
lines.push('| --- | --- | --- | --- | --- | --- |');
for (const key of allKeys) {
  const u = usages.get(key);
  const fileList = u ? [...u.files].slice(0, 3).join('<br>') + (u.files.size > 3 ? `<br>…(${u.files.size})` : '') : '-';
  const status = !(key in en) ? '缺少 en' : !(key in zh) ? '缺少 zh-CN' : u ? 'ok' : '未引用';
  lines.push(`| ${cell(key)} | ${cell(en[key] ?? '')} | ${cell(zh[key] ?? '')} | ${u ? u.count : 0} | ${fileList} | ${status} |`);
}
lines.push('');
if (usedUndefined.length > 0) {
  lines.push('## 源码引用但未定义');
  lines.push('');
  for (const key of usedUndefined) lines.push(`- ${key}`);
  lines.push('');
}
if (missingEn.length > 0 || missingZh.length > 0) {
  lines.push('## 缺失翻译');
  lines.push('');
  for (const key of missingEn) lines.push(`- ${key}（缺少 en）`);
  for (const key of missingZh) lines.push(`- ${key}（缺少 zh-CN）`);
  lines.push('');
}
lines.push('## 说明');
lines.push('');
lines.push('- 代码中的 `messageKey` 必须同时存在于 en 与 zh-CN；缺失项需在 M1 补齐。');
lines.push('- 未引用键可能是动态拼接或历史遗留，需人工确认后删除或保留。');
lines.push('- 本表由 `contracts/tools/extract-i18n.mjs` 生成，可复算。');
lines.push('');

fs.mkdirSync(OUT_DIR, { recursive: true });
fs.writeFileSync(path.join(OUT_DIR, 'errors-i18n.md'), lines.join('\n'));
console.log(JSON.stringify({ en: Object.keys(en).length, zh: Object.keys(zh).length, keys: allKeys.length, missingEn: missingEn.length, missingZh: missingZh.length, usedUndefined: usedUndefined.length, unused: unused.length }, null, 2));
