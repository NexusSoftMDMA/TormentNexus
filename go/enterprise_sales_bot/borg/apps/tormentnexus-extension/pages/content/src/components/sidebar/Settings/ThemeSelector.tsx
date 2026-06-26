import React from 'react';
import { useUserPreferences } from '@src/hooks';
import { cn } from '@src/lib/utils';
import { Icon } from '../ui';

const ACCENT_COLORS = [
  { id: 'blue', name: 'Blue', class: 'bg-blue-600' },
  { id: 'indigo', name: 'Indigo', class: 'bg-primary-600' },
  { id: 'violet', name: 'Violet', class: 'bg-violet-600' },
  { id: 'purple', name: 'Purple', class: 'bg-purple-600' },
  { id: 'fuchsia', name: 'Fuchsia', class: 'bg-fuchsia-600' },
  { id: 'pink', name: 'Pink', class: 'bg-pink-600' },
  { id: 'rose', name: 'Rose', class: 'bg-rose-600' },
  { id: 'red', name: 'Red', class: 'bg-red-600' },
  { id: 'orange', name: 'Orange', class: 'bg-orange-600' },
  { id: 'amber', name: 'Amber', class: 'bg-amber-600' },
  { id: 'yellow', name: 'Yellow', class: 'bg-yellow-500' },
  { id: 'lime', name: 'Lime', class: 'bg-lime-600' },
  { id: 'green', name: 'Green', class: 'bg-green-600' },
  { id: 'emerald', name: 'Emerald', class: 'bg-emerald-600' },
  { id: 'teal', name: 'Teal', class: 'bg-teal-600' },
  { id: 'cyan', name: 'Cyan', class: 'bg-cyan-600' },
  { id: 'sky', name: 'Sky', class: 'bg-sky-600' },
  { id: 'slate', name: 'Slate', class: 'bg-slate-600' },
];

export const ThemeSelector: React.FC = () => {
  const { preferences, updatePreferences } = useUserPreferences();
  const currentAccent = preferences.accentColor || 'indigo';

  const handleSelect = (colorId: string) => {
    updatePreferences({ accentColor: colorId });
    // Apply CSS variable immediately (though store update usually triggers re-render/effect)
    // The actual CSS var application happens in Sidebar.tsx via useEffect
    document.documentElement.style.setProperty('--mcp-accent-color', colorId); // Placeholder mechanism
  };

  return (
    <div className="space-y-3">
      <label className="text-sm font-medium text-slate-700 dark:text-slate-300">Accent Color</label>
      <div className="grid grid-cols-6 gap-2">
        {ACCENT_COLORS.map((color) => (
          <button
            key={color.id}
            onClick={() => handleSelect(color.id)}
            className={cn(
              "w-8 h-8 rounded-full flex items-center justify-center transition-all hover:scale-110 focus:outline-none focus:ring-2 focus:ring-offset-2 dark:focus:ring-offset-slate-900",
              color.class,
              currentAccent === color.id ? "ring-2 ring-offset-2 ring-slate-400 dark:ring-offset-slate-900 scale-110" : ""
            )}
            title={color.name}
            aria-label={`Select ${color.name} accent`}
          >
            {currentAccent === color.id && <Icon name="check" size="xs" className="text-white" />}
          </button>
        ))}
      </div>
    </div>
  );
};
