import { Link } from 'react-router-dom';
import { Plus, Moon, Sun } from 'lucide-react';
import { Suspense, useEffect, useState } from 'react';
import HeaderStats from './HeaderStats';

const THEME_KEY = 'fluxo-theme';

function readStoredLight(): boolean {
  try {
    const saved = localStorage.getItem(THEME_KEY);
    if (saved === 'light') return true;
    if (saved === 'dark') return false;
  } catch {
    // private mode / blocked storage — fall through
  }
  if (typeof window !== 'undefined' && window.matchMedia) {
    return window.matchMedia('(prefers-color-scheme: light)').matches;
  }
  return false;
}

function applyTheme(isLight: boolean) {
  const theme = isLight ? 'light' : 'dark';
  document.documentElement.setAttribute('data-theme', theme);
  try {
    localStorage.setItem(THEME_KEY, theme);
  } catch {
    // ignore quota / private mode
  }
}

export default function Header() {
  const [isLight, setIsLight] = useState(() => {
    const light = readStoredLight();
    // Apply before paint once JS runs so DaisyUI data-theme matches the toggle.
    applyTheme(light);
    return light;
  });

  useEffect(() => {
    applyTheme(isLight);
  }, [isLight]);

  return (
    <div className="navbar bg-base-100 shadow-lg sticky top-0 z-50 px-4">
      <div className="flex-1">
        <Link to="/" className="btn btn-ghost text-xl font-bold text-primary px-2">Fluxo</Link>
      </div>

      <div className="flex-none flex items-center gap-3 md:gap-4">
        <div className="opacity-80">
          <Suspense fallback={<div className="w-20 h-8 bg-base-200 rounded animate-pulse"></div>}>
            <HeaderStats />
          </Suspense>
        </div>

        <Link to="/add" className="btn btn-primary btn-sm md:btn-md btn-circle md:btn-wide md:px-4 md:w-auto" aria-label="Add Torrent">
          <Plus size={20} />
          <span className="hidden md:inline ml-2">Add Torrent</span>
        </Link>

        <label className="swap swap-rotate btn btn-ghost btn-circle btn-sm md:btn-md">
          <input
            type="checkbox"
            className="theme-controller"
            value="light"
            checked={isLight}
            onChange={(e) => setIsLight(e.target.checked)}
            aria-label={isLight ? 'Switch to dark theme' : 'Switch to light theme'}
          />
          <Sun className="swap-on fill-current w-5 h-5 md:w-6 md:h-6" />
          <Moon className="swap-off fill-current w-5 h-5 md:w-6 md:h-6" />
        </label>
      </div>
    </div>
  );
}
