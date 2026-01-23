import * as fs from 'fs';
import * as path from 'path';
import { EventEmitter } from 'events';

// 配置文件路径：支持从环境变量指定，否则使用默认路径
// 开发环境：server/configService.ts → ../config.json
// 生产环境：server/dist/configService.js → ../../config.json
const CONFIG_PATH = process.env.CONFIG_PATH || path.join(__dirname, '../../config.json');

// 类型定义
interface ElementCounts {
  [key: string]: number;
}

interface SpecialCards {
  [key: string]: string;
}

interface CardConfig {
  element_counts: ElementCounts;
  special_cards: SpecialCards;
}

interface Metadata {
  elements?: string[];
  note?: string;
}

interface CommonCompounds {
  [category: string]: any;
}

interface Reaction {
  type: string;
  reactions?: string[];
}

interface GameConfig {
  metadata: Metadata;
  card_config: CardConfig;
  common_compounds: CommonCompounds;
  representative_reactions: Reaction[];
  reactivity_series?: any;
  solubility_rules?: any;
  elemental_substances?: any;
  game_settings?: any;
}

// 默认卡牌配置，作为缺失字段时的兜底
const DEFAULT_CARD_CONFIG: CardConfig = {
  element_counts: {
    H: 12,
    O: 12,
    C: 4,
    N: 4,
    F: 4,
    Na: 4,
    Mg: 4,
    Al: 4,
    Si: 4,
    P: 4,
    S: 4,
    Cl: 4,
    K: 4,
    Ca: 4,
    Mn: 4,
    Fe: 4,
    Cu: 4,
    Zn: 4,
    Br: 4,
    I: 4,
    Ag: 4,
    '+4': 4,
    '+2': 8,
    He: 1,
    Ne: 1,
    Ar: 1,
    Kr: 1,
    Au: 4
  },
  special_cards: {
    He: 'reverse',
    Ne: 'reverse',
    Ar: 'reverse',
    Kr: 'reverse',
    Au: 'skip',
    '+4': 'draw4',
    '+2': 'draw2'
  }
};

function readJson(filePath: string): any | null {
  if (!fs.existsSync(filePath)) return null;
  try {
    const raw = fs.readFileSync(filePath, 'utf8');
    return JSON.parse(raw);
  } catch (err: any) {
    // 配置文件读取失败，使用默认值
    return null;
  }
}

function loadInitialConfig(): GameConfig {
  // 读取 config.json，不存在则使用内置默认值
  const configFromFile = readJson(CONFIG_PATH);
  if (configFromFile) {
    return applyDefaults(configFromFile);
  }

  return {
    metadata: { elements: [], note: 'fallback config' },
    card_config: DEFAULT_CARD_CONFIG,
    common_compounds: {},
    representative_reactions: [],
    reactivity_series: {},
    solubility_rules: {}
  };
}

function applyDefaults(config: Partial<GameConfig>): GameConfig {
  const next = { ...config } as GameConfig;
  next.card_config = {
    element_counts: {
      ...DEFAULT_CARD_CONFIG.element_counts,
      ...(config.card_config?.element_counts || {})
    },
    special_cards: {
      ...DEFAULT_CARD_CONFIG.special_cards,
      ...(config.card_config?.special_cards || {})
    }
  };
  return next;
}

class ConfigService extends EventEmitter {
  private config: GameConfig;

  constructor() {
    super();
    this.config = loadInitialConfig();
  }

  getConfig(): GameConfig {
    return this.config;
  }

  getElementCounts(): ElementCounts {
    return this.config.card_config?.element_counts || DEFAULT_CARD_CONFIG.element_counts;
  }

  getSpecialCards(): SpecialCards {
    return this.config.card_config?.special_cards || DEFAULT_CARD_CONFIG.special_cards;
  }

  getElementsList(): string[] {
    return this.config.metadata?.elements || [];
  }

  refreshFromDisk(): GameConfig {
    this.config = loadInitialConfig();
    this.emit('changed', this.config);
    return this.config;
  }

  saveConfig(newConfig: GameConfig): GameConfig {
    if (!newConfig || typeof newConfig !== 'object') {
      throw new Error('配置格式无效');
    }
    console.log('💾 服务器收到配置更新请求');
    console.log('  - elemental_substances 存在:', !!newConfig.elemental_substances);
    if (newConfig.elemental_substances) {
      console.log('  - metal_elements:', newConfig.elemental_substances.metal_elements?.length, '项');
      console.log('  - non_metal_elements:', Object.keys(newConfig.elemental_substances.non_metal_elements || {}).length, '个类别');
    }
    
    this.config = applyDefaults(newConfig);
    
    console.log('  应用默认值后 elemental_substances 存在:', !!this.config.elemental_substances);
    if (this.config.elemental_substances) {
      console.log('  - 保留了 metal_elements:', this.config.elemental_substances.metal_elements?.length, '项');
    }
    
    fs.writeFileSync(CONFIG_PATH, JSON.stringify(this.config, null, 2), 'utf8');
    console.log('✅ 配置已写入磁盘');
    this.emit('changed', this.config);
    return this.config;
  }

  updateConfig(updater: (config: GameConfig) => GameConfig): GameConfig {
    const draft = JSON.parse(JSON.stringify(this.config));
    const updated = updater(draft);
    return this.saveConfig(updated);
  }

  onChange(callback: (config: GameConfig) => void): void {
    this.on('changed', callback);
  }
}

export = new ConfigService();

