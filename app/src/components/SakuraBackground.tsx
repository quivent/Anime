import { useEffect, useState } from 'react'

interface Branch {
  id: number
  x: number
  y: number
  rotation: number
  scale: number
  side: 'left' | 'right'
}

export default function SakuraBackground() {
  const [petals, setPetals] = useState<Array<{ id: number; left: number; delay: number }>>([])
  const [branches, setBranches] = useState<Branch[]>([])

  useEffect(() => {
    // Generate sakura petals
    const newPetals = Array.from({ length: 15 }, (_, i) => ({
      id: i,
      left: Math.random() * 100,
      delay: Math.random() * 10,
    }))
    setPetals(newPetals)

    // Generate decorative branches on the sides
    const newBranches: Branch[] = [
      // Left side branches
      { id: 1, x: 2, y: 5, rotation: 15, scale: 1.2, side: 'left' },
      { id: 2, x: 5, y: 25, rotation: 25, scale: 1.0, side: 'left' },
      { id: 3, x: 3, y: 50, rotation: 10, scale: 1.3, side: 'left' },
      { id: 4, x: 6, y: 75, rotation: 20, scale: 0.9, side: 'left' },
      // Right side branches
      { id: 5, x: 98, y: 10, rotation: -20, scale: 1.1, side: 'right' },
      { id: 6, x: 95, y: 35, rotation: -15, scale: 1.2, side: 'right' },
      { id: 7, x: 97, y: 60, rotation: -25, scale: 1.0, side: 'right' },
      { id: 8, x: 94, y: 85, rotation: -18, scale: 1.3, side: 'right' },
    ]
    setBranches(newBranches)
  }, [])

  return (
    <div className="fixed inset-0 pointer-events-none overflow-hidden z-0">
      {/* Sakura Branches */}
      <div className="absolute inset-0 opacity-20">
        {branches.map((branch) => (
          <div
            key={branch.id}
            className="absolute"
            style={{
              left: `${branch.x}%`,
              top: `${branch.y}%`,
              transform: `rotate(${branch.rotation}deg) scale(${branch.scale})`,
            }}
          >
            <svg
              width="200"
              height="300"
              viewBox="0 0 200 300"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              {/* Main branch */}
              <path
                d="M100 300 Q90 200 80 150 Q75 100 70 50"
                stroke="#ec4899"
                strokeWidth="8"
                strokeLinecap="round"
                fill="none"
                opacity="0.6"
              />
              {/* Side branches */}
              <path
                d="M80 150 Q60 140 40 120"
                stroke="#ec4899"
                strokeWidth="5"
                strokeLinecap="round"
                fill="none"
                opacity="0.5"
              />
              <path
                d="M75 100 Q55 95 35 85"
                stroke="#ec4899"
                strokeWidth="4"
                strokeLinecap="round"
                fill="none"
                opacity="0.5"
              />

              {/* Blossoms */}
              <circle cx="40" cy="120" r="8" fill="#fda4af" opacity="0.8" />
              <circle cx="45" cy="115" r="6" fill="#fda4af" opacity="0.7" />
              <circle cx="35" cy="85" r="7" fill="#fda4af" opacity="0.8" />
              <circle cx="42" cy="88" r="5" fill="#fda4af" opacity="0.6" />
              <circle cx="70" cy="50" r="9" fill="#fda4af" opacity="0.9" />
              <circle cx="75" cy="55" r="6" fill="#fda4af" opacity="0.7" />
              <circle cx="65" cy="45" r="7" fill="#fda4af" opacity="0.8" />
            </svg>
          </div>
        ))}
      </div>

      {/* Falling petals */}
      <div className="absolute inset-0 opacity-30">
        {petals.map((petal) => (
          <div
            key={petal.id}
            className="sakura-petal"
            style={{
              left: `${petal.left}%`,
              animationDelay: `${petal.delay}s`,
            }}
          />
        ))}
      </div>
    </div>
  )
}
