import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { CoverageReport } from '../types/coverage'

interface CoverageStore {
  coverages: CoverageReport[]
  addCoverage: (coverage: Omit<CoverageReport, 'id' | 'created_at' | 'updated_at' | 'version'>) => void
  updateCoverage: (id: string, updates: Partial<CoverageReport>) => void
  deleteCoverage: (id: string) => void
  getCoverageById: (id: string) => CoverageReport | undefined
}

export const useCoverageStore = create<CoverageStore>()(
  persist(
    (set, get) => ({
      coverages: [],

      addCoverage: (coverage) => {
        const newCoverage: CoverageReport = {
          ...coverage,
          id: crypto.randomUUID(),
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          version: 1,
        }
        set((state) => ({ coverages: [...state.coverages, newCoverage] }))
      },

      updateCoverage: (id, updates) => {
        set((state) => ({
          coverages: state.coverages.map((coverage) =>
            coverage.id === id
              ? {
                  ...coverage,
                  ...updates,
                  updated_at: new Date().toISOString(),
                  version: coverage.version + 1,
                }
              : coverage
          ),
        }))
      },

      deleteCoverage: (id) => {
        set((state) => ({
          coverages: state.coverages.filter((coverage) => coverage.id !== id),
        }))
      },

      getCoverageById: (id) => {
        return get().coverages.find((coverage) => coverage.id === id)
      },
    }),
    {
      name: 'coverage-storage',
    }
  )
)
