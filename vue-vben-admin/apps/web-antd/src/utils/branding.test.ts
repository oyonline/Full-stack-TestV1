import { describe, expect, it } from 'vitest';

import { extractInitial } from './branding';

describe('extractInitial', () => {
  it('English string returns first letter uppercased', () => {
    expect(extractInitial('Eposeidon')).toBe('E');
  });

  it('leading spaces are skipped', () => {
    expect(extractInitial('  hello')).toBe('H');
  });

  it('Chinese string returns first character', () => {
    expect(extractInitial('中文系统')).toBe('中');
  });

  it('leading punctuation is skipped', () => {
    expect(extractInitial('!!@@abc')).toBe('A');
  });

  it('empty string returns S', () => {
    expect(extractInitial('')).toBe('S');
  });

  it('whitespace-only string returns S', () => {
    expect(extractInitial('   ')).toBe('S');
  });

  it('punctuation-only string returns S', () => {
    expect(extractInitial('!!@@')).toBe('S');
  });

  it('numeric string returns first digit', () => {
    expect(extractInitial('123abc')).toBe('1');
  });
});
