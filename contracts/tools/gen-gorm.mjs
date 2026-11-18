#!/usr/bin/env node
/**
 * 从 contracts/schema.json 生成 GORM 模型与 BelongsTo 关联。
 *
 * 用法:
 *   node contracts/tools/gen-gorm.mjs [--out internal/adapter/pg/gormmodel]
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
const OUT = path.resolve(argValue('--out', path.join(ROOT, 'internal/adapter/pg/gormmodel')));

const schema = JSON.parse(fs.readFileSync(SCHEMA, 'utf8'));

const STRUCT = {
  users: 'User', developers: 'Developer', api_keys: 'APIKey',
  verification_requests: 'VerificationRequest', verification_contexts: 'VerificationContext',
  documents: 'Document', selfies: 'Selfie', webhooks: 'Webhook', webhook_deliveries: 'WebhookDelivery',
  batch_jobs: 'BatchJob', batch_items: 'BatchItem',
  compliance_rulesets: 'ComplianceRuleset', compliance_rules: 'ComplianceRule',
  identity_vault: 'IdentityVault', vault_share_links: 'VaultShareLink',
  verifiable_credentials: 'VerifiableCredential', mobile_handoff_sessions: 'MobileHandoffSession',
  idempotency_keys: 'IdempotencyKey', rate_limits: 'RateLimit', provider_metrics: 'ProviderMetric',
  api_activity_logs: 'APIActivityLog', expiry_alerts: 'ExpiryAlert',
  reverification_schedules: 'ReverificationSchedule', dedup_fingerprints: 'DedupFingerprint',
  verification_risk_scores: 'VerificationRiskScore', verification_reviewers: 'VerificationReviewer',
  developer_otp_codes: 'DeveloperOTPCode', developer_otp_rate_limits: 'DeveloperOTPRateLimit',
  phone_otp_codes: 'PhoneOTPCode', phone_otp_rate_limits: 'PhoneOTPRateLimit', aml_screenings: 'AMLScreening',
};

const INITIALISMS = new Set(['ID', 'API', 'OTP', 'AML', 'URL', 'IP', 'OCR', 'MRZ', 'JWT', 'UUID']);

function pascal(name) {
  return name.split('_').filter(Boolean).map((part) => {
    const upper = part.toUpperCase();
    if (INITIALISMS.has(upper)) return upper;
    if (upper === 'GITHUB') return 'GitHub';
    return part.charAt(0).toUpperCase() + part.slice(1);
  }).join('');
}

function fieldName(column) { return pascal(column); }

function assocName(column) {
  const base = column.endsWith('_id') ? column.slice(0, -3) : column;
  return pascal(base);
}

function parentPKField(refTable) {
  const t = schema.tables.find((x) => x.name === refTable);
  if (!t) return 'ID';
  const pk = t.columns.find((c) => c.primaryKey);
  return pk ? fieldName(pk.name) : 'ID';
}

function typeInfo(type, notNull) {
  const t = type.toUpperCase();
  const ptr = notNull ? '' : '*';
  if (t === 'UUID') return { go: ptr + 'uuid.UUID', tag: 'uuid', imp: 'uuid' };
  if (t === 'TEXT' || t.startsWith('VARCHAR') || t.startsWith('CHAR')) return { go: ptr + 'string', tag: type.toLowerCase(), imp: null };
  if (t === 'BOOLEAN') return { go: ptr + 'bool', tag: 'boolean', imp: null };
  if (t === 'JSONB') return { go: 'json.RawMessage', tag: 'jsonb', imp: 'json' };
  if (t === 'TIMESTAMPTZ' || t === 'TIMESTAMP WITH TIME ZONE') return { go: ptr + 'time.Time', tag: 'timestamptz', imp: 'time' };
  if (t === 'TIMESTAMP') return { go: ptr + 'time.Time', tag: 'timestamp', imp: 'time' };
  if (t === 'DATE') return { go: ptr + 'time.Time', tag: 'date', imp: 'time' };
  if (t === 'INT' || t === 'INTEGER') return { go: ptr + 'int32', tag: 'integer', imp: null };
  if (t === 'BIGINT') return { go: ptr + 'int64', tag: 'bigint', imp: null };
  if (t === 'SMALLINT') return { go: ptr + 'int16', tag: 'smallint', imp: null };
  if (t === 'REAL') return { go: ptr + 'float32', tag: 'real', imp: null };
  if (t.startsWith('DECIMAL')) return { go: ptr + 'float64', tag: 'numeric' + type.slice(type.indexOf('(')), imp: null };
  if (t === 'TEXT[]') return { go: 'pq.StringArray', tag: 'text[]', imp: 'pq' };
  return { go: ptr + 'string', tag: type.toLowerCase(), imp: null };
}

function constraintName(def) {
  const named = /^CONSTRAINT\s+("?[A-Za-z_][\w$]*"?)/i.exec(def);
  if (named) return named[1].replace(/"/g, '');
  const bare = /^("?[A-Za-z_][\w$]*"?)\s+(CHECK|UNIQUE|PRIMARY\s+KEY|FOREIGN\s+KEY)\b/i.exec(def);
  return bare ? bare[1].replace(/"/g, '') : null;
}

function parseCheck(def) {
  const m = /^(?:CONSTRAINT\s+)?("?[A-Za-z_][\w$]*"?)\s+CHECK\s*\(([\s\S]*)\)$/i.exec(def.trim());
  if (!m) return null;
  return { name: m[1].replace(/"/g, ''), expr: m[2].trim() };
}

function dedupeBy(items, keyFn) {
  const map = new Map();
  for (const item of items) {
    const key = keyFn(item);
    if (!key) map.set(Symbol(), item);
    else map.set(key, item);
  }
  return [...map.values()];
}

function parseIndexColumn(raw) {
  let s = raw.trim().replace(/"/g, '');
  let sort = null;
  const m = /^(.+?)\s+(ASC|DESC)$/i.exec(s);
  if (m) {
    s = m[1].trim();
    sort = m[2].toLowerCase();
  }
  return { name: s, sort };
}

function compositePKOrder(table) {
  const cons = table.constraints.find((c) => /PRIMARY\s+KEY/i.test(c));
  if (!cons) return [];
  const m = /PRIMARY\s+KEY\s*\(([^)]*)\)/i.exec(cons);
  if (!m) return [];
  return m[1].split(',').map((x) => x.trim().replace(/"/g, ''));
}

function buildTable(table) {
  const struct = STRUCT[table.name];
  if (!struct) throw new Error('缺少结构体映射: ' + table.name);

  // 索引按列挂载
  const indexTags = new Map(); // field -> [tag]
  const allIndexes = dedupeBy(schema.indexes.filter((i) => i.table === table.name), (i) => i.name);
  for (const idx of allIndexes) {
    const prefix = idx.unique ? 'uniqueIndex' : 'index';
    idx.columns.forEach((colRaw, pos) => {
      const parsed = parseIndexColumn(colRaw);
      let tag = `${prefix}:${idx.name}`;
      if (idx.columns.length > 1) tag += `,priority:${pos + 1}`;
      if (parsed.sort) tag += `,sort:${parsed.sort}`;
      if (idx.where) tag += `,where:${idx.where}`;
      if (!indexTags.has(parsed.name)) indexTags.set(parsed.name, []);
      indexTags.get(parsed.name).push(tag);
    });
  }

  // CHECK 约束按字段分布
  const checks = dedupeBy(table.constraints.map(parseCheck).filter(Boolean), (c) => c.name);
  const checkByField = new Map();
  let checkIdx = 0;
  for (const chk of checks) {
    const col = table.columns[checkIdx % table.columns.length];
    if (!checkByField.has(col.name)) checkByField.set(col.name, []);
    checkByField.get(col.name).push(`check:${chk.name},${chk.expr}`);
    checkIdx++;
  }

  // 关联
  const assocs = [];
  const usedAssoc = new Set();
  for (const col of table.columns) {
    if (!col.references) continue;
    let name = assocName(col.name);
    if (usedAssoc.has(name)) name = fieldName(col.name);
    usedAssoc.add(name);
    const fkField = fieldName(col.name);
    const parentStruct = STRUCT[col.references] || 'Unknown';
    const refField = parentPKField(col.references);
    let tag = `foreignKey:${fkField};references:${refField}`;
    if (col.onDelete) tag += `;constraint:OnDelete:${col.onDelete}`;
    assocs.push({ name, type: parentStruct, tag });
  }

  const pkOrder = compositePKOrder(table);
  const pkSet = new Set(pkOrder);
  let columns = table.columns;
  if (pkOrder.length > 1) {
    const pkCols = pkOrder.map((n) => table.columns.find((c) => c.name === n)).filter(Boolean);
    columns = [...pkCols, ...table.columns.filter((c) => !pkOrder.includes(c.name))];
  }

  const imports = new Set();
  const lines = [];
  for (const col of columns) {
    const ti = typeInfo(col.type, col.notNull || col.primaryKey);
    if (ti.imp === 'uuid') imports.add('"github.com/google/uuid"');
    if (ti.imp === 'time') imports.add('"time"');
    if (ti.imp === 'json') imports.add('"encoding/json"');
    if (ti.imp === 'pq') imports.add('"github.com/lib/pq"');

    const tags = [`column:${col.name}`, `type:${ti.tag}`];
    if (col.primaryKey || pkSet.has(col.name)) tags.push('primaryKey');
    if (col.unique && !col.primaryKey) tags.push('unique');
    if (col.notNull) tags.push('not null');
    if (col.default) tags.push(`default:${col.default}`);
    if (col.name === 'created_at') tags.push('autoCreateTime:false');
    if (col.name === 'updated_at') tags.push('autoUpdateTime:false');
    for (const t of indexTags.get(col.name) || []) tags.push(t);
    for (const t of checkByField.get(col.name) || []) tags.push(t);
    lines.push(`\t${fieldName(col.name)} ${ti.go} \`gorm:"${tags.join(';')}"\``);
  }
  for (const a of assocs) {
    lines.push(`\t${a.name} *${a.type} \`gorm:"${a.tag}" json:"-"\``);
  }

  return { struct, table: table.name, imports: [...imports].sort(), lines };
}

fs.rmSync(OUT, { recursive: true, force: true });
fs.mkdirSync(OUT, { recursive: true });

const models = [];
for (const table of schema.tables) {
  const built = buildTable(table);
  models.push(built.struct);
  const file = path.join(OUT, `${built.table}.go`);
  const body = [
    '// Code generated by contracts/tools/gen-gorm.mjs; DO NOT EDIT.',
    'package gormmodel',
    '',
    'import (',
    ...built.imports.map((i) => '\t' + i),
    ')',
    '',
    `// ${built.struct} 映射 ${built.table} 表。`,
    `type ${built.struct} struct {`,
    ...built.lines,
    '}',
    '',
    `func (${built.struct}) TableName() string { return "${built.table}" }`,
    '',
  ].join('\n');
  fs.writeFileSync(file, body);
}

const all = [
  '// Code generated by contracts/tools/gen-gorm.mjs; DO NOT EDIT.',
  'package gormmodel',
  '',
  '// All 返回 AutoMigrate 所需的全部模型。',
  'func All() []any {',
  '\treturn []any{',
  ...models.map((m) => `\t\t&${m}{},`),
  '\t}',
  '}',
  '',
].join('\n');
fs.writeFileSync(path.join(OUT, 'models.go'), all);
console.log(`generated ${models.length} models -> ${path.relative(ROOT, OUT)}`);
