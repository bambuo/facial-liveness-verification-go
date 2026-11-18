#!/usr/bin/env node
/**
 * 从 contracts/schema.json 生成 31 表基线 DDL（C1 新库独立、C5 应用层重写）。
 *
 * 用法:
 *   node contracts/tools/gen-ddl.mjs [--out contracts/ddl/001_baseline.sql]
 */
import fs from 'node:fs';
import path from 'node:path';

const args = process.argv.slice(2);
function argValue(name, fallback) {
  const i = args.indexOf(name);
  return i >= 0 && args[i + 1] ? args[i + 1] : fallback;
}

const ROOT = path.resolve(path.dirname(new URL(import.meta.url).pathname), '../..');
const SCHEMA = path.join(ROOT, 'contracts/schema.json');
const OUT = path.resolve(argValue('--out', path.join(ROOT, 'contracts/ddl/001_baseline.sql')));

const schema = JSON.parse(fs.readFileSync(SCHEMA, 'utf8'));

function quote(name) {
  return '"' + String(name).replace(/"/g, '""') + '"';
}

function constraintName(def) {
  const named = /^CONSTRAINT\s+("?[A-Za-z_][\w$]*"?)/i.exec(def);
  if (named) return named[1].replace(/"/g, '');
  const bare = /^("?[A-Za-z_][\w$]*"?)\s+(CHECK|UNIQUE|PRIMARY\s+KEY|FOREIGN\s+KEY)\b/i.exec(def);
  return bare ? bare[1].replace(/"/g, '') : null;
}

// formatConstraint 为 ALTER ADD CONSTRAINT 产生的裸定义补回 CONSTRAINT 关键字。
function formatConstraint(def) {
  if (/^CONSTRAINT\b/i.test(def)) return def;
  const name = constraintName(def);
  if (name) {
    const rest = def.replace(/^("?[A-Za-z_][\w$]*"?)\s+/, '');
    return `CONSTRAINT ${quote(name)} ${rest}`;
  }
  return def;
}

function dedupeBy(items, keyFn) {
  const map = new Map();
  for (const item of items) {
    const key = keyFn(item);
    if (!key) {
      map.set(Symbol(), item);
      continue;
    }
    map.set(key, item);
  }
  return [...map.values()];
}

const lines = [];
lines.push('-- 由 contracts/tools/gen-ddl.mjs 从 contracts/schema.json 生成。');
lines.push('-- 依据：C1 新库独立；C5 SQL 函数/触发器/RLS 由应用层重写。');
lines.push('-- 注意：外键 ON DELETE 语义来自列定义；未捕获的约束需在 M1 对账。');
lines.push('');

for (const table of schema.tables) {
  const defs = [];
  for (const col of table.columns) {
    let def = `${quote(col.name)} ${col.type}`;
    if (col.notNull) def += ' NOT NULL';
    if (col.default) def += ` DEFAULT ${col.default}`;
    if (col.primaryKey) def += ' PRIMARY KEY';
    if (col.unique) def += ' UNIQUE';
    defs.push(def);
  }
  const constraints = dedupeBy(table.constraints, constraintName);
  for (const c of constraints) defs.push(formatConstraint(c));
  lines.push(`CREATE TABLE ${quote(table.name)} (`);
  lines.push(defs.map((d) => '  ' + d).join(',\n'));
  lines.push(');');
  lines.push('');
}

const fkLines = [];
for (const table of schema.tables) {
  for (const col of table.columns) {
    if (!col.references) continue;
    let fk = `ALTER TABLE ${quote(table.name)} ADD FOREIGN KEY (${quote(col.name)}) REFERENCES ${quote(col.references)}`;
    if (col.onDelete) fk += ` ON DELETE ${col.onDelete}`;
    fkLines.push(fk + ';');
  }
}
if (fkLines.length > 0) {
  lines.push('-- 外键在所有表创建后添加，避免表顺序依赖。');
  lines.push(...fkLines);
  lines.push('');
}

const indexes = dedupeBy(schema.indexes || [], (idx) => idx.name);
for (const idx of indexes) {
  const unique = idx.unique ? 'UNIQUE ' : '';
  let sql = `CREATE ${unique}INDEX ${quote(idx.name)} ON ${quote(idx.table)} (${idx.columns.join(', ')})`;
  if (idx.where) sql += ` WHERE ${idx.where}`;
  lines.push(sql + ';');
}
lines.push('');

fs.mkdirSync(path.dirname(OUT), { recursive: true });
fs.writeFileSync(OUT, lines.join('\n'));
console.log(`wrote ${OUT} tables=${schema.tables.length} indexes=${indexes.length} fks=${fkLines.length}`);
