import { create } from 'zustand'
import type { Instance } from '../types/lambda'

interface InstanceStore {
  selectedInstance: Instance | null
  setSelectedInstance: (instance: Instance | null) => void
}

export const useInstanceStore = create<InstanceStore>((set) => ({
  selectedInstance: null,
  setSelectedInstance: (instance) => set({ selectedInstance: instance }),
}))
