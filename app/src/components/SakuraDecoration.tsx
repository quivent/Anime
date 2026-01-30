export default function SakuraDecoration() {
  return (
    <div className="absolute inset-0 pointer-events-none overflow-hidden">
      {/* Top left corner branch */}
      <svg
        className="absolute top-0 left-0 opacity-40"
        width="400"
        height="200"
        viewBox="0 0 400 200"
        fill="none"
      >
        <path
          d="M0 0 Q50 20 100 40 Q150 60 200 70"
          stroke="#ec4899"
          strokeWidth="6"
          strokeLinecap="round"
          fill="none"
        />
        <path
          d="M100 40 Q120 60 140 80"
          stroke="#ec4899"
          strokeWidth="4"
          strokeLinecap="round"
          fill="none"
        />
        <path
          d="M150 60 Q170 50 190 40"
          stroke="#ec4899"
          strokeWidth="4"
          strokeLinecap="round"
          fill="none"
        />
        {/* Blossoms */}
        <circle cx="140" cy="80" r="8" fill="#fda4af" opacity="0.9" />
        <circle cx="148" cy="75" r="6" fill="#fda4af" opacity="0.8" />
        <circle cx="135" cy="85" r="7" fill="#fda4af" opacity="0.85" />
        <circle cx="190" cy="40" r="9" fill="#fda4af" opacity="0.9" />
        <circle cx="185" cy="35" r="6" fill="#fda4af" opacity="0.8" />
        <circle cx="195" cy="45" r="7" fill="#fda4af" opacity="0.85" />
        <circle cx="100" cy="40" r="8" fill="#fda4af" opacity="0.9" />
        <circle cx="105" cy="45" r="6" fill="#fda4af" opacity="0.8" />
      </svg>

      {/* Top right corner branch */}
      <svg
        className="absolute top-0 right-0 opacity-40"
        width="400"
        height="200"
        viewBox="0 0 400 200"
        fill="none"
      >
        <path
          d="M400 0 Q350 20 300 40 Q250 60 200 70"
          stroke="#ec4899"
          strokeWidth="6"
          strokeLinecap="round"
          fill="none"
        />
        <path
          d="M300 40 Q280 60 260 80"
          stroke="#ec4899"
          strokeWidth="4"
          strokeLinecap="round"
          fill="none"
        />
        <path
          d="M250 60 Q230 50 210 40"
          stroke="#ec4899"
          strokeWidth="4"
          strokeLinecap="round"
          fill="none"
        />
        {/* Blossoms */}
        <circle cx="260" cy="80" r="8" fill="#fda4af" opacity="0.9" />
        <circle cx="252" cy="75" r="6" fill="#fda4af" opacity="0.8" />
        <circle cx="265" cy="85" r="7" fill="#fda4af" opacity="0.85" />
        <circle cx="210" cy="40" r="9" fill="#fda4af" opacity="0.9" />
        <circle cx="215" cy="35" r="6" fill="#fda4af" opacity="0.8" />
        <circle cx="205" cy="45" r="7" fill="#fda4af" opacity="0.85" />
        <circle cx="300" cy="40" r="8" fill="#fda4af" opacity="0.9" />
        <circle cx="295" cy="45" r="6" fill="#fda4af" opacity="0.8" />
      </svg>
    </div>
  )
}
