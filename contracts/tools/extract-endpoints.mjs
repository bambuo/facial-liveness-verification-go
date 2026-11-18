#!/usr/bin/env node
/**
 * 从 idswyft 后端源码用 TypeScript AST 抽取 HTTP 接口清单，作为契约冻结输入。
 *
 * 用法:
 *   node contracts/tools/extract-endpoints.mjs [--repo <idswyft-repo>] [--out <contracts-dir>]
 *
 * 输出:
 *   contracts/endpoints.json  机器可读清单
 *   contracts/endpoints.md    人工评审表
 *
 * 统计口径:
 *   - 只统计 router.<get|post|put|patch|delete> 注册，不含 router.use 挂载本身
 *   - 同一路由模块挂载到多个前缀时展开为多条可达路径（credentials 挂载两次）
 *   - cloud-only 动态挂载单独记录，不计入 154
 *   - app 级路由（如 /api/public/assets/*）单独分组
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
const SERVER = path.join(REPO, 'backend/src/server.ts');
const ROUTES_DIR = path.join(REPO, 'backend/src/routes');
const HTTP_METHODS = new Set(['get', 'post', 'put', 'patch', 'delete', 'head', 'options', 'all']);

const AUTH_MIDDLEWARE = new Set([
  'authenticateServiceToken',
  'authenticateAPIKey',
  'authenticateHandoffToken',
  'authenticateSessionToken',
  'authenticateAPIKeyOrHandoff',
  'authenticateJWT',
  'authenticateUser',
  'requireAdminRole',
  'requireOrgAdminOrPlatformAdmin',
  'authenticateDeveloperJWT',
  'authenticateDeveloperJWTOrServiceKey',
  'authenticateReviewerJWT',
  'authenticateServiceOperatorJWT',
  'authenticateDashboard',
  'authenticateAdminOrReviewer',
  'authenticateReviewPrincipal',
  'authenticateComplianceRequest',
]);

const AUTH_LABEL = {
  authenticateServiceToken: 'service-token',
  authenticateAPIKey: 'api-key',
  authenticateHandoffToken: 'handoff-token',
  authenticateSessionToken: 'session-token',
  authenticateAPIKeyOrHandoff: 'api-key|handoff-token',
  authenticateJWT: 'jwt',
  authenticateUser: 'user-jwt',
  requireAdminRole: 'admin-jwt',
  requireOrgAdminOrPlatformAdmin: 'org-admin|platform-admin',
  authenticateDeveloperJWT: 'developer-jwt',
  authenticateDeveloperJWTOrServiceKey: 'developer-jwt|service-key',
  authenticateReviewerJWT: 'reviewer-jwt',
  authenticateServiceOperatorJWT: 'service-operator-jwt',
  authenticateDashboard: 'dashboard-jwt',
  authenticateAdminOrReviewer: 'admin-jwt|reviewer-jwt',
  authenticateReviewPrincipal: 'review-principal',
  authenticateComplianceRequest: 'compliance-request',
};

const RATE_LIMIT_MIDDLEWARE = new Set(['verificationRateLimit', 'developerRateLimit', 'authRateLimit', 'apiKeyRateLimit']);
const GLOBAL_MIDDLEWARE = new Set(['conditionalCsrf', 'apiActivityLogger', 'localeMiddleware', 'limiter']);

const sourceCache = new Map();
function loadSource(file) {
  if (sourceCache.has(file)) return sourceCache.get(file);
  const text = fs.readFileSync(file, 'utf8');
  const sf = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS);
  const entry = { file, text, sf };
  sourceCache.set(file, entry);
  return entry;
}

function propertyAccessName(node) {
  if (ts.isPropertyAccessExpression(node) && ts.isIdentifier(node.expression)) {
    return { obj: node.expression.text, prop: node.name.text };
  }
  return null;
}

function stringOf(node) {
  if (!node) return null;
  if (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node)) return node.text;
  return null;
}

function identifierOf(node) {
  if (!node) return null;
  if (ts.isAsExpression(node) || ts.isParenthesizedExpression(node) || ts.isSatisfiesExpression(node)) {
    return identifierOf(node.expression);
  }
  if (ts.isIdentifier(node)) return node.text;
  if (ts.isPropertyAccessExpression(node) && ts.isIdentifier(node.expression)) return node.expression.text;
  if (ts.isCallExpression(node)) {
    const pa = propertyAccessName(node.expression);
    if (pa) return pa.obj;
    if (ts.isIdentifier(node.expression)) return node.expression.text;
  }
  return null;
}

function isInlineFunction(node) {
  return ts.isArrowFunction(node) || ts.isFunctionExpression(node);
}

function flattenMiddleware(node, acc = []) {
  if (!node) return acc;
  if (ts.isArrayLiteralExpression(node)) {
    for (const el of node.elements) flattenMiddleware(el, acc);
    return acc;
  }
  if (isInlineFunction(node)) {
    acc.push('inline');
    return acc;
  }
  const id = identifierOf(node);
  if (id) acc.push(id);
  return acc;
}

function collectValidators(call) {
  const found = new Set();
  for (const arg of call.arguments) {
    const visit = (node) => {
      if (ts.isCallExpression(node)) {
        const name = identifierOf(node);
        if (name && ['validate', 'body', 'param', 'query', 'checkExact', 'checkSchema'].includes(name)) found.add(name);
      }
      node.forEachChild(visit);
    };
    visit(arg);
  }
  return [...found];
}

function lineOf(entry, node) {
  return entry.sf.getLineAndCharacterOfPosition(node.getStart(entry.sf)).line + 1;
}

function parseImports(entry) {
  const map = new Map();
  for (const stmt of entry.sf.statements) {
    if (!ts.isImportDeclaration(stmt)) continue;
    const spec = stringOf(stmt.moduleSpecifier);
    if (!spec) continue;
    const clause = stmt.importClause;
    if (clause?.name) map.set(clause.name.text, spec);
    if (clause?.namedBindings && ts.isNamespaceImport(clause.namedBindings)) {
      map.set(clause.namedBindings.name.text, spec);
    }
  }
  return map;
}

function resolveModule(fromFile, spec) {
  if (!spec.startsWith('.')) return null;
  const base = path.resolve(path.dirname(fromFile), spec);
  const candidates = [base, base.replace(/\.js$/, '.ts'), path.join(base, 'index.ts')];
  for (const candidate of candidates) {
    if (fs.existsSync(candidate) && fs.statSync(candidate).isFile()) return candidate;
  }
  return null;
}

function moduleName(file) {
  let rel = path.relative(ROUTES_DIR, file).replace(/\\/g, '/').replace(/\.ts$/, '');
  if (rel.endsWith('/index')) rel = rel.slice(0, -6);
  return rel;
}

function joinPath(prefix, suffix) {
  const p = `${prefix || ''}/${suffix || ''}`.replace(/\/+/g, '/');
  if (p.length > 1 && p.endsWith('/')) return p.slice(0, -1);
  return p || '/';
}

function extractRoutes(entry) {
  const routes = [];
  const walk = (node) => {
    if (ts.isCallExpression(node)) {
      const pa = propertyAccessName(node.expression);
      if (pa?.obj === 'router' && HTTP_METHODS.has(pa.prop)) {
        const routePath = stringOf(node.arguments[0]);
        if (routePath !== null) {
          const rest = node.arguments.slice(1);
          const middleware = [];
          for (const arg of rest.slice(0, Math.max(0, rest.length - 1))) flattenMiddleware(arg, middleware);
          const last = rest[rest.length - 1];
          const handler = isInlineFunction(last) ? 'inline' : identifierOf(last) || '<expr>';
          routes.push({
            method: pa.prop.toUpperCase(),
            routePath,
            middleware: [...new Set(middleware)],
            handler,
            validators: collectValidators(node),
            source: { file: path.relative(REPO, entry.file), line: lineOf(entry, node) },
          });
        }
      }
    }
    node.forEachChild(walk);
  };
  walk(entry.sf);
  return routes;
}

function extractUseMounts(entry) {
  const mounts = [];
  const walk = (node) => {
    if (ts.isCallExpression(node)) {
      const pa = propertyAccessName(node.expression);
      if (pa?.obj === 'router' && pa.prop === 'use') {
        const routeArgs = [...node.arguments];
        let prefix = '';
        const first = stringOf(routeArgs[0]);
        if (first !== null) {
          prefix = first;
          routeArgs.shift();
        }
        const target = routeArgs[routeArgs.length - 1];
        if (routeArgs.length > 0 && ts.isIdentifier(target)) {
          mounts.push({ prefix, target: target.text, line: lineOf(entry, node) });
        }
      }
    }
    node.forEachChild(walk);
  };
  walk(entry.sf);
  return mounts;
}

function collectModule(file, seen) {
  if (seen.has(file)) return { routes: [], children: [] };
  seen.add(file);
  const entry = loadSource(file);
  const routes = extractRoutes(entry);
  const imports = parseImports(entry);
  const children = [];
  for (const use of extractUseMounts(entry)) {
    const spec = imports.get(use.target);
    const childFile = spec ? resolveModule(file, spec) : null;
    if (childFile) children.push({ prefix: use.prefix, file: childFile });
  }
  return { routes, children };
}

function flattenModule(file, mountPrefix, acc, seen) {
  const { routes, children } = collectModule(file, seen);
  for (const route of routes) {
    acc.push({
      ...route,
      mount: mountPrefix,
      fullPath: joinPath(mountPrefix, route.routePath),
      module: moduleName(file),
    });
  }
  for (const child of children) flattenModule(child.file, joinPath(mountPrefix, child.prefix), acc, seen);
}

function parseServer() {
  const entry = loadSource(SERVER);
  const imports = parseImports(entry);
  const mounts = [];
  const cloud = [];
  const appRoutes = [];

  const walk = (node) => {
    if (ts.isCallExpression(node)) {
      const pa = propertyAccessName(node.expression);
      if (pa?.obj === 'app' && pa.prop === 'use') {
        const routeArgs = [...node.arguments];
        const prefix = stringOf(routeArgs[0]);
        if (prefix !== null && prefix !== '*') {
          routeArgs.shift();
          const last = routeArgs[routeArgs.length - 1];
          const middleware = [];
          for (const arg of routeArgs.slice(0, Math.max(0, routeArgs.length - 1))) flattenMiddleware(arg, middleware);
          if (ts.isIdentifier(last)) {
            const spec = imports.get(last.text);
            const file = spec ? resolveModule(SERVER, spec) : null;
            if (file) {
              mounts.push({
                prefix,
                router: last.text,
                sourceFile: path.relative(REPO, file),
                line: lineOf(entry, node),
                middleware: [...new Set(middleware)],
              });
            } else if (!AUTH_MIDDLEWARE.has(last.text) && !GLOBAL_MIDDLEWARE.has(last.text)) {
              cloud.push({ prefix, router: last.text, line: lineOf(entry, node) });
            }
          } else if (ts.isPropertyAccessExpression(last)) {
            cloud.push({ prefix, router: last.getText(entry.sf), line: lineOf(entry, node) });
          }
        }
      }
      if (pa?.obj === 'app' && HTTP_METHODS.has(pa.prop)) {
        const routePath = stringOf(node.arguments[0]);
        if (routePath !== null && routePath !== '*') {
          const last = node.arguments[node.arguments.length - 1];
          const preceding = entry.text.slice(Math.max(0, node.getStart(entry.sf) - 240), node.getStart(entry.sf));
          appRoutes.push({
            method: pa.prop.toUpperCase(),
            fullPath: routePath,
            handler: isInlineFunction(last) ? 'inline' : identifierOf(last) || '<expr>',
            module: 'app-level',
            source: { file: path.relative(REPO, SERVER), line: lineOf(entry, node) },
            conditional: preceding.includes('config.storage.provider') ? 'storage.provider=local|s3' : undefined,
          });
        }
      }
    }
    node.forEachChild(walk);
  };
  walk(entry.sf);

  return { mounts, cloud, appRoutes };
}

function deriveAuth(endpoint) {
  const found = endpoint.middleware.filter((x) => AUTH_MIDDLEWARE.has(x));
  if (found.length > 0) return [...new Set(found.map((x) => AUTH_LABEL[x] || x))].join(' + ');
  return 'none-or-inline';
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
  lines.push('# idswyft 接口清单（M0 契约冻结输入）');
  lines.push('');
  lines.push(`- 来源仓库：\`${result.source.repo}\``);
  lines.push(`- 基线提交：\`${result.source.commit}\`（${result.source.commitDate}）`);
  lines.push(`- 生成时间：${result.generatedAt}`);
  lines.push(`- 解析器：TypeScript ${result.source.typescriptVersion} AST`);
  lines.push('- 统计口径：`router.<get|post|put|patch|delete>` 注册；同一模块多前缀挂载展开为多条路径；cloud-only 动态挂载不计入');
  lines.push('');
  lines.push('## 汇总');
  lines.push('');
  lines.push('| 指标 | 数值 |');
  lines.push('| --- | --- |');
  lines.push(`| 路由注册数（文档口径 154） | ${result.statistics.routeRegistrations} |`);
  lines.push(`| 展开后 method+path | ${result.statistics.expandedMethodPath} |`);
  lines.push(`| app 级路由 | ${result.statistics.appLevelRoutes} |`);
  lines.push(`| cloud-only 挂载 | ${result.statistics.cloudOnlyMounts} |`);
  lines.push(`| 待人工确认鉴权 | ${result.statistics.needsReview} |`);
  lines.push('');
  lines.push('### 按模块');
  lines.push('');
  lines.push('| 模块 | 接口数 |');
  lines.push('| --- | --- |');
  for (const [k, v] of Object.entries(result.statistics.byModule).sort()) lines.push(`| ${k} | ${v} |`);
  lines.push('');
  lines.push('### 按鉴权');
  lines.push('');
  lines.push('| 鉴权 | 接口数 |');
  lines.push('| --- | --- |');
  for (const [k, v] of Object.entries(result.statistics.byAuth).sort()) lines.push(`| ${k} | ${v} |`);
  lines.push('');
  lines.push('## 接口明细');
  lines.push('');
  lines.push('> 明细共 158 行：154 条注册 + `credentials` 双挂载（`/api/v2/verify` 与 `/api/v2`）展开的 4 行。');
  lines.push('');
  lines.push('| # | 模块 | 方法 | 路径 | 挂载 | 鉴权 | 中间件 | 源码位置 | 待复核 |');
  lines.push('| --- | --- | --- | --- | --- | --- | --- | --- | --- |');
  result.endpoints.forEach((ep, i) => {
    const mw = ep.middleware.filter((x) => x !== 'conditionalCsrf').join(', ') || '-';
    lines.push(
      `| ${i + 1} | ${ep.module} | ${ep.method} | \`${ep.path}\` | \`${ep.mount}\` | ${ep.auth} | ${mw} | \`${ep.source.file}:${ep.source.line}\` | ${ep.needsReview ? '是' : ''} |`,
    );
  });
  lines.push('');
  lines.push('## app 级路由');
  lines.push('');
  lines.push('| 方法 | 路径 | 处理函数 | 条件 | 源码位置 |');
  lines.push('| --- | --- | --- | --- | --- |');
  for (const r of result.appLevel) {
    lines.push(`| ${r.method} | \`${r.fullPath}\` | ${r.handler} | ${r.conditional || '-'} | \`${r.source.file}:${r.source.line}\` |`);
  }
  lines.push('');
  lines.push('## 说明与待确认');
  lines.push('');
  for (const c of result.caveats) lines.push(`- ${c}`);
  lines.push('');
  return lines.join('\n');
}

function main() {
  const { mounts, cloud, appRoutes } = parseServer();
  const endpoints = [];
  const mountMap = new Map();

  for (const mount of mounts) {
    const file = path.join(REPO, mount.sourceFile);
    const list = [];
    flattenModule(file, mount.prefix, list, new Set());
    for (const ep of list) {
      endpoints.push({
        method: ep.method,
        path: ep.fullPath,
        mount: ep.mount,
        module: ep.module,
        auth: deriveAuth(ep),
        middleware: ep.middleware,
        validators: ep.validators,
        rateLimit: ep.middleware.find((x) => RATE_LIMIT_MIDDLEWARE.has(x)) || null,
        idempotency: ep.middleware.includes('idempotencyMiddleware') || null,
        upload: ep.middleware.includes('upload') ? 'upload' : null,
        csrf: ep.middleware.includes('conditionalCsrf') || mount.middleware.includes('conditionalCsrf') || null,
        handler: ep.handler,
        source: ep.source,
        needsReview: deriveAuth(ep) === 'none-or-inline',
      });
      const key = `${mount.prefix} -> ${mount.sourceFile}`;
      mountMap.set(key, (mountMap.get(key) || 0) + 1);
    }
  }

  const dedup = new Map();
  for (const ep of endpoints) dedup.set(`${ep.method} ${ep.path}`, ep);
  const unique = [...dedup.values()];

  const registrationKeys = new Set(endpoints.map((ep) => `${ep.source.file}:${ep.source.line}:${ep.method}`));

  const byModule = {};
  for (const ep of unique) byModule[ep.module] = (byModule[ep.module] || 0) + 1;
  const byAuth = {};
  for (const ep of unique) byAuth[ep.auth] = (byAuth[ep.auth] || 0) + 1;

  const result = {
    generatedBy: 'contracts/tools/extract-endpoints.mjs',
    generatedAt: new Date().toISOString(),
    source: {
      repo: REPO,
      commit: execGit(REPO, ['rev-parse', 'HEAD']),
      commitDate: execGit(REPO, ['log', '-1', '--format=%cI']),
      server: 'backend/src/server.ts',
      routesDir: 'backend/src/routes',
      typescriptVersion: ts.version,
    },
    statistics: {
      routeRegistrations: registrationKeys.size,
      expandedMethodPath: endpoints.length,
      uniqueMethodPath: unique.length,
      appLevelRoutes: appRoutes.length,
      cloudOnlyMounts: cloud.length,
      byModule,
      byAuth,
      needsReview: unique.filter((x) => x.needsReview).length,
    },
    endpoints: unique,
    appLevel: appRoutes,
    cloudOnly: cloud.map((c) => ({
      mount: c.prefix,
      router: c.router,
      source: { file: 'backend/src/server.ts', line: c.line },
      note: 'IDSWYFT_EDITION=cloud 动态挂载；社区镜像中不存在，未计入 154',
    })),
    mounts: [...mountMap.entries()].map(([key, count]) => ({ key, count })),
    validation: {
      documentedEndpoints: 154,
      extractedRegistrations: registrationKeys.size,
      matches: registrationKeys.size === 154,
    },
    caveats: [
      '静态解析；鉴权仅按中间件名推断，handler 内联鉴权标记为 none-or-inline，需要逐条人工确认',
      '同一路由模块挂载到多个前缀时展开为多条路径（credentials: /api/v2/verify 与 /api/v2）',
      'cloud-only 动态挂载不计入 154',
      '请求/响应字段未在本清单展开，见 M0 后续契约 schema 抽取',
    ],
  };

  fs.mkdirSync(OUT_DIR, { recursive: true });
  fs.writeFileSync(path.join(OUT_DIR, 'endpoints.json'), JSON.stringify(result, null, 2) + '\n');
  fs.writeFileSync(path.join(OUT_DIR, 'endpoints.md'), renderMarkdown(result));
  console.log(JSON.stringify(result.statistics, null, 2));
  console.log('validation:', JSON.stringify(result.validation));
  if (!result.validation.matches) process.exitCode = 1;
}

main();
