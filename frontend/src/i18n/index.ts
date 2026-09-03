import en, { type Messages } from './en';
import fa from './fa';
import type { Lang } from '../lib/types';

const tables: Record<Lang, Messages> = { en, fa };

export function t(lang: Lang): Messages {
  return tables[lang] ?? en;
}

export function detectLang(): Lang {
  const nav = typeof navigator !== 'undefined' ? navigator.language : 'en';
  return nav.toLowerCase().startsWith('fa') ? 'fa' : 'en';
}
