const USER_AVATAR_PRESET_COLORS = [
  '#1D4ED8',
  '#0F766E',
  '#B45309',
  '#7C3AED',
  '#BE123C',
  '#0369A1',
  '#15803D',
  '#9A3412',
  '#0F766E',
  '#4338CA',
  '#CA8A04',
  '#C2410C',
] as const;

const USER_AVATAR_ACCEPTED_TYPES = ['image/jpeg', 'image/png', 'image/webp'] as const;

type UserAvatarInput = {
  avatar?: string;
  avatarColor?: string;
  avatarType?: string;
  realName?: string;
  username?: string;
};

function getDisplaySeed(input: UserAvatarInput) {
  const value = input.realName?.trim() || input.username?.trim() || 'User';
  return value;
}

function expandShortHex(hex: string) {
  if (hex.length !== 4) {
    return hex;
  }
  const [hash, r, g, b] = hex;
  return `${hash}${r}${r}${g}${g}${b}${b}`;
}

export function normalizeAvatarColor(color?: string) {
  const value = color?.trim();
  if (!value) {
    return '';
  }
  const normalized = value.startsWith('#') ? value : `#${value}`;
  if (!/^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$/.test(normalized)) {
    return '';
  }
  return expandShortHex(normalized).toUpperCase();
}

export function getAvatarText(name?: string) {
  const seed = name?.trim();
  if (!seed) {
    return 'U';
  }
  const [firstChar = 'U'] = Array.from(seed);
  return /^[A-Za-z]$/.test(firstChar) ? firstChar.toUpperCase() : firstChar;
}

export function getDeterministicAvatarColor(seed?: string) {
  const source = seed?.trim() || 'User';
  const hash = Array.from(source).reduce((total, char, index) => {
    return total + char.charCodeAt(0) * (index + 1);
  }, 0);
  return USER_AVATAR_PRESET_COLORS[Math.abs(hash) % USER_AVATAR_PRESET_COLORS.length];
}

export function resolveUserAvatar(input: UserAvatarInput) {
  const displaySeed = getDisplaySeed(input);
  const explicitMode = input.avatarType?.trim().toLowerCase();
  const avatar = input.avatar?.trim() || '';
  const hasImage = Boolean(avatar);
  const avatarText = getAvatarText(displaySeed);
  const avatarBackgroundColor =
    normalizeAvatarColor(input.avatarColor) || getDeterministicAvatarColor(displaySeed);
  const useImage = explicitMode === 'image' ? hasImage : explicitMode === 'letter' ? false : hasImage;

  return {
    avatar: useImage ? avatar : '',
    avatarBackgroundColor,
    avatarText,
    mode: (useImage ? 'image' : 'letter') as 'image' | 'letter',
  };
}

export function isAvatarFileTypeSupported(file: File) {
  if (
    USER_AVATAR_ACCEPTED_TYPES.includes(
      file.type as (typeof USER_AVATAR_ACCEPTED_TYPES)[number],
    )
  ) {
    return true;
  }
  const extension = file.name.split('.').pop()?.toLowerCase();
  return ['jpeg', 'jpg', 'png', 'webp'].includes(extension || '');
}

export function buildAvatarAcceptAttribute() {
  return '.jpg,.jpeg,.png,.webp';
}

export { USER_AVATAR_ACCEPTED_TYPES, USER_AVATAR_PRESET_COLORS };
