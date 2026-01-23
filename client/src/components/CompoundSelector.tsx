import React, { useState, useEffect } from 'react';
import { formatFormula } from '../utils/chemistryFormatter';
import './CompoundSelector.css';

interface CompoundSelectorProps {
  compounds: string[];
  selectedCard: string;
  onSelect: (compound: string) => void;
  onClose: () => void;
}

const CompoundSelector: React.FC<CompoundSelectorProps> = ({ compounds, selectedCard, onSelect, onClose }) => {
  const [filteredCompounds, setFilteredCompounds] = useState<string[]>(compounds);
  const [searchTerm, setSearchTerm] = useState<string>('');

  useEffect(() => {
    setFilteredCompounds(
      compounds.filter(compound =>
        compound.toLowerCase().includes(searchTerm.toLowerCase())
      )
    );
  }, [searchTerm, compounds]);

  return (
    <div className="compound-selector-overlay" onClick={onClose}>
      <div className="compound-selector-modal" onClick={(e) => e.stopPropagation()}>
        <div className="selector-header">
          <h2>选择要打出的物质</h2>
          <span className="selected-card">使用元素: {selectedCard}</span>
        </div>

        <div className="selector-search">
          <input
            type="text"
            placeholder="搜索物质..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
            autoFocus
          />
        </div>

        {filteredCompounds.length > 0 ? (
          <div className="compounds-grid">
            {filteredCompounds.map((compound, idx) => (
              <button
                key={idx}
                className="compound-btn"
                onClick={() => onSelect(compound)}
              >
                <div className="compound-name">{formatFormula(compound)}</div>
              </button>
            ))}
          </div>
        ) : (
          <div className="no-compounds">
            {searchTerm ? '没有匹配的物质' : '该元素无法组成任何物质'}
          </div>
        )}

        <button className="close-btn" onClick={onClose}>取消</button>
      </div>
    </div>
  );
};

export default CompoundSelector;
