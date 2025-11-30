import React from 'react';
import { TIME_RANGES } from '../config/config';
import './TimeRangeSelector.css';

function TimeRangeSelector({ selectedRange, onRangeChange }) {
  const handleChange = (e) => {
    onRangeChange(parseInt(e.target.value));
  };

  return (
    <div className="time-range-info">
      <span className="time-range-label">Time Range</span>
      <div className="time-range-wrapper">
        <select
          value={selectedRange}
          onChange={handleChange}
          className="time-range-select"
        >
          {TIME_RANGES.map((range) => (
            <option key={range.value} value={range.value}>
              {range.label}
            </option>
          ))}
        </select>
      </div>
    </div>
  );
}

export default TimeRangeSelector;