#!/usr/bin/env node

/**
 * 版本号同步脚本
 * 从 .env 文件读取版本号并同步到所有 package.json 文件
 *
 * 使用方法：
 * node sync-version.js
 */

const fs = require('fs');
const path = require('path');

// 读取 .env 文件
function loadEnv() {
  const envPath = path.join(__dirname, '.env');
  if (!fs.existsSync(envPath)) {
    console.error('❌ .env 文件不存在！');
    process.exit(1);
  }

  const envContent = fs.readFileSync(envPath, 'utf-8');
  const env = {};

  envContent.split('\n').forEach(line => {
    const trimmed = line.trim();
    if (trimmed && !trimmed.startsWith('#')) {
      const [key, ...values] = trimmed.split('=');
      if (key && values.length > 0) {
        env[key.trim()] = values.join('=').trim();
      }
    }
  });

  return env;
}

// 更新 package.json 文件
function updatePackageJson(filePath, version, versionName) {
  if (!fs.existsSync(filePath)) {
    console.warn(`⚠️  文件不存在: ${filePath}`);
    return false;
  }

  try {
    const content = fs.readFileSync(filePath, 'utf-8');
    const pkg = JSON.parse(content);

    // 更新版本号
    pkg.version = version;

    // 如果有 description 字段包含版本信息，也更新它
    if (pkg.description && pkg.description.includes('V')) {
      pkg.description = pkg.description.replace(/V[\d.]+/g, `V${version}`);
    }

    // 写回文件，保持格式
    fs.writeFileSync(filePath, JSON.stringify(pkg, null, 2) + '\n', 'utf-8');
    console.log(`✅ 已更新: ${filePath}`);
    return true;
  } catch (error) {
    console.error(`❌ 更新失败 ${filePath}:`, error.message);
    return false;
  }
}

// 主函数
function main() {
  console.log('🚀 开始同步版本号...\n');

  // 读取环境变量
  const env = loadEnv();
  const version = env.APP_VERSION;
  const versionName = env.APP_VERSION_NAME;

  if (!version) {
    console.error('❌ .env 文件中未找到 APP_VERSION 配置！');
    process.exit(1);
  }

  console.log(`📦 版本号: ${version}`);
  console.log(`📛 版本名: ${versionName || 'N/A'}\n`);

  // 更新根目录的 package.json
  updatePackageJson(path.join(__dirname, 'package.json'), version, versionName);

  // 更新前端的 package.json
  updatePackageJson(path.join(__dirname, 'frontend', 'package.json'), version, versionName);

  console.log('\n✨ 版本号同步完成！');
  console.log(`\n💡 提示: 记得提交更改并重新构建项目`);
}

// 运行
main();
