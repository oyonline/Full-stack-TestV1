export function extractInitial(raw: string): string {
  for (const char of raw) {
    if (/\s/.test(char)) continue;
    if (/[\p{P}\p{S}]/u.test(char)) continue;
    if (/[一-鿿㐀-䶿\u{20000}-\u{2A6DF}]/u.test(char)) {
      return char;
    }
    return char.toUpperCase();
  }
  return 'S';
}

/** Pick black or white foreground based on WCAG relative luminance. */
export function pickForegroundColor(bgHex: string): '#FFFFFF' | '#1F2937' {
  const hex = bgHex.startsWith('#') ? bgHex.slice(1) : bgHex;
  let r: number, g: number, b: number;
  if (hex.length === 3) {
    const [rh, gh, bh] = hex;
    r = Number.parseInt(rh! + rh!, 16);
    g = Number.parseInt(gh! + gh!, 16);
    b = Number.parseInt(bh! + bh!, 16);
  } else if (hex.length === 6) {
    r = Number.parseInt(hex.slice(0, 2), 16);
    g = Number.parseInt(hex.slice(2, 4), 16);
    b = Number.parseInt(hex.slice(4, 6), 16);
  } else {
    return '#1F2937';
  }
  if (Number.isNaN(r) || Number.isNaN(g) || Number.isNaN(b)) return '#1F2937';

  const toLinear = (c: number): number => {
    const s = c / 255;
    return s <= 0.04045 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4;
  };

  const L = 0.2126 * toLinear(r) + 0.7152 * toLinear(g) + 0.0722 * toLinear(b);
  return L < 0.5 ? '#FFFFFF' : '#1F2937';
}
