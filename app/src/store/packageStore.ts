import { create } from 'zustand'
import type { Package, InstallProgress } from '../types/package'

interface PackageStore {
  packages: Package[]
  selectedPackages: Set<string>
  installProgress: Map<string, InstallProgress>
  installing: boolean

  togglePackage: (id: string) => void
  selectPackages: (ids: string[]) => void
  clearSelection: () => void

  setPackages: (packages: Package[]) => void
  startInstall: (packageIds: string[]) => void
  updateProgress: (packageId: string, progress: Partial<InstallProgress>) => void
  completeInstall: () => void
}

export const usePackageStore = create<PackageStore>((set, get) => ({
  packages: [],
  selectedPackages: new Set(),
  installProgress: new Map(),
  installing: false,

  togglePackage: (id: string) => {
    const selected = new Set(get().selectedPackages)
    if (selected.has(id)) {
      selected.delete(id)
    } else {
      selected.add(id)
    }
    set({ selectedPackages: selected })
  },

  selectPackages: (ids: string[]) => {
    set({ selectedPackages: new Set(ids) })
  },

  clearSelection: () => {
    set({ selectedPackages: new Set() })
  },

  setPackages: (packages: Package[]) => {
    set({ packages })
  },

  startInstall: (packageIds: string[]) => {
    const progress = new Map<string, InstallProgress>()
    packageIds.forEach(id => {
      progress.set(id, {
        packageId: id,
        status: 'pending',
        progress: 0,
        message: 'Waiting to start...',
        startTime: Date.now(),
      })
    })
    set({ installProgress: progress, installing: true })
  },

  updateProgress: (packageId: string, update: Partial<InstallProgress>) => {
    const progress = new Map(get().installProgress)
    const current = progress.get(packageId)
    if (current) {
      progress.set(packageId, { ...current, ...update })
      set({ installProgress: progress })
    }
  },

  completeInstall: () => {
    set({ installing: false })
  },
}))
