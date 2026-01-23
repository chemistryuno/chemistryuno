import configService = require('./configService');

function getConfig() {
  return configService.getConfig();
}

function getElementCounts() {
  return configService.getElementCounts();
}

function getSpecialCards() {
  return configService.getSpecialCards();
}

// 建立物质到元素的映射
function buildCompoundToElements(): Record<string, string[]> {
  const map: Record<string, string[]> = {};
  
  const allCompounds: string[] = [];
  
  const db = getConfig();

  // 收集所有物质
  for (const category in db.common_compounds || {}) {
    if (Array.isArray(db.common_compounds[category])) {
      allCompounds.push(...db.common_compounds[category]);
    } else if (typeof db.common_compounds[category] === 'object') {
      for (const subcat in db.common_compounds[category]) {
        if (Array.isArray(db.common_compounds[category][subcat])) {
          allCompounds.push(...db.common_compounds[category][subcat]);
        }
      }
    }
  }
  
  // 解析每个物质的元素
  allCompounds.forEach(compound => {
    const elements = extractElements(compound);
    map[compound] = elements;
  });
  
  return map;
}

// 从化学式中提取元素
function extractElements(formula: string): string[] {
  const elements = new Set<string>();
  
  // 匹配大写字母（可能后跟小写字母）
  const elementPattern = /[A-Z][a-z]?/g;
  const matches = formula.match(elementPattern);
  
  if (matches) {
    matches.forEach(element => {
      if (element && element.length <= 2) {
        elements.add(element);
      }
    });
  }
  
  return Array.from(elements);
}

// 获取两个物质之间的反应关系
function getReactionBetweenCompounds(compound1: string, compound2: string): any {
  const db = getConfig();
  for (const reaction of db.representative_reactions || []) {
    if (reaction.reactions) {
      for (const rxn of reaction.reactions) {
        // 简化匹配：只检查两个物质是否都在反应中
        if ((rxn.includes(compound1) && rxn.includes(compound2)) ||
            (rxn.includes(compound2) && rxn.includes(compound1))) {
          return {
            type: reaction.type,
            reaction: rxn
          };
        }
      }
    }
  }
  return null;
}

// 根据当前元素列表获取可以组成的物质
function getCompoundsByElements(elements: string[]): string[] {
  const elementSet = new Set(elements);
  const compoundMap = buildCompoundToElements();
  const possibleCompounds: string[] = [];
  
  for (const [compound, requiredElements] of Object.entries(compoundMap)) {
    // 检查是否所有必需的元素都在玩家的元素中
    if (requiredElements.every(elem => elementSet.has(elem))) {
      possibleCompounds.push(compound);
    }
  }
  
  return possibleCompounds;
}

// 初始化卡牌堆
function initializeDeck(): string[] {
  const deck: string[] = [];
  const elementCounts = getElementCounts();

  for (const [card, count] of Object.entries(elementCounts)) {
    for (let i = 0; i < count; i++) {
      deck.push(card);
    }
  }
  
  // 洗牌
  return shuffleDeck(deck);
}

// 根据玩家数量初始化卡牌堆（每2人增加一组牌）
export function initializeDeckForPlayers(playerCount: number, multiplier: number | null = null): string[] {
  const deck: string[] = [];
  const elementCounts = getElementCounts();
  
  // 如果没有指定倍数，根据玩家数量计算（每2人一组）
  const deckMultiplier = multiplier || Math.ceil(playerCount / 2);
  
  for (const [card, count] of Object.entries(elementCounts)) {
    for (let i = 0; i < count * deckMultiplier; i++) {
      deck.push(card);
    }
  }
  
  // 洗牌
  return shuffleDeck(deck);
}

// 洗牌
function shuffleDeck(deck: string[]): string[] {
  const shuffled = [...deck];
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
  }
  return shuffled;
}

// 初始化游戏
export function initializeGame(playerCount: number): any {
  const deck = initializeDeck();
  const players = [];
  
  for (let i = 0; i < playerCount; i++) {
    const hand: string[] = [];
    for (let j = 0; j < 10; j++) {
      const card = deck.pop();
      if (card) hand.push(card);
    }
    players.push({
      id: i,
      hand: hand,
      compounds: [] // 玩家打出的物质历史
    });
  }
  
  return {
    deck: deck,
    players: players,
    currentPlayer: 0,
    direction: 1, // 1 顺时针，-1 逆时针
    lastCompound: null,
    gameActive: true
  };
}

// 检查是否可以打出一个物质（基于上一个物质）
export function canPlayCompound(currentCompound: string, lastCompound: string | null, compoundMap: Record<string, string[]>): boolean {
  if (!lastCompound) return true; // 第一轮
  
  const currentElements = compoundMap[currentCompound] || [];
  const lastElements = compoundMap[lastCompound] || [];
  
  // 检查是否有共同元素
  const currentSet = new Set(currentElements);
  const hasCommonElement = lastElements.some(elem => currentSet.has(elem));
  
  if (hasCommonElement) {
    // 有共同元素，检查是否能反应
    const reaction = getReactionBetweenCompounds(lastCompound, currentCompound);
    return reaction !== null;
  }
  
  return false;
}

// 处理特殊卡牌
export function applySpecialCard(card: string, gameState: any): void {
  const special = getSpecialCards()[card];
  
  if (special === 'reverse') {
    gameState.direction *= -1;
  } else if (special === 'skip') {
    gameState.currentPlayer = (gameState.currentPlayer + gameState.direction + gameState.players.length) % gameState.players.length;
  } else if (special === 'draw2') {
    const nextPlayer = (gameState.currentPlayer + gameState.direction + gameState.players.length) % gameState.players.length;
    for (let i = 0; i < 2; i++) {
      if (gameState.deck.length > 0) {
        gameState.players[nextPlayer].hand.push(gameState.deck.pop());
      }
    }
  } else if (special === 'draw4') {
    const nextPlayer = (gameState.currentPlayer + gameState.direction + gameState.players.length) % gameState.players.length;
    for (let i = 0; i < 4; i++) {
      if (gameState.deck.length > 0) {
        gameState.players[nextPlayer].hand.push(gameState.deck.pop());
      }
    }
  }
}

// 移到下一个玩家
export function nextPlayer(gameState: any): void {
  gameState.currentPlayer = (gameState.currentPlayer + gameState.direction + gameState.players.length) % gameState.players.length;
}

export {
  getCompoundsByElements,
  getReactionBetweenCompounds,
  extractElements,
  buildCompoundToElements,
  getSpecialCards,
  getElementCounts
};
