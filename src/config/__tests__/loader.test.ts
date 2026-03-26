import { describe, it, expect } from 'vitest';
import { loadConfig, DEFAULT_CONFIG } from '../loader.js';

describe('loadConfig', () => {
  it('returns default config when no config file exists', () => {
    const config = loadConfig('/nonexistent/path');
    expect(config.weights).toEqual(DEFAULT_CONFIG.weights);
    expect(config.thresholds).toEqual(DEFAULT_CONFIG.thresholds);
    expect(config.languages).toEqual(DEFAULT_CONFIG.languages);
  });

  it('default weights sum to 100', () => {
    const total = Object.values(DEFAULT_CONFIG.weights).reduce((s, w) => s + w, 0);
    expect(total).toBe(100);
  });
});
