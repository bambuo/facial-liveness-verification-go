/**
 * SQL 静态解析工具：注释/字符串/美元引用块掩码 + 语句切分 + 常用 DDL 解析。
 * 掩码文本与原文等长、保留换行，结构解析在掩码上做，取值从原文切片读取。
 */

export function maskSql(src) {
  const out = src.split('');
  const n = src.length;
  let i = 0;
  const maskRange = (start, end) => {
    for (let k = start; k < end; k++) {
      if (out[k] !== '\n') out[k] = ' ';
    }
  };
  while (i < n) {
    const c = src[i];
    const next = src[i + 1];
    if (c === '-' && next === '-') {
      let k = i;
      while (k < n && src[k] !== '\n') k++;
      maskRange(i, k);
      i = k;
      continue;
    }
    if (c === '/' && next === '*') {
      let k = i + 2;
      while (k < n && !(src[k] === '*' && src[k + 1] === '/')) k++;
      k = Math.min(n, k + 2);
      maskRange(i, k);
      i = k;
      continue;
    }
    if (c === "'") {
      let k = i + 1;
      while (k < n) {
        if (src[k] === "'" && src[k + 1] === "'") {
          k += 2;
          continue;
        }
        if (src[k] === "'") {
          k++;
          break;
        }
        k++;
      }
      maskRange(i + 1, k - 1);
      i = k;
      continue;
    }
    if (c === '$') {
      const m = /^\$[A-Za-z_]*\$/.exec(src.slice(i));
      if (m) {
        const tag = m[0];
        const end = src.indexOf(tag, i + tag.length);
        if (end !== -1) {
          maskRange(i + tag.length, end);
          i = end + tag.length;
          continue;
        }
      }
    }
    i++;
  }
  return out.join('');
}

export function stripSqlComments(src) {
  let out = '';
  const n = src.length;
  let i = 0;
  while (i < n) {
    const c = src[i];
    const next = src[i + 1];
    if (c === '-' && next === '-') {
      while (i < n && src[i] !== '\n') i++;
      continue;
    }
    if (c === '/' && next === '*') {
      i += 2;
      while (i < n && !(src[i] === '*' && src[i + 1] === '/')) i++;
      i = Math.min(n, i + 2);
      continue;
    }
    if (c === "'") {
      out += c;
      i++;
      while (i < n) {
        out += src[i];
        if (src[i] === "'" && src[i + 1] === "'") {
          out += src[i + 1];
          i += 2;
          continue;
        }
        if (src[i] === "'") {
          i++;
          break;
        }
        i++;
      }
      continue;
    }
    if (c === '$') {
      const m = /^\$[A-Za-z_]*\$/.exec(src.slice(i));
      if (m) {
        const tag = m[0];
        const end = src.indexOf(tag, i + tag.length);
        if (end !== -1) {
          out += src.slice(i, end + tag.length);
          i = end + tag.length;
          continue;
        }
      }
    }
    out += c;
    i++;
  }
  return out;
}

export function splitStatements(masked, raw) {
  const statements = [];
  let start = 0;
  for (let i = 0; i < masked.length; i++) {
    if (masked[i] !== ';') continue;
    const m = masked.slice(start, i);
    if (m.trim()) statements.push({ masked: m, raw: raw.slice(start, i), offset: start });
    start = i + 1;
  }
  const tailMasked = masked.slice(start);
  if (tailMasked.trim()) statements.push({ masked: tailMasked, raw: raw.slice(start), offset: start });
  return statements;
}

export function findMatchingParen(masked, open) {
  let depth = 0;
  for (let i = open; i < masked.length; i++) {
    if (masked[i] === '(') depth++;
    else if (masked[i] === ')') {
      depth--;
      if (depth === 0) return i;
    }
  }
  return -1;
}

export function splitTopLevel(inner) {
  const parts = [];
  let depth = 0;
  let start = 0;
  for (let i = 0; i < inner.length; i++) {
    const c = inner[i];
    if (c === '(' || c === '[' || c === '{') depth++;
    else if (c === ')' || c === ']' || c === '}') depth--;
    else if (c === ',' && depth === 0) {
      parts.push({ masked: inner.slice(start, i), start });
      start = i + 1;
    }
  }
  const tail = inner.slice(start);
  if (tail.trim()) parts.push({ masked: tail, start });
  return parts;
}

export function unquoteIdent(value) {
  const t = value.trim();
  if (t.startsWith('"') && t.endsWith('"')) return t.slice(1, -1).replace(/""/g, '"');
  return t;
}

const TABLE_CONSTRAINT = /^(CONSTRAINT|PRIMARY\s+KEY|UNIQUE|FOREIGN\s+KEY|CHECK|EXCLUDE)\b/i;

export function parseColumn(defRaw) {
  const text = defRaw.trim();
  if (!text || TABLE_CONSTRAINT.test(text)) return null;
  const nameMatch = /^("(?:[^"]|"")+"|[A-Za-z_][\w$]*)/.exec(text);
  if (!nameMatch) return null;
  const name = unquoteIdent(nameMatch[1]);
  let rest = text.slice(nameMatch[1].length).trim();
  const constraintRe = /\b(NOT\s+NULL|NULL|DEFAULT|PRIMARY\s+KEY|UNIQUE|REFERENCES|CHECK|GENERATED|COLLATE)\b/i;
  const cm = constraintRe.exec(rest);
  const type = (cm ? rest.slice(0, cm.index) : rest).replace(/\s+/g, ' ').trim();
  const tail = cm ? rest.slice(cm.index) : '';
  const flags = {
    notNull: /\bNOT\s+NULL\b/i.test(tail),
    primaryKey: /\bPRIMARY\s+KEY\b/i.test(tail),
    unique: /\bUNIQUE\b/i.test(tail),
    references: null,
    default: null,
    check: null,
  };
  const ref = /\bREFERENCES\s+([\w."]+)/i.exec(tail);
  if (ref) flags.references = ref[1];
  const def = /\bDEFAULT\s+(.+?)(?=\bNOT\s+NULL\b|\bNULL\b|\bPRIMARY\s+KEY\b|\bUNIQUE\b|\bREFERENCES\b|\bCHECK\b|$)/i.exec(tail);
  if (def) flags.default = def[1].trim().replace(/\s+/g, ' ');
  const chk = /\bCHECK\s*(\([^;]*?\))/i.exec(tail);
  if (chk) flags.check = chk[1].trim();
  return { name, type, ...flags };
}

export function parseCreateTable(stmt) {
  const m = /^\s*CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?((?:"[^"]+"|[A-Za-z_][\w$]*)(?:\.(?:"[^"]+"|[A-Za-z_][\w$]*))?)/i.exec(stmt.masked);
  if (!m) return null;
  const fullName = m[1];
  const parts = fullName.split('.');
  const schema = parts.length > 1 ? unquoteIdent(parts[0]) : 'public';
  const name = unquoteIdent(parts[parts.length - 1]);
  const open = stmt.masked.indexOf('(', m[0].length - 1);
  if (open < 0) return null;
  const close = findMatchingParen(stmt.masked, open);
  if (close < 0) return null;
  const innerMasked = stmt.masked.slice(open + 1, close);
  const innerRaw = stmt.raw.slice(open + 1, close);
  const defs = splitTopLevel(innerMasked).map((part) =>
    stripSqlComments(innerRaw.slice(part.start, part.start + part.masked.length)),
  );
  const columns = [];
  const constraints = [];
  for (const def of defs) {
    const col = parseColumn(def);
    if (col) columns.push(col);
    else if (def.trim()) constraints.push(def.trim().replace(/\s+/g, ' '));
  }
  return { schema, name, columns, constraints };
}

export function parseAlterTable(stmt) {
  const m = /^\s*ALTER\s+TABLE\s+(?:ONLY\s+)?((?:"[^"]+"|[A-Za-z_][\w$]*)(?:\.(?:"[^"]+"|[A-Za-z_][\w$]*))?)\s+([\s\S]*)$/i.exec(stmt.masked);
  if (!m) return null;
  const parts = m[1].split('.');
  const name = unquoteIdent(parts[parts.length - 1]);
  const restMasked = m[2];
  const restRaw = stmt.raw.slice(stmt.raw.length - restMasked.length);
  const op = { table: name, actions: [] };
  for (const part of splitTopLevel(restMasked)) {
    const actionMasked = part.masked.trim();
    const actionRaw = stripSqlComments(restRaw.slice(part.start, part.start + part.masked.length)).trim();
    if (!actionMasked) continue;
    const addCol = /^ADD\s+COLUMN\s+(?:IF\s+NOT\s+EXISTS\s+)?([\s\S]+)$/i.exec(actionMasked);
    if (addCol) {
      const defRaw = actionRaw.slice(actionRaw.length - addCol[1].length);
      const col = parseColumn(defRaw);
      if (col) op.actions.push({ type: 'add-column', column: col });
      continue;
    }
    const dropCol = /^DROP\s+COLUMN\s+(?:IF\s+EXISTS\s+)?((?:"[^"]+"|[A-Za-z_][\w$]*))/i.exec(actionMasked);
    if (dropCol) {
      op.actions.push({ type: 'drop-column', column: unquoteIdent(dropCol[1]) });
      continue;
    }
    const renameCol = /^RENAME\s+COLUMN\s+((?:"[^"]+"|[A-Za-z_][\w$]*))\s+TO\s+((?:"[^"]+"|[A-Za-z_][\w$]*))/i.exec(actionMasked);
    if (renameCol) {
      op.actions.push({ type: 'rename-column', from: unquoteIdent(renameCol[1]), to: unquoteIdent(renameCol[2]) });
      continue;
    }
    const addConstraint = /^ADD\s+CONSTRAINT\s+([\s\S]+)$/i.exec(actionMasked);
    if (addConstraint) {
      op.actions.push({ type: 'add-constraint', definition: addConstraint[1].trim().replace(/\s+/g, ' ') });
      continue;
    }
    const setDefault = /^ALTER\s+COLUMN\s+((?:"[^"]+"|[A-Za-z_][\w$]*))\s+SET\s+DEFAULT\s+([\s\S]+)$/i.exec(actionMasked);
    if (setDefault) {
      op.actions.push({ type: 'set-default', column: unquoteIdent(setDefault[1]), value: setDefault[2].trim() });
      continue;
    }
    if (/^(ENABLE|DISABLE)\s+ROW\s+LEVEL\s+SECURITY/i.test(actionMasked)) {
      op.actions.push({ type: 'rls', enabled: /^ENABLE/i.test(actionMasked) });
      continue;
    }
    op.actions.push({ type: 'other', sql: actionMasked.replace(/\s+/g, ' ').slice(0, 200) });
  }
  return op;
}

export function parseCreateIndex(stmt) {
  const m = /^\s*CREATE\s+(UNIQUE\s+)?INDEX\s+(?:IF\s+NOT\s+EXISTS\s+)?((?:"[^"]+"|[A-Za-z_][\w$]*))\s+ON\s+(?:ONLY\s+)?((?:"[^"]+"|[A-Za-z_][\w$]*)(?:\.(?:"[^"]+"|[A-Za-z_][\w$]*))?)\s*(?:USING\s+[A-Za-z_]+\s*)?\(/i.exec(stmt.masked);
  if (!m) return null;
  const open = stmt.masked.indexOf('(', m.index + m[0].length - 1);
  const close = open >= 0 ? findMatchingParen(stmt.masked, open) : -1;
  const columnsRaw = close > open ? stmt.raw.slice(open + 1, close) : '';
  const whereMatch = /\bWHERE\b([\s\S]*)$/i.exec(stmt.masked.slice(close + 1));
  const parts = m[3].split('.');
  return {
    name: unquoteIdent(m[2]),
    table: unquoteIdent(parts[parts.length - 1]),
    unique: Boolean(m[1]),
    columns: columnsRaw.split(',').map((x) => x.trim()).filter(Boolean),
    where: whereMatch ? whereMatch[1].trim().replace(/\s+/g, ' ').slice(0, 160) : null,
  };
}

export function parseCreateFunction(stmt) {
  const m = /^\s*CREATE\s+(?:OR\s+REPLACE\s+)?FUNCTION\s+((?:"[^"]+"|[A-Za-z_][\w$]*)(?:\.(?:"[^"]+"|[A-Za-z_][\w$]*))*)\s*\(/i.exec(stmt.masked);
  if (!m) return null;
  const open = stmt.masked.indexOf('(', m.index + m[0].length - 1);
  const close = open >= 0 ? findMatchingParen(stmt.masked, open) : -1;
  const argsRaw = close > open ? stmt.raw.slice(open + 1, close) : '';
  const after = close > 0 ? stmt.masked.slice(close + 1) : '';
  const returns = /\bRETURNS\s+([\s\S]*?)\b(?:LANGUAGE|AS|IMMUTABLE|STABLE|VOLATILE|SECURITY)\b/i.exec(after);
  const lang = /\bLANGUAGE\s+([A-Za-z_]+)/i.exec(after);
  return {
    name: unquoteIdent(m[1].split('.').pop()),
    args: argsRaw.replace(/\s+/g, ' ').trim(),
    returns: returns ? returns[1].replace(/\s+/g, ' ').trim() : null,
    language: lang ? lang[1].toLowerCase() : null,
  };
}

export function parseCreateTrigger(stmt) {
  const m = /^\s*CREATE\s+(?:OR\s+REPLACE\s+)?TRIGGER\s+((?:"[^"]+"|[A-Za-z_][\w$]*))\s+([\s\S]*?)\bON\s+((?:"[^"]+"|[A-Za-z_][\w$]*)(?:\.(?:"[^"]+"|[A-Za-z_][\w$]*))?)/i.exec(stmt.masked);
  if (!m) return null;
  const timing = m[2].replace(/\s+/g, ' ').trim();
  const tableParts = m[3].split('.');
  return {
    name: unquoteIdent(m[1]),
    timing,
    table: unquoteIdent(tableParts[tableParts.length - 1]),
  };
}

export function parseCreatePolicy(stmt) {
  const m = /^\s*CREATE\s+POLICY\s+((?:"[^"]+"|[A-Za-z_][\w$]*))\s+ON\s+((?:"[^"]+"|[A-Za-z_][\w$]*)(?:\.(?:"[^"]+"|[A-Za-z_][\w$]*))?)([\s\S]*)$/i.exec(stmt.masked);
  if (!m) return null;
  const tail = m[3];
  const tableParts = m[2].split('.');
  return {
    name: unquoteIdent(m[1]),
    table: tableParts.map((p) => unquoteIdent(p)).join('.'),
    command: (/\bFOR\s+(SELECT|INSERT|UPDATE|DELETE|ALL)\b/i.exec(tail) || [])[1] || 'ALL',
    using: /\bUSING\s*\(/i.test(tail),
    withCheck: /\bWITH\s+CHECK\s*\(/i.test(tail),
  };
}

export function parseCreateType(stmt) {
  const m = /^\s*CREATE\s+TYPE\s+((?:"[^"]+"|[A-Za-z_][\w$]*)(?:\.(?:"[^"]+"|[A-Za-z_][\w$]*))?)\s+AS\s+ENUM\s*\(/i.exec(stmt.masked);
  if (!m) return null;
  const open = stmt.masked.indexOf('(', m.index + m[0].length - 1);
  const close = open >= 0 ? findMatchingParen(stmt.masked, open) : -1;
  const valuesRaw = close > open ? stmt.raw.slice(open + 1, close) : '';
  return { name: unquoteIdent(m[1].split('.').pop()), values: valuesRaw.split(',').map((x) => x.trim().replace(/^'|'$/g, '')) };
}

export function parseCreateExtension(stmt) {
  const m = /^\s*CREATE\s+EXTENSION\s+(?:IF\s+NOT\s+EXISTS\s+)?((?:"[^"]+"|[A-Za-z_][\w$]*))/i.exec(stmt.masked);
  return m ? { name: unquoteIdent(m[1]) } : null;
}

export function parseDropTable(stmt) {
  const m = /^\s*DROP\s+TABLE\s+(?:IF\s+EXISTS\s+)?((?:"[^"]+"|[A-Za-z_][\w$]*)(?:\s*,\s*(?:"[^"]+"|[A-Za-z_][\w$]*))*)/i.exec(stmt.masked);
  if (!m) return null;
  return m[1].split(',').map((x) => unquoteIdent(x.trim().split('.').pop()));
}

export function parseDoBlock(stmt) {
  return /^\s*DO\s+\$/i.test(stmt.masked.slice(0, 40));
}

export function parseStatement(stmt) {
  const t = stmt.masked.trimStart().toUpperCase();
  if (t.startsWith('CREATE TABLE')) return { kind: 'table', value: parseCreateTable(stmt) };
  if (t.startsWith('ALTER TABLE')) return { kind: 'alter-table', value: parseAlterTable(stmt) };
  if (t.startsWith('CREATE INDEX') || t.startsWith('CREATE UNIQUE INDEX')) return { kind: 'index', value: parseCreateIndex(stmt) };
  if (t.startsWith('CREATE FUNCTION') || t.startsWith('CREATE OR REPLACE FUNCTION')) return { kind: 'function', value: parseCreateFunction(stmt) };
  if (t.startsWith('CREATE TRIGGER')) return { kind: 'trigger', value: parseCreateTrigger(stmt) };
  if (t.startsWith('CREATE POLICY')) return { kind: 'policy', value: parseCreatePolicy(stmt) };
  if (t.startsWith('CREATE TYPE')) return { kind: 'type', value: parseCreateType(stmt) };
  if (t.startsWith('CREATE EXTENSION')) return { kind: 'extension', value: parseCreateExtension(stmt) };
  if (t.startsWith('DROP TABLE')) return { kind: 'drop-table', value: parseDropTable(stmt) };
  if (t.startsWith('DO ')) return { kind: 'do-block', value: true };
  return { kind: 'other', value: null };
}
