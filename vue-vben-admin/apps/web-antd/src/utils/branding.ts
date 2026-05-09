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
