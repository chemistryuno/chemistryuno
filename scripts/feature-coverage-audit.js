#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

const FEATURE_PRIORITY = [
  'anticheat',
  'feedback',
  'survey',
  'plugin',
  'reaction',
  'substance',
  'announcement',
  'chat',
  'leaderboard',
  'ranking',
  'profile',
  'replay',
  'oauth',
  'webauthn',
  'login',
  'register',
  'auth',
  'level',
  'point',
  'room',
  'game',
  'data',
  'friend',
  'user',
  'admin',
  'health',
  'version',
  'lobby'
];

const TOKEN_ALIASES = new Map([
  ['feedbacks', 'feedback'],
  ['surveys', 'survey'],
  ['substances', 'substance'],
  ['reactions', 'reaction'],
  ['announcements', 'announcement'],
  ['plugins', 'plugin'],
  ['rooms', 'room'],
  ['points', 'point'],
  ['profiles', 'profile'],
  ['leaderboard', 'ranking'],
  ['friends', 'friend'],
  ['oauthcallback', 'oauth'],
  ['webauthn', 'webauthn'],
  ['gamehistory', 'replay']
]);

const IGNORE_TOKENS = new Set([
  'api',
  'get',
  'post',
  'put',
  'delete',
  'patch',
  'head',
  'options',
  'list',
  'detail',
  'history',
  'config',
  'settings',
  'status',
  'active',
  'all',
  'my',
  'begin',
  'finish',
  'check',
  'export',
  'approve',
  'reject',
  'review',
  'submit',
  'create',
  'update',
  'save',
  'load',
  'search',
  'log',
  'logs',
  'response',
  'responses',
  'script',
  'page',
  'pages',
  'component',
  'components',
  'route',
  'routes',
  'handler',
  'handlers',
  'service',
  'services',
  'controller',
  'controllers',
  'ping',
  'pong',
  'ws'
]);

function toPosix(relPath) {
  return relPath.split(path.sep).join('/');
}

function readText(filePath) {
  return fs.existsSync(filePath) ? fs.readFileSync(filePath, 'utf8') : '';
}

function walkFiles(dirPath, predicate = () => true) {
  if (!fs.existsSync(dirPath)) return [];
  const result = [];
  for (const entry of fs.readdirSync(dirPath, { withFileTypes: true })) {
    const fullPath = path.join(dirPath, entry.name);
    if (entry.isDirectory()) {
      result.push(...walkFiles(fullPath, predicate));
    } else if (predicate(fullPath)) {
      result.push(fullPath);
    }
  }
  return result;
}

function lineNumberAt(content, index) {
  let line = 1;
  for (let i = 0; i < index; i += 1) {
    if (content.charCodeAt(i) === 10) {
      line += 1;
    }
  }
  return line;
}

function unquote(value) {
  const trimmed = String(value || '').trim();
  const quote = trimmed[0];
  if ((quote === '"' || quote === '\'' || quote === '`') && trimmed[trimmed.length - 1] === quote) {
    return trimmed.slice(1, -1);
  }
  return trimmed;
}

function singularize(token) {
  if (token.endsWith('ies') && token.length > 3) {
    return token.slice(0, -3) + 'y';
  }
  if (token.endsWith('ses') && token.length > 3) {
    return token.slice(0, -2);
  }
  if (token.endsWith('s') && !token.endsWith('ss') && token.length > 3) {
    return token.slice(0, -1);
  }
  return token;
}

function splitTokens(rawValue) {
  const raw = String(rawValue || '')
    .replace(/([a-z0-9])([A-Z])/g, '$1 $2')
    .replace(/[^a-zA-Z0-9]+/g, ' ')
    .toLowerCase();

  const tokens = raw
    .split(/\s+/)
    .map((token) => token.trim())
    .filter(Boolean)
    .map((token) => TOKEN_ALIASES.get(token) || singularize(token))
    .filter((token) => token && !IGNORE_TOKENS.has(token));

  return Array.from(new Set(tokens));
}

function familyKeyFromTokens(tokens) {
  for (const family of FEATURE_PRIORITY) {
    if (tokens.includes(family)) {
      return family;
    }
  }
  return tokens[0] || '';
}

function pathTokens(routePath) {
  return splitTokens(String(routePath || '').split('/').filter(Boolean).join(' '));
}

function makeLocation(filePath, line, note) {
  const location = { file: filePath, line };
  if (note) {
    location.note = note;
  }
  return location;
}

function normalizePath(value) {
  const collapsed = String(value || '').replace(/\\/g, '/').replace(/\/+/g, '/');
  if (!collapsed) {
    return '/';
  }
  const trimmed = collapsed.length > 1 ? collapsed.replace(/\/+$|\/+$/g, '') : collapsed;
  return trimmed.startsWith('/') ? trimmed : `/${trimmed}`;
}

function joinPaths(prefix, suffix) {
  const left = String(prefix || '').trim();
  const right = String(suffix || '').trim();
  if (!left) return right || '/';
  if (!right) return left;
  return `${left.replace(/\/+$/, '')}/${right.replace(/^\/+/, '')}`;
}

function collectFrontendInventory(rootDir) {
  const frontendDir = path.join(rootDir, 'frontend', 'src');
  const routerFile = path.join(frontendDir, 'router.ts');
  const routerContent = readText(routerFile);
  const items = [];
  const componentRouteItems = new Map();

  const importRegex = /const\s+([A-Za-z0-9_]+)\s*=\s*\(\)\s*=>\s*import\((['"`])(.+?\.vue)\2\)/g;
  const componentImports = new Map();
  let importMatch;
  while ((importMatch = importRegex.exec(routerContent))) {
    componentImports.set(importMatch[1], importMatch[3]);
  }

  const routeBlockRegex = /\{\s*path:[\s\S]*?\n\s*\},?/g;
  let routeMatch;
  while ((routeMatch = routeBlockRegex.exec(routerContent))) {
    const block = routeMatch[0];
    const blockStart = routeMatch.index;
    const pathMatch = block.match(/path:\s*(['"`])([^'"`]+)\1/);
    if (!pathMatch) continue;

    const routePath = pathMatch[2];
    const routeNameMatch = block.match(/name:\s*(['"`])([^'"`]+)\1/);
    const componentMatch = block.match(/component:\s*([A-Za-z0-9_]+)/);
    const routeLabel = routeNameMatch ? routeNameMatch[2] : routePath;
    const tokens = pathTokens(routePath);
    const family = familyKeyFromTokens(tokens.length > 0 ? tokens : splitTokens(routeLabel));

    const routeItem = {
      side: 'frontend',
      kind: 'route',
      family,
      tokens,
      label: routeLabel,
      path: routePath,
      locations: [makeLocation(toPosix(path.relative(rootDir, routerFile)), lineNumberAt(routerContent, blockStart + 1), routeLabel)],
    };

    items.push(routeItem);

    if (componentMatch && componentImports.has(componentMatch[1])) {
      componentRouteItems.set(path.resolve(frontendDir, componentImports.get(componentMatch[1])), routeItem);
    }
  }

  const pageFiles = walkFiles(path.join(frontendDir, 'pages'), (filePath) => filePath.endsWith('.vue'));
  for (const filePath of pageFiles) {
    const name = path.basename(filePath, '.vue');
    const routeItem = componentRouteItems.get(path.resolve(filePath));
    const tokens = routeItem ? routeItem.tokens : splitTokens(name);
    const family = routeItem ? routeItem.family : familyKeyFromTokens(tokens.length > 0 ? tokens : splitTokens(name));

    items.push({
      side: 'frontend',
      kind: 'page',
      family,
      tokens,
      label: name,
      locations: [makeLocation(toPosix(path.relative(rootDir, filePath)), 1)],
    });
  }

  const componentFiles = walkFiles(path.join(frontendDir, 'components'), (filePath) => filePath.endsWith('.vue'));
  for (const filePath of componentFiles) {
    const name = path.basename(filePath, '.vue');
    const tokens = splitTokens(name);
    const family = familyKeyFromTokens(tokens.length > 0 ? tokens : splitTokens(name));

    items.push({
      side: 'frontend',
      kind: 'component',
      family,
      tokens,
      label: name,
      locations: [makeLocation(toPosix(path.relative(rootDir, filePath)), 1)],
    });
  }

  return items;
}

function collectGoRouteItems(rootDir, filePath, kind = 'api') {
  const content = readText(filePath);
  const lines = content.split(/\r?\n/);
  const prefixes = new Map([
    ['r', ''],
    ['router', ''],
  ]);
  const items = [];

  const groupAssignRegex = /^\s*([A-Za-z_][A-Za-z0-9_]*)\s*:=\s*([A-Za-z_][A-Za-z0-9_]*)\.Group\(([^)]*)\)/;
  const routeCallRegex = /^\s*([A-Za-z_][A-Za-z0-9_]*)\.(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\((.+)$/;

  for (let index = 0; index < lines.length; index += 1) {
    const line = lines[index];
    const groupMatch = line.match(groupAssignRegex);
    if (groupMatch) {
      const alias = groupMatch[1];
      const parentAlias = groupMatch[2];
      const groupPath = normalizePath(joinPaths(prefixes.get(parentAlias) || '', unquote(groupMatch[3])));
      prefixes.set(alias, groupPath);
      continue;
    }

    const routeMatch = line.match(routeCallRegex);
    if (!routeMatch) continue;

    const alias = routeMatch[1];
    const method = routeMatch[2];
    const pathMatch = routeMatch[3].match(/(['"`])([^'"`]+)\1/);
    if (!pathMatch) continue;

    const routePath = normalizePath(joinPaths(prefixes.get(alias) || '', unquote(pathMatch[2])));
    const tokens = pathTokens(routePath);
    const family = familyKeyFromTokens(tokens.length > 0 ? tokens : splitTokens(routePath));

    items.push({
      side: 'backend',
      kind,
      family,
      tokens,
      label: `${method} ${routePath}`,
      path: routePath,
      locations: [makeLocation(toPosix(path.relative(rootDir, filePath)), index + 1, method)],
    });
  }

  return items;
}

function collectBackendInventory(rootDir) {
  const items = [];
  const routeFiles = [
    path.join(rootDir, 'backend', 'router', 'api_routes.go'),
    path.join(rootDir, 'backend', 'router', 'data_routes.go'),
    path.join(rootDir, 'backend', 'router', 'level_routes.go'),
    path.join(rootDir, 'main.go'),
  ];

  for (const filePath of routeFiles) {
    if (fs.existsSync(filePath)) {
      items.push(...collectGoRouteItems(rootDir, filePath, 'api'));
    }
  }

  const supportDirs = [
    path.join(rootDir, 'backend', 'handlers'),
    path.join(rootDir, 'backend', 'anticheat'),
    path.join(rootDir, 'backend', 'game'),
    path.join(rootDir, 'backend', 'plugins'),
  ];

  for (const dirPath of supportDirs) {
    const goFiles = walkFiles(dirPath, (filePath) => filePath.endsWith('.go') && !filePath.endsWith('_test.go'));
    for (const filePath of goFiles) {
      const content = readText(filePath);
      const relPath = toPosix(path.relative(rootDir, filePath));
      const kind = filePath.includes(`${path.sep}handlers${path.sep}`) ? 'handler' : 'service';
      const funcRegex = /^\s*func\s+(?:\([^)]*\)\s*)?([A-Z][A-Za-z0-9_]*)\s*\(/gm;
      let funcMatch;
      while ((funcMatch = funcRegex.exec(content))) {
        const funcName = funcMatch[1];
        const tokens = splitTokens(funcName);
        const family = familyKeyFromTokens(tokens.length > 0 ? tokens : splitTokens(funcName));

        items.push({
          side: 'backend',
          kind,
          family,
          tokens,
          label: funcName,
          locations: [makeLocation(relPath, lineNumberAt(content, funcMatch.index + 1), funcName)],
        });
      }
    }
  }

  return items;
}

function summarizeLocations(items) {
  return items
    .flatMap((item) => item.locations || [])
    .slice(0, 4)
    .map((location) => ({
      file: location.file,
      line: location.line,
      note: location.note,
    }));
}

function classifyFamilies(frontendItems, backendItems) {
  const familyMap = new Map();

  const add = (item) => {
    if (!item.family) return;
    if (!familyMap.has(item.family)) {
      familyMap.set(item.family, { family: item.family, frontend: [], backend: [] });
    }
    familyMap.get(item.family)[item.side].push(item);
  };

  frontendItems.forEach(add);
  backendItems.forEach(add);

  const findings = [];
  const coverage = [];
  const summary = {
    matched: 0,
    'frontend-only': 0,
    'backend-only': 0,
    ambiguous: 0,
  };

  for (const family of Array.from(familyMap.values()).sort((a, b) => a.family.localeCompare(b.family))) {
    const hasFrontendDirect = family.frontend.some((item) => item.kind === 'route' || item.kind === 'page');
    const hasBackendDirect = family.backend.some((item) => item.kind === 'api');
    const hasFrontendSupport = family.frontend.some((item) => item.kind === 'component');
    const hasBackendSupport = family.backend.some((item) => item.kind === 'service' || item.kind === 'handler');

    if (!hasFrontendDirect && !hasBackendDirect) {
      continue;
    }

    let classification = 'ambiguous';
    let reason = 'The family has supporting evidence on at least one side, but no direct route/API pairing was established.';

    if (hasFrontendDirect && hasBackendDirect) {
      classification = 'matched';
      reason = 'Feature family has both frontend and backend direct coverage in the audited sources.';
    } else if (hasFrontendDirect && hasBackendSupport) {
      classification = 'matched';
      reason = 'Frontend routes/pages exist and backend handler/service evidence was found for the same family.';
    } else if (hasFrontendDirect) {
      classification = 'frontend-only';
      reason = hasBackendSupport
        ? 'Frontend routes/pages exist, but the backend evidence did not include a direct API route for the same family.'
        : 'Feature family was found on the frontend but no identified backend counterpart was found.';
    } else if (hasBackendDirect && hasFrontendSupport) {
      classification = 'matched';
      reason = 'Backend API routes exist and frontend component evidence was found for the same family.';
    } else if (hasBackendDirect) {
      classification = 'backend-only';
      reason = 'Feature family was found on the backend but no identified frontend counterpart was found.';
    }

    summary[classification] += 1;
    coverage.push({
      family: family.family,
      classification,
      frontend: {
        direct: hasFrontendDirect,
        support: hasFrontendSupport,
        count: family.frontend.length,
      },
      backend: {
        direct: hasBackendDirect,
        support: hasBackendSupport,
        count: family.backend.length,
      },
    });

    if (classification === 'matched') {
      continue;
    }

    findings.push({
      family: family.family,
      classification,
      reason,
      frontend: {
        count: family.frontend.length,
        evidence: summarizeLocations(family.frontend),
      },
      backend: {
        count: family.backend.length,
        evidence: summarizeLocations(family.backend),
      },
    });
  }

  return { findings, summary, families: familyMap.size, coverage };
}

function generateAuditReport(rootDir) {
  const frontendItems = collectFrontendInventory(rootDir);
  const backendItems = collectBackendInventory(rootDir);
  const analysis = classifyFamilies(frontendItems, backendItems);

  return {
    schema: 'spec-driven',
    change: 'audit-feature-fullstack-implementation',
    sourceDefinitions: {
      frontend: ['frontend/src/router.ts', 'frontend/src/pages/**/*.vue', 'frontend/src/components/**/*.vue'],
      backend: ['backend/router/*.go', 'backend/handlers/**/*.go', 'backend/anticheat/**/*.go', 'backend/game/**/*.go', 'backend/plugins/**/*.go'],
    },
    inventory: {
      frontend: frontendItems.length,
      backend: backendItems.length,
      families: analysis.families,
    },
    summary: analysis.summary,
    coverage: analysis.coverage,
    findings: analysis.findings,
  };
}

function formatMarkdown(report) {
  const lines = [];
  lines.push('## Feature Coverage Audit');
  lines.push('');
  lines.push(`- Schema: ${report.schema}`);
  lines.push(`- Inventory: ${report.inventory.frontend} frontend items, ${report.inventory.backend} backend items, ${report.inventory.families} feature families`);
  lines.push(`- Summary: matched ${report.summary.matched}, frontend-only ${report.summary['frontend-only']}, backend-only ${report.summary['backend-only']}, ambiguous ${report.summary.ambiguous}`);
  lines.push('');

  if (report.findings.length === 0) {
    lines.push('No coverage gaps were found in the current heuristic pass.');
    return lines.join('\n');
  }

  lines.push('### Findings');
  lines.push('');
  for (const finding of report.findings) {
    lines.push(`- **${finding.family}**: ${finding.classification}`);
    lines.push(`  - Reason: ${finding.reason}`);
    if (finding.frontend.evidence.length > 0) {
      lines.push('  - Frontend evidence:');
      for (const item of finding.frontend.evidence) {
        lines.push(`    - ${item.file}:${item.line}${item.note ? ` (${item.note})` : ''}`);
      }
    }
    if (finding.backend.evidence.length > 0) {
      lines.push('  - Backend evidence:');
      for (const item of finding.backend.evidence) {
        lines.push(`    - ${item.file}:${item.line}${item.note ? ` (${item.note})` : ''}`);
      }
    }
    lines.push('');
  }

  return lines.join('\n');
}

function runCli(argv = process.argv.slice(2)) {
  const rootDir = path.resolve(__dirname, '..');
  const asJson = argv.includes('--json');
  const report = generateAuditReport(rootDir);

  if (asJson) {
    process.stdout.write(`${JSON.stringify(report, null, 2)}\n`);
    return report;
  }

  process.stdout.write(`${formatMarkdown(report)}\n`);
  return report;
}

if (require.main === module) {
  runCli();
}

module.exports = {
  collectFrontendInventory,
  collectBackendInventory,
  classifyFamilies,
  generateAuditReport,
  formatMarkdown,
  runCli,
};