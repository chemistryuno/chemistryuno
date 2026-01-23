"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
const fs = __importStar(require("fs"));
const path = __importStar(require("path"));
const events_1 = require("events");
// 配置文件路径：支持从环境变量指定，否则使用默认路径
// 开发环境：server/configService.ts → ../config.json
// 生产环境：server/dist/configService.js → ../../config.json
const CONFIG_PATH = process.env.CONFIG_PATH || path.join(__dirname, '../../config.json');
// 默认卡牌配置，作为缺失字段时的兜底
const DEFAULT_CARD_CONFIG = {
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
function readJson(filePath) {
    if (!fs.existsSync(filePath))
        return null;
    try {
        const raw = fs.readFileSync(filePath, 'utf8');
        return JSON.parse(raw);
    }
    catch (err) {
        // 配置文件读取失败，使用默认值
        return null;
    }
}
function loadInitialConfig() {
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
function applyDefaults(config) {
    const next = { ...config };
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
class ConfigService extends events_1.EventEmitter {
    constructor() {
        super();
        this.config = loadInitialConfig();
    }
    getConfig() {
        return this.config;
    }
    getElementCounts() {
        return this.config.card_config?.element_counts || DEFAULT_CARD_CONFIG.element_counts;
    }
    getSpecialCards() {
        return this.config.card_config?.special_cards || DEFAULT_CARD_CONFIG.special_cards;
    }
    getElementsList() {
        return this.config.metadata?.elements || [];
    }
    refreshFromDisk() {
        this.config = loadInitialConfig();
        this.emit('changed', this.config);
        return this.config;
    }
    saveConfig(newConfig) {
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
    updateConfig(updater) {
        const draft = JSON.parse(JSON.stringify(this.config));
        const updated = updater(draft);
        return this.saveConfig(updated);
    }
    onChange(callback) {
        this.on('changed', callback);
    }
}
module.exports = new ConfigService();
