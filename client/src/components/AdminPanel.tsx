import React, { useEffect, useMemo, useState } from 'react';
import axios from 'axios';
import './AdminPanel.css';
import API_ENDPOINTS from '../config/api';

interface EffectOption {
  value: string;
  label: string;
}

interface CardConfig {
  element_counts: Record<string, number>;
  special_cards: Record<string, string>;
}

interface GameSettings {
  reconnect_timeout: number;
  host_timeout: number;
  note?: string;
}

interface ElementalSubstances {
  non_metal_elements?: {
    diatomic_molecules?: string[] | string;
    polyatomic_molecules?: string[] | string;
    atomic_crystals?: string[] | string;
    noble_gases?: string[] | string;
  };
  metal_elements?: string[] | string;
  note?: string;
}

interface Config {
  card_config?: CardConfig;
  game_settings?: GameSettings;
  common_compounds?: Record<string, any>;
  elemental_substances?: ElementalSubstances;
  representative_reactions?: Record<string, string[]>;
}

const EFFECT_OPTIONS: EffectOption[] = [
  { value: 'reverse', label: '反转 (reverse)' },
  { value: 'skip', label: '跳过 (skip)' },
  { value: 'draw2', label: '+2 (draw2)' },
  { value: 'draw4', label: '+4 (draw4)' }
];

const deepClone = <T,>(value: T): T => JSON.parse(JSON.stringify(value));

const AdminPanel: React.FC = () => {
  const [config, setConfig] = useState<Config | null>(null);
  const [draft, setDraft] = useState<Config | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [saving, setSaving] = useState<boolean>(false);
  const [message, setMessage] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [newCard, setNewCard] = useState({ name: '', count: 1 });
  const [newSpecial, setNewSpecial] = useState({ name: '', effect: 'reverse' });
  const [newReaction, setNewReaction] = useState({ reactant: '', partners: '' });
  const [lastSavedAt, setLastSavedAt] = useState<Date | null>(null);

  useEffect(() => {
    fetchConfig();
  }, []);

  useEffect(() => {
    if (draft) {
      if (draft.elemental_substances) {
      }
    }
  }, [draft]);

  const fetchConfig = async (): Promise<void> => {
    setLoading(true);
    setError('');
    try {
      const res = await axios.get(API_ENDPOINTS.config);
      const loadedConfig: Config = res.data.config;
      
      
      // 确保 elemental_substances 存在
      if (!loadedConfig.elemental_substances) {
        // 创建默认物质配置
        loadedConfig.elemental_substances = {
          non_metal_elements: {
            diatomic_molecules: [],
            polyatomic_molecules: [],
            atomic_crystals: [],
            noble_gases: []
          },
          metal_elements: [],
          note: "单质分为金属单质、非金属单质和稀有气体单质"
        };
      } else {
        Object.keys(loadedConfig.elemental_substances).forEach(key => {
          if (key !== 'note') {
            const value = (loadedConfig.elemental_substances as any)[key];
            if (Array.isArray(value)) {
            } else if (typeof value === 'object') {
              const subKeys = Object.keys(value);
            }
          }
        });
      }
      
      // 确保 game_settings 存在
      if (!loadedConfig.game_settings) {
        // 创建默认游戏设置
        loadedConfig.game_settings = {
          reconnect_timeout: 30000,
          host_timeout: 30000,
          note: "reconnect_timeout: 普通玩家断线后保留位置的时间（毫秒），host_timeout: 房主断线后关闭房间的时间（毫秒）"
        };
      }
      
      setConfig(loadedConfig);
      setDraft(deepClone(loadedConfig));
      setLastSavedAt(new Date());
    } catch (err) {
      // 加载配置失败
      setError((err as any).response?.data?.error || '加载配置失败');
    } finally {
      setLoading(false);
    }
  };

  const handleRefreshFromDisk = async (): Promise<void> => {
    setLoading(true);
    setError('');
    try {
      const res = await axios.post(API_ENDPOINTS.configRefresh);
      const loadedConfig: Config = res.data.config;
      
      
      setConfig(loadedConfig);
      setDraft(deepClone(loadedConfig));
      setLastSavedAt(new Date());
      setMessage('配置已从磁盘重新加载');
      setTimeout(() => setMessage(''), 3000);
    } catch (err) {
      // 刷新配置失败
      setError((err as any).response?.data?.error || '刷新配置失败');
    } finally {
      setLoading(false);
    }
  };

  const resetDraft = (): void => {
    if (config) {
      setDraft(deepClone(config));
    }
  };

  const handleSave = async (): Promise<void> => {
    if (!draft) return;
    
    // 深拷贝以便清理
    const cleanedDraft = deepClone(draft);
    
    // 清理 common_compounds：将字符串转为数组
    if (cleanedDraft.common_compounds) {
      Object.keys(cleanedDraft.common_compounds).forEach(category => {
        const value = cleanedDraft.common_compounds![category];
        if (typeof value === 'string') {
          cleanedDraft.common_compounds![category] = value
            .split(',')
            .map(s => s.trim())
            .filter(Boolean);
        } else if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
          // 嵌套对象
          Object.keys(value).forEach(sub => {
            if (typeof value[sub] === 'string') {
              value[sub] = value[sub]
                .split(',')
                .map((s: string) => s.trim())
                .filter(Boolean);
            }
          });
        }
      });
    }
    
    // 清理 elemental_substances：将字符串转为数组
    if (cleanedDraft.elemental_substances) {
      Object.keys(cleanedDraft.elemental_substances).forEach(category => {
        if (category === 'note') return; // 跳过 note 字段
        
        const value = (cleanedDraft.elemental_substances as any)[category];
        if (typeof value === 'string') {
          (cleanedDraft.elemental_substances as any)[category] = value
            .split(',')
            .map(s => s.trim())
            .filter(Boolean);
        } else if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
          // 嵌套对象
          Object.keys(value).forEach(sub => {
            if (typeof value[sub] === 'string') {
              value[sub] = value[sub]
                .split(',')
                .map((s: string) => s.trim())
                .filter(Boolean);
            }
          });
        }
      });
    }
    
    // 清理 representative_reactions 中的空行
    cleanedDraft.representative_reactions = Object.fromEntries(
      Object.entries(cleanedDraft.representative_reactions || {}).map(([reactant, partners]: [string, any]) => {
        const cleanedPartners = Array.isArray(partners)
          ? partners.map((p: string) => p.trim()).filter(Boolean)
          : (typeof partners === 'string'
              ? (partners as string).split(',').map((p: string) => p.trim()).filter(Boolean)
              : []);
        return [reactant, cleanedPartners];
      })
    );
    
    if (cleanedDraft.elemental_substances) {
    }
    
    setSaving(true);
    setMessage('');
    setError('');
    try {
      const res = await axios.put(API_ENDPOINTS.config, cleanedDraft);
      setConfig(res.data.config);
      setDraft(deepClone(res.data.config));
      setMessage('配置已保存 ✓');
      setLastSavedAt(new Date());
    } catch (err) {
      const errorMsg = (err as any).response?.data?.error || '保存失败';
      setError(errorMsg);
      // 保存失败
    } finally {
      setSaving(false);
    }
  };

  const updateCardCount = (card: string, value: string | number): void => {
    const count = Number(value);
    setDraft((prev) => {
      const next = deepClone(prev!);
      next.card_config = next.card_config || { element_counts: {}, special_cards: {} };
      next.card_config.element_counts[card] = Number.isFinite(count) ? count : 0;
      return next;
    });
  };

  const handleAddCard = (): void => {
    if (!newCard.name.trim()) return;
    updateCardCount(newCard.name.trim(), newCard.count);
    setNewCard({ name: '', count: 1 });
  };

  const updateSpecialCard = (card: string, effect: string): void => {
    setDraft((prev) => {
      const next = deepClone(prev!);
      next.card_config = next.card_config || { element_counts: {}, special_cards: {} };
      next.card_config.special_cards[card] = effect;
      return next;
    });
  };

  const handleAddSpecial = (): void => {
    if (!newSpecial.name.trim()) return;
    updateSpecialCard(newSpecial.name.trim(), newSpecial.effect);
    setNewSpecial({ name: '', effect: 'reverse' });
  };

  const updateCompoundList = (path: string[], textValue: string): void => {
    // 保存原始文本，不立即解析，避免输入逗号时被过滤掉
    setDraft((prev) => {
      const next = deepClone(prev!);
      
      // 确保 elemental_substances 和 common_compounds 存在
      if (!next.elemental_substances) {
        next.elemental_substances = {
          non_metal_elements: {
            diatomic_molecules: [],
            polyatomic_molecules: [],
            atomic_crystals: [],
            noble_gases: []
          },
          metal_elements: [],
          note: "单质分为金属单质、非金属单质和稀有气体单质"
        };
      }
      if (!next.common_compounds) {
        next.common_compounds = {};
      }
      
      // 如果路径第一个是 elemental_substances，从那里开始
      let cursor: any = path[0] === 'elemental_substances' ? next.elemental_substances : next.common_compounds;
      const startIndex = path[0] === 'elemental_substances' ? 1 : 0;
      
      for (let i = startIndex; i < path.length - 1; i += 1) {
        const key = path[i];
        if (!cursor[key]) {
          cursor[key] = {};
        }
        cursor = cursor[key];
      }
      
      const lastKey = path[path.length - 1];
      // 直接保存原始文本，不做任何处理
      cursor[lastKey] = textValue;
      
      return next;
    });
  };

  const updateReactionPartners = (reactant: string, textValue: string): void => {
    // 直接将文本按行分割，不进行 trim 和 filter
    // 这样用户可以正常输入和编辑
    const partners = textValue.split('\n');

    setDraft((prev) => {
      return {
        ...prev!,
        representative_reactions: {
          ...prev!.representative_reactions,
          [reactant]: partners
        }
      };
    });
  };

  const addReactant = (): void => {
    if (!newReaction.reactant.trim()) return;
    const partners = newReaction.partners
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean);


    setDraft((prev) => {
      const next = deepClone(prev!);
      if (!next.representative_reactions) {
        next.representative_reactions = {};
      }
      next.representative_reactions[newReaction.reactant.trim()] = partners;
      return next;
    });
    setNewReaction({ reactant: '', partners: '' });
  };

  const orderedElementCounts = useMemo(() => {
    if (!draft?.card_config?.element_counts) return [];
    return Object.entries(draft.card_config.element_counts).sort((a, b) => a[0].localeCompare(b[0]));
  }, [draft]);

  const hasChanges = useMemo(() => {
    if (!config || !draft) return false;
    return JSON.stringify(config) !== JSON.stringify(draft);
  }, [config, draft]);

  const stats = useMemo(() => {
    const totalCards = orderedElementCounts.reduce((sum, [, count]) => sum + Number(count || 0), 0);
    const specialCount = draft?.card_config?.special_cards ? Object.keys(draft.card_config.special_cards).length : 0;
    const compoundGroups = draft?.common_compounds ? Object.keys(draft.common_compounds).length : 0;
    const elementalGroups = draft?.elemental_substances ? Object.keys(draft.elemental_substances).filter(k => k !== 'note').length : 0;
    const reactionTypes = draft?.representative_reactions ? Object.keys(draft.representative_reactions).length : 0;
    return { totalCards, specialCount, compoundGroups, elementalGroups, reactionTypes };
  }, [draft, orderedElementCounts]);

  if (loading) {
    return (
      <div className="admin-page">
        <div className="admin-card">加载配置中...</div>
      </div>
    );
  }

  if (!draft) {
    return (
      <div className="admin-page">
        <div className="admin-card error">无法加载配置</div>
      </div>
    );
  }

  return (
    <div className="admin-page">
      <div className="admin-header">
        <div>
          <h1>化学UNO 配置后台</h1>
          <p className="subtitle">在此调整卡牌数量、特殊牌效果、物质列表与反应规则</p>
          <div className="pill">已通过密码验证</div>
        </div>
        <div className="header-actions">
          <a className="ghost-btn" href="/">返回游戏</a>
          <button className="ghost-btn" onClick={handleRefreshFromDisk} disabled={loading || saving} title="从config.json重新加载配置">
            {loading ? '加载中...' : '刷新配置'}
          </button>
          <button className="ghost-btn" onClick={resetDraft} disabled={saving || !hasChanges}>重置更改</button>
          <button className="primary-btn" onClick={handleSave} disabled={saving || !hasChanges}>
            {saving ? '保存中...' : hasChanges ? '保存配置' : '无更改'}
          </button>
        </div>
      </div>

      <div className="meta-bar">
        <div className="meta-group">
          <span className={`badge ${hasChanges ? 'warn' : 'ok'}`}>
            {hasChanges ? '待保存' : '已同步'}
          </span>
          {lastSavedAt && (
            <span className="meta-text">上次保存：{lastSavedAt.toLocaleString()}</span>
          )}
        </div>
        <div className="meta-group">
          <span className="meta-text">总牌数：{stats.totalCards}</span>
          <span className="meta-text">特殊牌：{stats.specialCount}</span>
          <span className="meta-text">物质分类：{stats.compoundGroups}</span>
          <span className="meta-text">单质分类：{stats.elementalGroups}</span>
          <span className="meta-text">反应类型：{stats.reactionTypes}</span>
        </div>
      </div>

      {message && <div className="notice success">{message}</div>}
      {error && <div className="notice error">{error}</div>}

      <div className="admin-grid">
        <section className="admin-card">
          <div className="section-header">
            <h2>游戏设置</h2>
            <span className="section-desc">配置游戏基本参数</span>
          </div>
          <div className="settings-group">
            <div className="setting-item">
              <label>
                <span className="setting-label">玩家断线保留时间</span>
                <span className="setting-hint">普通玩家断线后保留位置的时间（秒）</span>
              </label>
              <input
                type="number"
                min={5}
                max={300}
                step={5}
                value={draft.game_settings?.reconnect_timeout ? draft.game_settings.reconnect_timeout / 1000 : 30}
                onChange={(e) => {
                  const seconds = Number(e.target.value);
                  setDraft((prev) => ({
                    ...prev!,
                    game_settings: {
                      ...prev!.game_settings!,
                      reconnect_timeout: seconds * 1000,
                      host_timeout: prev!.game_settings?.host_timeout || 30000
                    }
                  }));
                }}
              />
              <span className="setting-unit">秒</span>
            </div>
            <div className="setting-item">
              <label>
                <span className="setting-label">房主断线关闭房间时间</span>
                <span className="setting-hint">房主断线后自动关闭房间的时间（秒）</span>
              </label>
              <input
                type="number"
                min={5}
                max={300}
                step={5}
                value={draft.game_settings?.host_timeout ? draft.game_settings.host_timeout / 1000 : 30}
                onChange={(e) => {
                  const seconds = Number(e.target.value);
                  setDraft((prev) => ({
                    ...prev!,
                    game_settings: {
                      ...prev!.game_settings!,
                      host_timeout: seconds * 1000,
                      reconnect_timeout: prev!.game_settings?.reconnect_timeout || 30000
                    }
                  }));
                }}
              />
              <span className="setting-unit">秒</span>
            </div>
          </div>
        </section>

        <section className="admin-card">
          <div className="section-header">
            <h2>卡牌与数量</h2>
            <span className="section-desc">编辑所有卡牌的库存数量</span>
          </div>
          <div className="table">
            <div className="table-head">
              <span>卡牌</span>
              <span>数量</span>
            </div>
            {orderedElementCounts.map(([card, count]) => (
              <div key={card} className="table-row">
                <span className="mono">{card}</span>
                <input
                  type="number"
                  min={0}
                  value={count}
                  onChange={(e) => updateCardCount(card, e.target.value)}
                />
              </div>
            ))}
          </div>
          <div className="inline-form">
            <input
              type="text"
              placeholder="新增卡牌符号"
              value={newCard.name}
              onChange={(e) => setNewCard({ ...newCard, name: e.target.value })}
            />
            <input
              type="number"
              min={0}
              value={newCard.count}
              onChange={(e) => setNewCard({ ...newCard, count: Number(e.target.value) })}
            />
            <button onClick={handleAddCard}>添加卡牌</button>
          </div>
        </section>

        <section className="admin-card">
          <div className="section-header">
            <h2>特殊牌效果</h2>
            <span className="section-desc">配置特殊牌对应的效果</span>
          </div>
          <div className="table">
            <div className="table-head">
              <span>卡牌</span>
              <span>效果</span>
            </div>
            {draft.card_config?.special_cards &&
              Object.entries(draft.card_config.special_cards).map(([card, effect]) => (
                <div key={card} className="table-row">
                  <span className="mono">{card}</span>
                  <select value={effect} onChange={(e) => updateSpecialCard(card, e.target.value)}>
                    {EFFECT_OPTIONS.map((opt) => (
                      <option key={opt.value} value={opt.value}>
                        {opt.label}
                      </option>
                    ))}
                  </select>
                </div>
              ))}
          </div>
          <div className="inline-form">
            <input
              type="text"
              placeholder="新增特殊牌"
              value={newSpecial.name}
              onChange={(e) => setNewSpecial({ ...newSpecial, name: e.target.value })}
            />
            <select
              value={newSpecial.effect}
              onChange={(e) => setNewSpecial({ ...newSpecial, effect: e.target.value })}
            >
              {EFFECT_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
            <button onClick={handleAddSpecial}>添加特殊牌</button>
          </div>
        </section>

        <section className="admin-card wide">
          <div className="section-header">
            <h2>常见物质</h2>
            <span className="section-desc">用逗号分隔每一类物质</span>
          </div>
          <div className="compound-grid">
            {Object.entries(draft.common_compounds || {}).map(([category, value]) => {
              if (Array.isArray(value)) {
                return (
                  <div key={category} className="compound-block">
                    <div className="compound-title">{category}</div>
                    <textarea
                      value={typeof value === 'string' ? value : value.join(', ')}
                      onChange={(e) => updateCompoundList([category], e.target.value)}
                      rows={4}
                    />
                  </div>
                );
              }

              if (typeof value === 'object' && value !== null) {
                return (
                  <div key={category} className="compound-block">
                    <div className="compound-title">{category}</div>
                    {Object.entries(value).map(([sub, list]) => (
                      <div key={`${category}-${sub}`} className="compound-subblock">
                        <div className="compound-subtitle">{sub}</div>
                        <textarea
                          value={typeof list === 'string' ? list : (Array.isArray(list) ? list.join(', ') : '')}
                          onChange={(e) => updateCompoundList(['elemental_substances', category, sub], e.target.value)}
                          rows={3}
                        />
                      </div>
                    ))}
                  </div>
                );
              }

              return null;
            })}
          </div>
        </section>

        <section className="admin-card wide">
          <div className="section-header">
            <h2>单质列表</h2>
            <span className="section-desc">用逗号分隔各类单质</span>
          </div>
          {!draft?.elemental_substances && (
            <div style={{ padding: '20px', color: '#666', textAlign: 'center' }}>
              配置加载中或单质列表不存在...
            </div>
          )}
          <div className="compound-grid">
            {draft?.elemental_substances && Object.entries(draft.elemental_substances).map(([category, value]) => {
              // 跳过 note 字段
              if (category === 'note') return null;
              
              if (Array.isArray(value)) {
                return (
                  <div key={category} className="compound-block">
                    <div className="compound-title">{category}</div>
                    <textarea
                      value={typeof value === 'string' ? value : value.join(', ')}
                      onChange={(e) => updateCompoundList(['elemental_substances', category], e.target.value)}
                      rows={4}
                      placeholder="输入单质，用逗号分隔"
                    />
                  </div>
                );
              }

              if (typeof value === 'object' && value !== null) {
                return (
                  <div key={category} className="compound-block">
                    <div className="compound-title">{category}</div>
                    {Object.entries(value).map(([sub, list]) => (
                      <div key={`elemental-${category}-${sub}`} className="compound-subblock">
                        <div className="compound-subtitle">{sub}</div>
                        <textarea
                          value={typeof list === 'string' ? list : (Array.isArray(list) ? list.join(', ') : '')}
                          onChange={(e) => updateCompoundList(['elemental_substances', category, sub], e.target.value)}
                          rows={3}
                          placeholder="输入单质，用逗号分隔"
                        />
                      </div>
                    ))}
                  </div>
                );
              }

              return null;
            })}
          </div>
        </section>

        <section className="admin-card wide">
          <div className="section-header">
            <h2>反应规则</h2>
            <span className="section-desc">每个反应物对应可与之反应的物质列表（修改后请点击顶部"保存配置"按钮）</span>
          </div>
          {hasChanges && (
            <div style={{ 
              padding: '12px', 
              background: '#fef3c7', 
              border: '1px solid #fbbf24',
              borderRadius: '8px',
              marginBottom: '12px',
              fontSize: '13px',
              color: '#92400e'
            }}>
              ⚠️ 您有未保存的更改，请点击顶部"保存配置"按钮保存
            </div>
          )}
          <div className="reaction-grid">
            {Object.entries(draft.representative_reactions || {}).sort((a, b) => a[0].localeCompare(b[0])).map(([reactant, partners]) => {
              // 确保 partners 是数组
              const partnersList = Array.isArray(partners) ? partners : [];
              const textValue = partnersList.join('\n');
              
              return (
              <div key={reactant} className="reaction-block">
                <div className="reaction-title">{reactant}</div>
                <textarea
                  rows={4}
                  value={textValue}
                  onChange={(e) => updateReactionPartners(reactant, e.target.value)}
                  placeholder="一行一个物质"
                  style={{ resize: 'vertical' }}
                />
                <button 
                  className="btn-delete"
                  onClick={() => {
                    if (window.confirm(`确定要删除反应物 "${reactant}" 吗？`)) {
                      setDraft(prev => {
                        const next = deepClone(prev!);
                        delete next.representative_reactions![reactant];
                        return next;
                      });
                    }
                  }}
                >删除</button>
              </div>
              );
            })}
          </div>
          <div className="inline-form">
            <input
              type="text"
              placeholder="反应物（如 HCl）"
              value={newReaction.reactant}
              onChange={(e) => setNewReaction({ ...newReaction, reactant: e.target.value })}
            />
            <textarea
              rows={2}
              placeholder="可反应物质（一行一个）"
              value={newReaction.partners}
              onChange={(e) => setNewReaction({ ...newReaction, partners: e.target.value })}
            />
            <button onClick={addReactant}>添加反应物</button>
          </div>
        </section>
      </div>
    </div>
  );
};

export default AdminPanel;
