import { describe, expect, it } from 'vitest';

import { extractInitial, pickForegroundColor } from './branding';

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

describe('pickForegroundColor', () => {
  it('pure black returns white foreground', () => {
    expect(pickForegroundColor('#000000')).toBe('#FFFFFF');
  });

  it('pure white returns dark foreground', () => {
    expect(pickForegroundColor('#FFFFFF')).toBe('#1F2937');
  });

  it('mid gray returns a valid foreground', () => {
    const result = pickForegroundColor('#7F7F7F');
    expect(['#FFFFFF', '#1F2937']).toContain(result);
  });

  it('red (L≈0.21) returns white foreground', () => {
    expect(pickForegroundColor('#FF0000')).toBe('#FFFFFF');
  });

  it('green (L≈0.72) returns dark foreground', () => {
    expect(pickForegroundColor('#00FF00')).toBe('#1F2937');
  });

  it('3-digit shorthand does not throw', () => {
    expect(() => pickForegroundColor('abc')).not.toThrow();
  });

  it('invalid input returns dark foreground', () => {
    expect(pickForegroundColor('!!!')).toBe('#1F2937');
  });
});
