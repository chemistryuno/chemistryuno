import configService = require('./configService');

/**
 * 物质和反应库 - 提供优化的查询性能
 */
class ChemistryDatabase {
  private config: any;
  public compoundToElements: Record<string, string[]>;
  private reactionMap: Map<string, boolean>;

  constructor() {
    this.config = configService.getConfig();
    this.compoundToElements = this.buildCompoundToElements();
    this.reactionMap = this.buildReactionMap();
    
    console.log('🧪 Chemistry Database 初始化');
    console.log('  - 已加载物质数量:', Object.keys(this.compoundToElements).length);
    console.log('  - elemental_substances 存在:', !!this.config.elemental_substances);
    if (this.config.elemental_substances) {
      const elementToSimple = this.getElementToSimpleSubstance();
      console.log('  - 单质映射数量:', Object.keys(elementToSimple).length);
      console.log('  - 单质示例:', Object.entries(elementToSimple).slice(0, 5));
    }

    configService.onChange((nextConfig: any) => {
      this.config = nextConfig;
      this.compoundToElements = this.buildCompoundToElements();
      this.reactionMap = this.buildReactionMap();
      console.log('🔄 配置已更新，重新构建数据库');
    });
  }

  /**
   * 判断物质是否在配置的常见物质列表中
   */
  isKnownCompound(compound: string): boolean {
    return Boolean(this.compoundToElements[compound]);
  }

  /**
   * 构建物质到元素的映射
   */
  buildCompoundToElements(): Record<string, string[]> {
    const map: Record<string, string[]> = {};
    const allCompounds = this.getAllCompounds();

    allCompounds.forEach(compound => {
      const elements = this.extractElements(compound);
      map[compound] = elements;
    });

    return map;
  }

  /**
   * 获取所有物质
   */
  getAllCompounds(): string[] {
    const compounds: string[] = [];

    const addCompounds = (obj: any): void => {
      if (Array.isArray(obj)) {
        compounds.push(...obj);
      } else if (typeof obj === 'object' && obj !== null) {
        Object.values(obj).forEach(val => {
          if (Array.isArray(val)) {
            compounds.push(...val);
          } else if (typeof val === 'object') {
            addCompounds(val);
          }
        });
      }
    };

    // 添加化合物
    addCompounds(this.config.common_compounds || {});
    
    // 添加单质
    if (this.config.elemental_substances) {
      const elementalSubstances = this.config.elemental_substances;
      
      // 添加金属单质
      if (Array.isArray(elementalSubstances.metal_elements)) {
        compounds.push(...elementalSubstances.metal_elements);
      }
      
      // 添加非金属单质
      if (elementalSubstances.non_metal_elements) {
        addCompounds(elementalSubstances.non_metal_elements);
      }
    }
    
    return compounds;
  }

  /**
   * 从化学式中提取元素符号
   * @param formula - 化学式
   * @returns 元素列表
   */
  extractElements(formula: string): string[] {
    const elements = new Set<string>();
    const elementPattern = /[A-Z][a-z]?/g;
    const matches = formula.match(elementPattern);

    if (matches) {
      matches.forEach(element => {
        if (element && element.length <= 2) {
          // 验证是否为有效的元素符号
          if (this.config.metadata?.elements?.includes(element) ||
              Object.keys(this.getSpecialCards()).includes(element)) {
            elements.add(element);
          }
        }
      });
    }

    return Array.from(elements);
  }

  /**
   * 获取特殊卡牌
   */
  getSpecialCards(): Record<string, string> {
    return configService.getSpecialCards();
  }

  /**
   * 构建反应映射 - 优化查询性能
   */
  buildReactionMap(): Map<string, boolean> {
    const map = new Map<string, boolean>();

    if (this.config.representative_reactions && typeof this.config.representative_reactions === 'object') {
      // 新格式：{ "HCl": ["NaOH", "Ca(OH)2"], ... }
      for (const [reactant, partners] of Object.entries(this.config.representative_reactions)) {
        const normalizedReactant = reactant.replace(/[↓↑]/g, '').trim();
        
        if (Array.isArray(partners)) {
          partners.forEach((partner: string) => {
            const normalizedPartner = partner.replace(/[↓↑]/g, '').trim();
            const key = `${normalizedReactant}|${normalizedPartner}`;
            map.set(key, true);
          });
        }
      }
    }

    return map;
  }

  /**
   * 检查两个物质是否能反应
   * @param compound1
   * @param compound2
   * @returns 能否反应
   */
  getReactionBetweenCompounds(compound1: string, compound2: string): boolean {
    // 规范化物质名称（移除箭头、向上箭头等符号）
    const normalize = (str: string) => str.replace(/[↓↑]/g, '').trim();
    const c1 = normalize(compound1);
    const c2 = normalize(compound2);

    // 检查双向反应
    const key = `${c1}|${c2}`;
    const reverseKey = `${c2}|${c1}`;
    
    return this.reactionMap.has(key) || this.reactionMap.has(reverseKey);
  }

  /**
   * 获取元素对应的单质映射（从配置动态生成）
   */
  getElementToSimpleSubstance(): Record<string, string> {
    const mapping: Record<string, string> = {};
    
    // 如果配置中有 elemental_substances，从配置中读取
    if (this.config.elemental_substances) {
      const elemental = this.config.elemental_substances;
      
      // 处理非金属单质
      if (elemental.non_metal_elements) {
        const nonMetals = elemental.non_metal_elements;
        
        // 双原子分子：H2, O2, N2, F2, Cl2, Br2, I2
        if (Array.isArray(nonMetals.diatomic_molecules)) {
          nonMetals.diatomic_molecules.forEach((molecule: string) => {
            const element = molecule.replace(/\d+/g, ''); // 去除数字，如 H2 -> H
            if (element) {
              mapping[element] = molecule;
            }
          });
        }
        
        // 多原子分子：P4, S8
        if (Array.isArray(nonMetals.polyatomic_molecules)) {
          nonMetals.polyatomic_molecules.forEach((molecule: string) => {
            const element = molecule.replace(/\d+/g, '');
            if (element) {
              mapping[element] = molecule;
            }
          });
        }
        
        // 原子晶体：C, Si（元素符号本身就是单质）
        if (Array.isArray(nonMetals.atomic_crystals)) {
          nonMetals.atomic_crystals.forEach((element: string) => {
            mapping[element] = element;
          });
        }
        
        // 稀有气体：He, Ne, Ar, Kr（元素符号本身就是单质）
        if (Array.isArray(nonMetals.noble_gases)) {
          nonMetals.noble_gases.forEach((element: string) => {
            mapping[element] = element;
          });
        }
      }
      
      // 处理金属单质（元素符号本身就是单质）
      if (Array.isArray(elemental.metal_elements)) {
        elemental.metal_elements.forEach((element: string) => {
          mapping[element] = element;
        });
      }
    } else {
      // 回退到硬编码（向后兼容）
      return {
        'H': 'H2',
        'O': 'O2',
        'N': 'N2',
        'Cl': 'Cl2',
        'F': 'F2',
        'Br': 'Br2',
        'I': 'I2',
        'P': 'P4',
        'S': 'S8',
        'C': 'C',
        'Fe': 'Fe',
        'Cu': 'Cu',
        'Zn': 'Zn',
        'Al': 'Al',
        'Mg': 'Mg',
        'Ca': 'Ca',
        'Na': 'Na',
        'K': 'K',
        'Ag': 'Ag',
        'Mn': 'Mn',
        'Si': 'Si'
      };
    }
    
    return mapping;
  }

  /**
   * 根据持有的元素获取可能的物质
   * @param elements - 持有的元素
   * @returns 可能的物质列表
   */
  getCompoundsByElements(elements: string[]): string[] {
    const elementSet = new Set(elements);
    const possibleCompounds: string[] = [];
    const elementToSimple = this.getElementToSimpleSubstance();

    // 如果是单个元素，添加对应的单质
    if (elements.length === 1) {
      const element = elements[0];
      const simpleSubstance = elementToSimple[element];
      if (simpleSubstance) {
        possibleCompounds.push(simpleSubstance);
      }
    }

    // 添加所有包含该元素的化合物
    for (const [compound, requiredElements] of Object.entries(this.compoundToElements)) {
      // 检查是否至少有一个必需的元素在玩家的元素中
      // 这样玩家打出 H 时可以看到所有含 H 的物质
      if (requiredElements.some(elem => elementSet.has(elem))) {
        // 避免重复添加单质
        if (!possibleCompounds.includes(compound)) {
          possibleCompounds.push(compound);
        }
      }
    }

    return possibleCompounds;
  }

  /**
   * 检查是否可以打出物质
   * @param currentCompound - 当前物质
   * @param lastCompound - 上一个物质
   * @returns
   */
  canPlayCompound(currentCompound: string, lastCompound: string | null): boolean {
    if (!lastCompound) return true; // 第一轮
    return this.getReactionBetweenCompounds(lastCompound, currentCompound) !== false;
  }

  /**
   * 获取物质的类别
   * @param compound
   * @returns 类别名称
   */
  getCompoundCategory(compound: string): string {
    for (const [category, items] of Object.entries(this.config.common_compounds || {})) {
      if (Array.isArray(items) && items.includes(compound)) {
        return category;
      }
      if (typeof items === 'object' && items !== null) {
        for (const [subcat, subitems] of Object.entries(items)) {
          if (Array.isArray(subitems) && (subitems as string[]).includes(compound)) {
            return `${category}/${subcat}`;
          }
        }
      }
    }
    return 'unknown';
  }

  /**
   * 验证化学式的格式
   * @param formula
   * @returns
   */
  isValidFormula(formula: string): boolean {
    // 简单的化学式验证
    const pattern = /^[A-Z][a-z]?(\d+)?([A-Z][a-z]?(\d+)?)*(\([A-Z][a-z]?(\d+)?\)(\d+)?)*$/;
    return pattern.test(formula);
  }
}

// 创建单例
const database = new ChemistryDatabase();

export = database;
