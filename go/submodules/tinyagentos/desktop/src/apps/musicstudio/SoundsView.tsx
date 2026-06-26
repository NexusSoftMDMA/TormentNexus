const FILTER_PILLS = ["All", "Drums", "Bass", "Keys", "Synths", "FX"];

interface SoundCard {
  name: string;
  category: string;
  detail: string;
  padGradient: string;
  bars: number[];
}

const SOUNDS: SoundCard[] = [
  {
    name: "Boom Bap Kit",
    category: "Drums",
    detail: "24 hits",
    padGradient: "linear-gradient(140deg, #3a2c24, #1d1611)",
    bars: [40, 90, 30, 70, 45, 95, 35],
  },
  {
    name: "Analog Bass",
    category: "Bass",
    detail: "synth",
    padGradient: "linear-gradient(140deg, #23303f, #121b24)",
    bars: [70, 50, 85, 45, 65],
  },
  {
    name: "Rhodes Mk I",
    category: "Keys",
    detail: "electric piano",
    padGradient: "linear-gradient(140deg, #223529, #101c15)",
    bars: [55, 65, 50, 75, 55, 60],
  },
  {
    name: "Warm Pad",
    category: "Synths",
    detail: "ambient",
    padGradient: "linear-gradient(140deg, #2f2839, #171320)",
    bars: [35, 45, 40, 50, 42],
  },
  {
    name: "Pluck Lead",
    category: "Synths",
    detail: "mono",
    padGradient: "linear-gradient(140deg, #3a3526, #1d1a11)",
    bars: [80, 45, 90, 55, 70],
  },
  {
    name: "Vinyl FX",
    category: "FX",
    detail: "texture",
    padGradient: "linear-gradient(140deg, #34232c, #1a1116)",
    bars: [30, 60, 35, 55],
  },
  {
    name: "Sub 808",
    category: "Bass",
    detail: "808",
    padGradient: "linear-gradient(140deg, #23303f, #121b24)",
    bars: [60, 80, 50, 70, 55],
  },
  {
    name: "Felt Piano",
    category: "Keys",
    detail: "acoustic",
    padGradient: "linear-gradient(140deg, #223529, #101c15)",
    bars: [50, 55, 65, 45, 60],
  },
];

export function SoundsView() {
  return (
    <div className="flex min-h-0 flex-1 flex-col">
      {/* view header */}
      <div
        className="flex flex-none items-center gap-3 border-b border-shell-border px-[22px]"
        style={{ height: "54px" }}
      >
        <h2 className="text-[17px] font-bold tracking-[-0.02em]">Sounds</h2>
        <span className="text-[12px] text-shell-text-tertiary">
          Instruments and samples, all offline
        </span>
      </div>

      <div className="flex-1 overflow-auto p-[22px]">
        {/* filter pills */}
        <div className="mb-[18px] flex flex-wrap gap-[9px]">
          {FILTER_PILLS.map((pill, i) => (
            <button
              key={pill}
              type="button"
              className={`rounded-full border px-3.5 py-[7px] text-[12px] font-semibold ${
                i === 0
                  ? "border-transparent bg-accent text-white"
                  : "border-shell-border bg-shell-surface text-shell-text-secondary"
              }`}
            >
              {pill}
            </button>
          ))}
        </div>

        {/* sound card grid */}
        <div className="grid grid-cols-4 gap-[13px]">
          {SOUNDS.map((sound) => (
            <div
              key={sound.name}
              className="flex cursor-pointer flex-col gap-[11px] rounded-[14px] border border-shell-border bg-shell-surface p-[14px] transition-all hover:-translate-y-[3px] hover:border-shell-border-strong"
            >
              {/* pad preview */}
              <div
                className="flex items-end gap-[2px] rounded-[10px] p-2"
                style={{ height: "54px", background: sound.padGradient }}
              >
                {sound.bars.map((h, i) => (
                  <span
                    key={i}
                    className="flex-1 rounded-[1px]"
                    style={{ height: `${h}%`, background: "rgba(255,255,255,0.8)" }}
                  />
                ))}
              </div>

              <div className="text-[13px] font-bold">{sound.name}</div>
              <div className="mt-[-6px] text-[11px] text-shell-text-tertiary">
                {sound.category} - {sound.detail}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
